package minify

import (
	"bytes"
	"io"
	"io/ioutil"
	"math/rand"
	"regexp"
	"testing"
)

func helperHTML(t *testing.T, m *Minifier, input, expected string) {
	b := &bytes.Buffer{}
	if err := m.HTML(b, bytes.NewBufferString(input)); err != nil {
		t.Error(err)
	}

	if b.String() != expected {
		t.Error(b.String(), "!=", expected)
	}
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
	m := NewMinifier()
	helperHTML(t, m, "html", "html")
	helperHTML(t, m, "<!DOCTYPE html PUBLIC \"-//W3C//DTD XHTML+RDFa 1.0//EN\" \"http://www.w3.org/MarkUp/DTD/xhtml-rdfa-1.dtd\">", "<!doctype html>")
	helperHTML(t, m, "<!-- comment -->", "")
	helperHTML(t, m, "<!--[if IE 6]>html<![endif]-->", "<!--[if IE 6]>html<![endif]-->")
	helperHTML(t, m, "<!--[if IE 6]><!--html--><![endif]-->", "<!--[if IE 6]><!--html--><![endif]-->")
	helperHTML(t, m, "<!--[if IE 6]><style><!--\ncss\n--></style><![endif]-->", "<!--[if IE 6]><style><!--\ncss\n--></style><![endif]-->")
	helperHTML(t, m, "<style><!--\ncss\n--></style>", "<style><!-- css --></style>")
	helperHTML(t, m, "<html><head></head><body>x</body></html>", "x")
	helperHTML(t, m, "<meta http-equiv=\"content-type\" content=\"text/html; charset=utf-8\">", "<meta charset=utf-8>")
	helperHTML(t, m, "<meta name=\"keywords\" content=\"a, b\">", "<meta name=keywords content=a,b>")
	helperHTML(t, m, "<meta name=\"viewport\" content=\"width = 996\" />", "<meta name=viewport content=\"width=996\">")
	helperHTML(t, m, "<span attr=\"test\"></span>", "<span attr=test></span>")
	helperHTML(t, m, "<span attr='test&apos;test'></span>", "<span attr=\"test'test\"></span>")
	helperHTML(t, m, "<span attr=\"test&quot;test\"></span>", "<span attr='test\"test'></span>")
	helperHTML(t, m, "<span attr=\"test/test\"></span>", "<span attr=\"test/test\"></span>")
	helperHTML(t, m, "<span method=GET></span>", "<span></span>")
	helperHTML(t, m, "<span onload=\"javascript:x;\"></span>", "<span onload=x;></span>")
	helperHTML(t, m, "<span href=\"http://test\"></span>", "<span href=\"//test\"></span>")
	helperHTML(t, m, "<span selected=\"selected\"></span>", "<span selected></span>")
	helperHTML(t, m, "<noscript><html></noscript>", "<noscript></noscript>")
	helperHTML(t, m, "<body id=\"main\"></body>", "<body id=main>")

	//helperHTML(t, m, "<!--[if IE 6]>some   spaces<![endif]-->", "<!--[if IE 6]>some spaces<![endif]-->") // TODO: make this work by changing the tokenizer code, see other TODO

	// increase coverage
	helperHTML(t, m, "<script style=\"css\">js</script>", "<script style=css>js</script>")
	helperHTML(t, m, "<meta http-equiv=\"content-type\" content=\"text/plain, text/html\">", "<meta http-equiv=content-type content=\"text/plain,text/html\">")
	helperHTML(t, m, "<meta http-equiv=\"content-style-type\" content=\"text/less\">", "<meta http-equiv=content-style-type content=\"text/less\">")
	helperHTML(t, m, "<meta http-equiv=\"content-script-type\" content=\"application/js\">", "<meta http-equiv=content-script-type content=\"application/js\">")
	helperHTML(t, m, "<span attr=\"\"></span>", "<span attr></span>")
	helperHTML(t, m, "<code>x</code>", "<code>x</code>")
	helperHTML(t, m, "<br/>", "<br>")
	helperHTML(t, m, "<p></p><p></p>", "<p><p>")
	helperHTML(t, m, "<ul><li></li> <li></li></ul>", "<ul><li><li></ul>")
	helperHTML(t, m, "<ul><li></li><a></a></ul>", "<ul><li></li><a></a></ul>")
	helperHTML(t, m, "<p></p><a></a>", "<p></p><a></a>")
	helperHTML(t, m, "<p></p>x<a></a>", "<p></p>x<a></a>")

	// whitespace
	helperHTML(t, m, "cats  and 	dogs", "cats and dogs")
	helperHTML(t, m, " <div> <i> test </i> <b> test </b> </div> ", "<div><i>test</i> <b>test</b></div>")
	helperHTML(t, m, "<strong>x </strong>y", "<strong>x </strong>y")
	helperHTML(t, m, "<strong>x </strong> y", "<strong>x</strong> y")
	helperHTML(t, m, "<p>x </p>y", "<p>x</p>y")
	helperHTML(t, m, "x <p>y</p>", "x<p>y")

	// from HTML Minifier
	helperHTML(t, m, "<DIV TITLE=\"blah\">boo</DIV>", "<div title=blah>boo</div>")
	helperHTML(t, m, "<p title\n\n\t  =\n     \"bar\">foo</p>", "<p title=bar>foo")
	helperHTML(t, m, "<p class=\" foo      \">foo bar baz</p>", "<p class=foo>foo bar baz")
	helperHTML(t, m, "<a href=\"   http://example.com  \">x</a>", "<a href=\"//example.com\">x</a>")
	helperHTML(t, m, "<input maxlength=\"     5 \">", "<input maxlength=5>")
	helperHTML(t, m, "<input type=\"text\">", "<input>")
	helperHTML(t, m, "<form method=\"get\">", "<form>")
	helperHTML(t, m, "<script></script>", "")
	helperHTML(t, m, "<p onclick=\" JavaScript: x\">x</p>", "<p onclick=\" x\">x")
	helperHTML(t, m, "<link rel=\"stylesheet\" type=\"text/css\" href=\"http://example.com\">", "<link rel=stylesheet href=\"//example.com\">")
	helperHTML(t, m, "<span Selected=\"selected\"></span>", "<span selected></span>")
	helperHTML(t, m, "<table><thead><tr><th>foo</th><th>bar</th></tr></thead><tfoot><tr><th>baz</th><th>qux</th></tr></tfoot><tbody><tr><td>boo</td><td>moo</td></tr></tbody></table>",
		"<table><thead><tr><th>foo<th>bar<tfoot><tr><th>baz<th>qux<tbody><tr><td>boo<td>moo</table>")
	helperHTML(t, m, "<select><option>foo</option><option>bar</option></select>", "<select><option>foo<option>bar</select>")
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
	m := NewMinifier()
	m.Add("text/css", func(m Minifier, w io.Writer, r io.Reader) error {
		b, _ := ioutil.ReadAll(r)
		if string(b) != "</script>" {
			t.Error(string(b), "!= </script>")
		}
		w.Write(b)
		return nil
	})
	helperHTML(t, m, "<style></script></style>", "<style></script></style>")
}
