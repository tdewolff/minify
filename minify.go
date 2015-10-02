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

// Func is the function interface for minifiers.
// The Minifier parameter is used for embedded resources, such as JS within HTML.
// The mimetype string is for wildcard minifiers so they know what they minify and for parameter passing (charset for example).
type Func func(m *Minifier, w io.Writer, r io.Reader, mimetype string, params map[string]string) error

////////////////////////////////////////////////////////////////

type regexpFunc struct {
	re *regexp.Regexp
	Func
}

func cmdFunc(origCmd *exec.Cmd) func(_ *Minifier, w io.Writer, r io.Reader, _ string, _ map[string]string) error {
	return func(_ *Minifier, w io.Writer, r io.Reader, _ string, _ map[string]string) error {
		cmd := &exec.Cmd{}
		*cmd = *origCmd // concurrency safety
		cmd.Stdout = w
		cmd.Stdin = r
		return cmd.Run()
	}
}

////////////////////////////////////////////////////////////////

// Minifier holds a map of mimetype => function to allow recursive minifier calls of the minifier functions.
type Minifier struct {
	options map[string]string
	literal map[string]Func
	regexp  []regexpFunc
}

// New returns a new Minifier.
func New() *Minifier {
	return &Minifier{
		map[string]string{},
		map[string]Func{},
		[]regexpFunc{},
	}
}

// Set sets a key-value pair as option passed to the minifiers.
func (m *Minifier) Set(key, val string) {
	m.options[key] = val
}

// Get retrieves the value of a key in options.
func (m *Minifier) Get(key string) string {
	val, _ := m.options[key]
	return val
}

// AddFunc adds a minify function to the mimetype => function map (unsafe for concurrent use).
// It allows one to implement a custom minifier for a specific mimetype.
func (m *Minifier) AddFunc(mimetype string, minifyFunc Func) {
	m.literal[mimetype] = minifyFunc
}

// AddFuncRegexp adds a minify function to the mimetype => function map (unsafe for concurrent use).
// It allows one to implement a custom minifier for a specific mimetype regular expression.
func (m *Minifier) AddFuncRegexp(mimetype *regexp.Regexp, minifyFunc Func) {
	m.regexp = append(m.regexp, regexpFunc{mimetype, minifyFunc})
}

// AddCmd adds a minify function to the mimetype => function map (unsafe for concurrent use) that executes a command to process the minification.
// It allows the use of external tools like ClosureCompiler, UglifyCSS, etc. for a specific mimetype.
// Be aware that running external tools will slow down minification a lot!
func (m *Minifier) AddCmd(mimetype string, cmd *exec.Cmd) {
	m.literal[mimetype] = cmdFunc(cmd)
}

// AddCmd adds a minify function to the mimetype => function map (unsafe for concurrent use) that executes a command to process the minification.
// It allows the use of external tools like ClosureCompiler, UglifyCSS, etc. for a specific mimetype regular expression.
// Be aware that running external tools will slow down minification a lot!
func (m *Minifier) AddCmdRegexp(mimetype *regexp.Regexp, cmd *exec.Cmd) {
	m.regexp = append(m.regexp, regexpFunc{mimetype, cmdFunc(cmd)})
}

// Minify minifies the content of a Reader and writes it to a Writer (safe for concurrent use).
// An error is returned when no such mimetype exists (ErrNotExist) or when an error occurred in the minifier function.
// Mimetype may take the form of 'text/plain', 'text/*' or '*/*'.
func (m *Minifier) Minify(w io.Writer, r io.Reader, mimetype string, params map[string]string) error {
	if minifyFunc, ok := m.literal[mimetype]; ok {
		if err := minifyFunc(m, w, r, mimetype, params); err != nil {
			return err
		}
		return nil
	}
	for _, minifyRegexp := range m.regexp {
		if minifyRegexp.re.MatchString(mimetype) {
			if err := minifyRegexp.Func(m, w, r, mimetype, params); err != nil {
				return err
			}
			return nil
		}
	}
	return ErrNotExist
}

// Bytes minifies an array of bytes (safe for concurrent use). When an error occurs it return the original array and the error.
// It returns an error when no such mimetype exists (ErrNotExist) or any error occurred in the minifier function.
func (m *Minifier) Bytes(v []byte, mimetype string, params map[string]string) ([]byte, error) {
	out := buffer.NewWriter(make([]byte, 0, len(v)))
	if err := m.Minify(out, buffer.NewReader(v), mimetype, params); err != nil {
		return v, err
	}
	return out.Bytes(), nil
}

// String minifies a string (safe for concurrent use). When an error occurs it return the original string and the error.
// It returns an error when no such mimetype exists (ErrNotExist) or any error occurred in the minifier function.
func (m *Minifier) String(v string, mimetype string, params map[string]string) (string, error) {
	out := buffer.NewWriter(make([]byte, 0, len(v)))
	if err := m.Minify(out, buffer.NewReader([]byte(v)), mimetype, params); err != nil {
		return v, err
	}
	return string(out.Bytes()), nil
}
