package xml

import (
	"fmt"
	"io"
	"testing"

	"github.com/tdewolff/minify/v2/parse"
	"github.com/tdewolff/test"
)

type TTs []TokenType

func TestTokens(t *testing.T) {
	var tokenTests = []struct {
		xml      string
		expected []TokenType
	}{
		{"", TTs{}},
		{"<!-- comment -->", TTs{CommentToken}},
		{"<!-- comment \n multi \r line -->", TTs{CommentToken}},
		{"<foo/>", TTs{StartTagToken, StartTagCloseVoidToken}},
		{"<foo \t\r\n/>", TTs{StartTagToken, StartTagCloseVoidToken}},
		{"<foo:bar.qux-norf/>", TTs{StartTagToken, StartTagCloseVoidToken}},
		{"<foo></foo>", TTs{StartTagToken, StartTagCloseToken, EndTagToken}},
		{"<foo>text</foo>", TTs{StartTagToken, StartTagCloseToken, TextToken, EndTagToken}},
		{"<foo/> text", TTs{StartTagToken, StartTagCloseVoidToken, TextToken}},
		{"<a> <b> <c>text</c> </b> </a>", TTs{StartTagToken, StartTagCloseToken, TextToken, StartTagToken, StartTagCloseToken, TextToken, StartTagToken, StartTagCloseToken, TextToken, EndTagToken, TextToken, EndTagToken, TextToken, EndTagToken}},
		{"<foo a='a' b=\"b\" c=c/>", TTs{StartTagToken, AttributeToken, AttributeToken, AttributeToken, StartTagCloseVoidToken}},
		{"<foo a=\"\"/>", TTs{StartTagToken, AttributeToken, StartTagCloseVoidToken}},
		{"<foo a-b=\"\"/>", TTs{StartTagToken, AttributeToken, StartTagCloseVoidToken}},
		{"<foo \nchecked \r\n value\r=\t'=/>\"' />", TTs{StartTagToken, AttributeToken, AttributeToken, StartTagCloseVoidToken}},
		{"<?xml?>", TTs{StartTagPIToken, StartTagClosePIToken}},
		{"<?xml a=\"a\" ?>", TTs{StartTagPIToken, AttributeToken, StartTagClosePIToken}},
		{"<?xml a=a?>", TTs{StartTagPIToken, AttributeToken, StartTagClosePIToken}},
		{"<![CDATA[ test ]]>", TTs{CDATAToken}},
		{"<!DOCTYPE>", TTs{DOCTYPEToken}},
		{"<!DOCTYPE note SYSTEM \"Note.dtd\">", TTs{DOCTYPEToken}},
		{`<!DOCTYPE note [<!ENTITY nbsp "&#xA0;"><!ENTITY writer "Writer: Donald Duck."><!ENTITY copyright "Copyright:]> W3Schools.">]>`, TTs{DOCTYPEToken}},
		{"<!foo>", TTs{StartTagToken, StartTagCloseToken}},

		// early endings
		{"<!-- comment", TTs{CommentToken}},
		{"<foo", TTs{StartTagToken}},
		{"</foo", TTs{EndTagToken}},
		{"<foo x", TTs{StartTagToken, AttributeToken}},
		{"<foo x=", TTs{StartTagToken, AttributeToken}},
		{"<foo x='", TTs{StartTagToken, AttributeToken}},
		{"<foo x=''", TTs{StartTagToken, AttributeToken}},
		{"<?xml", TTs{StartTagPIToken}},
		{"<![CDATA[ test", TTs{CDATAToken}},
		{"<!DOCTYPE note SYSTEM", TTs{DOCTYPEToken}},

		// go fuzz
		{"</", TTs{EndTagToken}},
		{"</\n", TTs{EndTagToken}},
	}
	for _, tt := range tokenTests {
		t.Run(tt.xml, func(t *testing.T) {
			l := NewLexer(parse.NewInputString(tt.xml))
			i := 0
			for {
				token, _ := l.Next()
				if token == ErrorToken {
					test.T(t, l.Err(), io.EOF)
					test.T(t, i, len(tt.expected), "when error occurred we must be at the end")
					break
				}
				test.That(t, i < len(tt.expected), "index", i, "must not exceed expected token types size", len(tt.expected))
				if i < len(tt.expected) {
					test.T(t, token, tt.expected[i], "token types must match")
				}
				i++
			}
		})
	}

	// coverage
	for i := 0; ; i++ {
		if TokenType(i).String() == fmt.Sprintf("Invalid(%d)", i) {
			break
		}
	}
}

func TestTags(t *testing.T) {
	var tagTests = []struct {
		xml      string
		expected string
	}{
		{"<foo:bar.qux-norf/>", "foo:bar.qux-norf"},
		{"<?xml?>", "xml"},
		{"<foo?bar/qux>", "foo?bar/qux"},
		{"<!DOCTYPE note SYSTEM \"Note.dtd\">", " note SYSTEM \"Note.dtd\""},

		// early endings
		{"<foo ", "foo"},
	}
	for _, tt := range tagTests {
		t.Run(tt.xml, func(t *testing.T) {
			l := NewLexer(parse.NewInputString(tt.xml))
			for {
				token, _ := l.Next()
				if token == ErrorToken {
					test.T(t, l.Err(), io.EOF)
					test.Fail(t, "when error occurred we must be at the end")
					break
				} else if token == StartTagToken || token == StartTagPIToken || token == EndTagToken || token == DOCTYPEToken {
					test.String(t, string(l.Text()), tt.expected, "tags must match")
					break
				}
			}
		})
	}
}

