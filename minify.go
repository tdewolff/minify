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

	// Minifies HTML code from stdin to stdout
	// Note that reading the file into a buffer first and writing to a buffer would be faster.
	func main() {
		m := minify.NewMinifierDefault()
		m.AddCmd("text/javascript", exec.Command("java", "-jar", "build/compiler.jar"))

		if err := m.Minify("text/html", os.Stdout, os.Stdin); err != nil {
			fmt.Println("minify.Minify:", err)
		}
	}

*/
package minify // import "github.com/tdewolff/minify"

import (
	"bytes"
	"errors"
	"io"
	"os/exec"
	"strings"
)

// ErrNotExist is returned when no minifier exists for a given mediatype
var ErrNotExist = errors.New("minifier does not exist for mediatype")

////////////////////////////////////////////////////////////////

// Func is the function interface for minifiers.
// The Minifier parameter is used for embedded resources, such as JS within HTML.
type Func func(Minifier, io.Writer, io.Reader) error

// Minifier holds a map of mediatype => function to allow recursive minifier calls of the minifier functions.
type Minifier struct {
	minify map[string]Func
	Info   map[string]string
}

// NewMinifier returns a new Minifier struct with initialized map.
func NewMinifier() *Minifier {
	return &Minifier{
		map[string]Func{},
		map[string]string{},
	}
}

// NewMinifierDefault returns a new Minifier struct with initialized map.
// It loads in the default minifier functions for HTML and CSS (test/html and text/css mediatypes respectively).
func NewMinifierDefault() *Minifier {
	return &Minifier{
		map[string]Func{
			"text/html": (Minifier).HTML,
			"text/css":  (Minifier).CSS,
			"*/*":       (Minifier).Default,
		},
		map[string]string{},
	}
}

// Add adds a minify function to the mediatype => function map.
// It allows one to implement a custom minifier for a specific mediatype.
func (m *Minifier) Add(mediatype string, f Func) {
	m.minify[mediatype] = f
}

// AddCmd adds a minify function that executes a command to process the minification.
// It allows the use of external tools like ClosureCompiler, UglifyCSS, etc.
// Be aware that running external tools will slow down minification a lot!
func (m *Minifier) AddCmd(mediatype string, cmd *exec.Cmd) error {
	m.minify[mediatype] = func(m Minifier, w io.Writer, r io.Reader) error {
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
// An error is returned when no such mediatype exists (ErrNotExist) or any error occurred in the minifier function.
// Mediatype may take the form of 'text', 'text/css', 'text/css;utf8'
func (m Minifier) Minify(mediatype string, w io.Writer, r io.Reader) error {
	parentInfo := m.Info
	m.Info = make(map[string]string)
	defer func() {
		m.Info = parentInfo
	}()

	m.Info["mediatype"] = mediatype
	params := strings.Split(mediatype, ";")
	for _, p := range params[1:] {
		if i := strings.IndexByte(p, '='); i != -1 {
			m.Info[strings.TrimSpace(p[:i])] = strings.TrimSpace(p[i+1:])
		}
	}

	if f, ok := m.minify[params[0]]; ok {
		if err := f(m, w, r); err != nil {
			return err
		}
		return nil
	} else if i := strings.IndexByte(params[0], '/'); i != -1 {
		if f, ok := m.minify[params[0][:i]+"/*"]; ok {
			if err := f(m, w, r); err != nil {
				return err
			}
			return nil
		} else if f, ok := m.minify["*/*"]; ok {
			if err := f(m, w, r); err != nil {
				return err
			}
			return nil
		}
	}
	return ErrNotExist
}

// MinifyBytes minifies an array of bytes. When an error occurs it return the original array and the error.
// It return an error when no such mediatype exists (ErrNotExist) or any error occurred in the minifier function.
func (m Minifier) MinifyBytes(mediatype string, v []byte) ([]byte, error) {
	b := &bytes.Buffer{}
	if err := m.Minify(mediatype, b, bytes.NewBuffer(v)); err != nil {
		return v, err
	}
	return b.Bytes(), nil
}

// MinifyString minifies a string. When an error occurs it return the original string and the error.
// It return an error when no such mediatype exists (ErrNotExist) or any error occurred in the minifier function.
func (m Minifier) MinifyString(mediatype string, v string) (string, error) {
	b := &bytes.Buffer{}
	if err := m.Minify(mediatype, b, bytes.NewBufferString(v)); err != nil {
		return v, err
	}
	return b.String(), nil
}
