package html // import "github.com/tdewolff/minify/html"

import (
	"bytes"
	"io"
	"io/ioutil"
	"math/rand"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tdewolff/minify"
)

func assertHTML(t *testing.T, m minify.Minifier, input, expected string) {
	b := &bytes.Buffer{}
	assert.Nil(t, Minify(m, b, bytes.NewBufferString(input)), "Minify must not return error in "+input)
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

////////////////////////////////////////////////////////////////

func TestHTML(t *testing.T) {
	m := minify.NewMinifier()
	assertHTML(t, m, "html", "html")
	assertHTML(t, m, "<!DOCTYPE html PUBLIC \"-//W3C//DTD XHTML+RDFa 1.0//EN\" \"http://www.w3.org/MarkUp/DTD/xhtml-rdfa-1.dtd\">", "<!doctype html>")
	assertHTML(t, m, "<!-- comment -->", "")
	assertHTML(t, m, "<!--[if IE 6]>html<![endif]-->", "<!--[if IE 6]>html<![endif]-->")
	assertHTML(t, m, "<!--[if IE 6]><!--html--><![endif]-->", "<!--[if IE 6]><!--html--><![endif]-->")
	assertHTML(t, m, "<!--[if IE 6]><style><!--\ncss\n--></style><![endif]-->", "<!--[if IE 6]><style><!--\ncss\n--></style><![endif]-->")
	assertHTML(t, m, "<style><!--\ncss\n--></style>", "<style><!--\ncss\n--></style>")
	assertHTML(t, m, "<style>&</style>", "<style>&</style>")
	assertHTML(t, m, "<html><head></head><body>x</body></html>", "x")
	assertHTML(t, m, "<meta http-equiv=\"content-type\" content=\"text/html; charset=utf-8\">", "<meta charset=utf-8>")
	assertHTML(t, m, "<meta name=\"keywords\" content=\"a, b\">", "<meta name=keywords content=a,b>")
	assertHTML(t, m, "<meta name=\"viewport\" content=\"width = 996\" />", "<meta name=viewport content=\"width=996\">")
	assertHTML(t, m, "<span attr=\"test\"></span>", "<span attr=test></span>")
	assertHTML(t, m, "<span attr='test&apos;test'></span>", "<span attr=\"test'test\"></span>")
	assertHTML(t, m, "<span attr=\"test&quot;test\"></span>", "<span attr='test\"test'></span>")
	assertHTML(t, m, "<span attr='test\"\"&apos;&amp;test'></span>", "<span attr='test\"\"&#39;&amp;test'></span>")
	assertHTML(t, m, "<span attr=\"test/test\"></span>", "<span attr=\"test/test\"></span>")
	assertHTML(t, m, "<span>&amp;</span>", "<span>&amp;</span>")
	assertHTML(t, m, "<span clear=none method=GET></span>", "<span></span>")
	assertHTML(t, m, "<span onload=\"javascript:x;\"></span>", "<span onload=x;></span>")
	assertHTML(t, m, "<span href=\"http://test\"></span>", "<span href=\"//test\"></span>")
	assertHTML(t, m, "<span selected=\"selected\"></span>", "<span selected></span>")
	assertHTML(t, m, "<noscript><html></noscript>", "<noscript></noscript>")
	assertHTML(t, m, "<body id=\"main\"></body>", "<body id=main>")

	//assertHTML(t, m, "<!--[if IE 6]>some   spaces<![endif]-->", "<!--[if IE 6]>some spaces<![endif]-->") // TODO: make this work by changing the tokenizer code, see other TODO

	// increase coverage
	assertHTML(t, m, "<script style=\"css\">js</script>", "<script style=css>js</script>")
	assertHTML(t, m, "<meta http-equiv=\"content-type\" content=\"text/plain, text/html\">", "<meta http-equiv=content-type content=\"text/plain,text/html\">")
	assertHTML(t, m, "<meta http-equiv=\"content-style-type\" content=\"text/less\">", "<meta http-equiv=content-style-type content=\"text/less\">")
	assertHTML(t, m, "<meta http-equiv=\"content-script-type\" content=\"application/js\">", "<meta http-equiv=content-script-type content=\"application/js\">")
	assertHTML(t, m, "<span attr=\"\"></span>", "<span attr></span>")
	assertHTML(t, m, "<code>x</code>", "<code>x</code>")
	assertHTML(t, m, "<br/>", "<br>")
	assertHTML(t, m, "<p></p><p></p>", "<p><p>")
	assertHTML(t, m, "<ul><li></li> <li></li></ul>", "<ul><li><li></ul>")
	assertHTML(t, m, "<ul><li></li><a></a></ul>", "<ul><li></li><a></a></ul>")
	assertHTML(t, m, "<p></p><a></a>", "<p></p><a></a>")
	assertHTML(t, m, "<p></p>x<a></a>", "<p></p>x<a></a>")

	// whitespace
	assertHTML(t, m, "cats  and 	dogs", "cats and dogs")
	assertHTML(t, m, " <div> <i> test </i> <b> test </b> </div> ", "<div><i>test</i> <b>test</b></div>")
	assertHTML(t, m, "<strong>x </strong>y", "<strong>x </strong>y")
	assertHTML(t, m, "<strong>x </strong> y", "<strong>x</strong> y")
	assertHTML(t, m, "<strong>x </strong>\ny", "<strong>x</strong> y")
	assertHTML(t, m, "<p>x </p>y", "<p>x</p>y")
	assertHTML(t, m, "x <p>y</p>", "x<p>y")

	// from HTML Minifier
	assertHTML(t, m, "<DIV TITLE=\"blah\">boo</DIV>", "<div title=blah>boo</div>")
	assertHTML(t, m, "<p title\n\n\t  =\n     \"bar\">foo</p>", "<p title=bar>foo")
	assertHTML(t, m, "<p class=\" foo      \">foo bar baz</p>", "<p class=foo>foo bar baz")
	assertHTML(t, m, "<a href=\"   http://example.com  \">x</a>", "<a href=\"//example.com\">x</a>")
	assertHTML(t, m, "<input maxlength=\"     5 \">", "<input maxlength=5>")
	assertHTML(t, m, "<input type=\"text\">", "<input>")
	assertHTML(t, m, "<form method=\"get\">", "<form>")
	assertHTML(t, m, "<script language=\"Javascript\">alert(1)</script>", "<script>alert(1)</script>")
	assertHTML(t, m, "<script></script>", "")
	assertHTML(t, m, "<p onclick=\" JavaScript: x\">x</p>", "<p onclick=\" x\">x")
	assertHTML(t, m, "<link rel=\"stylesheet\" type=\"text/css\" href=\"http://example.com\">", "<link rel=stylesheet href=\"//example.com\">")
	assertHTML(t, m, "<span Selected=\"selected\"></span>", "<span selected></span>")
	assertHTML(t, m, "<table><thead><tr><th>foo</th><th>bar</th></tr></thead><tfoot><tr><th>baz</th><th>qux</th></tr></tfoot><tbody><tr><td>boo</td><td>moo</td></tr></tbody></table>",
		"<table><thead><tr><th>foo<th>bar<tfoot><tr><th>baz<th>qux<tbody><tr><td>boo<td>moo</table>")
	assertHTML(t, m, "<select><option>foo</option><option>bar</option></select>", "<select><option>foo<option>bar</select>")
}

func TestWhitespace(t *testing.T) {
	multipleWhitespaceRegexp := regexp.MustCompile("\\s+")
	array := helperRand(100, 20, []byte("abcdefg \n\r\f\t"))
	for _, e := range array {
		reference := multipleWhitespaceRegexp.ReplaceAll([]byte(e), []byte(" "))
		assert.Equal(t, reference, replaceMultipleWhitespace([]byte(e)), "must remove all multiple whitespace")
	}
}

func TestSpecialTagClosing(t *testing.T) {
	m := minify.NewMinifier()
	m.Add("text/css", func(m minify.Minifier, w io.Writer, r io.Reader) error {
		b, _ := ioutil.ReadAll(r)
		assert.Equal(t, "</script>", string(b))
		w.Write(b)
		return nil
	})
	assertHTML(t, m, "<style></script></style>", "<style></script></style>")
}

func TestHelpers(t *testing.T) {
	assert.Equal(t, []byte("xyz"), escapeText([]byte("xyz")))
	assert.Equal(t, []byte("x&amp;z"), escapeText([]byte("x&z")), "ampersand must be escaped")

	assert.Equal(t, []byte("xyz"), escapeAttrVal([]byte("xyz")))
	assert.Equal(t, []byte("\"\""), escapeAttrVal([]byte("")))
	assert.Equal(t, []byte("x&amp;z"), escapeAttrVal([]byte("x&z")))
	assert.Equal(t, []byte("\"x/z\""), escapeAttrVal([]byte("x/z")))
	assert.Equal(t, []byte("\"x'z\""), escapeAttrVal([]byte("x'z")))
	assert.Equal(t, []byte("'x\"z'"), escapeAttrVal([]byte("x\"z")))

	assert.Equal(t, []byte("text/html"), normalizeContentType([]byte("text/html")))
	assert.Equal(t, []byte("text/html;charset=utf-8"), normalizeContentType([]byte("text/html; charset=UTF-8")))
	assert.Equal(t, []byte("text/html,text/css"), normalizeContentType([]byte("text/html, text/css")))
}
