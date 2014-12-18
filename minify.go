/*
Package minify is a minifier written in Go that has a built-in HTML and CSS minifier.

Usage example:

	package main

	import (
		"fmt"
		"os"
		"os/exec"

		"github.com/tdewolff/minify"
	)

	// Minifies HTML code from stdin to stdout (note that using buffer is faster).
	func main() {
		m := minify.NewMinifierDefault()
		m.AddCmd("text/javascript", exec.Command("java", "-jar", "path/to/compiler.jar"))

		if err := m.Minify("text/html", os.Stdout, os.Stdin); err != nil {
			fmt.Println("minify.Minify:", err)
		}
	}

*/
package minify

import (
	"bytes"
	"errors"
	"io"
	"os/exec"
)

// ErrNotExist is returned when no minifier exists for a given mime type
var ErrNotExist = errors.New("minifier does not exist for mime type")

// ErrWrite is returned when an error occurred when writing to the writer
var ErrWrite = errors.New("write error")

////////////////////////////////////////////////////////////////

// Func is the function interface for minifiers
// The Minifier parameter is used for embedded resources, such as JS within HTML.
type Func func(Minifier, io.Writer, io.Reader) error

// Minifier holds a map of mime => function to allow recursive minifier calls of the minifier functions.
type Minifier struct {
	Mime map[string]Func
}

// NewMinifier returns a new Minifier struct with initialized map.
func NewMinifier() *Minifier {
	return &Minifier{map[string]Func{}}
}

// NewMinifierDefault returns a new Minifier struct with initialized map.
// It loads in the default minifier functions for HTML and CSS (test/html and text/css mime types respectively).
func NewMinifierDefault() *Minifier {
	return &Minifier{
		map[string]Func{
			"text/html": (Minifier).HTML,
			"text/css":  (Minifier).CSS,
		},
	}
}

// Add adds a minify function to the mime => function map.
// It allows one to implement a custom minifier for a specific mime type.
func (m *Minifier) Add(mime string, f Func) {
	m.Mime[mime] = f
}

// AddCmd adds a minify function that executes a command to process the minification.
// It allows the use of external tools like ClosureCompiler, UglifyCSS, etc.
// Be aware that running external tools will slow down minification a lot!
func (m *Minifier) AddCmd(mime string, cmd *exec.Cmd) error {
	m.Mime[mime] = func(m Minifier, w io.Writer, r io.Reader) error {
		stdOut, err := cmd.StdoutPipe()
		if err != nil {
			return err
		}
		defer stdOut.Close()

		stdIn, err := cmd.StdinPipe()
		if err != nil {
			return err
		}
		defer stdIn.Close()

		if err = cmd.Start(); err != nil {
			return err
		}
		if _, err := io.Copy(stdIn, r); err != nil {
			return err
		}
		stdIn.Close()
		if _, err = io.Copy(w, stdOut); err != nil {
			return err
		}
		return cmd.Wait()
	}
	return nil
}

// Minify minifies the content of a Reader and writes it to a Writer.
// An error is returned when no such mime type exists (ErrNotExist) or any error occurred in the minifier function.
func (m Minifier) Minify(mime string, w io.Writer, r io.Reader) error {
	if f, ok := m.Mime[mime]; ok {
		if err := f(m, w, r); err != nil {
			return err
		}
		return nil
	}
	return ErrNotExist
}

// MinifyBytes minifies an array of bytes. When an error occurs it return the original array and the error.
// It return an error when no such mime type exists (ErrNotExist) or any error occurred in the minifier function.
func (m Minifier) MinifyBytes(mime string, v []byte) ([]byte, error) {
	b := &bytes.Buffer{}
	if err := m.Minify(mime, b, bytes.NewBuffer(v)); err != nil {
		return v, err
	}
	return b.Bytes(), nil
}

// MinifyString minifies a string. When an error occurs it return the original string and the error.
// It return an error when no such mime type exists (ErrNotExist) or any error occurred in the minifier function.
func (m Minifier) MinifyString(mime string, v string) (string, error) {
	b := &bytes.Buffer{}
	if err := m.Minify(mime, b, bytes.NewBufferString(v)); err != nil {
		return v, err
	}
	return b.String(), nil
}