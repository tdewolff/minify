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

var Version = "master"
var Commit = ""
var Date = ""

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
)

type Task struct {
	srcs []string
	dst  string
}

func NewTask(root, input, output string) Task {
	t := Task{[]string{input}, output}
	if 0 < len(output) && output[len(output)-1] == '/' {
		rel, err := filepath.Rel(root, input)
		if err != nil {
			Error.Fatalln(err)
		}
		t.dst = path.Clean(filepath.ToSlash(path.Join(output, rel)))
	}
	return t
}

var (
	Error *log.Logger
	Info  *log.Logger
)

func main() {
	output := ""
	mimetype := ""
	filetype := ""
	match := ""
	siteurl := ""

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
		fmt.Fprintf(os.Stderr, "\nInput:\n  Files or directories, leave blank to use stdin\n")
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
	flag.BoolVarP(&version, "version", "", false, "Version")

	flag.StringVar(&siteurl, "url", "", "URL of file to enable URL minification")
	flag.IntVar(&cssMinifier.Decimals, "css-decimals", 0, "Number of significant digits to preserve in numbers, 0 is all (DEPRECATED")
	flag.IntVar(&cssMinifier.Precision, "css-precision", 0, "Number of significant digits to preserve in numbers, 0 is all")
	flag.BoolVar(&htmlMinifier.KeepConditionalComments, "html-keep-conditional-comments", false, "Preserve all IE conditional comments")
	flag.BoolVar(&htmlMinifier.KeepDefaultAttrVals, "html-keep-default-attrvals", false, "Preserve default attribute values")
	flag.BoolVar(&htmlMinifier.KeepDocumentTags, "html-keep-document-tags", false, "Preserve html, head and body tags")
	flag.BoolVar(&htmlMinifier.KeepEndTags, "html-keep-end-tags", false, "Preserve all end tags")
	flag.BoolVar(&htmlMinifier.KeepWhitespace, "html-keep-whitespace", false, "Preserve whitespace characters but still collapse multiple into one")
	flag.BoolVar(&htmlMinifier.KeepQuotes, "html-keep-quotes", false, "Preserve quotes around attribute values")
	flag.IntVar(&svgMinifier.Decimals, "svg-decimals", 0, "Number of significant digits to preserve in numbers, 0 is all (DEPRECATED")
	flag.IntVar(&svgMinifier.Precision, "svg-precision", 0, "Number of significant digits to preserve in numbers, 0 is all")
	flag.BoolVar(&xmlMinifier.KeepWhitespace, "xml-keep-whitespace", false, "Preserve whitespace characters but still collapse multiple into one")
	if err := flag.Parse(os.Args[1:]); err != nil {
		fmt.Printf("Error: %v\n\n", err)
		flag.Usage()
		os.Exit(2)
	}
	rawInputs := flag.Args()
	useStdin := len(rawInputs) == 0

	Error = log.New(os.Stderr, "ERROR: ", 0)
	if verbose {
		Info = log.New(os.Stderr, "INFO: ", 0)
	} else {
		Info = log.New(ioutil.Discard, "INFO: ", 0)
	}

	if help {
		flag.Usage()
		os.Exit(0)
	}

	if version {
		if Version == "devel" {
			fmt.Printf("minify version devel+%.7s %s\n", Commit, Date)
		} else {
			fmt.Printf("minify version %s\n", Version)
		}
		os.Exit(0)
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
		os.Exit(0)
	}

	var err error
	if match != "" {
		pattern, err = regexp.Compile(match)
		if err != nil {
			Error.Fatalln(err)
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
		os.Exit(1)
	}
	if mimetype == "" && filetype == "" && useStdin {
		Error.Fatalln("must specify --mime or --type for stdin")
	}

	// detect mimetype, mimetype=="" means we'll infer mimetype from file extensions
	if mimetype == "" && filetype != "" {
		var ok bool
		if mimetype, ok = filetypeMime[filetype]; !ok {
			Error.Fatalln("cannot find mimetype for filetype", filetype)
		}
	}
	if verbose {
		if mimetype == "" {
			Info.Println("infer mimetype from file extensions")
		} else {
			Info.Println("use mimetype", mimetype)
		}
	}

	////////////////

	// set output, empty means stdout, ending in slash means a directory, otherwise a file
	dirDst := false
	if output != "" {
		output = sanitizePath(output)
		if output[len(output)-1] == '/' {
			dirDst = true
			if err := os.MkdirAll(output, 0777); err != nil {
				Error.Fatalln(err)
			}
		}
	}
	if !dirDst && (sync || watch) {
		if sync {
			Error.Println("--sync requires destination to be a directory")
		}
		if watch {
			Error.Println("--watch requires destination to be a directory")
		}
		os.Exit(1)
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

	var tasks []Task
	var roots []string
	if useStdin {
		tasks = append(tasks, NewTask("", "", output))
		roots = append(roots, "")
	} else {
		tasks, roots = createTasks(rawInputs, output)
	}

	// concatenate
	if 1 < len(tasks) && !dirDst {
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
		Error.Fatalln(err)
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
			Error.Fatalln(err)
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

				// skip files in output directory (which is also an input directory) for the first change
				// skips files that are not minified and stay put, as they are not explicitly copied, but that's ok
				if autoDir && root == output {
					if _, ok := skip[file]; !ok {
						skip[file] = true
						break
					}
				}

				if !verbose {
					Info.Println(file, "changed")
				}
				chanTasks <- NewTask(root, file, output)
			}
		}
	}

	fails := 0
	close(chanTasks)
	for n := 0; n < numWorkers; n++ {
		fails += <-chanFails
	}

	if verbose && !watch {
		Info.Println(time.Since(start), "total")
	}
	if fails > 0 {
		os.Exit(1)
	}
	os.Exit(0)
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
	p = filepath.ToSlash(p)
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
	if info.Mode().IsRegular() && len(info.Name()) > 0 && (hidden || info.Name()[0] != '.') {
		if pattern != nil && !pattern.MatchString(info.Name()) {
			return false
		}

		if !sync {
			ext := path.Ext(info.Name())
			if len(ext) > 0 {
				ext = ext[1:]
			}

			if _, ok := filetypeMime[ext]; !ok {
				return false
			}
		}
		return true
	}
	return false
}

