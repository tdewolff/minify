package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/js"
	"github.com/tdewolff/minify/json"
	"github.com/tdewolff/minify/svg"
	"github.com/tdewolff/minify/xml"
)

var extMime = map[string]string{
	".css":  "text/css",
	".html": "text/html",
	".js":   "application/javascript",
	".json": "application/json",
	".svg":  "image/svg+xml",
	".xml":  "text/xml",
}

func main() {
	input := ""
	output := ""
	ext := ""
	directory := ""
	recursive := false

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] [file]\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.StringVar(&output, "o", "", "Output file (stdout when empty)")
	flag.StringVar(&ext, "x", "", "File extension (css, html, js, json, svg or xml), optional for input files")
	flag.StringVar(&directory, "d", "", "Directory to search for files")
	flag.BoolVar(&recursive, "r", false, "Recursively minify everything")
	flag.Parse()
	if len(flag.Args()) > 0 {
		input = flag.Arg(0)
	}

	mediatype := ""
	r := io.Reader(os.Stdin)
	w := io.Writer(os.Stdout)
	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/html", html.Minify)
	m.AddFunc("application/javascript", js.Minify)
	m.AddFunc("image/svg+xml", svg.Minify)
	m.AddFuncRegexp(regexp.MustCompile("[/+]json$"), json.Minify)
	m.AddFuncRegexp(regexp.MustCompile("[/+]xml$"), xml.Minify)

	filenames := make(map[string]string)
	if directory != "" {
		filenames = dirFilenames(directory, recursive)
	} else {
		filenames[input] = output
	}

	for input, output := range filenames {
		if input != "" {
			in, err := os.Open(input)
			if err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}
			defer in.Close()
			r = in
			if input == output {
				b := &bytes.Buffer{}
				io.Copy(b, r)
				r = b
			}
			if ext == "" {
				ext = filepath.Ext(input)
			}
		}
		if output != "" {
			out, err := os.Create(output)
			if err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}
			defer out.Close()
			w = out
		}
		if ext != "" {
			mediatype = extMime[ext]
		}
		if err := m.Minify(mediatype, w, r); err != nil {
			if err == minify.ErrNotExist {
				io.Copy(w, r)
			} else {
				fmt.Println("Error:", err)
				os.Exit(1)
			}
		}
	}
}

// dirFilenames returns a map of input paths and output paths.
func dirFilenames(root string, recursive bool) map[string]string {
	names := map[string]string{}
	if recursive {
		filepath.Walk(root, func(path string, info os.FileInfo, _ error) error {
			if isValidFile(info) {
				names[path] = insertMinExt(path)
			}
			return nil
		})
	} else if infos, err := ioutil.ReadDir(root); err == nil {
		for _, info := range infos {
			if isValidFile(info) {
				path := filepath.Join(root, info.Name())
				names[path] = insertMinExt(path)
			}
		}
	}
	return names
}

// isValidFile checks to see if a file is a directory, hidden, already has the
// minified extension, or if it's one of the minifiable extensions.
func isValidFile(info os.FileInfo) bool {
	if info.IsDir() || len(info.Name()) > 0 && (info.Name()[0] == '.' || strings.Contains(info.Name(), ".min.")) {
		return false
	}
	_, exists := extMime[strings.ToLower(filepath.Ext(info.Name()))]
	return exists
}

// insertMinExt adds .min before a file's extension. If a file doesn't have an
// extension then .min will become the file's extension.
func insertMinExt(path string) string {
	if dot := strings.LastIndex(path, "."); dot != -1 {
		return path[:dot] + ".min" + path[dot:]
	}
	return path + ".min"
}
