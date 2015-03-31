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

func assertHTML(t *testing.T, input, expected string) {
	m := minify.New()
	m.AddFunc("text/html", Minify)
	b := &bytes.Buffer{}
	assert.Nil(t, m.Minify("text/html", b, bytes.NewBufferString(input)), "Minify must not return error in "+input)
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
	assert.Equal(t, expected, string(escapeAttrVal([]byte(input))))
}

////////////////////////////////////////////////////////////////

func TestHTML(t *testing.T) {
	assertHTML(t, "html", "html")
	assertHTML(t, "<!DOCTYPE html PUBLIC \"-//W3C//DTD XHTML+RDFa 1.0//EN\" \"http://www.w3.org/MarkUp/DTD/xhtml-rdfa-1.dtd\">", "<!doctype html>")
	assertHTML(t, "<!-- comment -->", "")
	assertHTML(t, "<!--[if IE 6]>html<![endif]-->", "<!--[if IE 6]>html<![endif]-->")
	assertHTML(t, "<!--[if IE 6]><!--html--><![endif]-->", "<!--[if IE 6]><!--html--><![endif]-->")
	assertHTML(t, "<!--[if IE 6]><style><!--\ncss\n--></style><![endif]-->", "<!--[if IE 6]><style><!--\ncss\n--></style><![endif]-->")
	assertHTML(t, "<style><!--\ncss\n--></style>", "<style><!--\ncss\n--></style>")
	assertHTML(t, "<style>&</style>", "<style>&</style>")
	assertHTML(t, "<html><head></head><body>x</body></html>", "x")
	assertHTML(t, "<meta http-equiv=\"content-type\" content=\"text/html; charset=utf-8\">", "<meta charset=utf-8>")
	assertHTML(t, "<meta http-equiv=\"Content-Type\" content=\"text/html; charset=UTF-8\" />", "<meta charset=utf-8>")
	assertHTML(t, "<meta name=\"keywords\" content=\"a, b\">", "<meta name=keywords content=a,b>")
	assertHTML(t, "<meta name=\"viewport\" content=\"width = 996\" />", "<meta name=viewport content=\"width=996\">")
	assertHTML(t, "<span attr=\"test\"></span>", "<span attr=test></span>")
	assertHTML(t, "<span attr='test&apos;test'></span>", "<span attr=\"test'test\"></span>")
	assertHTML(t, "<span attr=\"test&quot;test\"></span>", "<span attr='test\"test'></span>")
	assertHTML(t, "<span attr='test\"\"&apos;&amp;test'></span>", "<span attr='test\"\"&#39;&amp;test'></span>")
	assertHTML(t, "<span attr=\"test/test\"></span>", "<span attr=test/test></span>")
	assertHTML(t, "<span>&amp;</span>", "<span>&amp;</span>")
	assertHTML(t, "<span clear=none method=GET></span>", "<span></span>")
	assertHTML(t, "<span onload=\"javascript:x;\"></span>", "<span onload=x;></span>")
	assertHTML(t, "<span href=\"http://test\"></span>", "<span href=//test></span>")
	assertHTML(t, "<span selected=\"selected\"></span>", "<span selected></span>")
	assertHTML(t, "<noscript><html><img id=\"x\"></noscript>", "<noscript><img id=x></noscript>")
	assertHTML(t, "<body id=\"main\"></body>", "<body id=main>")

	//assertHTML(t, "<!--[if IE 6]>some   spaces<![endif]-->", "<!--[if IE 6]>some spaces<![endif]-->") // TODO: make this work by changing the tokenizer code, see other TODO

	// increase coverage
	assertHTML(t, "<script style=\"css\">js</script>", "<script style=css>js</script>")
	assertHTML(t, "<meta http-equiv=\"content-type\" content=\"text/plain, text/html\">", "<meta http-equiv=content-type content=text/plain,text/html>")
	assertHTML(t, "<meta http-equiv=\"content-style-type\" content=\"text/less\">", "<meta http-equiv=content-style-type content=text/less>")
	assertHTML(t, "<meta http-equiv=\"content-script-type\" content=\"application/js\">", "<meta http-equiv=content-script-type content=application/js>")
	assertHTML(t, "<span attr=\"\"></span>", "<span attr></span>")
	assertHTML(t, "<code>x</code>", "<code>x</code>")
	assertHTML(t, "<br/>", "<br>")
	assertHTML(t, "<p></p><p></p>", "<p><p>")
	assertHTML(t, "<ul><li></li> <li></li></ul>", "<ul><li><li></ul>")
	//assertHTML(t, "<ul><li></li><a></a></ul>", "<ul><li></li><a></a></ul>")
	assertHTML(t, "<p></p><a></a>", "<p></p><a></a>")
	assertHTML(t, "<p></p>x<a></a>", "<p></p>x<a></a>")

	// whitespace
	assertHTML(t, "cats  and 	dogs", "cats and dogs")
	assertHTML(t, " <div> <i> test </i> <b> test </b> </div> ", "<div><i>test</i> <b>test</b></div>")
	assertHTML(t, "<strong>x </strong>y", "<strong>x </strong>y")
	assertHTML(t, "<strong>x </strong> y", "<strong>x</strong> y")
	assertHTML(t, "<strong>x </strong>\ny", "<strong>x</strong> y")
	assertHTML(t, "<p>x </p>y", "<p>x</p>y")
	assertHTML(t, "x <p>y</p>", "x<p>y")
	assertHTML(t, " <!doctype html> <!--comment--> <html> <body><p></p></body></html>", "<!doctype html><p>") // spaces before html and at the start of html are dropped

	// from HTML Minifier
	assertHTML(t, "<DIV TITLE=\"blah\">boo</DIV>", "<div title=blah>boo</div>")
	assertHTML(t, "<p title\n\n\t  =\n     \"bar\">foo</p>", "<p title=bar>foo")
	assertHTML(t, "<p class=\" foo      \">foo bar baz</p>", "<p class=foo>foo bar baz")
	assertHTML(t, "<a href=\"   http://example.com  \">x</a>", "<a href=//example.com>x</a>")
	assertHTML(t, "<input maxlength=\"     5 \">", "<input maxlength=5>")
	assertHTML(t, "<input type=\"text\">", "<input>")
	assertHTML(t, "<form method=\"get\">", "<form>")
	assertHTML(t, "<script language=\"Javascript\">alert(1)</script>", "<script>alert(1)</script>")
	assertHTML(t, "<script></script>", "")
	assertHTML(t, "<p onclick=\" JavaScript: x\">x</p>", "<p onclick=\" x\">x")
	assertHTML(t, "<link rel=\"stylesheet\" type=\"text/css\" href=\"http://example.com\">", "<link rel=stylesheet href=//example.com>")
	assertHTML(t, "<span Selected=\"selected\"></span>", "<span selected></span>")
	assertHTML(t, "<table><thead><tr><th>foo</th><th>bar</th></tr></thead><tfoot><tr><th>baz</th><th>qux</th></tr></tfoot><tbody><tr><td>boo</td><td>moo</td></tr></tbody></table>",
		"<table><thead><tr><th>foo<th>bar<tfoot><tr><th>baz<th>qux<tbody><tr><td>boo<td>moo</table>")
	assertHTML(t, "<select><option>foo</option><option>bar</option></select>", "<select><option>foo<option>bar</select>")

	assertHTML(t, `<!doctype html> <html xmlns="http://www.w3.org/1999/xhtml" xml:lang="en"> <head profile="http://dublincore.org/documents/dcq-html/"> <!-- Barlesque 2.75.0 --> <meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />`, `<!doctype html><html xmlns=//www.w3.org/1999/xhtml xml:lang=en><head profile=//dublincore.org/documents/dcq-html/><meta charset=utf-8>`)
	assertHTML(t, `<meta name="keywords" content="A, B">`, `<meta name=keywords content=A,B>`)
	assertHTML(t, `<script type="text/html"><![CDATA[ <img id="x"> ]]></script>`, `<script type=text/html><![CDATA[<img id=x>]]></script>`)
	assertHTML(t, `<iframe><html> <p> x </p> </html></iframe>`, `<iframe><p>x</iframe>`)
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
	m := minify.New()
	m.AddFunc("text/css", func(m minify.Minifier, mediatype string, w io.Writer, r io.Reader) error {
		b, _ := ioutil.ReadAll(r)
		assert.Equal(t, "</script>", string(b))
		w.Write(b)
		return nil
	})
	assertHTML(t, "<style></script></style>", "<style></script></style>")
}

func TestHelpers(t *testing.T) {
	assertAttrVal(t, "xyz", "xyz")
	assertAttrVal(t, "", "")
	assertAttrVal(t, "x&amp;z", "x&amp;z")
	assertAttrVal(t, "x/z", "x/z")
	assertAttrVal(t, "x'z", "\"x'z\"")
	assertAttrVal(t, "x\"z", "'x\"z'")
	assertAttrVal(t, "You&#039;re encouraged to log in; however, it&#039;s not mandatory. [o]", "\"You're encouraged to log in; however, it's not mandatory. [o]\"")

	assert.Equal(t, "text/html", string(normalizeContentType([]byte("text/html"))))
	assert.Equal(t, "text/html;charset=utf-8", string(normalizeContentType([]byte("text/html; charset=UTF-8"))))
	assert.Equal(t, "text/html,text/css", string(normalizeContentType([]byte("text/html, text/css"))))
}
