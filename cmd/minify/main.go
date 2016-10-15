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
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/matryer/try"
	flag "github.com/spf13/pflag"
	min "github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/js"
	"github.com/tdewolff/minify/json"
	"github.com/tdewolff/minify/svg"
	"github.com/tdewolff/minify/xml"
)

var filetypeMime = map[string]string{
	"css":  "text/css",
	"htm":  "text/html",
	"html": "text/html",
	"js":   "text/javascript",
	"json": "application/json",
	"svg":  "image/svg+xml",
	"xml":  "text/xml",
}

var (
	m         *min.M
	recursive bool
	concat    bool
	hidden    bool
	list      bool
	verbose   bool
	watch     bool
	update    bool
	pattern   *regexp.Regexp
)

type task struct {
	srcs   []string
	srcDir string
	dst    string
}

type countingReader struct {
	io.Reader
	N int
}

func (w *countingReader) Read(p []byte) (int, error) {
	n, err := w.Reader.Read(p)
	w.N += n
	return n, err
}

type countingWriter struct {
	io.Writer
	N int
}

func (w *countingWriter) Write(p []byte) (int, error) {
	n, err := w.Writer.Write(p)
	w.N += n
	return n, err
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

	htmlMinifier := &html.Minifier{}
	xmlMinifier := &xml.Minifier{}

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] [input]\n\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nInput:\n  Files or directories, leave blank to use stdin\n")
	}
	flag.StringVarP(&output, "output", "o", "", "Output file or directory, leave blank to use stdout")
	flag.StringVar(&mimetype, "mime", "", "Mimetype (text/css, application/javascript, ...), optional for input filenames, has precendence over -type")
	flag.StringVar(&filetype, "type", "", "Filetype (css, html, js, ...), optional for input filenames")
	flag.StringVar(&match, "match", "", "Filename pattern matching using regular expressions, see https://github.com/google/re2/wiki/Syntax")
	flag.BoolVarP(&recursive, "recursive", "r", false, "Recursively minify directories")
	flag.BoolVarP(&concat, "concat", "", false, "Concatenate inputs to single output file")
	flag.BoolVarP(&hidden, "all", "a", false, "Minify all files, including hidden files and files in hidden directories")
	flag.BoolVarP(&list, "list", "l", false, "List all accepted filetypes")
	flag.BoolVarP(&verbose, "verbose", "v", false, "Verbose")
	flag.BoolVarP(&watch, "watch", "w", false, "Watch files and minify upon changes")
	flag.BoolVarP(&update, "update", "u", false, "Update binary")

	flag.StringVar(&siteurl, "url", "", "URL of file to enable URL minification")
	flag.BoolVar(&htmlMinifier.KeepDefaultAttrVals, "html-keep-default-attrvals", false, "Preserve default attribute values")
	flag.BoolVar(&htmlMinifier.KeepWhitespace, "html-keep-whitespace", false, "Preserve whitespace characters but still collapse multiple whitespace into one")
	flag.BoolVar(&xmlMinifier.KeepWhitespace, "xml-keep-whitespace", false, "Preserve whitespace characters but still collapse multiple whitespace into one")
	flag.Parse()
	rawInputs := flag.Args()

	Error = log.New(os.Stderr, "ERROR: ", 0)
	if verbose {
		Info = log.New(os.Stderr, "INFO: ", 0)
	} else {
		Info = log.New(ioutil.Discard, "INFO: ", 0)
	}

	if update {
		if err := equinoxUpdate(); err != nil {
			Error.Fatalln(err)
		}
		return
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
		return
	}

	usePipe := len(rawInputs) == 0
	mimetype = getMimetype(mimetype, filetype, usePipe)

	var err error
	if match != "" {
		pattern, err = regexp.Compile(match)
		if err != nil {
			Error.Fatalln(err)
		}
	}

	tasks, dirDst, ok := expandInputs(rawInputs)
	if !ok {
		os.Exit(1)
	}

	if output != "" {
		output = sanitizePath(output)
		if dirDst {
			if output[len(output)-1] != '/' {
				output += "/"
			}
			if err := os.MkdirAll(output, 0777); err != nil {
				Error.Fatalln(err)
			}
		}
	}

	if ok = expandOutputs(output, usePipe, &tasks); !ok {
		os.Exit(1)
	}

	m = min.New()
	m.AddFunc("text/css", css.Minify)
	m.Add("text/html", htmlMinifier)
	m.AddFunc("text/javascript", js.Minify)
	m.AddFunc("image/svg+xml", svg.Minify)
	m.AddFuncRegexp(regexp.MustCompile("[/+]json$"), json.Minify)
	m.AddRegexp(regexp.MustCompile("[/+]xml$"), xmlMinifier)

	if m.URL, err = url.Parse(siteurl); err != nil {
		Error.Fatalln(err)
	}

	var watcher *RecursiveWatcher
	if watch {
		if usePipe || output == "" {
			Error.Fatalln("watch only works with files that do not overwrite themselves")
		} else if len(rawInputs) > 1 {
			Error.Fatalln("watch only works with one input directory")
		} else if concat {
			Error.Fatalln("watch doesn't work together with concatenation")
		}

		input := sanitizePath(rawInputs[0])
		info, err := os.Stat(input)
		if err != nil || !info.Mode().IsDir() {
			Error.Fatalln("watch only works with an input directory")
		}

		watcher, err = NewRecursiveWatcher(input, recursive)
		if err != nil {
			Error.Fatalln(err)
		}
		defer watcher.Close()
	}

	start := time.Now()

	var fails int32
	if verbose {
		for _, t := range tasks {
			if ok := minify(mimetype, t); !ok {
				fails++
			}
		}
	} else {
		var wg sync.WaitGroup
		for _, t := range tasks {
			wg.Add(1)
			go func(t task) {
				defer wg.Done()
				if ok := minify(mimetype, t); !ok {
					atomic.AddInt32(&fails, 1)
				}
			}(t)
		}
		wg.Wait()
	}

	if watch {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)

		input := sanitizePath(rawInputs[0])
		files := watcher.Run()
		for files != nil {
			select {
			case <-c:
				watcher.Close()
			case file, ok := <-files:
				if !ok {
					files = nil
					break
				}
				file = sanitizePath(file)
				if strings.HasPrefix(file, input) {
					t := task{srcs: []string{file}, srcDir: input}
					if t.dst, err = getOutputFilename(output, t); err != nil {
						Error.Println(err)
						continue
					}
					if !verbose {
						fmt.Fprintln(os.Stderr, file, "changed")
					}
					if ok := minify(mimetype, t); !ok {
						fails++
					}
				}
			}
		}
	}

	if verbose {
		Info.Println(time.Since(start), "total")
	}

	if fails > 0 {
		os.Exit(1)
	}
}

