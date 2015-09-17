package xml // import "github.com/tdewolff/minify/xml"

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tdewolff/minify"
)

func TestXML(t *testing.T) {
	var xmlTests = []struct {
		xml      string
		expected string
	}{
		{"<!-- comment -->", ""},
		{"<A>x</A>", "<A>x</A>"},
		{"<a><b>x</b></a>", "<a><b>x</b></a>"},
		{"<a><b>x\ny</b></a>", "<a><b>x y</b></a>"},
		{"<a><![CDATA[<b>]]></a>", "<a>&lt;b></a>"},
		{"<a><![CDATA[abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz]]></a>", "<a>abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz</a>"},
		{"<a><![CDATA[ <b> ]]></a>", "<a>&lt;b></a>"},
		{"<a><![CDATA[<<<<<]]></a>", "<a><![CDATA[<<<<<]]></a>"},
		{"<a><![CDATA[&]]></a>", "<a>&amp;</a>"},
		{"<a><![CDATA[&&&&]]></a>", "<a><![CDATA[&&&&]]></a>"},
		{"<a> <![CDATA[ a ]]> </a>", "<a>a</a>"},
		{"<?xml version=\"1.0\" ?>", "<?xml version=\"1.0\"?>"},
		{"<x></x>", "<x/>"},
		{"<x> </x>", "<x/>"},
		{"<x a=\"b\"></x>", "<x a=\"b\"/>"},
		{"<x a=\"\"></x>", "<x a=\"\"/>"},
		{"<x a=a></x>", "<x a=a/>"},
		{"<x a=\" a \n\r\t b \"/>", "<x a=\" a     b \"/>"},
		{"<x a=\"&apos;b&quot;\"></x>", "<x a=\"'b&#34;\"/>"},
		{"<x a=\"&quot;&quot;'\"></x>", "<x a='\"\"&#39;'/>"},
		{"<!DOCTYPE foo SYSTEM \"Foo.dtd\">", "<!DOCTYPE foo SYSTEM \"Foo.dtd\">"},
	}

	m := minify.New()
	for _, tt := range xmlTests {
		b := &bytes.Buffer{}
		assert.Nil(t, Minify(m, "text/xml", b, bytes.NewBufferString(tt.xml)), "Minify must not return error in "+tt.xml)
		assert.Equal(t, tt.expected, b.String(), "Minify must give expected result in "+tt.xml)
	}
}

////////////////////////////////////////////////////////////////

func ExampleMinify() {
	m := minify.New()
	m.AddFuncRegexp(regexp.MustCompile("[/+]xml$"), Minify)

	if err := m.Minify("text/xml", os.Stdout, os.Stdin); err != nil {
		fmt.Println("minify.Minify:", err)
	}
}
