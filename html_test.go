package minify

import (
	"bytes"
	"math/rand"
	"regexp"
	"testing"
)

func helperHTML(t *testing.T, input, expected string) {
	m := &Minifier{}
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
	helperHTML(t, "html", "html")
	helperHTML(t, "<!DOCTYPE html PUBLIC \"-//W3C//DTD XHTML+RDFa 1.0//EN\" \"http://www.w3.org/MarkUp/DTD/xhtml-rdfa-1.dtd\">", "<!doctype html>")
	helperHTML(t, "<!-- comment -->", "")
	helperHTML(t, "<!--[if IE 6]>html<![endif]-->", "<!--[if IE 6]>html<![endif]-->")
	helperHTML(t, "<!--[if IE 6]><!--html--><![endif]-->", "<!--[if IE 6]><!--html--><![endif]-->") // TODO: tokenizer doesn't deal with conditionary comments
	helperHTML(t, "<!--[if IE 6]><style><!--\ncss\n--></style><![endif]-->", "<!--[if IE 6]><style><!--\ncss\n--></style><![endif]-->")
	helperHTML(t, "<style><!--\ncss\n--></style>", "<style><!--\ncss\n--></style>")
	helperHTML(t, "cats  and 	dogs", "cats and dogs")
	helperHTML(t, " <div> <i> test </i> <b> test </b> </div> ", "<div><i>test</i><b> test</b></div>")
	helperHTML(t, "<html><head></head><body>html</body></html>", "html")
	helperHTML(t, "<meta http-equiv=\"content-type\" content=\"text/html; charset=utf-8\">", "<meta charset=utf-8>")
	helperHTML(t, "<meta name=\"keywords\" content=\"a, b\">", "<meta name=keywords content=a,b>")
	helperHTML(t, "<meta name=\"viewport\" content=\"width = 996\" />", "<meta name=viewport content=\"width=996\">")
	helperHTML(t, "<span attr=\"test\"></span>", "<span attr=test></span>")
	helperHTML(t, "<span attr='test&apos;test'></span>", "<span attr=\"test'test\"></span>")
	helperHTML(t, "<span attr=\"test&quot;test\"></span>", "<span attr='test\"test'></span>")
	helperHTML(t, "<span attr=\"test/test\"></span>", "<span attr=\"test/test\"></span>")
	helperHTML(t, "<span clear=none method=get></span>", "<span></span>")
	helperHTML(t, "<span onload=\"javascript:x;\"></span>", "<span onload=x;></span>") // TODO: remove semicolon?
	helperHTML(t, "<span href=\"http://test\"></span>", "<span href=\"//test\"></span>")
	helperHTML(t, "<span selected=\"selected\"></span>", "<span selected></span>")

	// increase coverage
	helperHTML(t, "<script style=\"css\">js</script>", "<script style=css>js</script>")
	helperHTML(t, "<meta http-equiv=\"content-type\" content=\"text/plain, text/html\">", "<meta http-equiv=content-type content=\"text/plain,text/html\">")
	helperHTML(t, "<meta http-equiv=\"content-style-type\" content=\"text/less\">", "<meta http-equiv=content-style-type content=\"text/less\">")
	helperHTML(t, "<meta http-equiv=\"content-script-type\" content=\"application/js\">", "<meta http-equiv=content-script-type content=\"application/js\">")
	helperHTML(t, "<span attr=\"\"></span>", "<span attr></span>")
	helperHTML(t, "<code>x</code>", "<code>x</code>")
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