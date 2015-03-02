package trim // import "github.com/tdewolff/minify/trim"

import (
	"bytes"
	"io"
	"testing"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/parse"
)

// Don't implement Bytes() to test for buffer exceeding.
type readerMockup struct {
	r io.Reader
}

func (r *readerMockup) Read(p []byte) (int, error) {
	n, err := r.r.Read(p)
	if n < len(p) {
		err = io.EOF
	}
	return n, err
}

////////////////////////////////////////////////////////////////

func helperTestDefault(t *testing.T, m minify.Minifier, input, expected string) {
	b := &bytes.Buffer{}
	if err := Minify(m, b, &readerMockup{bytes.NewBufferString(input)}); err != nil {
		t.Error(err)
		return
	}

	if b.String() != expected {
		t.Error(b.String(), "!=", expected)
	}
}

func helperTestDefaultError(t *testing.T, m minify.Minifier, input string, expErr error) {
	b := &bytes.Buffer{}
	if err := Minify(m, b, &readerMockup{bytes.NewBufferString(input)}); err != expErr {
		t.Error(err, "!=", expErr, "for", input)
	}
}

////////////////////////////////////////////////////////////////

func TestDefault(t *testing.T) {
	m := minify.NewMinifier()
	helperTestDefault(t, m, "  x  ", "x")

	parse.MinBuf = 2
	parse.MaxBuf = 4
	helperTestDefault(t, m, "  y  ", "y")
	helperTestDefaultError(t, m, "  y   ", nil) // EOF
	helperTestDefaultError(t, m, "  y    ", parse.ErrBufferExceeded)
	helperTestDefaultError(t, m, "    y  ", parse.ErrBufferExceeded)
}
