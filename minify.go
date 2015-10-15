// Package minify relates MIME type to minifiers. Several minifiers are provided in the subpackages.
package minify // import "github.com/tdewolff/minify"

import (
	"errors"
	"io"
	"os/exec"
	"regexp"

	"github.com/tdewolff/buffer"
)

// ErrNotExist is returned when no minifier exists for a given mediatype.
var ErrNotExist = errors.New("minify function does not exist for mediatype")

////////////////////////////////////////////////////////////////

// Func is the function interface for minifiers.
// The Minifier parameter is used for embedded resources, such as JS within HTML.
// The mediatype string is for wildcard minifiers so they know what they minify and for parameter passing (charset for example).
type Func func(m Minifier, mediatype string, w io.Writer, r io.Reader) error

// Minifier is the interface which all minifier functions accept as first parameter.
// It's used to extract parameter values of the mediatype and to recursively call other minifier functions.
type Minifier interface {
	Minify(mediatype string, w io.Writer, r io.Reader) error
}

////////////////////////////////////////////////////////////////

type regexpFunc struct {
	re *regexp.Regexp
	Func
}

func cmdFunc(origCmd *exec.Cmd) func(_ Minifier, _ string, w io.Writer, r io.Reader) error {
	return func(_ Minifier, _ string, w io.Writer, r io.Reader) error {
		cmd := &exec.Cmd{}
		*cmd = *origCmd // concurrency safety
		cmd.Stdout = w
		cmd.Stdin = r
		return cmd.Run()
	}
}

////////////////////////////////////////////////////////////////

// Minify holds a map of mediatype => function to allow recursive minifier calls of the minifier functions.
type Minify struct {
	literal map[string]Func
	regexp  []regexpFunc
}

// New returns a new Minify.
func New() *Minify {
	return &Minify{
		map[string]Func{},
		[]regexpFunc{},
	}
}

// AddFunc adds a minify function to the mediatype => function map (unsafe for concurrent use).
// It allows one to implement a custom minifier for a specific mediatype.
func (m *Minify) AddFunc(mediatype string, minifyFunc Func) {
	m.literal[mediatype] = minifyFunc
}

// AddFuncRegexp adds a minify function to the mediatype => function map (unsafe for concurrent use).
// It allows one to implement a custom minifier for a specific mediatype regular expression.
func (m *Minify) AddFuncRegexp(mediatype *regexp.Regexp, minifyFunc Func) {
	m.regexp = append(m.regexp, regexpFunc{mediatype, minifyFunc})
}

// AddCmd adds a minify function to the mediatype => function map (unsafe for concurrent use) that executes a command to process the minification.
// It allows the use of external tools like ClosureCompiler, UglifyCSS, etc. for a specific mediatype.
// Be aware that running external tools will slow down minification a lot!
func (m *Minify) AddCmd(mediatype string, cmd *exec.Cmd) {
	m.literal[mediatype] = cmdFunc(cmd)
}

// AddCmdRegexp adds a minify function to the mediatype => function map (unsafe for concurrent use) that executes a command to process the minification.
// It allows the use of external tools like ClosureCompiler, UglifyCSS, etc. for a specific mediatype regular expression.
// Be aware that running external tools will slow down minification a lot!
func (m *Minify) AddCmdRegexp(mediatype *regexp.Regexp, cmd *exec.Cmd) {
	m.regexp = append(m.regexp, regexpFunc{mediatype, cmdFunc(cmd)})
}

// Minify minifies the content of a Reader and writes it to a Writer (safe for concurrent use).
// An error is returned when no such mediatype exists (ErrNotExist) or when an error occurred in the minifier function.
// Mediatype may take the form of 'text/plain', 'text/*', '*/*' or 'text/plain; charset=UTF-8; version=2.0'.
func (m Minify) Minify(mediatype string, w io.Writer, r io.Reader) error {
	mimetype := mediatype
	for i, c := range mediatype {
		if c == ';' {
			mimetype = mediatype[:i]
			break
		}
	}
	if minifyFunc, ok := m.literal[mimetype]; ok {
		if err := minifyFunc(m, mediatype, w, r); err != nil {
			return err
		}
		return nil
	}
	for _, minifyRegexp := range m.regexp {
		if minifyRegexp.re.MatchString(mimetype) {
			if err := minifyRegexp.Func(m, mediatype, w, r); err != nil {
				return err
			}
			return nil
		}
	}
	return ErrNotExist
}

// Bytes minifies an array of bytes (safe for concurrent use). When an error occurs it return the original array and the error.
// It returns an error when no such mediatype exists (ErrNotExist) or any error occurred in the minifier function.
func Bytes(m Minifier, mediatype string, v []byte) ([]byte, error) {
	out := buffer.NewWriter(make([]byte, 0, len(v)))
	if err := m.Minify(mediatype, out, buffer.NewReader(v)); err != nil {
		return v, err
	}
	return out.Bytes(), nil
}

// String minifies a string (safe for concurrent use). When an error occurs it return the original string and the error.
// It returns an error when no such mediatype exists (ErrNotExist) or any error occurred in the minifier function.
func String(m Minifier, mediatype string, v string) (string, error) {
	out := buffer.NewWriter(make([]byte, 0, len(v)))
	if err := m.Minify(mediatype, out, buffer.NewReader([]byte(v))); err != nil {
		return v, err
	}
	return string(out.Bytes()), nil
}
