// Package minify relates MIME type to minifiers. Several minifiers are provided in the subpackages.
package minify // import "github.com/tdewolff/minify"

import (
	"errors"
	"io"
	"os/exec"
	"regexp"

	"github.com/tdewolff/buffer"
)

// ErrNotExist is returned when no minifier exists for a given mimetype.
var ErrNotExist = errors.New("minify function does not exist for mimetype")

////////////////////////////////////////////////////////////////

type minifierFunc func(*M, io.Writer, io.Reader, map[string]string) error

func (f minifierFunc) Minify(m *M, w io.Writer, r io.Reader, params map[string]string) error {
	return f(m, w, r, params)
}

// Minifier is the interface for minifiers.
// The *M parameter is used for minifying embedded resources, such as JS within HTML.
// The map[string]string is for parameter passing (charset=UTF-8 for example).
type Minifier interface {
	Minify(*M, io.Writer, io.Reader, map[string]string) error
}

////////////////////////////////////////////////////////////////

type patternMinifier struct {
	pattern *regexp.Regexp
	Minifier
}

type cmdMinifier struct {
	cmd *exec.Cmd
}

func (c *cmdMinifier) Minify(_ *M, w io.Writer, r io.Reader, _ map[string]string) error {
	cmd := &exec.Cmd{}
	*cmd = *c.cmd // concurrency safety
	cmd.Stdout = w
	cmd.Stdin = r
	return cmd.Run()
}

////////////////////////////////////////////////////////////////

// M holds a map of mimetype => function to allow recursive minifier calls of the minifier functions.
type M struct {
	literal map[string]Minifier
	pattern []patternMinifier
}

// New returns a new M.
func New() *M {
	return &M{
		map[string]Minifier{},
		[]patternMinifier{},
	}
}

// Add adds a minifier to the mimetype => function map (unsafe for concurrent use).
func (m *M) Add(mimetype string, minifier Minifier) {
	m.literal[mimetype] = minifier
}

// AddFunc adds a minify function to the mimetype => function map (unsafe for concurrent use).
func (m *M) AddFunc(mimetype string, minifier minifierFunc) {
	m.literal[mimetype] = minifier
}

// AddPattern adds a minifier to the mimetype => function map (unsafe for concurrent use).
func (m *M) AddPattern(pattern *regexp.Regexp, minifier Minifier) {
	m.pattern = append(m.pattern, patternMinifier{pattern, minifier})
}

// AddFuncPattern adds a minify function to the mimetype => function map (unsafe for concurrent use).
func (m *M) AddFuncPattern(pattern *regexp.Regexp, minifier minifierFunc) {
	m.pattern = append(m.pattern, patternMinifier{pattern, minifier})
}

// AddCmd adds a minify function to the mimetype => function map (unsafe for concurrent use) that executes a command to process the minification.
// It allows the use of external tools like ClosureCompiler, UglifyCSS, etc. for a specific mimetype.
func (m *M) AddCmd(mimetype string, cmd *exec.Cmd) {
	m.literal[mimetype] = &cmdMinifier{cmd}
}

// AddCmdRegexp adds a minify function to the mimetype => function map (unsafe for concurrent use) that executes a command to process the minification.
// It allows the use of external tools like ClosureCompiler, UglifyCSS, etc. for a specific mimetype regular expression.
func (m *M) AddCmdPattern(pattern *regexp.Regexp, cmd *exec.Cmd) {
	m.pattern = append(m.pattern, patternMinifier{pattern, &cmdMinifier{cmd}})
}

// Minify minifies the content of a Reader and writes it to a Writer (safe for concurrent use).
// An error is returned when no such mimetype exists (ErrNotExist) or when an error occurred in the minifier function.
// Mimetype may take the form of 'text/plain', 'text/*' or '*/*'.
func (m *M) Minify(w io.Writer, r io.Reader, mimetype string, params map[string]string) error {
	if minifier, ok := m.literal[mimetype]; ok {
		if err := minifier.Minify(m, w, r, params); err != nil {
			return err
		}
		return nil
	}
	for _, minifier := range m.pattern {
		if minifier.pattern.MatchString(mimetype) {
			if err := minifier.Minify(m, w, r, params); err != nil {
				return err
			}
			return nil
		}
	}
	return ErrNotExist
}

// Bytes minifies an array of bytes (safe for concurrent use). When an error occurs it return the original array and the error.
// It returns an error when no such mimetype exists (ErrNotExist) or any error occurred in the minifier function.
func (m *M) Bytes(v []byte, mimetype string, params map[string]string) ([]byte, error) {
	out := buffer.NewWriter(make([]byte, 0, len(v)))
	if err := m.Minify(out, buffer.NewReader(v), mimetype, params); err != nil {
		return v, err
	}
	return out.Bytes(), nil
}

// String minifies a string (safe for concurrent use). When an error occurs it return the original string and the error.
// It returns an error when no such mimetype exists (ErrNotExist) or any error occurred in the minifier function.
func (m *M) String(v string, mimetype string, params map[string]string) (string, error) {
	out := buffer.NewWriter(make([]byte, 0, len(v)))
	if err := m.Minify(out, buffer.NewReader([]byte(v)), mimetype, params); err != nil {
		return v, err
	}
	return string(out.Bytes()), nil
}

// Reader wraps a Reader interface and minifies the stream.
// Errors from the minifier are returned by the reader.
func (m *M) Reader(r io.Reader, mimetype string, params map[string]string) io.Reader {
	pr, pw := io.Pipe()
	go func() {
		if err := m.Minify(pw, r, mimetype, params); err != nil {
			pw.CloseWithError(err)
		}
		pw.Close()
	}()
	return pr
}

// Writers wraps a Writer interface and minifies the stream.
// Errors from the minifier are returned by the writer.
// The writer must be closed explicitly.
func (m *M) Writer(w io.Writer, mimetype string, params map[string]string) io.WriteCloser {
	pr, pw := io.Pipe()
	go func() {
		if err := m.Minify(w, pr, mimetype, params); err != nil {
			pr.CloseWithError(err)
		}
		pr.Close()
	}()
	return pw
}