func validDir(info os.FileInfo) bool {
	return info.Mode().IsDir() && len(info.Name()) > 0 && (hidden || info.Name()[0] != '.')
}

func createTasks(inputs []string, output string) ([]Task, []string) {
	tasks := []Task{}
	roots := []string{}
	for _, input := range inputs {
		input = sanitizePath(input)
		info, err := os.Stat(input)
		if err != nil {
			Error.Fatalln(err)
		}

		if info.Mode().IsRegular() {
			tasks = append(tasks, NewTask("", input, output))
		} else if info.Mode().IsDir() {
			roots = append(roots, input)
			if !recursive {
				infos, err := ioutil.ReadDir(input)
				if err != nil {
					Error.Fatalln(err)
				}
				for _, info := range infos {
					if validFile(info) {
						tasks = append(tasks, NewTask(input, path.Join(input, info.Name()), output))
					}
				}
			} else {
				err := filepath.Walk(input, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if validFile(info) {
						tasks = append(tasks, NewTask(input, path, output))
					} else if info.Mode().IsDir() && !validDir(info) && info.Name() != "." && info.Name() != ".." { // check for IsDir, so we don't skip the rest of the directory when we have an invalid file
						return filepath.SkipDir
					}
					return nil
				})
				if err != nil {
					Error.Fatalln(err)
				}
			}
		} else {
			Error.Fatalln("not a file or directory", input)
		}
	}
	return tasks, roots
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
	if mimetype == "" {
		for _, src := range t.srcs {
			if len(path.Ext(src)) > 0 {
				srcMimetype, ok := filetypeMime[path.Ext(src)[1:]]
				if !ok && sync {
					break // is sync==true, then len(t.srcs)==1
				} else if !ok {
					Error.Println("cannot infer mimetype from extension in", src)
					return false
				}
				if mimetype == "" {
					mimetype = srcMimetype
				} else if srcMimetype != mimetype {
					Error.Println("inferred mimetype", srcMimetype, "of", src, "for concatenation unequal to previous mimetypes", mimetype)
					return false
				}
			}
		}
	}

	// synchronizing files that are not minified but just copied to the same directory, no action needed
	if sync && mimetype == "" && t.srcs[0] == t.dst {
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

	fr, err := NewConcatFileReader(t.srcs, openInputFile)
	if err != nil {
		Error.Println(err)
		return false
	}
	if mimetype == filetypeMime["js"] {
		fr.SetSeparator([]byte("\n"))
	}
	r := NewCountingReader(fr)

	fw, err := openOutputFile(t.dst)
	if err != nil {
		Error.Println(err)
		fr.Close()
		return false
	}
	var w *countingWriter
	if fw == os.Stdout {
		w = NewCountingWriter(fw)
	} else {
		w = NewCountingWriter(bufio.NewWriter(fw))
	}

	// synchronize file
	if sync && mimetype == "" {
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

	success := true
	startTime := time.Now()
	if err = m.Minify(mimetype, w, r); err != nil {
		Error.Println("cannot minify "+srcName+":", err)
		success = false
	}
	if verbose {
		dur := time.Since(startTime)
		speed := "Inf MB"
		if dur > 0 {
			speed = humanize.Bytes(uint64(float64(r.N) / dur.Seconds()))
		}
		ratio := 1.0
		if r.N > 0 {
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
