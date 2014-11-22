package minify

import (
	"bytes"
	"testing"
)

func minifyCSS(t *testing.T, s string) string {
	m := &Minify{}
	b := &bytes.Buffer{}
	if err := m.CSS(b, bytes.NewBufferString(s)); err != nil {
		t.Error(err)
	}
	return b.String()
}

func TestCSSBasic(t *testing.T) {
	r := "key:value"
	if w := minifyCSS(t, r); w != r {
		t.Error(w, "!=", r)
	}
}