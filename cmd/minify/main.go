package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/pprof"
	"sort"
	"strings"
	"time"

	"github.com/djherbis/atime"
	humanize "github.com/dustin/go-humanize"
	"github.com/matryer/try"
	flag "github.com/spf13/pflag"
	min "github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
	"github.com/tdewolff/minify/v2/json"
	"github.com/tdewolff/minify/v2/svg"
	"github.com/tdewolff/minify/v2/xml"
)

// Version is the current minify version.
var Version = "built from source"

var filetypeMime = map[string]string{
	"css":  "text/css",
	"htm":  "text/html",
	"html": "text/html",
	"js":   "application/javascript",
	"mjs":  "application/javascript",
	"json": "application/json",
	"svg":  "image/svg+xml",
	"xml":  "text/xml",
}

var (
	help               bool
	hidden             bool
	list               bool
	m                  *min.M
	matches            []string
	filters            []string
	recursive          bool
	quiet              bool
	verbose            int
	version            bool
	watch              bool
	sync               bool
	bundle             bool
	preserve           []string
	preserveMode       bool
	preserveOwnership  bool
	preserveTimestamps bool
	preserveLinks      bool
	mimetype           string
)

// Task is a minify task.
type Task struct {
	srcs []string
	root string
	dst  string
	sync bool
}

// NewTask returns a new Task.
func NewTask(root, input, output string, sync bool) (Task, error) {
	if IsDir(output) {
		root = filepath.Dir(root)
	}
	t := Task{[]string{input}, root, output, sync}
	if 0 < len(output) && output[len(output)-1] == os.PathSeparator {
		rel, err := filepath.Rel(root, input)
		if err != nil {
			return Task{}, err
		}
		t.dst = filepath.Join(output, rel)
	}
	return t, nil
}

// Loggers.
var (
	Error   *log.Logger
	Warning *log.Logger
	Info    *log.Logger
)

func main() {
	// os.Exit doesn't execute pending defer calls, this is fixed by encapsulating run()
	os.Exit(run())
}

