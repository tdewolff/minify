package html // import "github.com/tdewolff/minify/html"

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/js"
	"github.com/tdewolff/minify/json"
	"github.com/tdewolff/minify/svg"
	"github.com/tdewolff/minify/xml"
	"github.com/tdewolff/test"
)

func TestHTML(t *testing.T) {
	var htmlTests = []struct {
		html     string
		expected string
	}{
		{`html`, `html`},
		{`<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML+RDFa 1.0//EN" "http://www.w3.org/MarkUp/DTD/xhtml-rdfa-1.dtd">`, `<!doctype html>`},
		{`<!-- comment -->`, ``},
		{`<!--[if IE 6]>html<![endif]-->`, `<!--[if IE 6]>html<![endif]-->`},
		{`<!--[if IE 6]><!--html--><![endif]-->`, `<!--[if IE 6]><!--html--><![endif]-->`},
		{`<!--[if IE 6]><style><!--\ncss\n--></style><![endif]-->`, `<!--[if IE 6]><style><!--\ncss\n--></style><![endif]-->`},
		{`<style><!--\ncss\n--></style>`, `<style><!--\ncss\n--></style>`},
		{`<style>&</style>`, `<style>&</style>`},
		{`<html><head></head><body>x</body></html>`, `x`},
		{`<meta http-equiv="content-type" content="text/html; charset=utf-8">`, `<meta charset=utf-8>`},
		{`<meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />`, `<meta charset=utf-8>`},
		{`<meta name="keywords" content="a, b">`, `<meta name=keywords content=a,b>`},
		{`<meta name="viewport" content="width = 996" />`, `<meta name=viewport content="width=996">`},
		{`<span attr="test"></span>`, `<span attr=test></span>`},
		{`<span attr='test&apos;test'></span>`, `<span attr="test'test"></span>`},
		{`<span attr="test&quot;test"></span>`, `<span attr='test"test'></span>`},
		{`<span attr='test""&apos;&amp;test'></span>`, `<span attr='test""&#39;&amp;test'></span>`},
		{`<span attr="test/test"></span>`, `<span attr=test/test></span>`},
		{`<span>&amp;</span>`, `<span>&amp;</span>`},
		{`<span clear=none method=GET></span>`, `<span></span>`},
		{`<span onload="javascript:x;"></span>`, `<span onload=x;></span>`},
		{`<span selected="selected"></span>`, `<span selected></span>`},
		{`<noscript><html><img id="x"></noscript>`, `<noscript><img id=x></noscript>`},
		{`<body id="main"></body>`, `<body id=main>`},
		{`<style><![CDATA[x]]></style>`, `<style>x</style>`},
		{`<link href="data:text/plain, data">`, `<link href=data:,+data>`},
		{`<svg width="100" height="100"><circle cx="50" cy="50" r="40" stroke="green" stroke-width="4" fill="yellow" /></svg>`, `<svg width=100 height=100><circle cx="50" cy="50" r="40" stroke="green" stroke-width="4" fill="yellow" /></svg>`},
		{`</span >`, `</span>`},

		// increase coverage
		{`<script style="css">js</script>`, `<script style=css>js</script>`},
		{`<meta http-equiv="content-type" content="text/plain, text/html">`, `<meta http-equiv=content-type content=text/plain,text/html>`},
		{`<meta http-equiv="content-style-type" content="text/less">`, `<meta http-equiv=content-style-type content=text/less>`},
		{`<meta http-equiv="content-script-type" content="application/js">`, `<meta http-equiv=content-script-type content=application/js>`},
		{`<span attr=""></span>`, `<span attr></span>`},
		{`<code>x</code>`, `<code>x</code>`},
		{`<br/>`, `<br>`},
		{`<p></p><p></p>`, `<p><p>`},
		{`<ul><li></li> <li></li></ul>`, `<ul><li><li></ul>`},
		{`<p></p><a></a>`, `<p></p><a></a>`},
		{`<p></p>x<a></a>`, `<p></p>x<a></a>`},
		{`<span style=>`, `<span>`},
		{`<button onclick=>`, `<button>`},

		// whitespace
		{`cats  and 	dogs `, `cats and dogs`},
		{` <div> <i> test </i> <b> test </b> </div> `, `<div><i>test</i> <b>test</b></div>`},
		{`<strong>x </strong>y`, `<strong>x </strong>y`},
		{`<strong>x </strong> y`, `<strong>x</strong> y`},
		{"<strong>x </strong>\ny", "<strong>x</strong>\ny"},
		{`<p>x </p>y`, `<p>x</p>y`},
		{`x <p>y</p>`, `x<p>y`},
		{` <!doctype html> <!--comment--> <html> <body><p></p></body></html>`, `<!doctype html><p>`}, // spaces before html and at the start of html are dropped
		{`<p>x<br> y`, `<p>x<br>y`},
		{`<p>x </b> <b> y`, `<p>x</b> <b>y`},
		{`a <code>code</code> b`, `a <code>code</code> b`},
		{`a <code></code> b`, `a <code></code>b`},
		{`a <script>script</script> b`, `a <script>script</script>b`},

		// from HTML Minifier
		{`<DIV TITLE="blah">boo</DIV>`, `<div title=blah>boo</div>`},
		{"<p title\n\n\t  =\n     \"bar\">foo</p>", `<p title=bar>foo`},
		{`<p class=" foo      ">foo bar baz</p>`, `<p class=foo>foo bar baz`},
		{`<input maxlength="     5 ">`, `<input maxlength=5>`},
		{`<input type="text">`, `<input>`},
		{`<form method="get">`, `<form>`},
		{`<script language="Javascript">alert(1)</script>`, `<script>alert(1)</script>`},
		{`<script></script>`, ``},
		{`<p onclick=" JavaScript: x">x</p>`, `<p onclick=" x">x`},
		{`<span Selected="selected"></span>`, `<span selected></span>`},
		{`<table><thead><tr><th>foo</th><th>bar</th></tr></thead><tfoot><tr><th>baz</th><th>qux</th></tr></tfoot><tbody><tr><td>boo</td><td>moo</td></tr></tbody></table>`,
			`<table><thead><tr><th>foo<th>bar<tfoot><tr><th>baz<th>qux<tbody><tr><td>boo<td>moo</table>`},
		{`<select><option>foo</option><option>bar</option></select>`, `<select><option>foo<option>bar</select>`},
		{`<meta name="keywords" content="A, B">`, `<meta name=keywords content=A,B>`},
		{`<script type="text/html"><![CDATA[ <img id="x"> ]]></script>`, `<script type=text/html><img id=x></script>`},
		{`<iframe><html> <p> x </p> </html></iframe>`, `<iframe><p>x</iframe>`},
		{`<math> &int;_a_^b^{f(x)<over>1+x} dx </math>`, `<math> &int;_a_^b^{f(x)<over>1+x} dx </math>`},
		{`<script language="x" charset="x" src="y"></script>`, `<script src=y></script>`},
		{`<style media="all">x</style>`, `<style>x</style>`},
		{`<a id="abc" name="abc">y</a>`, `<a id=abc>y</a>`},
		{`<a id="" value="">y</a>`, `<a value>y</a>`},

		// go-fuzz
		{`<meta e t n content=ful><a b`, `<meta e t n content=ful><a b>`},
		{`<img alt=a'b="">`, `<img alt='a&#39;b=""'>`},
	}

	m := minify.New()
	m.AddFunc("text/html", Minify)
	m.AddFunc("text/css", func(_ *minify.M, w io.Writer, r io.Reader, _ map[string]string) error {
		_, err := io.Copy(w, r)
		return err
	})
	m.AddFunc("text/javascript", func(_ *minify.M, w io.Writer, r io.Reader, _ map[string]string) error {
		_, err := io.Copy(w, r)
		return err
	})
	for _, tt := range htmlTests {
		r := bytes.NewBufferString(tt.html)
		w := &bytes.Buffer{}
		assert.Nil(t, Minify(m, w, r, nil), "Minify must not return error in "+tt.html)
		assert.Equal(t, tt.expected, w.String(), "Minify must give expected result in "+tt.html)
	}
}