func getMimetype(mimetype, filetype string, usePipe bool) string {
	if mimetype == "" && filetype != "" {
		var ok bool
		if mimetype, ok = filetypeMime[filetype]; !ok {
			Error.Fatalln("cannot find mimetype for filetype", filetype)
		}
	}
	if mimetype == "" && usePipe {
		Error.Fatalln("must specify mime or type for stdin")
	}

	if verbose {
		if mimetype == "" {
			Info.Println("infer mimetype from file extensions")
		} else {
			Info.Println("use mimetype", mimetype)
		}
	}
	return mimetype
}

func sanitizePath(p string) string {
	p = path.Clean(filepath.ToSlash(p))
	if p[len(p)-1] != '/' {
		if info, err := os.Stat(p); err == nil && info.Mode().IsDir() {
			p += "/"
		}
	}
	return p
}

func validFile(info os.FileInfo) bool {
	if info.Mode().IsRegular() && len(info.Name()) > 0 && (hidden || info.Name()[0] != '.') {
		if pattern != nil && !pattern.MatchString(info.Name()) {
			return false
		}

		ext := path.Ext(info.Name())
		if len(ext) > 0 {
			ext = ext[1:]
		}

		if _, ok := filetypeMime[ext]; !ok {
			return false
		}
		return true
	}
	return false
}

func validDir(info os.FileInfo) bool {
	return info.Mode().IsDir() && len(info.Name()) > 0 && (hidden || info.Name()[0] != '.')
}

func expandInputs(inputs []string) ([]task, bool, bool) {
	ok := true
	dirDst := false
	tasks := []task{}
	for _, input := range inputs {
		input = sanitizePath(input)
		info, err := os.Stat(input)
		if err != nil {
			Error.Println(err)
			ok = false
			continue
		}

		if info.Mode().IsRegular() {
			tasks = append(tasks, task{[]string{filepath.ToSlash(input)}, "", ""})
		} else if info.Mode().IsDir() {
			expandDir(input, &tasks, &ok)
			dirDst = true
		} else {
			Error.Println("not a file or directory", input)
			ok = false
		}
	}
	if len(tasks) > 1 {
		if concat {
			tasks[0].srcDir = ""
			for _, task := range tasks[1:] {
				tasks[0].srcs = append(tasks[0].srcs, task.srcs[0])
			}
			tasks = tasks[:1]
		} else {
			dirDst = true
		}
	}

	if verbose {
		if len(inputs) == 0 {
			Info.Println("minify from stdin")
		} else if len(tasks) == 1 {
			if concat {
				Info.Println("minify and concatenate input files (" + strings.Join(tasks[0].srcs, "+") + ")")
			} else {
				Info.Println("minify input file", tasks[0].srcs[0])
			}
		} else {
			Info.Println("minify", len(tasks), "input files")
		}
	}
	return tasks, dirDst, ok
}

func expandDir(input string, tasks *[]task, ok *bool) {
	if !recursive {
		if verbose {
			Info.Println("expanding directory", input)
		}

		infos, err := ioutil.ReadDir(input)
		if err != nil {
			Error.Println(err)
			*ok = false
		}
		for _, info := range infos {
			if validFile(info) {
				*tasks = append(*tasks, task{[]string{path.Join(input, info.Name())}, input, ""})
			}
		}
	} else {
		if verbose {
			Info.Println("expanding directory", input, "recursively")
		}

		err := filepath.Walk(input, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if validFile(info) {
				*tasks = append(*tasks, task{[]string{filepath.ToSlash(path)}, input, ""})
			} else if info.Mode().IsDir() && !validDir(info) && info.Name() != "." && info.Name() != ".." { // check for IsDir, so we don't skip the rest of the directory when we have an invalid file
				return filepath.SkipDir
			}
			return nil
		})
		if err != nil {
			Error.Println(err)
			*ok = false
		}
	}
}