func TestAttributes(t *testing.T) {
	var attributeTests = []struct {
		attr     string
		expected []string
	}{
		{"<foo a=\"b\" />", []string{"a", "\"b\""}},
		{"<foo \nchecked \r\n value\r=\t'=/>\"' />", []string{"checked", "", "value", "'=/>\"'"}},
		{"<foo bar=\" a \n\t\r b \" />", []string{"bar", "\" a     b \""}},
		{"<?xml a=b?>", []string{"a", "b"}},
		{"<foo /=? >", []string{"/", "?"}},

		// early endings
		{"<foo x", []string{"x", ""}},
		{"<foo x=", []string{"x", ""}},
		{"<foo x='", []string{"x", "'"}},
	}
	for _, tt := range attributeTests {
		t.Run(tt.attr, func(t *testing.T) {
			l := NewLexer(parse.NewInputString(tt.attr))
			i := 0
			for {
				token, _ := l.Next()
				if token == ErrorToken {
					test.T(t, l.Err(), io.EOF)
					test.T(t, i, len(tt.expected), "when error occurred we must be at the end")
					break
				} else if token == AttributeToken {
					test.That(t, i+1 < len(tt.expected), "index", i+1, "must not exceed expected attributes size", len(tt.expected))
					if i+1 < len(tt.expected) {
						test.String(t, string(l.Text()), tt.expected[i], "attribute keys must match")
						test.String(t, string(l.AttrVal()), tt.expected[i+1], "attribute keys must match")
						i += 2
					}
				}
			}
		})
	}
}

func TestErrors(t *testing.T) {
	var errorTests = []struct {
		xml string
		col int
	}{
		{"a\x00b", 2},
		{"<\x00 b='5'>", 2},
		{"<a\x00b='5'>", 3},
		{"<a \x00='5'>", 4},
		{"<a b\x00'5'>", 5},
		{"<a b=\x005'>", 6},
		{"<a b='\x00'>", 7},
		{"<a b='5\x00>", 8},
		{"<a b='5'\x00", 9},
		{"</\x00a>", 3},
		{"</ \x00>", 4},
		{"</ a\x00", 5},
		{"<!\x00", 3},
		{"<![CDATA[\x00", 10},
		{"/*\x00", 3},
	}
	for _, tt := range errorTests {
		t.Run(tt.xml, func(t *testing.T) {
			l := NewLexer(parse.NewInputString(tt.xml))
			for {
				token, _ := l.Next()
				if token == ErrorToken {
					if perr, ok := l.Err().(*parse.Error); ok {
						_, col, _ := perr.Position()
						test.T(t, col, tt.col)
					} else {
						test.Fail(t, "bad error:", l.Err())
					}
					break
				}
			}
		})
	}
}

func TestTextAndAttrVal(t *testing.T) {
	l := NewLexer(parse.NewInputString(`<xml attr="val" >text<!--comment--><!DOCTYPE doctype><![CDATA[cdata]]>`))
	_, data := l.Next()
	test.Bytes(t, data, []byte("<xml"))
	test.Bytes(t, l.Text(), []byte("xml"))
	test.Bytes(t, l.AttrVal(), nil)

	_, data = l.Next()
	test.Bytes(t, data, []byte(` attr="val"`))
	test.Bytes(t, l.Text(), []byte("attr"))
	test.Bytes(t, l.AttrVal(), []byte(`"val"`))

	_, data = l.Next()
	test.Bytes(t, data, []byte(">"))
	test.Bytes(t, l.Text(), nil)
	test.Bytes(t, l.AttrVal(), nil)

	_, data = l.Next()
	test.Bytes(t, data, []byte("text"))
	test.Bytes(t, l.Text(), []byte("text"))
	test.Bytes(t, l.AttrVal(), nil)

	_, data = l.Next()
	test.Bytes(t, data, []byte("<!--comment-->"))
	test.Bytes(t, l.Text(), []byte("comment"))
	test.Bytes(t, l.AttrVal(), nil)

	_, data = l.Next()
	test.Bytes(t, data, []byte("<!DOCTYPE doctype>"))
	test.Bytes(t, l.Text(), []byte(" doctype"))
	test.Bytes(t, l.AttrVal(), nil)

	_, data = l.Next()
	test.Bytes(t, data, []byte("<![CDATA[cdata]]>"))
	test.Bytes(t, l.Text(), []byte("cdata"))
	test.Bytes(t, l.AttrVal(), nil)
}

func TestOffset(t *testing.T) {
	z := parse.NewInputString(`<div attr="val">text</div>`)
	l := NewLexer(z)
	test.T(t, z.Offset(), 0)
	_, _ = l.Next()
	test.T(t, z.Offset(), 4) // <div
	_, _ = l.Next()
	test.T(t, z.Offset(), 15) // attr="val"
	_, _ = l.Next()
	test.T(t, z.Offset(), 16) // >
	_, _ = l.Next()
	test.T(t, z.Offset(), 20) // text
	_, _ = l.Next()
	test.T(t, z.Offset(), 26) // </div>
}

////////////////////////////////////////////////////////////////

func ExampleNewLexer() {
	l := NewLexer(parse.NewInputString("<span class='user'>John Doe</span>"))
	out := ""
	for {
		tt, data := l.Next()
		if tt == ErrorToken {
			break
		}
		out += string(data)
	}
	fmt.Println(out)
	// Output: <span class='user'>John Doe</span>
}
