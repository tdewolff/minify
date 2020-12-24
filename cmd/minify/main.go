package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/pprof"
	"sort"
	"strings"
	"time"

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
	"json": "application/json",
	"svg":  "image/svg+xml",
	"xml":  "text/xml",
}

var (
	help      bool
	hidden    bool
	list      bool
	m         *min.M
	pattern   *regexp.Regexp
	recursive bool
	verbose   bool
	version   bool
	watch     bool
	sync      bool
	bundle    bool
)

// Task is a minify task.
type Task struct {
	srcs []string
	dst  string
	sync bool
}

// NewTask returns a new Task.
func NewTask(root, input, output string, sync bool) (Task, error) {
	t := Task{[]string{input}, output, sync}
	if 0 < len(output) && output[len(output)-1] == '/' {
		rel, err := filepath.Rel(root, input)
		if err != nil {
			return Task{}, err
		}
		t.dst = path.Clean(filepath.ToSlash(path.Join(output, rel)))
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
	mimetype := ""
	filetype := ""
	match := ""
	siteurl := ""
	cpuprofile := ""
	memprofile := ""

	cssMinifier := &css.Minifier{}
	htmlMinifier := &html.Minifier{}
	jsMinifier := &js.Minifier{}
	jsonMinifier := &json.Minifier{}
	svgMinifier := &svg.Minifier{}
	xmlMinifier := &xml.Minifier{}

	flag := flag.NewFlagSet("minify", flag.ContinueOnError)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] [input]\n\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nInput:\n  Files or directories, leave blank to use stdin. Specify --mime or --type to use stdin and stdout.\n")
	}

	flag.BoolVarP(&help, "help", "h", false, "Show usage")
	flag.StringVarP(&output, "output", "o", "", "Output file or directory (must have trailing slash), leave blank to use stdout")
	flag.StringVar(&mimetype, "mime", "", "Mimetype (eg. text/css), optional for input filenames, has precedence over -type")
	flag.StringVar(&filetype, "type", "", "Filetype (eg. css), optional for input filenames")
	flag.StringVar(&match, "match", "", "Filename pattern matching using regular expressions")
	flag.BoolVarP(&recursive, "recursive", "r", false, "Recursively minify directories")
	flag.BoolVarP(&hidden, "all", "a", false, "Minify all files, including hidden files and files in hidden directories")
	flag.BoolVarP(&list, "list", "l", false, "List all accepted filetypes")
	flag.BoolVarP(&verbose, "verbose", "v", false, "Verbose")
	flag.BoolVarP(&watch, "watch", "w", false, "Watch files and minify upon changes")
	flag.BoolVarP(&sync, "sync", "s", false, "Copy all files to destination directory and minify when filetype matches")
	flag.BoolVarP(&bundle, "bundle", "b", false, "Bundle files by concatenation into a single file")
	flag.BoolVarP(&version, "version", "", false, "Version")

	flag.StringVar(&siteurl, "url", "", "URL of file to enable URL minification")
	flag.StringVar(&cpuprofile, "cpuprofile", "", "Export CPU profile")
	flag.StringVar(&memprofile, "memprofile", "", "Export memory profile")
	flag.IntVar(&cssMinifier.Precision, "css-precision", 0, "Number of significant digits to preserve in numbers, 0 is all")
	flag.BoolVar(&htmlMinifier.KeepComments, "html-keep-comments", false, "Preserve all comments")
	flag.BoolVar(&htmlMinifier.KeepConditionalComments, "html-keep-conditional-comments", false, "Preserve all IE conditional comments")
	flag.BoolVar(&htmlMinifier.KeepDefaultAttrVals, "html-keep-default-attrvals", false, "Preserve default attribute values")
	flag.BoolVar(&htmlMinifier.KeepDocumentTags, "html-keep-document-tags", false, "Preserve html, head and body tags")
	flag.BoolVar(&htmlMinifier.KeepEndTags, "html-keep-end-tags", false, "Preserve all end tags")
	flag.BoolVar(&htmlMinifier.KeepWhitespace, "html-keep-whitespace", false, "Preserve whitespace characters but still collapse multiple into one")
	flag.BoolVar(&htmlMinifier.KeepQuotes, "html-keep-quotes", false, "Preserve quotes around attribute values")
	flag.IntVar(&jsMinifier.Precision, "js-precision", 0, "Number of significant digits to preserve in numbers, 0 is all")
	flag.BoolVar(&jsMinifier.KeepVarNames, "js-keep-var-names", false, "Preserve original variable names")
	flag.IntVar(&jsonMinifier.Precision, "json-precision", 0, "Number of significant digits to preserve in numbers, 0 is all")
	flag.IntVar(&svgMinifier.Precision, "svg-precision", 0, "Number of significant digits to preserve in numbers, 0 is all")
	flag.BoolVar(&xmlMinifier.KeepWhitespace, "xml-keep-whitespace", false, "Preserve whitespace characters but still collapse multiple into one")
	if len(os.Args) == 1 {
		fmt.Printf("minify: must specify --mime or --type in order to use stdin and stdout\n")
		fmt.Printf("Try 'minify --help' for more information\n")
		return 1
	} else if err := flag.Parse(os.Args[1:]); err != nil {
		fmt.Printf("minify: %v\n", err)
		fmt.Printf("Try 'minify --help' for more information\n")
		return 1
	}
	inputs := flag.Args()
	useStdin := len(inputs) == 0

	Error = log.New(os.Stderr, "ERROR: ", 0)
	Warning = log.New(os.Stderr, "WARNING: ", 0)
	if verbose {
		Info = log.New(os.Stderr, "", 0)
	} else {
		Info = log.New(ioutil.Discard, "", 0)
	}

	if help {
		flag.Usage()
		return 0
	}

	if version {
		fmt.Printf("minify %s\n", Version)
		return 0
	}

	if list {
		var keys []string
		for k := range filetypeMime {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Println(k + "\t" + filetypeMime[k])
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

	var err error
	if match != "" {
		pattern, err = regexp.Compile(match)
		if err != nil {
			Error.Println(err)
			return 1
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

	if (useStdin || output == "") && (watch || sync || recursive) {
		if watch {
			Error.Println("--watch doesn't work on stdin and stdout, specify input and output")
		}
		if sync {
			Error.Println("--sync doesn't work on stdin and stdout, specify input and output")
		}
		if recursive {
			Error.Println("--recursive doesn't work on stdin and stdout, specify input and output")
		}
		return 1
	}
	if mimetype == "" && useStdin {
		Error.Println("must specify --mime or --type for stdin")
		return 1
	} else if mimetype != "" && sync {
		Error.Println("must specify either --sync or --mime/--type")
		return 1
	}
	if verbose {
		if mimetype == "" {
			Info.Println("infer mimetype from file extensions")
		} else {
			Info.Println("use mimetype", mimetype)
		}
	}

	////////////////

	for i, input := range inputs {
		inputs[i] = sanitizePath(input)
	}

	// set output, empty means stdout, ending in slash means a directory, otherwise a file
	dirDst := false
	if output != "" {
		dirSrc := 1 < len(inputs) || len(inputs) == 1 && inputs[0][len(inputs[0])-1] == '/'
		if !bundle && dirSrc && output[len(output)-1] != '/' {
			output += "/"
		}
		output = sanitizePath(output)
		dirDst = output[len(output)-1] == '/'
		if dirDst && bundle {
			Error.Println("--bundle requires destination to be stdout or a file")
			return 1
		}
	}
	if !dirDst && (sync || watch) {
		if sync {
			Error.Println("--sync requires destination to be a directory")
		}
		if watch {
			Error.Println("--watch requires destination to be a directory")
		}
		return 1
	}
	if verbose {
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
	}
	if dirDst {
		if err := os.MkdirAll(output, 0777); err != nil {
			Error.Println(err)
			return 1
		}
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

	////////////////

	m = min.New()
	m.Add("text/css", cssMinifier)
	m.Add("text/html", htmlMinifier)
	m.Add("image/svg+xml", svgMinifier)
	m.AddRegexp(regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$"), jsMinifier)
	m.AddRegexp(regexp.MustCompile("[/+]json$"), jsonMinifier)
	m.AddRegexp(regexp.MustCompile("[/+]xml$"), xmlMinifier)

	if m.URL, err = url.Parse(siteurl); err != nil {
		Error.Println(err)
		return 1
	}

	numWorkers := 1
	if !verbose && len(tasks) > 1 {
		numWorkers = 4
		if n := runtime.NumCPU(); n > numWorkers {
			numWorkers = n
		}
	}

	start := time.Now()

	chanTasks := make(chan Task, 100)
	chanFails := make(chan int, numWorkers)
	for n := 0; n < numWorkers; n++ {
		go minifyWorker(mimetype, chanTasks, chanFails)
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
		for _, task := range tasks {
			chanTasks <- task
		}

		autoDir := false
		for _, root := range roots {
			watcher.AddPath(root)
			if root == output {
				autoDir = true
			}
		}
		skip := map[string]bool{}

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
				file = sanitizePath(file)

				// find longest common path among roots
				root := roots[0]
				for _, path := range roots[1:] {
					pathRel, err1 := filepath.Rel(path, file)
					rootRel, err2 := filepath.Rel(root, file)
					if err2 != nil || err1 == nil && len(pathRel) < len(rootRel) {
						root = path
					}
				}

				if autoDir && root == output {
					// skip files in output directory (which is also an input directory) for the first change
					// skips files that are not minified and stay put as they are not explicitly copied, but that's ok
					if _, ok := skip[file]; !ok {
						skip[file] = true
						break
					}
				}

				if !verbose {
					Info.Println(file, "changed")
				}
				task, err := NewTask(root, file, output, !fileMatches(file))
				if err != nil {
					Error.Println(err)
					return 1
				}
				chanTasks <- task
			}
		}
	}

	fails := 0
	close(chanTasks)
	for n := 0; n < numWorkers; n++ {
		fails += <-chanFails
	}

	if verbose && !watch {
		Info.Println("finished in", time.Since(start))
	}
	if 0 < fails {
		return 1
	}
	return 0
}

func minifyWorker(mimetype string, chanTasks <-chan Task, chanFails chan<- int) {
	fails := 0
	for task := range chanTasks {
		if ok := minify(mimetype, task); !ok {
			fails++
		}
	}
	chanFails <- fails
}

func sanitizePath(p string) string {
	p = filepath.ToSlash(p) // replace \ with / for Windows
	isDir := p[len(p)-1] == '/'
	p = path.Clean(p)
	if isDir {
		p += "/"
	} else if info, err := os.Stat(p); err == nil && info.Mode().IsDir() {
		p += "/"
	}
	return p
}

func validFile(info os.FileInfo) bool {
	return info.Mode().IsRegular() && 0 < len(info.Name()) && (hidden || info.Name()[0] != '.')
}

func validDir(info os.FileInfo) bool {
	return info.Mode().IsDir() && 0 < len(info.Name()) && (hidden || info.Name()[0] != '.')
}

func fileMatches(filename string) bool {
	if pattern != nil && !pattern.MatchString(filename) {
		return false
	}

	ext := path.Ext(filename)
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
		info, err := os.Stat(input)
		if err != nil {
			return nil, nil, err
		}

		if info.Mode().IsRegular() {
			valid := pattern == nil || pattern.MatchString(info.Name())
			if valid || sync {
				task, err := NewTask(filepath.Dir(input), input, output, !valid)
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
			roots = append(roots, input)
			err := filepath.Walk(input, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				path = sanitizePath(path)
				if validFile(info) {
					valid := fileMatches(info.Name())
					if valid || sync {
						task, err := NewTask(input, path, output, !valid)
						if err != nil {
							return err
						}
						tasks = append(tasks, task)
					}
				} else if info.Mode().IsDir() && !validDir(info) && info.Name() != "." && info.Name() != ".." { // check for IsDir, so we don't skip the rest of the directory when we have an invalid file
					return filepath.SkipDir
				}
				return nil
			})
			if err != nil {
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
			return nil, err
		}
	}
	return r, nil
}

func openOutputFile(output string) (*os.File, error) {
	var w *os.File
	if output == "" {
		w = os.Stdout
	} else {
		if err := os.MkdirAll(path.Dir(output), 0777); err != nil {
			return nil, err
		}
		err := try.Do(func(attempt int) (bool, error) {
			var ferr error
			w, ferr = os.OpenFile(output, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
			return attempt < 5, ferr
		})
		if err != nil {
			return nil, err
		}
	}
	return w, nil
}

func minify(mimetype string, t Task) bool {
	if mimetype == "" && !t.sync {
		for _, src := range t.srcs {
			ext := path.Ext(src)
			if 0 < len(ext) {
				ext = ext[1:]
			}
			srcMimetype, ok := filetypeMime[ext]
			if !ok {
				Error.Println("cannot infer mimetype from extension in", src, ", set --type or --mime explicitly")
				return false
			}
			if mimetype == "" {
				mimetype = srcMimetype
			} else if srcMimetype != mimetype {
				Error.Println("inferred mimetype", srcMimetype, "of", src, "for concatenation unequal to previous mimetypes, set --type or --mime explicitly", mimetype)
				return false
			}
		}
	}

	// synchronizing files that are not minified but just copied to the same directory, no action needed
	if t.sync && t.srcs[0] == t.dst {
		return true
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
		dstName = "stdin"
	} else {
		// rename original when overwriting
		for i := range t.srcs {
			if t.srcs[i] == t.dst {
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

	fr, err := newConcatFileReader(t.srcs, openInputFile)
	if err != nil {
		Error.Println(err)
		return false
	}
	if mimetype == filetypeMime["js"] {
		fr.SetSeparator([]byte("\n"))
	}
	fw, err := openOutputFile(t.dst)
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
		Info.Println("copy", srcName, "to", dstName)
		return true
	}

	r := newCountingReader(fr)
	var w *countingWriter
	if fw == os.Stdout {
		w = newCountingWriter(fw)
	} else {
		w = newCountingWriter(bufio.NewWriter(fw))
	}

	success := true
	startTime := time.Now()
	if err = m.Minify(mimetype, w, r); err != nil {
		if t.dst == "" {
			fmt.Fprintf(os.Stdout, "\n\n")
		}
		Error.Println("cannot minify "+srcName+":", err)
		success = false
	}
	if verbose {
		dur := time.Since(startTime)
		speed := "Inf MB"
		if 0 < dur {
			speed = humanize.Bytes(uint64(float64(r.N) / dur.Seconds()))
		}
		ratio := 1.0
		if 0 < r.N {
			ratio = float64(w.N) / float64(r.N)
		}

		stats := fmt.Sprintf("(%9v, %6v, %6v, %5.1f%%, %6v/s)", dur, humanize.Bytes(uint64(r.N)), humanize.Bytes(uint64(w.N)), ratio*100, speed)
		if srcName != dstName {
			Info.Println(stats, "-", srcName, "to", dstName)
		} else {
			Info.Println(stats, "-", srcName)
		}
	}

	fr.Close()
	if bw, ok := w.Writer.(*bufio.Writer); ok {
		bw.Flush()
	}
	fw.Close()

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
	return success
}
