package minify

import (
	"bytes"
	"testing"
)

func helperCSS(t *testing.T, input, expected string) {
	m := &Minifier{}
	b := &bytes.Buffer{}
	if err := m.CSS(b, bytes.NewBufferString(input)); err != nil {
		t.Error(err)
	}

	if b.String() != expected {
		t.Error(b.String(), "!=", expected)
	}
}

func TestCSS(t *testing.T) {
	helperCSS(t, "key: value;", "key:value")
	helperCSS(t, "i { key: value; }", "i{key:value}")
	helperCSS(t, "color: #FF0000;", "color:red")
	helperCSS(t, "color: #000000;", "color:#000")
	helperCSS(t, "color: black;", "color:#000")
	helperCSS(t, "color: rgb(255,255,255);", "color:#FFF")
	helperCSS(t, "color: rgb(100%,100%,100%);", "color:#FFF")
	helperCSS(t, "color: rgba(255,255,255,1);", "color:#FFF")
	helperCSS(t, "font-weight: bold; font-weight: normal;", "font-weight:700;font-weight:400")
	helperCSS(t, "outline: none;", "outline:0")
}