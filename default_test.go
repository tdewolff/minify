package minify // import "github.com/tdewolff/minify"

import (
	"bytes"
	"testing"
)

func helperDefault(t *testing.T, m *Minifier, input, expected string) {
	b := &bytes.Buffer{}
	if err := m.Default(b, bytes.NewBufferString(input)); err != nil {
		t.Error(err)
	}

	if b.String() != expected {
		t.Error(b.String(), "!=", expected)
	}
}

////////////////////////////////////////////////////////////////

func TestDefault(t *testing.T) {
	m := NewMinifier()
	helperDefault(t, m, "  x  ", "x")
}
