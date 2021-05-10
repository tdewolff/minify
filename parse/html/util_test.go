package html

import (
	"testing"

	"github.com/tdewolff/test"
)

func TestEscapeAttrVal(t *testing.T) {
	var escapeAttrValTests = []struct {
		attrVal  string
		expected string
	}{
		{`xyz`, `xyz`},
		{``, ``},
		{`x/z`, `x/z`},
		{`x'z`, `"x'z"`},
		{`x"z`, `'x"z'`},
		{`'x"z'`, `'x"z'`},
		{`'x'"'z'`, `"x'&#34;'z"`},
		{`"x"'"z"`, `'x"&#39;"z'`},
		{`"x'z"`, `"x'z"`},
		{`'x'z'`, `"x'z"`},
		{`a'b=""`, `'a&#39;b=""'`},
		{`x<z`, `"x<z"`},
		{`'x"'"z'`, `'x"&#39;"z'`},
	}
	var buf []byte
	for _, tt := range escapeAttrValTests {
		t.Run(tt.attrVal, func(t *testing.T) {
			b := []byte(tt.attrVal)
			orig := b
			if len(b) > 1 && (b[0] == '"' || b[0] == '\'') && b[0] == b[len(b)-1] {
				b = b[1 : len(b)-1]
			}
			val := EscapeAttrVal(&buf, orig, []byte(b), false)
			test.String(t, string(val), tt.expected)
		})
	}
}

func TestEscapeAttrValXML(t *testing.T) {
	var escapeAttrValTests = []struct {
		attrVal  string
		expected string
	}{
		{`xyz`, `"xyz"`},
		{``, `""`},
	}
	var buf []byte
	for _, tt := range escapeAttrValTests {
		t.Run(tt.attrVal, func(t *testing.T) {
			b := []byte(tt.attrVal)
			orig := b
			if len(b) > 1 && (b[0] == '"' || b[0] == '\'') && b[0] == b[len(b)-1] {
				b = b[1 : len(b)-1]
			}
			val := EscapeAttrVal(&buf, orig, []byte(b), true)
			test.String(t, string(val), tt.expected)
		})
	}
}