func run() int {
	output := ""
	filetype := ""
	siteurl := ""
	cpuprofile := ""
	memprofile := ""

	cssMinifier := &css.Minifier{}
	htmlMinifier := &html.Minifier{}
	jsMinifier := &js.Minifier{}
	jsonMinifier := &json.Minifier{}
	svgMinifier := &svg.Minifier{}
	xmlMinifier := &xml.Minifier{}

	f := flag.NewFlagSet("minify", flag.ContinueOnError)
	f.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] [input]\n\nOptions:\n", os.Args[0])
		f.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nInput:\n  Files or directories, leave blank to use stdin. Specify --mime or --type to use stdin and stdout.\n")
	}

	f.BoolVarP(&help, "help", "h", false, "Show usage")
	f.StringVarP(&output, "output", "o", "", "Output file or directory (must have trailing slash), leave blank to use stdout")
	f.StringVar(&mimetype, "mime", "", "Mimetype (eg. text/css), optional for input filenames, has precedence over --type")
	f.StringVar(&filetype, "type", "", "Filetype (eg. css), optional for input filenames")
	f.String("match", "", "Filename matching pattern, only matching files are processed")
	f.String("include", "", "Filename inclusion pattern, includes files previously excluded")
	f.String("exclude", "", "Filename exclusion pattern, excludes files from being processed")
	f.BoolVarP(&recursive, "recursive", "r", false, "Recursively minify directories")
	f.BoolVarP(&hidden, "all", "a", false, "Minify all files, including hidden files and files in hidden directories")
	f.BoolVarP(&list, "list", "l", false, "List all accepted filetypes")
	f.BoolVarP(&quiet, "quiet", "q", false, "Quiet mode to suppress all output")
	f.CountVarP(&verbose, "verbose", "v", "Verbose mode, set twice for more verbosity")
	f.BoolVarP(&watch, "watch", "w", false, "Watch files and minify upon changes")
	f.BoolVarP(&sync, "sync", "s", false, "Copy all files to destination directory and minify when filetype matches")
	f.StringSliceVarP(&preserve, "preserve", "p", nil, "Preserve options (mode, ownership, timestamps, links, all)")
	f.Lookup("preserve").NoOptDefVal = "mode,ownership,timestamps"
	f.BoolVarP(&bundle, "bundle", "b", false, "Bundle files by concatenation into a single file")
	f.BoolVar(&version, "version", false, "Version")

	f.StringVar(&siteurl, "url", "", "URL of file to enable URL minification")
	f.StringVar(&cpuprofile, "cpuprofile", "", "Export CPU profile")
	f.StringVar(&memprofile, "memprofile", "", "Export memory profile")
	f.IntVar(&cssMinifier.Precision, "css-precision", 0, "Number of significant digits to preserve in numbers, 0 is all")
	f.BoolVar(&htmlMinifier.KeepComments, "html-keep-comments", false, "Preserve all comments")
	f.BoolVar(&htmlMinifier.KeepConditionalComments, "html-keep-conditional-comments", false, "Preserve all IE conditional comments")
	f.BoolVar(&htmlMinifier.KeepDefaultAttrVals, "html-keep-default-attrvals", false, "Preserve default attribute values")
	f.BoolVar(&htmlMinifier.KeepDocumentTags, "html-keep-document-tags", false, "Preserve html, head and body tags")
	f.BoolVar(&htmlMinifier.KeepEndTags, "html-keep-end-tags", false, "Preserve all end tags")
	f.BoolVar(&htmlMinifier.KeepWhitespace, "html-keep-whitespace", false, "Preserve whitespace characters but still collapse multiple into one")
	f.BoolVar(&htmlMinifier.KeepQuotes, "html-keep-quotes", false, "Preserve quotes around attribute values")
	f.IntVar(&jsMinifier.Precision, "js-precision", 0, "Number of significant digits to preserve in numbers, 0 is all")
	f.BoolVar(&jsMinifier.KeepVarNames, "js-keep-var-names", false, "Preserve original variable names")
	f.IntVar(&jsMinifier.Version, "js-version", 0, "ECMAScript version to toggle supported optimizations (e.g. 2019, 2020), by default 0 is the latest version")
	f.IntVar(&jsonMinifier.Precision, "json-precision", 0, "Number of significant digits to preserve in numbers, 0 is all")
	f.BoolVar(&jsonMinifier.KeepNumbers, "json-keep-numbers", false, "Preserve original numbers instead of minifying them")
	f.BoolVar(&svgMinifier.KeepComments, "svg-keep-comments", false, "Preserve all comments")
	f.IntVar(&svgMinifier.Precision, "svg-precision", 0, "Number of significant digits to preserve in numbers, 0 is all")
	f.BoolVar(&xmlMinifier.KeepWhitespace, "xml-keep-whitespace", false, "Preserve whitespace characters but still collapse multiple into one")
	if len(os.Args) == 1 {
		if !quiet {
			fmt.Printf("minify: must specify --mime or --type in order to use stdin and stdout\n")
			fmt.Printf("Try 'minify --help' for more information\n")
		}
		return 1
	}
	err := f.ParseAll(os.Args[1:], func(flag *flag.Flag, value string) error {
		if flag.Name == "match" || flag.Name == "include" || flag.Name == "exclude" {
			for _, filter := range strings.Split(value, ",") {
				if flag.Name == "match" {
					matches = append(matches, filter)
				} else if flag.Name == "include" {
					filters = append(filters, "+"+filter)
				} else {
					filters = append(filters, "-"+filter)
				}
			}
			return nil
		}
		return f.Set(flag.Name, value)
	})
	if err != nil {
		if !quiet {
			fmt.Printf("minify: %v\n", err)
			fmt.Printf("Try 'minify --help' for more information\n")
		}
		return 1
	}
	inputs := f.Args()
	if len(inputs) == 1 && inputs[0] == "-" {
		inputs = inputs[:0]
	} else if output == "-" {
		output = ""
	}
	useStdin := len(inputs) == 0

	Error = log.New(ioutil.Discard, "", 0)
	Warning = log.New(ioutil.Discard, "", 0)
	Info = log.New(ioutil.Discard, "", 0)
	if !quiet {
		Error = log.New(os.Stderr, "ERROR: ", 0)
		if 0 < verbose {
			Warning = log.New(os.Stderr, "WARNING: ", 0)
		}
		if 1 < verbose {
			Info = log.New(os.Stderr, "INFO: ", 0)
		}
	}

	if help {
		f.Usage()
		return 0
	}

	if version {
		if !quiet {
			fmt.Printf("minify %s\n", Version)
		}
		return 0
	}

	if list {
		if !quiet {
			var keys []string
			for k := range filetypeMime {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				fmt.Println(k + "\t" + filetypeMime[k])
			}
		}
		return 0
	}

	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		if err != nil {
			Error.Println(err)
			return 1
		}
		if err = pprof.StartCPUProfile(f); err != nil {
			Error.Println(err)
			return 1
		}
		defer func() {
			pprof.StopCPUProfile()
			if err = f.Close(); err != nil {
				Error.Println(err)
			}
		}()
	}

	if memprofile != "" {
		defer func() {
			f, err := os.Create(memprofile)
			if err != nil {
				Error.Println(err)
			}
			if err = pprof.WriteHeapProfile(f); err != nil {
				Error.Println(err)
			}
			if err = f.Close(); err != nil {
				Error.Println(err)
			}
		}()
	}

	if 0 < len(matches) {
		for _, filter := range matches {
			if _, err = regexp.Compile(filter); err != nil {
				Error.Println(err)
				return 1
			}
		}
	}
	if 0 < len(filters) {
		for _, filter := range filters {
			if _, err = regexp.Compile(filter[1:]); err != nil {
				Error.Println(err)
				return 1
			}
		}
	}

	// detect mimetype, mimetype=="" means we'll infer mimetype from file extensions
	if mimetype == "" && filetype != "" {
		var ok bool
		if mimetype, ok = filetypeMime[filetype]; !ok {
			Error.Println("cannot find mimetype for filetype", filetype)
			return 1
		}
	}

	if (useStdin || output == "") && (watch || sync) {
		if watch {
			Error.Println("--watch doesn't work with stdin and stdout, specify input and output")
		}
		if sync {
			Error.Println("--sync doesn't work with stdin and stdout, specify input and output")
		}
		return 1
	} else if useStdin && (bundle || recursive) {
		if bundle {
			Error.Println("--bundle doesn't work with stdin, specify input")
		}
		if recursive {
			Error.Println("--recursive doesn't work with stdin, specify input")
		}
		return 1
	} else if output == "" && recursive && !bundle {
		Error.Println("--recursive doesn't work with stdout, specify output or use --bundle")
		return 1
	}
	if mimetype == "" && useStdin {
		Error.Println("must specify --mime or --type for stdin")
		return 1
	} else if mimetype != "" && sync {
		Error.Println("must specify either --sync or --mime/--type")
		return 1
	}
	if mimetype == "" {
		Info.Println("infer mimetype from file extensions")
	} else {
		Info.Println("use mimetype", mimetype)
	}
	if len(preserve) != 0 {
		if bundle {
			Error.Println("--preserve cannot be used together with --bundle")
			return 1
		} else if useStdin || output == "" {
			Error.Println("--preserve cannot be used together with stdin or stdout")
			return 1
		}
	}
	for _, option := range preserve {
		switch option {
		case "all":
			preserveMode = true
			preserveOwnership = true
			preserveTimestamps = true
			preserveLinks = true
		case "mode":
			preserveMode = true
		case "ownership":
			preserveOwnership = true
		case "timestamps":
			preserveTimestamps = true
		case "links":
			preserveLinks = true
		}
	}

	////////////////

	for i, input := range inputs {
		if input == "-" {
			Error.Println("cannot mix files and stdin as input")
			return 1
		}
		inputs[i] = filepath.Clean(input)
		if input[len(input)-1] == '/' {
			inputs[i] += "/"
		}
	}

	// set output file or directory, empty means stdout
	dirDst := false
	if output != "" {
		dirDst = IsDir(output)
		if !dirDst {
			if 1 < len(inputs) && !bundle {
				Error.Printf("stat %v: no such file or directory\n", output)
				return 1
			} else if len(inputs) == 1 {
				if info, err := os.Lstat(inputs[0]); err == nil && !bundle && info.Mode().IsDir() && info.Mode()&os.ModeSymlink == 0 {
					dirDst = true
				}
			}
		}
		if dirDst && bundle {
			Error.Println("--bundle requires destination to be stdout or a file")
			return 1
		}

		output = filepath.Clean(output)
		if dirDst {
			output += string(os.PathSeparator)
		}
	} else if 1 < len(inputs) {
		Error.Println("must specify --bundle for multiple input files with stdout destination")
		return 1
	}
	if output == "" {
		Info.Println("minify to stdout")
	} else if !dirDst {
		Info.Println("minify to output file", output)
	} else if output == "./" {
		Info.Println("minify to current working directory")
	} else {
		Info.Println("minify to output directory", output)
	}
	if useStdin {
		Info.Println("minify from stdin")
	}

	var tasks []Task
	var roots []string
	if useStdin {
		task, err := NewTask("", "", output, false)
		if err != nil {
			Error.Println(err)
			return 1
		}
		tasks = append(tasks, task)
		roots = append(roots, "")
	} else {
		tasks, roots, err = createTasks(inputs, output)
		if err != nil {
			Error.Println(err)
			return 1
		}
	}

	// concatenate
	if 1 < len(tasks) && bundle {
		// Task.sync == false because dirDst == false
		for _, task := range tasks[1:] {
			tasks[0].srcs = append(tasks[0].srcs, task.srcs[0])
		}
		tasks = tasks[:1]
	}

	// make output directory
	if dirDst {
		if err := os.MkdirAll(output, 0777); err != nil {
			Error.Println(err)
			return 1
		}
	}

	////////////////

	m = min.New()
	m.Add("text/css", cssMinifier)
	m.Add("text/html", htmlMinifier)
	m.Add("image/svg+xml", svgMinifier)
	m.AddRegexp(regexp.MustCompile("^(application|text)/(x-)?(java|ecma|j|live)script(1\\.[0-5])?$|^module$"), jsMinifier)
	m.AddRegexp(regexp.MustCompile("[/+]json$"), jsonMinifier)
	m.AddRegexp(regexp.MustCompile("[/+]xml$"), xmlMinifier)

	if m.URL, err = url.Parse(siteurl); err != nil {
		Error.Println(err)
		return 1
	}

	fails := 0
	start := time.Now()
	if !watch && (len(tasks) == 1 || 0 < verbose) {
		for _, task := range tasks {
			if ok := minify(task); !ok {
				fails++
			}
		}
	} else {
		numWorkers := runtime.NumCPU()
		if 0 < verbose {
			numWorkers = 1
		} else if numWorkers < 4 {
			numWorkers = 4
		}

		chanTasks := make(chan Task, 20)
		chanFails := make(chan int, numWorkers)
		for n := 0; n < numWorkers; n++ {
			go minifyWorker(chanTasks, chanFails)
		}

		if !watch {
			for _, task := range tasks {
				chanTasks <- task
			}
		} else {
			watcher, err := NewWatcher(recursive)
			if err != nil {
				Error.Println(err)
				return 1
			}
			defer watcher.Close()
			changes := watcher.Run()

			for _, filename := range inputs {
				watcher.AddPath(filename)
			}

			for _, task := range tasks {
				watcher.IgnoreNext(task.dst)
				chanTasks <- task
			}

			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt)
			for changes != nil {
				select {
				case <-c:
					watcher.Close()
				case file, ok := <-changes:
					if !ok {
						changes = nil
						break
					}
					file = filepath.Clean(file)

					// find longest common path among roots
					root := ""
					for _, path := range roots {
						pathRel, err1 := filepath.Rel(path, file)
						rootRel, err2 := filepath.Rel(root, file)
						if err2 != nil || err1 == nil && len(pathRel) < len(rootRel) {
							root = path
						}
					}

					task, err := NewTask(root, file, output, !fileMatches(file))
					if err != nil {
						Error.Println(err)
						return 1
					}
					watcher.IgnoreNext(task.dst) // skip change on output
					chanTasks <- task
				}
			}
		}

		close(chanTasks)
		for n := 0; n < numWorkers; n++ {
			fails += <-chanFails
		}
	}

	if !watch {
		Info.Println("finished in", time.Since(start))
	}
	if 0 < fails {
		return 1
	}
	return 0
}

