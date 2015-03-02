package html // import "github.com/tdewolff/minify/html"

import (
	"bytes"
	"io"
	"io/ioutil"
	"math/rand"
	"regexp"
	"testing"

	"github.com/tdewolff/minify"
	hash "github.com/tdewolff/parse/html"
)

func helperTestHTML(t *testing.T, m minify.Minifier, input, expected string) {
	b := &bytes.Buffer{}
	if err := Minify(m, b, bytes.NewBufferString(input)); err != nil {
		t.Error(err)
	}

	if b.String() != expected {
		t.Error(b.String(), "!=", expected)
	}
}

func helperTestString(t *testing.T, s, r string) {
	if s != r {
		t.Error(s, "!=", r)
	}
}

func helperTestBytes(t *testing.T, s, r []byte) {
	helperTestString(t, string(s), string(r))
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
	helperTestHTML(t, m, "html", "html")
	helperTestHTML(t, m, "<!DOCTYPE html PUBLIC \"-//W3C//DTD XHTML+RDFa 1.0//EN\" \"http://www.w3.org/MarkUp/DTD/xhtml-rdfa-1.dtd\">", "<!doctype html>")
	helperTestHTML(t, m, "<!-- comment -->", "")
	helperTestHTML(t, m, "<!--[if IE 6]>html<![endif]-->", "<!--[if IE 6]>html<![endif]-->")
	helperTestHTML(t, m, "<!--[if IE 6]><!--html--><![endif]-->", "<!--[if IE 6]><!--html--><![endif]-->")
	helperTestHTML(t, m, "<!--[if IE 6]><style><!--\ncss\n--></style><![endif]-->", "<!--[if IE 6]><style><!--\ncss\n--></style><![endif]-->")
	helperTestHTML(t, m, "<style><!--\ncss\n--></style>", "<style><!--\ncss\n--></style>")
	helperTestHTML(t, m, "<style>&</style>", "<style>&</style>")
	helperTestHTML(t, m, "<html><head></head><body>x</body></html>", "x")
	helperTestHTML(t, m, "<meta http-equiv=\"content-type\" content=\"text/html; charset=utf-8\">", "<meta charset=utf-8>")
	helperTestHTML(t, m, "<meta name=\"keywords\" content=\"a, b\">", "<meta name=keywords content=a,b>")
	helperTestHTML(t, m, "<meta name=\"viewport\" content=\"width = 996\" />", "<meta name=viewport content=\"width=996\">")
	helperTestHTML(t, m, "<span attr=\"test\"></span>", "<span attr=test></span>")
	helperTestHTML(t, m, "<span attr='test&apos;test'></span>", "<span attr=\"test'test\"></span>")
	helperTestHTML(t, m, "<span attr=\"test&quot;test\"></span>", "<span attr='test\"test'></span>")
	helperTestHTML(t, m, "<span attr='test\"\"&apos;&amp;test'></span>", "<span attr='test\"\"&#39;&amp;test'></span>")
	helperTestHTML(t, m, "<span attr=\"test/test\"></span>", "<span attr=\"test/test\"></span>")
	helperTestHTML(t, m, "<span>&amp;</span>", "<span>&amp;</span>")
	helperTestHTML(t, m, "<span clear=none method=GET></span>", "<span></span>")
	helperTestHTML(t, m, "<span onload=\"javascript:x;\"></span>", "<span onload=x;></span>")
	helperTestHTML(t, m, "<span href=\"http://test\"></span>", "<span href=\"//test\"></span>")
	helperTestHTML(t, m, "<span selected=\"selected\"></span>", "<span selected></span>")
	helperTestHTML(t, m, "<noscript><html></noscript>", "<noscript></noscript>")
	helperTestHTML(t, m, "<body id=\"main\"></body>", "<body id=main>")

	//helperTestHTML(t, m, "<!--[if IE 6]>some   spaces<![endif]-->", "<!--[if IE 6]>some spaces<![endif]-->") // TODO: make this work by changing the tokenizer code, see other TODO

	// increase coverage
	helperTestHTML(t, m, "<script style=\"css\">js</script>", "<script style=css>js</script>")
	helperTestHTML(t, m, "<meta http-equiv=\"content-type\" content=\"text/plain, text/html\">", "<meta http-equiv=content-type content=\"text/plain,text/html\">")
	helperTestHTML(t, m, "<meta http-equiv=\"content-style-type\" content=\"text/less\">", "<meta http-equiv=content-style-type content=\"text/less\">")
	helperTestHTML(t, m, "<meta http-equiv=\"content-script-type\" content=\"application/js\">", "<meta http-equiv=content-script-type content=\"application/js\">")
	helperTestHTML(t, m, "<span attr=\"\"></span>", "<span attr></span>")
	helperTestHTML(t, m, "<code>x</code>", "<code>x</code>")
	helperTestHTML(t, m, "<br/>", "<br>")
	helperTestHTML(t, m, "<p></p><p></p>", "<p><p>")
	helperTestHTML(t, m, "<ul><li></li> <li></li></ul>", "<ul><li><li></ul>")
	helperTestHTML(t, m, "<ul><li></li><a></a></ul>", "<ul><li></li><a></a></ul>")
	helperTestHTML(t, m, "<p></p><a></a>", "<p></p><a></a>")
	helperTestHTML(t, m, "<p></p>x<a></a>", "<p></p>x<a></a>")

	// whitespace
	helperTestHTML(t, m, "cats  and 	dogs", "cats and dogs")
	helperTestHTML(t, m, " <div> <i> test </i> <b> test </b> </div> ", "<div><i>test</i> <b>test</b></div>")
	helperTestHTML(t, m, "<strong>x </strong>y", "<strong>x </strong>y")
	helperTestHTML(t, m, "<strong>x </strong> y", "<strong>x</strong> y")
	helperTestHTML(t, m, "<strong>x </strong>\ny", "<strong>x</strong> y")
	helperTestHTML(t, m, "<p>x </p>y", "<p>x</p>y")
	helperTestHTML(t, m, "x <p>y</p>", "x<p>y")

	// from HTML Minifier
	helperTestHTML(t, m, "<DIV TITLE=\"blah\">boo</DIV>", "<div title=blah>boo</div>")
	helperTestHTML(t, m, "<p title\n\n\t  =\n     \"bar\">foo</p>", "<p title=bar>foo")
	helperTestHTML(t, m, "<p class=\" foo      \">foo bar baz</p>", "<p class=foo>foo bar baz")
	helperTestHTML(t, m, "<a href=\"   http://example.com  \">x</a>", "<a href=\"//example.com\">x</a>")
	helperTestHTML(t, m, "<input maxlength=\"     5 \">", "<input maxlength=5>")
	helperTestHTML(t, m, "<input type=\"text\">", "<input>")
	helperTestHTML(t, m, "<form method=\"get\">", "<form>")
	helperTestHTML(t, m, "<script language=\"Javascript\">alert(1)</script>", "<script>alert(1)</script>")
	helperTestHTML(t, m, "<script></script>", "")
	helperTestHTML(t, m, "<p onclick=\" JavaScript: x\">x</p>", "<p onclick=\" x\">x")
	helperTestHTML(t, m, "<link rel=\"stylesheet\" type=\"text/css\" href=\"http://example.com\">", "<link rel=stylesheet href=\"//example.com\">")
	helperTestHTML(t, m, "<span Selected=\"selected\"></span>", "<span selected></span>")
	helperTestHTML(t, m, "<table><thead><tr><th>foo</th><th>bar</th></tr></thead><tfoot><tr><th>baz</th><th>qux</th></tr></tfoot><tbody><tr><td>boo</td><td>moo</td></tr></tbody></table>",
		"<table><thead><tr><th>foo<th>bar<tfoot><tr><th>baz<th>qux<tbody><tr><td>boo<td>moo</table>")
	helperTestHTML(t, m, "<select><option>foo</option><option>bar</option></select>", "<select><option>foo<option>bar</select>")
}

