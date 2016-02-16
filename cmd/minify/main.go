package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	flag "github.com/ogier/pflag"
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
	hidden    bool
	list      bool
	verbose   bool
	pattern   *regexp.Regexp
)

type task struct {
	src    string
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

func main() {
	output := ""
	mimetype := ""
	filetype := ""
	match := ""
	siteurl := ""

	htmlMinifier := &html.Minifier{}

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] [input]\n\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nInput:\n  Files or directories, optional when using piping\n")
	}
	flag.StringVarP(&output, "output", "o", "", "Output (concatenated) file (stdout when empty) or directory")
	flag.StringVar(&mimetype, "mime", "", "Mimetype (text/css, application/javascript, ...), optional for input filenames, has precendence over -type")
	flag.StringVar(&filetype, "type", "", "Filetype (css, html, js, ...), optional for input filenames")
	flag.StringVar(&match, "match", "", "Filename pattern matching using regular expressions, see https://github.com/google/re2/wiki/Syntax")
	flag.BoolVarP(&recursive, "recursive", "r", false, "Recursively minify directories")
	flag.BoolVarP(&hidden, "all", "a", false, "Minify all files, including hidden files and files in hidden directories")
	flag.BoolVarP(&list, "list", "l", false, "List all accepted filetypes")
	flag.BoolVarP(&verbose, "verbose", "v", false, "Verbose")

	flag.StringVar(&siteurl, "url", "", "URL of file to enable URL minification")
	flag.BoolVar(&htmlMinifier.KeepDefaultAttrVals, "html-keep-default-attrvals", false, "Preserve default attribute values")
	flag.BoolVar(&htmlMinifier.KeepWhitespace, "html-keep-whitespace", false, "Preserve whitespace characters but still collapse multiple whitespace into one")
	flag.Parse()
	rawInputs := flag.Args()

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

	if match != "" {
		var err error
		pattern, err = regexp.Compile(match)
		if err != nil {
			fmt.Fprintln(os.Stderr, "ERROR: "+err.Error())
			os.Exit(1)
		}
	}

	tasks, dirDst, ok := expandInputs(rawInputs)
	if !ok {
		os.Exit(1)
	}
	if ok = expandOutputs(output, dirDst, usePipe, &tasks); !ok {
		os.Exit(1)
	}

	m = min.New()
	m.AddFunc("text/css", css.Minify)
	m.Add("text/html", htmlMinifier)
	m.AddFunc("text/javascript", js.Minify)
	m.AddFunc("image/svg+xml", svg.Minify)
	m.AddFuncRegexp(regexp.MustCompile("[/+]json$"), json.Minify)
	m.AddFuncRegexp(regexp.MustCompile("[/+]xml$"), xml.Minify)

	var err error
	if m.URL, err = url.Parse(siteurl); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: "+err.Error())
		os.Exit(1)
	}

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
	if fails > 0 {
		os.Exit(1)
	}
}

func getMimetype(mimetype, filetype string, usePipe bool) string {
	if mimetype == "" && filetype != "" {
		var ok bool
		if mimetype, ok = filetypeMime[filetype]; !ok {
			fmt.Fprintln(os.Stderr, "ERROR: cannot find mimetype for filetype "+filetype)
			os.Exit(1)
		}
	}
	if mimetype == "" && usePipe {
		fmt.Fprintln(os.Stderr, "ERROR: must specify mime or type for stdin")
		os.Exit(1)
	}
	if verbose {
		if mimetype == "" {
			fmt.Fprintln(os.Stderr, "INFO:  infer mimetype from file extensions")
		} else {
			fmt.Fprintln(os.Stderr, "INFO:  use mimetype "+mimetype)
		}
	}
	return mimetype
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
			if verbose {
				fmt.Fprintln(os.Stderr, "WARN:  cannot infer mimetype from extension in "+info.Name())
			}
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
		input = path.Clean(filepath.ToSlash(input))
		info, err := os.Stat(input)
		if err != nil {
			fmt.Fprintln(os.Stderr, "ERROR: "+err.Error())
			ok = false
			continue
		}
		if info.Mode().IsRegular() {
			tasks = append(tasks, task{filepath.ToSlash(input), "", ""})
		} else if info.Mode().IsDir() {
			expandDir(input, &tasks, &ok)
			dirDst = true
		} else {
			fmt.Fprintln(os.Stderr, "ERROR: not a file or directory "+input)
			ok = false
		}
	}
	if len(tasks) > 1 {
		dirDst = true
	}

	if verbose {
		if len(inputs) == 0 {
			fmt.Fprintln(os.Stderr, "INFO:  minify from stdin")
		} else if len(tasks) == 1 {
			fmt.Fprintf(os.Stderr, "INFO:  minify input file %v\n", tasks[0].src)
		} else {
			fmt.Fprintf(os.Stderr, "INFO:  minify %v input files\n", len(tasks))
			for _, t := range tasks {
				fmt.Fprintf(os.Stderr, "  %v\n", t.src)
			}
		}
	}
	return tasks, dirDst, ok
}