func minifyWorker(chanTasks <-chan Task, chanFails chan<- int) {
	fails := 0
	for task := range chanTasks {
		if ok := minify(task); !ok {
			fails++
		}
	}
	chanFails <- fails
}

func fileFilter(filename string) bool {
	if 0 < len(matches) {
		match := false
		for _, filter := range matches {
			pattern := regexp.MustCompile(filter)
			if pattern.MatchString(filename) {
				match = true
				break
			}
		}
		if !match {
			return false
		}
	}
	for _, filter := range filters {
		pattern := regexp.MustCompile(filter[1:])
		if pattern.MatchString(filename) {
			return filter[0] == '+'
		}
	}
	return true
}

func fileMatches(filename string) bool {
	if !fileFilter(filename) {
		return false
	} else if mimetype != "" {
		return true
	}

	ext := filepath.Ext(filename)
	if 0 < len(ext) {
		ext = ext[1:]
	}
	if _, ok := filetypeMime[ext]; !ok {
		return false
	}
	return true
}

func createTasks(inputs []string, output string) ([]Task, []string, error) {
	tasks := []Task{}
	roots := []string{}
	for _, input := range inputs {
		var info os.FileInfo
		var err error
		if !preserveLinks {
			// follow and dereference symlinks
			info, err = os.Stat(input)
		} else {
			info, err = os.Lstat(input)
		}
		if err != nil {
			return nil, nil, err
		}

		root := filepath.Clean(filepath.Dir(input) + string(os.PathSeparator) + ".")
		if preserveLinks && info.Mode()&os.ModeSymlink != 0 {
			// copy symlink as is
			if !sync {
				Warning.Println("--sync not specified, omitting symbolic link", input)
				continue
			}
			task, err := NewTask(root, input, output, true)
			if err != nil {
				return nil, nil, err
			}
			tasks = append(tasks, task)
		} else if info.Mode().IsRegular() {
			valid := fileFilter(info.Name()) // don't filter mimetype
			if valid || sync {
				task, err := NewTask(root, input, output, !valid)
				if err != nil {
					return nil, nil, err
				}
				tasks = append(tasks, task)
			}
		} else if info.Mode().IsDir() {
			if !recursive {
				Warning.Println("--recursive not specified, omitting directory", input)
				continue
			}
			root = input
			roots = append(roots, root)

			var walkFn func(string, DirEntry, error) error
			walkFn = func(path string, d DirEntry, err error) error {
				if err != nil {
					return err
				} else if d.Name() == "." || d.Name() == ".." {
					return nil
				} else if d.Name() == "" || !hidden && d.Name()[0] == '.' {
					if d.IsDir() {
						return SkipDir
					}
					return nil
				}

				input := path
				if filepath.IsAbs(input) {
					// path is absolute, clean it
					input = filepath.Clean(input)
				} else {
					// path is relative, make an absolute path - Join() returns a clean result
					input = filepath.Join(root, input)
				}

				if !preserveLinks && d.Type()&os.ModeSymlink != 0 {
					// follow and dereference symlinks
					info, err := os.Stat(input)
					if err != nil {
						return err
					}
					if info.IsDir() {
						return WalkDir(DirFS(root), path, walkFn)
					}
					d = &statDirEntry{info}
				}

				if preserveLinks && d.Type()&os.ModeSymlink != 0 {
					// copy symlink as is
					if !sync {
						Warning.Println("--sync not specified, omitting symbolic link", path)
						return nil
					}
					task, err := NewTask(root, input, output, true)
					if err != nil {
						return err
					}
					tasks = append(tasks, task)
				} else if d.Type().IsRegular() {
					valid := fileMatches(d.Name())
					if valid || sync {
						task, err := NewTask(root, input, output, !valid)
						if err != nil {
							return err
						}
						tasks = append(tasks, task)
					}
				}
				return nil
			}
			if err := WalkDir(DirFS(root), ".", walkFn); err != nil {
				return nil, nil, err
			}
		} else {
			return nil, nil, fmt.Errorf("not a file or directory %s", input)
		}
	}
	return tasks, roots, nil
}

