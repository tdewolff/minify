/*
Package minify is a minifier written in Go that has a built-in HTML and CSS minifier.

Usage example:

	package main

	import (
		"fmt"
		"os"
		"os/exec"

		"github.com/tdewolff/minify"
		"github.com/tdewolff/minify/html"
		"github.com/tdewolff/minify/css"
		"github.com/tdewolff/minify/trim"
	)

	// Minifies HTML code from stdin to stdout
	// Note that reading the file into a buffer first and writing to a buffer would be faster.
	func main() {
		m := minify.NewMinifier()
		m.Add("text/html", html.Minify)
		m.Add("text/css", css.Minify)
		m.Add("*\/*", trim.Minify) // remove backslash
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

// Minifier is the interface which all minifier functions accept as first parameter.
// It's used to extract parameter values of the mediatype and to recursively call other minifier functions.
type Minifier interface {
	Param(string) string
	Minify(string, io.Writer, io.Reader) error
}

////////////////////////////////////////////////////////////////

// DefaultMinifier holds a map of mediatype => function to allow recursive minifier calls of the minifier functions.
type DefaultMinifier struct {
	minify map[string]Func
	params map[string]string
}

// NewMinifier returns a new Minifier.
func NewMinifier() *DefaultMinifier {
	return &DefaultMinifier{
		map[string]Func{},
		map[string]string{},
	}
}

// Add adds a minify function to the mediatype => function map.
// It allows one to implement a custom minifier for a specific mediatype.
func (m *DefaultMinifier) Add(mediatype string, f Func) {
	m.minify[mediatype] = f
}

// AddCmd adds a minify function that executes a command to process the minification.
// It allows the use of external tools like ClosureCompiler, UglifyCSS, etc.
// Be aware that running external tools will slow down minification a lot!
func (m *DefaultMinifier) AddCmd(mediatype string, cmd *exec.Cmd) error {
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

// Param returns the parameter value of the given key.
// If the key exists it returns a non-zero length string. If the value is an empty string, the key is returned.
func (m DefaultMinifier) Param(key string) string {
	return m.params[key]
}

// Minify minifies the content of a Reader and writes it to a Writer.
// An error is returned when no such mediatype exists (ErrNotExist) or any error occurred in the minifier function.
// Mediatype may take the form of 'text', 'text/css', 'text/css;utf8'
func (m DefaultMinifier) Minify(mediatype string, w io.Writer, r io.Reader) error {
	parentParams := m.params
	defer func() {
		m.params = parentParams
	}()

	m.params = map[string]string{"mediatype": mediatype}
	params := strings.Split(mediatype, ";")
	for _, p := range params[1:] {
		if i := strings.IndexByte(p, '='); i != -1 {
			m.params[strings.TrimSpace(p[:i])] = strings.TrimSpace(p[i+1:])
		} else {
			p = strings.TrimSpace(p)
			m.params[p] = p
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
func (m DefaultMinifier) MinifyBytes(mediatype string, v []byte) ([]byte, error) {
	b := &bytes.Buffer{}
	b.Grow(len(v))
	if err := m.Minify(mediatype, b, bytes.NewReader(v)); err != nil {
		return v, err
	}
	return b.Bytes(), nil
}

// MinifyString minifies a string. When an error occurs it return the original string and the error.
// It return an error when no such mediatype exists (ErrNotExist) or any error occurred in the minifier function.
func (m DefaultMinifier) MinifyString(mediatype string, v string) (string, error) {
	b := &bytes.Buffer{}
	b.Grow(len(v))
	if err := m.Minify(mediatype, b, bytes.NewReader([]byte(v))); err != nil {
		return v, err
	}
	return b.String(), nil
}
