package trim // import "github.com/tdewolff/minify/trim"

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/parse"
)

// Don't implement Bytes() to test for buffer exceeding.
type ReaderMockup struct {
	r io.Reader
}

func (r *ReaderMockup) Read(p []byte) (int, error) {
	n, err := r.r.Read(p)
	if n < len(p) {
		err = io.EOF
	}
	return n, err
}

////////////////////////////////////////////////////////////////

func helperTestDefault(t *testing.T, m minify.Minifier, input, expected string) {
	b := &bytes.Buffer{}
	assert.Nil(t, Minify(m, b, &ReaderMockup{bytes.NewBufferString(input)}), "Minify must not return error in "+input)
	assert.Equal(t, expected, b.String(), "Minify must give expected result in "+input)
}

func helperTestDefaultError(t *testing.T, m minify.Minifier, input string, expErr error) {
	assert.Equal(t, expErr, Minify(m, &bytes.Buffer{}, &ReaderMockup{bytes.NewBufferString(input)}), "Minify must give expected error in "+input)
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