func TestHTMLURL(t *testing.T) {
	var htmlTests = []struct {
		html     string
		expected string
	}{
		{`<a href=http://example.com/>link</a>`, `<a href=//example.com/>link</a>`},
		{`<a href=https://example.com/>link</a>`, `<a href=https://example.com/>link</a>`},
		{`<a href="   http://example.com  ">x</a>`, `<a href=//example.com>x</a>`},
		{`<link rel="stylesheet" type="text/css" href="http://example.com">`, `<link rel=stylesheet href=//example.com>`},
		{`<!doctype html> <html xmlns="http://www.w3.org/1999/xhtml" xml:lang="en"> <head profile="http://dublincore.org/documents/dcq-html/"> <!-- Barlesque 2.75.0 --> <meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />`,
			`<!doctype html><html xmlns=//www.w3.org/1999/xhtml xml:lang=en><head profile=//dublincore.org/documents/dcq-html/><meta charset=utf-8>`},
		{`<svg xmlns="http://www.w3.org/2000/svg"><path d="x"/></svg>`, `<svg xmlns=//www.w3.org/2000/svg><path d="x"/></svg>`},
	}

	m := minify.New()
	m.AddFunc("text/html", Minify)
	for _, tt := range htmlTests {
		r := bytes.NewBufferString(tt.html)
		w := &bytes.Buffer{}
		m.URL, _ = url.Parse("http://example.com/")
		assert.Nil(t, Minify(m, w, r, nil), "Minify must not return error in "+tt.html)
		assert.Equal(t, tt.expected, w.String(), "Minify must give expected result in "+tt.html)
	}
}

