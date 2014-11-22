package minify

import (
	"bytes"
	"testing"
)

func minifyHTML(t *testing.T, s string) string {
	m := &Minify{}
	b := &bytes.Buffer{}
	if err := m.HTML(b, bytes.NewBufferString(s)); err != nil {
		t.Error(err)
	}
	return b.String()
}

func TestHTMLBasic(t *testing.T) {
	r := "html"
	if w := minifyHTML(t, r); w != r {
		t.Error(w, "!=", r)
	}
}