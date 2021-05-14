package xml

import (
	"testing"

	"github.com/tdewolff/test"
)

func TestEscapeAttrVal(t *testing.T) {
	var attrValTests = []struct {
		attrVal  string
		expected string
	}{
		{`xyz`, `"xyz"`},
		{``, `""`},
		{`x'z`, `"x'z"`},
		{`x"z`, `'x"z'`},
		{`a'b=""`, `'a&#39;b=""'`},
		{`'x'"'z'`, `"x'&#34;'z"`},
		{`"x"'"z"`, `'x"&#39;"z'`},
		{`a'b=""`, `'a&#39;b=""'`},
	}
	var buf []byte
	for _, tt := range attrValTests {
		t.Run(tt.attrVal, func(t *testing.T) {
			b := []byte(tt.attrVal)
			if len(b) > 1 && (b[0] == '"' || b[0] == '\'') && b[0] == b[len(b)-1] {
				b = b[1 : len(b)-1]
			}
			val := EscapeAttrVal(&buf, b)
			test.String(t, string(val), tt.expected)
		})
	}
}

func TestEscapeCDATAVal(t *testing.T) {
	var CDATAValTests = []struct {
		CDATAVal string
		expected string
	}{
		{"<![CDATA[<b>]]>", "&lt;b>"},
		{"<![CDATA[abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz]]>", "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz"},
		{"<![CDATA[ <b> ]]>", " &lt;b> "},
		{"<![CDATA[<<<<<]]>", "<![CDATA[<<<<<]]>"},
		{"<![CDATA[&]]>", "&amp;"},
		{"<![CDATA[&&&&]]>", "<![CDATA[&&&&]]>"},
		{"<![CDATA[ a ]]>", " a "},
		{"<![CDATA[]]>", ""},
		{"<![CDATA[ a ]]&gt; b ]]>", " a ]]&amp;gt; b "},
	}
	var buf []byte
	for _, tt := range CDATAValTests {
		t.Run(tt.CDATAVal, func(t *testing.T) {
			b := []byte(tt.CDATAVal[len("<![CDATA[") : len(tt.CDATAVal)-len("]]>")])
			data, useText := EscapeCDATAVal(&buf, b)
			text := string(data)
			if !useText {
				text = "<![CDATA[" + text + "]]>"
			}
			test.String(t, text, tt.expected)
		})
	}
}