func TestSpecialTagClosing(t *testing.T) {
	m := minify.New()
	m.AddFunc("text/html", Minify)
	m.AddFunc("text/css", func(_ *minify.M, w io.Writer, r io.Reader, _ map[string]string) error {
		b, err := ioutil.ReadAll(r)
		assert.Nil(t, err, "ioutil.ReadAll must not return error")
		assert.Equal(t, "</script>", string(b))
		_, err = w.Write(b)
		return err
	})

	r := bytes.NewBufferString(`<style></script></style>`)
	w := &bytes.Buffer{}
	assert.Nil(t, Minify(m, w, r, nil), "Minify must not return error in <style></script></style>")
	assert.Equal(t, `<style></script></style>`, w.String(), "Minify must give expected result in <style></script></style>")
}

func TestReaderErrors(t *testing.T) {
	m := minify.New()
	r := test.NewErrorReader(0)
	w := &bytes.Buffer{}
	assert.Equal(t, test.ErrPlain, Minify(m, w, r, nil), "Minify must return error at first read")
}

func TestWriterErrors(t *testing.T) {
	var errorTests = []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 12, 14}

	m := minify.New()
	for _, n := range errorTests {
		// writes:                  0         1   2     34   56  78  9       0    12   3      4
		r := bytes.NewBufferString(`<!doctype>text<style attr=val>css</style><code>code</code><!--comment-->`)
		w := test.NewErrorWriter(n)
		assert.Equal(t, test.ErrPlain, Minify(m, w, r, nil), "Minify must return error at write "+strconv.FormatInt(int64(n), 10))
	}
}

////////////////////////////////////////////////////////////////

func ExampleMinify() {
	m := minify.New()
	m.AddFunc("text/html", Minify)
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/javascript", js.Minify)
	m.AddFunc("image/svg+xml", svg.Minify)
	m.AddFuncRegexp(regexp.MustCompile("[/+]json$"), json.Minify)
	m.AddFuncRegexp(regexp.MustCompile("[/+]xml$"), xml.Minify)

	// set URL to minify link locations too
	m.URL, _ = url.Parse("https://www.example.com/")
	if err := m.Minify("text/html", os.Stdout, os.Stdin); err != nil {
		panic(err)
	}
}

func ExampleMinify_options() {
	m := minify.New()
	m.Add("text/html", &Minifier{
		KeepDefaultAttrVals: true,
		KeepWhitespace:      true,
	})

	if err := m.Minify("text/html", os.Stdout, os.Stdin); err != nil {
		panic(err)
	}
}

func ExampleMinify_reader() {
	b := bytes.NewReader([]byte("<html><body><h1>Example</h1></body></html>"))

	m := minify.New()
	m.Add("text/html", &Minifier{})

	r := m.Reader("text/html", b)
	if _, err := io.Copy(os.Stdout, r); err != nil {
		panic(err)
	}
	// Output: <h1>Example</h1>
}

func ExampleMinify_writer() {
	m := minify.New()
	m.Add("text/html", &Minifier{})

	w := m.Writer("text/html", os.Stdout)
	w.Write([]byte("<html><body><h1>Example</h1></body></html>"))
	w.Close()
	// Output: <h1>Example</h1>
}
