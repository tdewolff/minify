package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"regexp"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/js"
	"github.com/tdewolff/minify/json"
	"github.com/tdewolff/minify/svg"
	"github.com/tdewolff/minify/xml"
)

func main() {
	input := ""
	output := ""
	ext := ""

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] [file]\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.StringVar(&output, "o", "", "Output file (stdout when empty)")
	flag.StringVar(&ext, "x", "", "File extension (css, html, js, json, svg or xml), optional for input files")
	flag.Parse()
	if len(flag.Args()) > 0 {
		input = flag.Arg(0)
	}

	mediatype := ""
	r := io.Reader(os.Stdin)
	w := io.Writer(os.Stdout)

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
			ext = path.Ext(input)[1:]
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
		switch ext {
		case "css":
			mediatype = "text/css"
		case "html":
			mediatype = "text/html"
		case "js":
			mediatype = "text/javascript"
		case "json":
			mediatype = "application/json"
		case "svg":
			mediatype = "image/svg+xml"
		case "xml":
			mediatype = "text/xml"
		}
	}

	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/html", html.Minify)
	m.AddFunc("text/javascript", js.Minify)
	m.AddFunc("image/svg+xml", svg.Minify)
	m.AddFuncRegexp(regexp.MustCompile("[/+]json$"), json.Minify)
	m.AddFuncRegexp(regexp.MustCompile("[/+]xml$"), xml.Minify)
	if err := m.Minify(mediatype, w, r); err != nil {
		if err == minify.ErrNotExist {
			io.Copy(w, r)
		} else {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	}
}
