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

func assertXML(t *testing.T, input, expected string) {
	m := minify.New()
	b := &bytes.Buffer{}
	assert.Nil(t, Minify(m, b, bytes.NewBufferString(input), "text/xml", nil), "Minify must not return error in "+input)
	assert.Equal(t, expected, b.String(), "Minify must give expected result in "+input)
}

////////////////////////////////////////////////////////////////

func TestXML(t *testing.T) {
	assertXML(t, "<!-- comment -->", "")
	assertXML(t, "<A>x</A>", "<A>x</A>")
	assertXML(t, "<a><b>x</b></a>", "<a><b>x</b></a>")
	assertXML(t, "<a><b>x\ny</b></a>", "<a><b>x y</b></a>")
	assertXML(t, "<a><![CDATA[<b>]]></a>", "<a>&lt;b></a>")
	assertXML(t, "<a><![CDATA[abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz]]></a>", "<a>abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz</a>")
	assertXML(t, "<a><![CDATA[ <b> ]]></a>", "<a>&lt;b></a>")
	assertXML(t, "<a><![CDATA[<<<<<]]></a>", "<a><![CDATA[<<<<<]]></a>")
	assertXML(t, "<a><![CDATA[&]]></a>", "<a>&amp;</a>")
	assertXML(t, "<a><![CDATA[&&&&]]></a>", "<a><![CDATA[&&&&]]></a>")
	assertXML(t, "<a> <![CDATA[ a ]]> </a>", "<a>a</a>")
	assertXML(t, "<?xml version=\"1.0\" ?>", "<?xml version=\"1.0\"?>")
	assertXML(t, "<x></x>", "<x/>")
	assertXML(t, "<x> </x>", "<x/>")
	assertXML(t, "<x a=\"b\"></x>", "<x a=\"b\"/>")
	assertXML(t, "<x a=\"\"></x>", "<x a=\"\"/>")
	assertXML(t, "<x a=a></x>", "<x a=a/>")
	assertXML(t, "<x a=\" a \n\r\t b \"/>", "<x a=\" a     b \"/>")
	assertXML(t, "<x a=\"&apos;b&quot;\"></x>", "<x a=\"'b&#34;\"/>")
	assertXML(t, "<x a=\"&quot;&quot;'\"></x>", "<x a='\"\"&#39;'/>")
	assertXML(t, "<!DOCTYPE foo SYSTEM \"Foo.dtd\">", "<!DOCTYPE foo SYSTEM \"Foo.dtd\">")
}

////////////////////////////////////////////////////////////////

func ExampleMinify() {
	m := minify.New()
	m.AddFuncRegexp(regexp.MustCompile("[/+]xml$"), Minify)

	if err := m.Minify(os.Stdout, os.Stdin, "text/xml", nil); err != nil {
		fmt.Println("minify.Minify:", err)
	}
}
