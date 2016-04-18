package xml // import "github.com/tdewolff/minify/xml"

import (
	"bytes"
	"os"
	"regexp"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/test"
)

func TestXML(t *testing.T) {
	var xmlTests = []struct {
		xml      string
		expected string
	}{
		{"<!-- comment -->", ""},
		{"<A>x</A>", "<A>x</A>"},
		{"<a><b>x</b></a>", "<a><b>x</b></a>"},
		{"<a><b>x\ny</b></a>", "<a><b>x\ny</b></a>"},
		{"<a> <![CDATA[ a ]]> </a>", "<a>a</a>"},
		{"<a >a</a >", "<a>a</a>"},
		{"<?xml  version=\"1.0\" ?>", "<?xml version=\"1.0\"?>"},
		{"<x></x>", "<x/>"},
		{"<x> </x>", "<x/>"},
		{"<x a=\"b\"></x>", "<x a=\"b\"/>"},
		{"<x a=\"\"></x>", "<x a=\"\"/>"},
		{"<x a=a></x>", "<x a=a/>"},
		{"<x a=\" a \n\r\t b \"/>", "<x a=\" a     b \"/>"},
		{"<x a=\"&apos;b&quot;\"></x>", "<x a=\"'b&#34;\"/>"},
		{"<x a=\"&quot;&quot;'\"></x>", "<x a='\"\"&#39;'/>"},
		{"<!DOCTYPE foo SYSTEM \"Foo.dtd\">", "<!DOCTYPE foo SYSTEM \"Foo.dtd\">"},
		{"text <!--comment--> text", "text text"},
		{"text\n<!--comment-->\ntext", "text\ntext"},
		{"<!doctype html>", "<!doctype html=>"}, // bad formatted, doctype must be uppercase and html must have attribute value
		{"<x>\n<!--y-->\n</x>", "<x></x>"},
		{"<style>lala{color:red}</style>", "<style>lala{color:red}</style>"},
		{`cats  and 	dogs `, `cats and dogs`},

		{`</0`, `</0`}, // go fuzz
	}

	m := minify.New()
	for _, tt := range xmlTests {
		b := &bytes.Buffer{}
		assert.Nil(t, Minify(m, b, bytes.NewBufferString(tt.xml), nil), "Minify must not return error in "+tt.xml)
		assert.Equal(t, tt.expected, b.String(), "Minify must give expected result in "+tt.xml)
	}
}

func TestXMLKeepWhitespace(t *testing.T) {
	var xmlTests = []struct {
		xml      string
		expected string
	}{
		{`cats  and 	dogs `, `cats and dogs`},
		{` <div> <i> test </i> <b> test </b> </div> `, `<div> <i> test </i> <b> test </b> </div>`},
		{"text\n<!--comment-->\ntext", "text\ntext"},
		{"text\n<!--comment-->text<!--comment--> text", "text\ntext text"},
		{"<x>\n<!--y-->\n</x>", "<x>\n</x>"},
		{"<style>lala{color:red}</style>", "<style>lala{color:red}</style>"},
		{"<x> <?xml?> </x>", "<x><?xml?> </x>"},
		{"<x> <![CDATA[ x ]]> </x>", "<x> x </x>"},
		{"<x> <![CDATA[ <<<<< ]]> </x>", "<x><![CDATA[ <<<<< ]]></x>"},
	}

	m := minify.New()
	xmlMinifier := &Minifier{KeepWhitespace: true}
	for _, tt := range xmlTests {
		b := &bytes.Buffer{}
		assert.Nil(t, xmlMinifier.Minify(m, b, bytes.NewBufferString(tt.xml), nil), "Minify must not return error in "+tt.xml)
		assert.Equal(t, tt.expected, b.String(), "Minify must give expected result in "+tt.xml)
	}
}

func TestReaderErrors(t *testing.T) {
	m := minify.New()
	r := test.NewErrorReader(0)
	w := &bytes.Buffer{}
	assert.Equal(t, test.ErrPlain, Minify(m, w, r, nil), "Minify must return error at first read")
}

func TestWriterErrors(t *testing.T) {
	var errorTests = []int{0, 1, 2, 3, 4, 5, 6, 7, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

	m := minify.New()
	for _, n := range errorTests {
		// writes:                  0             1    2 3 45678901    23 4 5 6    7   8                    9   0
		r := bytes.NewBufferString(`<!DOCTYPE foo><?xml?><a x=y z="val"><b/><c></c></a><![CDATA[data<<<<<]]>text</x`)
		w := test.NewErrorWriter(n)
		assert.Equal(t, test.ErrPlain, Minify(m, w, r, nil), "Minify must return error at write "+strconv.FormatInt(int64(n), 10))
	}
}

////////////////////////////////////////////////////////////////

func ExampleMinify() {
	m := minify.New()
	m.AddFuncRegexp(regexp.MustCompile("[/+]xml$"), Minify)

	if err := m.Minify("text/xml", os.Stdout, os.Stdin); err != nil {
		panic(err)
	}
}
