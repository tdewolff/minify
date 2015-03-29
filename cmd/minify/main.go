package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"mime"
	"os"
	"path"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/js"
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
	flag.StringVar(&ext, "x", "", "File extension (html, css or js), optional for input files")
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
		mediatype = mime.TypeByExtension(path.Ext(input))
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
		mediatype = mime.TypeByExtension("." + ext)
	}

	m := minify.New()
	m.AddFunc("text/html", html.Minify)
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/javascript", js.Minify)
	if err := m.Minify(mediatype, w, r); err != nil {
		if err == minify.ErrNotExist {
			io.Copy(w, r)
		} else {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	}
}