func expandOutputs(output string, usePipe bool, tasks *[]task) bool {
	if output == "" && concat {
		Error.Println("no output given for concatenation")
		return false
	}

	if verbose {
		if output == "" {
			if usePipe {
				Info.Println("minify to stdout")
			} else {
				Info.Println("minify and overwrite original")
			}
		} else if output[len(output)-1] != '/' {
			Info.Println("minify to output file", output)
		} else if output == "./" {
			Info.Println("minify to current working directory")
		} else {
			Info.Println("minify to output directory", output)
		}
	}

	ok := true
	for i, t := range *tasks {
		if !usePipe && output == "" {
			(*tasks)[i].dst = (*tasks)[i].srcs[0] // no concatenation
		} else {
			var err error
			(*tasks)[i].dst, err = getOutputFilename(output, t)
			if err != nil {
				Error.Println(err)
				ok = false
			}
		}
	}

	if usePipe && len(*tasks) == 0 {
		*tasks = append(*tasks, task{[]string{""}, "", output})
	}
	return ok
}

func getOutputFilename(output string, t task) (string, error) {
	if len(output) > 0 && output[len(output)-1] == '/' && !concat {
		rel, err := filepath.Rel(t.srcDir, t.srcs[0])
		if err != nil {
			return "", err
		}
		return path.Clean(filepath.ToSlash(path.Join(output, rel))), nil
	}
	return output, nil
}

func openInputFile(input string) (*os.File, bool) {
	var r *os.File
	if input == "" {
		r = os.Stdin
	} else {
		err := try.Do(func(attempt int) (bool, error) {
			var err error
			r, err = os.Open(input)
			return attempt < 5, err
		})
		if err != nil {
			Error.Println(err)
			return nil, false
		}
	}
	return r, true
}

func openOutputFile(output string) (*os.File, bool) {
	var w *os.File
	if output == "" {
		w = os.Stdout
	} else {
		if err := os.MkdirAll(path.Dir(output), 0777); err != nil {
			Error.Println(err)
			return nil, false
		}
		err := try.Do(func(attempt int) (bool, error) {
			var err error
			w, err = os.OpenFile(output, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
			return attempt < 5, err
		})
		if err != nil {
			Error.Println(err)
			return nil, false
		}
	}
	return w, true
}

func minify(mimetype string, t task) bool {
	if mimetype == "" {
		for _, src := range t.srcs {
			if len(path.Ext(src)) > 0 {
				srcMimetype, ok := filetypeMime[path.Ext(src)[1:]]
				if !ok {
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
	}

	for _, src := range t.srcs {
		if src == t.dst && src != "" {
			src += ".bak"
			if err := os.Rename(t.dst, src); err != nil {
				Error.Println(err)
				return false
			}
			break
		}
	}

	frs := make([]io.ReadCloser, 0, len(t.srcs))
	frsMulti := make([]io.Reader, 0, len(t.srcs))
	for _, src := range t.srcs {
		fr, ok := openInputFile(src)
		if !ok {
			for _, fr := range frs {
				fr.Close()
			}
			return false
		}
		frs = append(frs, fr)
		frsMulti = append(frsMulti, fr)
	}
	r := &countingReader{io.MultiReader(frsMulti...), 0}

	fw, ok := openOutputFile(t.dst)
	if !ok {
		for _, fr := range frs {
			fr.Close()
		}
		return false
	}
	var w *countingWriter
	if fw == os.Stdout {
		w = &countingWriter{fw, 0}
	} else {
		w = &countingWriter{bufio.NewWriter(fw), 0}
	}

	success := true
	startTime := time.Now()
	err := m.Minify(mimetype, w, r)
	if err != nil {
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

		stats := fmt.Sprintf("(%9v, %6v, %5.1f%%, %6v/s)", dur, humanize.Bytes(uint64(w.N)), ratio*100, speed)
		if srcName != dstName {
			Info.Println(stats, "-", srcName, "to", dstName)
		} else {
			Info.Println(stats, "-", srcName)
		}
	}

	for _, fr := range frs {
		fr.Close()
	}
	if bw, ok := w.Writer.(*bufio.Writer); ok {
		bw.Flush()
	}
	fw.Close()

	for _, src := range t.srcs {
		if src == t.dst+".bak" {
			if err == nil {
				if err = os.Remove(src); err != nil {
					Error.Println(err)
					return false
				}
			} else {
				if err = os.Remove(t.dst); err != nil {
					Error.Println(err)
					return false
				} else if err = os.Rename(src, t.dst); err != nil {
					Error.Println(err)
					return false
				}
			}
			break
		}
	}
	return success
}