func expandDir(input string, tasks *[]task, ok *bool) {
	if !recursive {
		if verbose {
			fmt.Fprintln(os.Stderr, "INFO:  expanding directory "+input)
		}
		infos, err := ioutil.ReadDir(input)
		if err != nil {
			fmt.Fprintln(os.Stderr, "ERROR: "+err.Error())
			*ok = false
		}
		for _, info := range infos {
			if validFile(info) {
				*tasks = append(*tasks, task{path.Join(input, info.Name()), input, ""})
			}
		}
	} else {
		if verbose {
			fmt.Fprintln(os.Stderr, "INFO:  expanding directory "+input+" recursively")
		}
		err := filepath.Walk(input, func(path string, info os.FileInfo, _ error) error {
			if validFile(info) {
				*tasks = append(*tasks, task{filepath.ToSlash(path), input, ""})
			} else if info.Mode().IsDir() && !validDir(info) && info.Name() != "." && info.Name() != ".." { // check for IsDir, so we don't skip the rest of the directory when we have an invalid file
				return filepath.SkipDir
			}
			return nil
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, "ERROR: "+err.Error())
			*ok = false
		}
	}
}

func expandOutputs(output string, dirDst, usePipe bool, tasks *[]task) bool {
	if output != "" {
		output = path.Clean(filepath.ToSlash(output))
		if output[len(output)-1] != '/' {
			info, err := os.Stat(output)
			if err == nil && info.Mode().IsDir() {
				output += "/"
			}
		}
		if dirDst {
			if output[len(output)-1] != '/' {
				output += "/"
			}
			if err := os.MkdirAll(output, 0777); err != nil {
				fmt.Fprintln(os.Stderr, "ERROR: "+err.Error())
				os.Exit(1)
			}
		} else if output[len(output)-1] == '/' {
			output += "out"
		}
	}

	if verbose {
		if output == "" {
			if usePipe {
				fmt.Fprintln(os.Stderr, "INFO:  minify to stdout")
			} else {
				fmt.Fprintln(os.Stderr, "INFO:  minify to overwrite itself")
			}
		} else if output[len(output)-1] != '/' {
			fmt.Fprintf(os.Stderr, "INFO:  minify to output file %v\n", output)
		} else if output == "./" {
			fmt.Fprintf(os.Stderr, "INFO:  minify to current working directory\n")
		} else {
			fmt.Fprintf(os.Stderr, "INFO:  minify to output directory %v\n", output)
		}
	}

	ok := true
	for i, t := range *tasks {
		if !usePipe && output == "" {
			(*tasks)[i].dst = (*tasks)[i].src
		} else {
			(*tasks)[i].dst = output
		}
		if len(output) > 0 && output[len(output)-1] == '/' {
			rel, err := filepath.Rel(t.srcDir, t.src)
			if err != nil {
				fmt.Fprintln(os.Stderr, "ERROR: "+err.Error())
				ok = false
			}
			(*tasks)[i].dst = path.Clean(filepath.ToSlash(path.Join(output, rel)))
		}
	}

	if usePipe && len(*tasks) == 0 {
		*tasks = append(*tasks, task{"", "", output})
	}
	return ok
}

func openInputFile(input string) (*os.File, bool) {
	var err error
	var r *os.File
	if input == "" {
		r = os.Stdin
	} else if r, err = os.Open(input); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: "+err.Error())
		return nil, false
	}
	return r, true
}

func openOutputFile(output string) (*os.File, bool) {
	var err error
	var w *os.File
	if output == "" {
		w = os.Stdout
	} else if w, err = os.OpenFile(output, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: "+err.Error())
		return nil, false
	}
	return w, true
}

func minify(mimetype string, t task) bool {
	if mimetype == "" {
		if len(path.Ext(t.src)) > 0 {
			var ok bool
			if mimetype, ok = filetypeMime[path.Ext(t.src)[1:]]; !ok {
				fmt.Fprintln(os.Stderr, "ERROR: cannot infer mimetype from extension in "+t.src)
				return false
			}
		}
	}

	srcName := t.src
	if srcName == "" {
		srcName = "stdin"
	}
	dstName := t.dst
	if dstName == "" {
		dstName = "stdin"
	}

	if t.src == t.dst && t.src != "" {
		t.src += ".bak"
		if err := os.Rename(t.dst, t.src); err != nil {
			fmt.Fprintln(os.Stderr, "ERROR: "+err.Error())
			return false
		}
	}

	fr, ok := openInputFile(t.src)
	if !ok {
		return false
	}
	r := &countingReader{fr, 0}

	fw, ok := openOutputFile(t.dst)
	if !ok {
		fr.Close()
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
		fmt.Fprintln(os.Stderr, "ERROR: cannot minify "+srcName+": "+err.Error())
		success = false
	}
	if verbose {
		dur := time.Since(startTime)
		fmt.Fprintf(os.Stderr, "INFO:  %v to %v\n  time:  %v\n  size:  %vB\n  ratio: %.1f%%\n  speed: %.1fMB/s\n", srcName, dstName, dur, w.N, float64(w.N)/float64(r.N)*100, float64(r.N)/dur.Seconds()/1000000)
	}

	fr.Close()
	if bw, ok := w.Writer.(*bufio.Writer); ok {
		bw.Flush()
	}
	fw.Close()

	if t.src == t.dst+".bak" {
		if err == nil {
			if err = os.Remove(t.src); err != nil {
				fmt.Fprintln(os.Stderr, "ERROR: "+err.Error())
				return false
			}
		} else {
			if err = os.Remove(t.dst); err != nil {
				fmt.Fprintln(os.Stderr, "ERROR: "+err.Error())
				return false
			} else if err = os.Rename(t.src, t.dst); err != nil {
				fmt.Fprintln(os.Stderr, "ERROR: "+err.Error())
				return false
			}
		}
	}
	return success
}