func TestWhitespace(t *testing.T) {
	multipleWhitespaceRegexp := regexp.MustCompile("\\s+")

	array := helperRand(100, 20, []byte("abcdefg \n\r\f\t"))
	for _, e := range array {
		reference := multipleWhitespaceRegexp.ReplaceAll([]byte(e), []byte(" "))
		actual := replaceMultipleWhitespace([]byte(e))
		if !bytes.Equal(actual, reference) {
			t.Error(actual, "!=", reference)
		}
	}
}

func TestSpecialTagClosing(t *testing.T) {
	m := minify.NewMinifier()
	m.Add("text/css", func(m minify.Minifier, w io.Writer, r io.Reader) error {
		b, _ := ioutil.ReadAll(r)
		if string(b) != "</script>" {
			t.Error(string(b), "!= </script>")
		}
		w.Write(b)
		return nil
	})
	helperTestHTML(t, m, "<style></script></style>", "<style></script></style>")
}

func TestHashtable(t *testing.T) {
	helperTestString(t, "address", hash.Address.String())
	helperTestString(t, "accept-charset", hash.Accept_Charset.String())
}

func TestHelpers(t *testing.T) {
	helperTestBytes(t, escapeText([]byte("xyz")), []byte("xyz"))
	helperTestBytes(t, escapeText([]byte("x&z")), []byte("x&amp;z"))

	helperTestBytes(t, escapeAttrVal([]byte("xyz")), []byte("xyz"))
	helperTestBytes(t, escapeAttrVal([]byte("")), []byte("\"\""))
	helperTestBytes(t, escapeAttrVal([]byte("x&z")), []byte("x&amp;z"))
	helperTestBytes(t, escapeAttrVal([]byte("x/z")), []byte("\"x/z\""))
	helperTestBytes(t, escapeAttrVal([]byte("x'z")), []byte("\"x'z\""))
	helperTestBytes(t, escapeAttrVal([]byte("x\"z")), []byte("'x\"z'"))

	helperTestBytes(t, normalizeContentType([]byte("text/html")), []byte("text/html"))
	helperTestBytes(t, normalizeContentType([]byte("text/html; charset=UTF-8")), []byte("text/html;charset=utf-8"))
	helperTestBytes(t, normalizeContentType([]byte("text/html, text/css")), []byte("text/html,text/css"))
}