func openInputFile(input string) (io.ReadCloser, error) {
	var r *os.File
	if input == "" {
		r = os.Stdin
	} else {
		err := try.Do(func(attempt int) (bool, error) {
			var ferr error
			r, ferr = os.Open(input)
			return attempt < 5, ferr
		})

		if err != nil {
			return nil, fmt.Errorf("open input file %q: %w", input, err)
		}
	}
	return r, nil
}

func openOutputFile(output string) (*os.File, error) {
	var w *os.File
	if output == "" {
		w = os.Stdout
	} else {
		dir := filepath.Dir(output)
		if err := os.MkdirAll(dir, 0777); err != nil {
			return nil, fmt.Errorf("creating directory %q: %w", dir, err)
		}

		err := try.Do(func(attempt int) (bool, error) {
			var ferr error
			w, ferr = os.OpenFile(output, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
			return attempt < 5, ferr
		})

		if err != nil {
			return nil, fmt.Errorf("open output file %q: %w", output, err)
		}
	}
	return w, nil
}

func createSymlink(input, output string) error {
	if _, err := os.Stat(output); err == nil {
		if err = os.Remove(output); err != nil {
			return err
		}
	}
	if err := os.MkdirAll(filepath.Dir(output), 0777); err != nil {
		return err
	}
	if err := os.Symlink(input, output); err != nil {
		return err
	}
	return nil
}

func minify(t Task) bool {
	// synchronizing files that are not minified but just copied to the same directory, no action needed
	if t.sync {
		if t.srcs[0] == t.dst {
			return true
		} else if info, err := os.Lstat(t.srcs[0]); preserveLinks && err == nil && info.Mode()&os.ModeSymlink != 0 {
			src, err := os.Readlink(t.srcs[0])
			if err != nil {
				Error.Println(err)
				return false
			}
			if err := createSymlink(src, t.dst); err != nil {
				Error.Println(err)
				return false
			}
			return true
		}
	}

	fileMimetype := mimetype
	if mimetype == "" && !t.sync {
		for _, src := range t.srcs {
			ext := filepath.Ext(src)
			if 0 < len(ext) {
				ext = ext[1:]
			}
			srcMimetype, ok := filetypeMime[ext]
			if !ok {
				Warning.Println("cannot infer mimetype from extension in", src, ", set --type or --mime explicitly")
				return false
			}
			if fileMimetype == "" {
				fileMimetype = srcMimetype
			} else if srcMimetype != fileMimetype {
				Warning.Println("inferred mimetype", srcMimetype, "of", src, "for concatenation unequal to previous mimetypes, set --type or --mime explicitly")
				return false
			}
		}
	}

	srcName := strings.Join(t.srcs, " + ")
	if len(t.srcs) > 1 {
		srcName = "(" + srcName + ")"
	}
	if srcName == "" {
		srcName = "stdin"
	}
	dstName := t.dst
	if dstName == "" {
		dstName = "stdout"
	} else {
		// rename original when overwriting
		for i := range t.srcs {
			if sameFile, _ := SameFile(t.srcs[i], t.dst); sameFile {
				t.srcs[i] += ".bak"
				err := try.Do(func(attempt int) (bool, error) {
					ferr := os.Rename(t.dst, t.srcs[i])
					return attempt < 5, ferr
				})
				if err != nil {
					Error.Println(err)
					return false
				}
				break
			}
		}
	}

	var err error
	var fr io.ReadCloser
	var fw io.WriteCloser
	if len(t.srcs) == 1 {
		fr, err = openInputFile(t.srcs[0])
	} else {
		fr, err = newConcatFileReader(t.srcs, openInputFile)
		if err == nil && fileMimetype == filetypeMime["js"] {
			fr.(*concatFileReader).SetSeparator([]byte(";"))
		}
	}
	if err != nil {
		Error.Println(err)
		return false
	}

	fw, err = openOutputFile(t.dst)
	if err != nil {
		Error.Println(err)
		fr.Close()
		return false
	}

	// synchronize file
	if t.sync {
		_, err = io.Copy(fw, fr)
		fr.Close()
		fw.Close()
		if err != nil {
			Error.Println(err)
			return false
		}
		preserveAttributes(t.srcs[0], t.root, t.dst)
		Info.Println("copy", srcName, "to", dstName)
		return true
	}

	b, err := ioutil.ReadAll(fr)
	if err != nil {
		fr.Close()
		fw.Close()
		Error.Println("cannot minify "+srcName+":", err)
		return false
	}
	w := bytes.NewBuffer(make([]byte, 0, len(b)))

	success := true
	startTime := time.Now()
	if err = m.Minify(fileMimetype, w, bytes.NewReader(b)); err != nil {
		w = bytes.NewBuffer(b) // copy original
		Error.Println("cannot minify "+srcName+":", err)
		success = false
	}

	rLen, wLen := len(b), w.Len()
	_, err = io.Copy(fw, w)
	fr.Close()
	fw.Close()

	if !quiet {
		dur := time.Since(startTime)
		speed := "Inf MB"
		if 0 < dur {
			speed = humanize.Bytes(uint64(float64(rLen) / dur.Seconds()))
		}
		ratio := 1.0
		if 0 < rLen {
			ratio = float64(wLen) / float64(rLen)
		}

		stats := fmt.Sprintf("(%9v, %6v, %6v, %5.1f%%, %6v/s)", dur, humanize.Bytes(uint64(rLen)), humanize.Bytes(uint64(wLen)), ratio*100, speed)
		if srcName != dstName {
			fmt.Println(stats, "-", srcName, "to", dstName)
		} else {
			fmt.Println(stats, "-", srcName)
		}
	}

	// remove original that was renamed, when overwriting files
	for i := range t.srcs {
		if t.srcs[i] == t.dst+".bak" {
			if err == nil {
				if err = os.Remove(t.srcs[i]); err != nil {
					Error.Println(err)
					return false
				}
			} else {
				if err = os.Remove(t.dst); err != nil {
					Error.Println(err)
					return false
				} else if err = os.Rename(t.srcs[i], t.dst); err != nil {
					Error.Println(err)
					return false
				}
			}
			t.srcs[i] = t.dst
			break
		}
	}
	preserveAttributes(t.srcs[0], t.root, t.dst)
	return success
}

func preserveAttributes(src, root, dst string) {
	if src == "" || dst == "" {
		return
	}
Next:
	srcInfo, err := os.Stat(src)
	if err != nil {
		Warning.Println(err)
		return
	}

	if preserveMode {
		err = os.Chmod(dst, srcInfo.Mode().Perm())
		if err != nil {
			Warning.Println(err)
		}
	}
	if preserveOwnership {
		if uid, gid, ok := getOwnership(srcInfo); ok {
			err = os.Chown(dst, uid, gid)
			if err != nil {
				Warning.Println(err)
			}
		} else {
			Warning.Println(fmt.Errorf("preserve ownership not supported on platform"))
		}
	}
	if preserveTimestamps {
		err = os.Chtimes(dst, atime.Get(srcInfo), srcInfo.ModTime())
		if err != nil {
			Warning.Println(err)
		}
	}
	if src != root {
		// preserve on until root path
		src = filepath.Dir(src)
		dst = filepath.Dir(dst)
		goto Next
	}
}

// IsDir returns true if the passed string looks like it specifies a directory, false otherwise.
func IsDir(dir string) bool {
	if 0 < len(dir) && dir[len(dir)-1] == '/' {
		return true
	}
	info, err := os.Lstat(dir)
	return err == nil && info.Mode().IsDir() && info.Mode()&os.ModeSymlink == 0
}

// SameFile returns true if the two file paths specify the same path.
// While Linux is case-preserving case-sensitive (and therefore a string comparison will work),
// Windows is case-preserving case-insensitive; we use os.SameFile() to work cross-platform.
func SameFile(filename1 string, filename2 string) (bool, error) {
	fi1, err := os.Stat(filename1)
	if err != nil {
		return false, err
	}

	fi2, err := os.Stat(filename2)
	if err != nil {
		return false, err
	}
	return os.SameFile(fi1, fi2), nil
}
