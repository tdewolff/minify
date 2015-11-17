package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"sync"
	"time"

	flag "github.com/ogier/pflag"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/js"
	"github.com/tdewolff/minify/json"
	"github.com/tdewolff/minify/svg"
	"github.com/tdewolff/minify/xml"
)

var FiletypeMime = map[string]string{
	"css":  "text/css",
	"htm":  "text/html",
	"html": "text/html",
	"js":   "application/javascript",
	"json": "application/json",
	"svg":  "image/svg+xml",
	"xml":  "text/xml",
}

var (
	recursive bool
	hidden    bool
	list      bool
	verbose   bool
	force     bool
	pattern   *regexp.Regexp
)

type CountingReader struct {
	io.Reader
	N int
}

func (w *CountingReader) Read(p []byte) (int, error) {
	n, err := w.Reader.Read(p)
	w.N += n
	return n, err
}

type CountingWriter struct {
	io.Writer
	N int
}

func (w *CountingWriter) Write(p []byte) (int, error) {
	n, err := w.Writer.Write(p)
	w.N += n
	return n, err
}

func main() {
	output := ""
	mimetype := ""
	filetype := ""
	match := ""
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
	flag.BoolVarP(&force, "force", "f", false, "Force overwriting existing files")
	flag.Parse()
	rawInputs := flag.Args()

	if list {
		var keys []string
		for k := range FiletypeMime {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Println(k + "\t" + FiletypeMime[k])
		}
		return
	}

	if mimetype == "" && filetype != "" {
		var ok bool
		if mimetype, ok = FiletypeMime[filetype]; !ok {
			fmt.Fprintln(os.Stderr, "ERROR: cannot find mimetype for filetype "+filetype)
			os.Exit(1)
		}
	}
	if mimetype == "" && len(rawInputs) == 0 {
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

	if match != "" {
		var err error
		pattern, err = regexp.Compile(match)
		if err != nil {
			fmt.Fprintln(os.Stderr, "ERROR: "+err.Error())
			os.Exit(1)
		}
	}

	inputs, dirDst, ok := expandInputs(rawInputs)
	if !ok {
		os.Exit(1)
	}
	if output != "" {
		output = path.Clean(filepath.ToSlash(output))
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
	} else if len(rawInputs) > 0 {
		output = "./"
	}
	if verbose {
		if len(rawInputs) == 0 {
			fmt.Fprintln(os.Stderr, "INFO:  minify from stdin")
		} else {
			fmt.Fprintf(os.Stderr, "INFO:  minify %v input files\n", len(inputs))
			for _, input := range inputs {
				fmt.Fprintf(os.Stderr, "  %v\n", input)
			}
		}
		if output == "" {
			fmt.Fprintln(os.Stderr, "INFO:  minify to stdout")
		} else if output[len(output)-1] != '/' {
			fmt.Fprintf(os.Stderr, "INFO:  minify to output file %v\n", output)
		} else if output == "./" {
			fmt.Fprintf(os.Stderr, "INFO:  minify to current working directory\n")
		} else {
			fmt.Fprintf(os.Stderr, "INFO:  minify to output directory %v\n", output)
		}
	}

	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/html", html.Minify)
	m.AddFunc("application/javascript", js.Minify)
	m.AddFunc("image/svg+xml", svg.Minify)
	m.AddFuncRegexp(regexp.MustCompile("[/+]json$"), json.Minify)
	m.AddFuncRegexp(regexp.MustCompile("[/+]xml$"), xml.Minify)

	if len(rawInputs) == 0 && len(inputs) == 0 {
		inputs = append(inputs, "")
	}
	success := true
	var wg sync.WaitGroup
	for _, input := range inputs {
		wg.Add(1)
		go func(output string, input string) {
			defer wg.Done()

			if len(output) > 0 && output[len(output)-1] == '/' {
				rel, err := filepath.Rel(output, input)
				if err != nil {
					fmt.Fprintln(os.Stderr, "ERROR: "+err.Error())
					return
				}
				output = path.Clean(filepath.ToSlash(path.Join(output, rel)))
			}

			mimetype := mimetype // don't shadow global variable
			if mimetype == "" {
				if len(path.Ext(input)) > 0 {
					if mimetype, ok = FiletypeMime[path.Ext(input)[1:]]; !ok {
						fmt.Fprintln(os.Stderr, "ERROR: cannot infer mimetype from extension in "+input)
						return
					}
				}
			}

			originalInput := input
			if input == output {
				input += ".bak"
				if err := os.Rename(output, input); err != nil {
					fmt.Fprintln(os.Stderr, "ERROR: "+err.Error())
					return
				}
			}

			fr, ok := openInputFile(input)
			if !ok {
				return
			}
			r := &CountingReader{fr, 0}

			fw, ok := openOutputFile(output)
			if !ok {
				fr.Close()
				return
			}
			var w *CountingWriter
			if fw == os.Stdout {
				w = &CountingWriter{fw, 0}
			} else {
				w = &CountingWriter{bufio.NewWriter(fw), 0}
			}

			t := time.Now()
			err := m.Minify(mimetype, w, r)
			if err != nil {
				fmt.Fprintln(os.Stderr, "ERROR: cannot minify input "+originalInput+": "+err.Error())
				success = false
			}
			if verbose {
				d := time.Since(t)
				fmt.Fprintf(os.Stderr, "INFO:  %v to %v\n  time:  %v\n  size:  %vB\n  ratio: %.1f%%\n  speed: %.1fMB/s\n", originalInput, output, d, w.N, float64(w.N)/float64(r.N)*100, float64(r.N)/d.Seconds()/1000000)
			}

			fr.Close()
			if bw, ok := w.Writer.(*bufio.Writer); ok {
				bw.Flush()
			}
			fw.Close()

			if input == output+".bak" {
				if err == nil {
					if err = os.Remove(input); err != nil {
						fmt.Fprintln(os.Stderr, "ERROR: "+err.Error())
					}
				} else {
					if err = os.Remove(output); err != nil {
						fmt.Fprintln(os.Stderr, "ERROR: "+err.Error())
					} else if err = os.Rename(input, output); err != nil {
						fmt.Fprintln(os.Stderr, "ERROR: "+err.Error())
					}
				}
			}
		}(output, input)
	}
	wg.Wait()
	if !success {
		os.Exit(1)
	}
}

func expandInputs(inputs []string) ([]string, bool, bool) {
	ok := true
	dirDst := false
	expanded := []string{}
	for _, input := range inputs {
		input = path.Clean(filepath.ToSlash(input))
		info, err := os.Stat(input)
		if err != nil {
			fmt.Fprintln(os.Stderr, "ERROR: "+err.Error())
			ok = false
			continue
		}
		if info.Mode().IsRegular() {
			expanded = append(expanded, filepath.ToSlash(input))
		} else if info.Mode().IsDir() {
			dirDst = true
			if !recursive {
				if verbose {
					fmt.Fprintln(os.Stderr, "INFO:  expanding directory "+input)
				}
				infos, err := ioutil.ReadDir(input)
				if err != nil {
					fmt.Fprintln(os.Stderr, "ERROR: "+err.Error())
					ok = false
				}
				for _, info := range infos {
					if validFile(info) {
						expanded = append(expanded, path.Join(input, info.Name()))
					}
				}
			} else {
				if verbose {
					fmt.Fprintln(os.Stderr, "INFO:  expanding directory "+input+" recursively")
				}
				err = filepath.Walk(input, func(path string, info os.FileInfo, _ error) error {
					if validFile(info) {
						expanded = append(expanded, filepath.ToSlash(path))
					} else if info.Mode().IsDir() && !validDir(info) && info.Name() != "." && info.Name() != ".." { // check for IsDir, so we don't skip the rest of the directory when we have an invalid file
						return filepath.SkipDir
					}
					return nil
				})
				if err != nil {
					fmt.Fprintln(os.Stderr, "ERROR: "+err.Error())
					ok = false
				}
			}
		} else {
			fmt.Fprintln(os.Stderr, "ERROR: not a file or directory "+input)
			ok = false
		}
	}
	if len(expanded) > 1 {
		dirDst = true
	}
	return expanded, dirDst, ok
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
	} else {
		if !force {
			if _, err = os.Stat(output); !os.IsNotExist(err) {
				fmt.Fprintln(os.Stderr, "ERROR: overwriting "+output+" (use -f)")
				return nil, false
			}
		}
		if w, err = os.OpenFile(output, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666); err != nil {
			fmt.Fprintln(os.Stderr, "ERROR: "+err.Error())
			return nil, false
		}
	}
	return w, true
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
		if _, ok := FiletypeMime[ext]; !ok {
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
