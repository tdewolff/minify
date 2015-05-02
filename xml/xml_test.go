package xml // import "github.com/tdewolff/minify/xml"

import (
	"bytes"
	"math/rand"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tdewolff/minify"
)

func assertXML(t *testing.T, input, expected string) {
	m := minify.New()
	m.AddFunc("text/xml", Minify)
	b := &bytes.Buffer{}
	assert.Nil(t, m.Minify("text/xml", b, bytes.NewBufferString(input)), "Minify must not return error in "+input)
	assert.Equal(t, expected, b.String(), "Minify must give expected result in "+input)
}

func helperRand(n, m int, chars []byte) []string {
	r := make([]string, n)
	for i := range r {
		for j := 0; j < m; j++ {
			r[i] += string(chars[rand.Intn(len(chars))])
		}
	}
	return r
}

func assertAttrVal(t *testing.T, input, expected string) {
	buf := make([]byte, len(input))
	assert.Equal(t, expected, string(escapeAttrVal(&buf, []byte(input))))
}

////////////////////////////////////////////////////////////////

func TestXML(t *testing.T) {
	assertXML(t, "<!-- comment -->", "")
	assertXML(t, "<a><b>x</b></a>", "<a><b>x</b></a>")
	assertXML(t, "<a><b>x\ny</b></a>", "<a><b>x y</b></a>")
	assertXML(t, "<a><![CDATA[<b>]]></a>", "<a>&lt;b></a>")
	assertXML(t, "<a><![CDATA[ <b> ]]></a>", "<a>&lt;b></a>")
	assertXML(t, "<a><![CDATA[<<<<<]]></a>", "<a><![CDATA[<<<<<]]></a>")
	assertXML(t, "<a><![CDATA[&&&&]]></a>", "<a><![CDATA[&&&&]]></a>")
	assertXML(t, "<?xml version=\"1.0\" ?>", "<?xml version=\"1.0\"?>")
	assertXML(t, "<x></x>", "<x/>")
	assertXML(t, "<x a=\"b\"></x>", "<x a=\"b\"/>")
	assertXML(t, "<x> </x>", "<x/>")
	assertXML(t, "<x a=\" a \n\r\t b \"/>", "<x a=\" a     b \"/>")
	assertXML(t, "<!DOCTYPE foo SYSTEM \"Foo.dtd\">", "<!DOCTYPE foo SYSTEM \"Foo.dtd\">") // lower-case?
}

func TestWhitespace(t *testing.T) {
	multipleWhitespaceRegexp := regexp.MustCompile("\\s+")
	array := helperRand(100, 20, []byte("abcdefg \n\r\f\t"))
	for _, e := range array {
		reference := multipleWhitespaceRegexp.ReplaceAll([]byte(e), []byte(" "))
		assert.Equal(t, reference, replaceMultipleWhitespace([]byte(e)), "must remove all multiple whitespace")
	}
}

func TestHelpers(t *testing.T) {
	assertAttrVal(t, "xyz", "\"xyz\"")
	assertAttrVal(t, "", "\"\"")
	assertAttrVal(t, "x&amp;z", "\"x&amp;z\"")
	assertAttrVal(t, "x'z", "\"x'z\"")
	assertAttrVal(t, "x\"z", "'x\"z'")
}
