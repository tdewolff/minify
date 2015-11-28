// Package minify relates MIME type to minifiers. Several minifiers are provided in the subpackages.
package minify // import "github.com/tdewolff/minify"

import (
	"errors"
	"io"
	"net/url"
	"os/exec"
	"regexp"
	"sync"

	"github.com/tdewolff/buffer"
	"github.com/tdewolff/parse"
)

// ErrNotExist is returned when no minifier exists for a given mimetype.
var ErrNotExist = errors.New("minifier does not exist for mimetype")

////////////////////////////////////////////////////////////////

type minifierFunc func(*M, io.Writer, io.Reader, map[string]string) error

func (f minifierFunc) Minify(m *M, w io.Writer, r io.Reader, params map[string]string) error {
	return f(m, w, r, params)
}

// Minifier is the interface for minifiers.
// The *M parameter is used for minifying embedded resources, such as JS within HTML.
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

	URL *url.URL
}

// New returns a new M.
func New() *M {
	return &M{
		map[string]Minifier{},
		[]patternMinifier{},
		nil,
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

// AddRegexp adds a minifier to the mimetype => function map (unsafe for concurrent use).
func (m *M) AddRegexp(pattern *regexp.Regexp, minifier Minifier) {
	m.pattern = append(m.pattern, patternMinifier{pattern, minifier})
}

// AddFuncRegexp adds a minify function to the mimetype => function map (unsafe for concurrent use).
func (m *M) AddFuncRegexp(pattern *regexp.Regexp, minifier minifierFunc) {
	m.pattern = append(m.pattern, patternMinifier{pattern, minifier})
}

// AddCmd adds a minify function to the mimetype => function map (unsafe for concurrent use) that executes a command to process the minification.
// It allows the use of external tools like ClosureCompiler, UglifyCSS, etc. for a specific mimetype.
func (m *M) AddCmd(mimetype string, cmd *exec.Cmd) {
	m.literal[mimetype] = &cmdMinifier{cmd}
}

// AddCmdRegexp adds a minify function to the mimetype => function map (unsafe for concurrent use) that executes a command to process the minification.
// It allows the use of external tools like ClosureCompiler, UglifyCSS, etc. for a specific mimetype regular expression.
func (m *M) AddCmdRegexp(pattern *regexp.Regexp, cmd *exec.Cmd) {
	m.pattern = append(m.pattern, patternMinifier{pattern, &cmdMinifier{cmd}})
}

// Minify minifies the content of a Reader and writes it to a Writer (safe for concurrent use).
// An error is returned when no such mimetype exists (ErrNotExist) or when an error occurred in the minifier function.
// Mediatype may take the form of 'text/plain', 'text/*', '*/*' or 'text/plain; charset=UTF-8; version=2.0'.
func (m *M) Minify(mediatype string, w io.Writer, r io.Reader) error {
	mimetype, params := parse.Mediatype([]byte(mediatype))
	return m.MinifyMimetype(mimetype, w, r, params)
}

// MinifyMimetype minifies the content of a Reader and writes it to a Writer (safe for concurrent use).
// It is a lower level version of Minify and requires the mediatype to be split up into mimetype and parameters.
// It is mostly used internally by minifiers because it is faster (no need to convert a byte-slice to string and vice versa).
func (m *M) MinifyMimetype(mimetype []byte, w io.Writer, r io.Reader, params map[string]string) error {
	err := ErrNotExist
	if minifier, ok := m.literal[string(mimetype)]; ok { // string conversion is optimized away
		err = minifier.Minify(m, w, r, params)
	} else {
		for _, minifier := range m.pattern {
			if minifier.pattern.Match(mimetype) {
				err = minifier.Minify(m, w, r, params)
				break
			}
		}
	}
	return err
}

// Bytes minifies an array of bytes (safe for concurrent use). When an error occurs it return the original array and the error.
// It returns an error when no such mimetype exists (ErrNotExist) or any error occurred in the minifier function.
func (m *M) Bytes(mediatype string, v []byte) ([]byte, error) {
	out := buffer.NewWriter(make([]byte, 0, len(v)))
	if err := m.Minify(mediatype, out, buffer.NewReader(v)); err != nil {
		return v, err
	}
	return out.Bytes(), nil
}

// String minifies a string (safe for concurrent use). When an error occurs it return the original string and the error.
// It returns an error when no such mimetype exists (ErrNotExist) or any error occurred in the minifier function.
func (m *M) String(mediatype string, v string) (string, error) {
	out := buffer.NewWriter(make([]byte, 0, len(v)))
	if err := m.Minify(mediatype, out, buffer.NewReader([]byte(v))); err != nil {
		return v, err
	}
	return string(out.Bytes()), nil
}

// Reader wraps a Reader interface and minifies the stream.
// Errors from the minifier are returned by the reader.
func (m *M) Reader(mediatype string, r io.Reader) io.Reader {
	pr, pw := io.Pipe()
	go func() {
		if err := m.Minify(mediatype, pw, r); err != nil {
			pw.CloseWithError(err)
		} else {
			pw.Close()
		}
	}()
	return pr
}

// minifyWriter makes sure that errors from the minifier are passed down through Close.
// It also makes sure that all data has been written on calling Close, it flushes.
type minifyWriter struct {
	pw  *io.PipeWriter
	wg  sync.WaitGroup
	err error
}

func (mw *minifyWriter) Write(b []byte) (int, error) {
	return mw.pw.Write(b)
}

func (mw *minifyWriter) Close() error {
	mw.pw.Close()
	mw.wg.Wait()
	return mw.err
}

// Writer wraps a Writer interface and minifies the stream.
// Errors from the minifier are returned by the writer.
// The writer must be closed explicitly.
func (m *M) Writer(mediatype string, w io.Writer) io.WriteCloser {
	pr, pw := io.Pipe()
	mw := &minifyWriter{pw, sync.WaitGroup{}, nil}
	mw.wg.Add(1)
	go func() {
		if err := m.Minify(mediatype, w, pr); err != nil {
			mw.err = err
			pr.CloseWithError(err)
		} else {
			pr.Close()
		}
		mw.wg.Done()
	}()
	return mw
}
