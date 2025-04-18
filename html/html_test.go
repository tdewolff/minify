package html

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/js"
	"github.com/tdewolff/minify/v2/json"
	"github.com/tdewolff/minify/v2/svg"
	"github.com/tdewolff/minify/v2/xml"
	"github.com/tdewolff/test"
)

func TestHTML(t *testing.T) {
	htmlTests := []struct {
		html     string
		expected string
	}{
		{`html`, `html`},
		//{"<title>title</title> <body>", `<title>title</title>`},
		{`<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML+RDFa 1.0//EN" "http://www.w3.org/MarkUp/DTD/xhtml-rdfa-1.dtd">`, `<!doctype html>`},
		{`<!-- comment -->`, ``},
		{`<!--# SSI Tag -->`, ``},
		{`<style><!--\ncss\n--></style>`, `<style><!--\ncss\n--></style>`},
		{`<style>&</style>`, `<style>&</style>`},
		{`<html><head></head><body>x</body></html>`, `x`},
		{"<html><head><title> a </title>\n<link></head></html>", `<title>a</title><link>`},
		{`<meta http-equiv="content-type" content="text/html; charset=utf-8">`, `<meta charset=utf-8>`},
		{`<meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />`, `<meta charset=utf-8>`},
		{`<meta http-equiv=" content-type " content=" text/html;  charset=utf-8 ">`, `<meta charset=utf-8>`},
		{`<meta http-equiv="Content-Security-Policy" content="default-src 'self'; img-src https://*; child-src 'none';">`, `<meta http-equiv=Content-Security-Policy content="default-src 'self'; img-src https://*; child-src 'none';">`},
		{`<meta name="keywords" content="a, b">`, `<meta name=keywords content="a,b">`},
		//{`<meta name="keywords" content=" a,  b ">`, `<meta name=keywords content="a,b">`}, // TODO
		{`<meta name="viewport" content="width = 996" />`, `<meta name=viewport content="width=996">`},
		{`<span attr="test"></span>`, `<span attr=test></span>`},
		{`<span attr='test&apos;test'></span>`, `<span attr="test'test"></span>`},
		{`<span attr="test&quot;test"></span>`, `<span attr='test"test'></span>`},
		{`<span attr='test""&apos;&amp;test'></span>`, `<span attr='test""&#39;&amp;test'></span>`},
		{`<span attr="test/test"></span>`, `<span attr=test/test></span>`},
		{`<span attr="test/"></span>`, `<span attr=test/></span>`},
		{`<span>&amp;</span>`, `<span>&</span>`},
		{`<span>&lt;</span>`, `<span>&lt;</span>`},
		{`<span>&gt;</span>`, `<span>></span>`},
		{`<code>&quot;&rsquor;</code>`, `<code>"&#8217;</code>`},
		{`<small>&#160;</small>`, `<small>&#160;</small>`},
		{`<span name="&lt;&apos;">a</span>`, `<span name="<'">a</span>`},
		{`<span method=GET></span>`, `<span></span>`},
		{`<span onload="javascript:x;"></span>`, `<span onload=x;></span>`},
		{`<span selected="selected"></span>`, `<span selected></span>`},
		{`<noscript><html><img id="x"></noscript>`, `<noscript><img id=x></noscript>`},
		{`<body id="main"></body>`, `<body id=main>`},
		{`<link href="data:text/plain, data">`, `<link href=data:,%20data>`},
		{`<svg width="100" height="100"><circle cx="50" cy="50" r="40" stroke="green" stroke-width="4" fill="yellow" /></svg>`, `<svg width="100" height="100"><circle cx="50" cy="50" r="40" stroke="green" stroke-width="4" fill="yellow" /></svg>`},
		{`<span ></span>`, `<span></span>`},
		{`</span >`, `</span>`},
		{`<meta name=viewport content="width=0.1, initial-scale=1.0 , maximum-scale=1000">`, `<meta name=viewport content="width=.1,initial-scale=1,maximum-scale=1e3">`},
		{`<br/>`, `<br>`},
		{`<input type="radio" value="">`, `<input type=radio value>`},
		{`<input type="radio" value="on">`, `<input type=radio>`},
		{`<input type="text" value="">`, `<input>`},
		{`<a rel="noopener">`, `<a rel=noopener>`},
		{`<a rel=" noopener  external ">`, `<a rel="noopener external">`},
		{`<input accept="image/png, image/jpeg">`, `<input accept=image/png,image/jpeg>`},
		{`<p></p></label>`, `<p></p></label>`}, // experimentally found

		// increase coverage
		{`<script style="css">js</script>`, `<script style=css>js</script>`},
		{`<script type="text/javascript">js</script>`, `<script>js</script>`},
		{`<script type="application/javascript">js</script>`, `<script>js</script>`},
		{`<meta http-equiv="content-type" content="text/plain, text/html">`, `<meta http-equiv=content-type content="text/plain,text/html">`},
		{`<meta property="rdfa" content="data">`, `<meta property="rdfa" content="data">`},
		{`<span attr=""></span>`, `<span attr></span>`},
		{`<code>x</code>`, `<code>x</code>`},
		{`<p></p><p></p>`, `<p><p>`},
		{`<ul><li></li> <li></li></ul>`, `<ul><li><li></ul>`},
		{`<p></p><a></a>`, `<p></p><a></a>`},
		{`<p></p>x<a></a>`, `<p></p>x<a></a>`},
		{`<span style=>`, `<span>`},
		{`<button onclick=>`, `<button>`},

		// whitespace
		{`cats  and 	dogs `, `cats and dogs`},
		{` <div> <i> test </i> <b> test </b> </div> `, `<div><i>test </i><b>test</b></div>`},
		{`<strong>x </strong>y`, `<strong>x </strong>y`},
		{`<strong>x </strong> y`, `<strong>x </strong>y`},
		{"<strong>x </strong>\ny", "<strong>x </strong>y"},
		{`<p>x </p>y`, `<p>x</p>y`},
		{`x <p>y</p>`, `x<p>y`},
		{` <!doctype html> <!--comment--> <html> <body><p></p></body></html> `, `<!doctype html><p>`}, // spaces before html and at the start of html are dropped
		{`<p>x<br> y`, `<p>x<br>y`},
		{`<p>x </b> <b> y`, `<p>x </b><b>y`},
		{`a <code></code> b`, `a <code></code> b`},
		{`a <code>code</code> b`, `a <code>code</code> b`},
		{`a <code> code </code> b`, `a <code>code </code>b`},
		{`a <script>script</script> b`, `a <script>script</script>b`},
		{"text\n<!--comment-->\ntext", "text\ntext"},
		{"abc\n</body>\ndef", "abc\ndef"},
		{"<x>\n<!--y-->\n</x>", "<x></x>"},
		{`<p class="  name  ">`, `<p class=name>`},
		{`<p class="  name  other  ">`, `<p class="name other">`},
		{`<p onclick="  javascript:lala  ">`, `<p onclick=lala>`},
		{`<p url="http://test">`, `<p url=http://test>`},
		{`a<span> <img> </span>b`, `a<span> <img> </span>b`},
		{`a <span> <img> </span>b`, `a <span><img> </span>b`},
		{`a<picture> <img> </picture>b`, `a<picture> <img> </picture>b`},
		{`a <picture> <img> </picture>b`, `a <picture><img> </picture>b`},
		{`<input placeholder=" a " value=" b ">`, `<input placeholder=" a " value=" b ">`},
		{`a <strike> b </strike> c`, `a <strike>b </strike>c`},
		{`a <svg>`, `a <svg>`},
		{"<h5>\n    Lorem & ipsum\n    dolor sid amet\n</h5>", "<h5>Lorem & ipsum\ndolor sid amet</h5>"},
		{`cats  <br>  dogs `, `cats<br>dogs`},

		// from HTML Minifier
		{`<DIV TITLE="blah">boo</DIV>`, `<div title=blah>boo</div>`},
		{"<p title\n\n\t  =\n     \"bar\">foo</p>", `<p title=bar>foo`},
		{`<p class=" foo      ">foo bar baz</p>`, `<p class=foo>foo bar baz`},
		{`<input maxlength="     5 ">`, `<input maxlength=5>`},
		{`<input type="text">`, `<input>`},
		{`<form method="get">`, `<form>`},
		{`<script></script>`, ``},
		{`<p onclick=" JavaScript: x">x</p>`, `<p onclick=" x">x`},
		{`<span Selected="selected"></span>`, `<span selected></span>`},
		{`<table><thead><tr><th>foo</th><th>bar</th></tr></thead><tfoot><tr><th>baz</th><th>qux</th></tr></tfoot><tbody><tr><td>boo</td><td>moo</td></tr></tbody></table>`,
			`<table><thead><tr><th>foo<th>bar<tfoot><tr><th>baz<th>qux<tbody><tr><td>boo<td>moo</table>`},
		{`<select><option>foo</option><option>bar</option></select>`, `<select><option>foo<option>bar</select>`},
		{`<select><optgroup><option>foo</option></optgroup><optgroup><option>bar</option></optgroup><option>zoo</option></select>`, `<select><optgroup><option>foo<optgroup><option>bar</optgroup><option>zoo</select>`},
		{`<select>text<option>foo</option>text<optgroup>text<option>bar</option>text</optgroup>text</select>`, `<select><option>foo<optgroup><option>bar</select>`},
		{`<meta name="keywords" content="A, B">`, `<meta name=keywords content="A,B">`},
		{`<iframe><html> <p> x </p> </html></iframe>`, `<iframe><p>x</iframe>`},
		{`<math> &int;_a_^b^{f(x)<over>1+x} dx </math>`, `<math> &int;_a_^b^{f(x)<over>1+x} dx </math>`},
		{`<script charset="x" src="y"></script>`, `<script src=y></script>`},
		{`<style media="all">x</style>`, `<style>x</style>`},
		{`<a id="abc" name="abc">y</a>`, `<a id=abc>y</a>`},
		{`<a name="abc" id="abc">y</a>`, `<a id=abc>y</a>`},
		{`<a id="" value="">y</a>`, `<a value>y</a>`},

		// from Kangax html-minfier
		{`<span style="font-family:&quot;Helvetica Neue&quot;,&quot;Helvetica&quot;,Helvetica,Arial,sans-serif">text</span>`, `<span style='font-family:"Helvetica Neue","Helvetica",Helvetica,Arial,sans-serif'>text</span>`},

		// go-fuzz
		{`<meta e t n content=ful><a b`, `<meta e t n content=ful><a b>`},
		{`<img alt=a'b="">`, `<img alt='a&#39;b=""'>`},
		{`</b`, `</b`},
		{`<title></`, `<title></`},
		{`<svg <`, `<svg <`},
		{`<svg "`, `<svg "`},
		{`<svg></`, `<svg></`},
		{`<script><!--<`, `<script><!--<`},

		// bugs
		{`<svg id="1"></svg>`, `<svg id="1"></svg>`},                                                                                     // #67
		{`<pre> <x> a  b </x> </pre>`, `<pre> <x> a  b </x> </pre>`},                                                                     // #82
		{`text <progress></progress> text`, `text <progress></progress> text`},                                                           // #89
		{`text <img> text`, `text <img> text`},                                                                                           // #89
		{`<p>text</p><br>text`, `<p>text</p><br>text`},                                                                                   // #122
		{`<amp-analytics type=adobeanalytics_nativeConfig>`, `<amp-analytics type=adobeanalytics_nativeConfig>`},                         // #270
		{`<input title=""><input lang="">`, `<input title><input lang>`},                                                                 // #331
		{`<span translate="no">`, `<span translate=no>`},                                                                                 // #352
		{`<i class="fas"> </i> Text`, `<i class=fas></i>Text`},                                                                           // #390
		{`<i class="fas"></i> Text`, `<i class=fas></i> Text`},                                                                           // #390
		{`a <span></span> b`, `a <span></span> b`},                                                                                       // #427
		{`<canvas><p>test</p></canvas>`, `<canvas><p>test</p></canvas>`},                                                                 // #440
		{`<source type='video/mp4; codecs="av01.0.05M.08"'>`, `<source type='video/mp4;codecs="av01.0.05M.08"'>`},                        // #545
		{`<html lang="mn-Cyrl-MN">`, `<html lang=mn-Cyrl-MN>`},                                                                           // #579
		{`<script src=a></script> <script src=b></script> <span></span>`, `<script src=a></script><script src=b></script><span></span>`}, // #592
		{`a <math>b</math> c`, `a <math>b</math> c`},                                                                                     // #611
		{`a <svg>b</svg> c`, `a <svg>b</svg> c`},                                                                                         // #611
	}

	m := minify.New()
	m.AddFunc("text/html", Minify)
	m.AddFunc("text/css", func(_ *minify.M, w io.Writer, r io.Reader, _ map[string]string) error {
		_, err := io.Copy(w, r)
		return err
	})
	m.AddFunc("application/javascript", func(_ *minify.M, w io.Writer, r io.Reader, _ map[string]string) error {
		_, err := io.Copy(w, r)
		return err
	})
	for _, tt := range htmlTests {
		t.Run(tt.html, func(t *testing.T) {
			r := bytes.NewBufferString(tt.html)
			w := &bytes.Buffer{}
			err := Minify(m, w, r, nil)
			test.Minify(t, tt.html, err, w.String(), tt.expected)
		})
	}
}

func TestHTMLCSSJS(t *testing.T) {
	htmlTests := []struct {
		html     string
		expected string
	}{
		// bugs
		{`<script>var a="<\/script>"</script>`, `<script>var a="<\/script>"</script>`},                                                          // #355
		{`<script>var a="</scr"+"ipt>"</script>`, `<script>var a="<\/script>"</script>`},                                                        // #355
		{`<div style="font-family: Arial, &#39;sans-serif&#39;; font-size: 22px;">`, `<div style=font-family:Arial,sans-serif;font-size:22px>`}, // #272
		{`<style amp-boilerplate>body{-webkit-animation:-amp-start 8s    steps(1,end) 0s 1 normal both;}</style>`, `<style amp-boilerplate>body{-webkit-animation:-amp-start 8s steps(1,end) 0s 1 normal both;}</style>`},
		{`<button onclick="return false">`, `<button onclick=return!1>`},                                                                                                                      // #631
		{`<html><svg id="foo" viewBox="0 0 1 1" xmlns="http://www.w3.org/2000/svg"><rect width="1" height="1"/></svg>`, `<svg id="foo" viewBox="0 0 1 1"><rect width="1" height="1"/></svg>`}, // #704
	}

	m := minify.New()
	m.AddFunc("text/html", Minify)
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("application/javascript", js.Minify)
	m.AddFunc("image/svg+xml", svg.Minify)
	for _, tt := range htmlTests {
		t.Run(tt.html, func(t *testing.T) {
			r := bytes.NewBufferString(tt.html)
			w := &bytes.Buffer{}
			err := Minify(m, w, r, nil)
			test.Minify(t, tt.html, err, w.String(), tt.expected)
		})
	}
}

func TestHTMLKeepEndTags(t *testing.T) {
	htmlTests := []struct {
		html     string
		expected string
	}{
		{`<p></p><p></p>`, `<p></p><p></p>`},
		{`<ul><li></li><li></li></ul>`, `<ul><li></li><li></li></ul>`},
	}

	m := minify.New()
	htmlMinifier := &Minifier{KeepEndTags: true}
	for _, tt := range htmlTests {
		t.Run(tt.html, func(t *testing.T) {
			r := bytes.NewBufferString(tt.html)
			w := &bytes.Buffer{}
			err := htmlMinifier.Minify(m, w, r, nil)
			test.Minify(t, tt.html, err, w.String(), tt.expected)
		})
	}
}

func TestHTMLKeepSpecialComments(t *testing.T) {
	htmlTests := []struct {
		html     string
		expected string
	}{
		{`<!--[if IE 6]> <b> </b> <![endif]-->`, `<!--[if IE 6]><b></b><![endif]-->`},
		{`<![if IE 6]> <b> </b> <![endif]>`, `<![if IE 6]><b></b><![endif]>`},
		{`<!--[if IE 6]--> <b> </b> <!--[endif]-->`, `<!--[if IE 6]--><b></b><!--[endif]-->`},
		{`<!--[if !mso]><!--> <b> </b> <!--<![endif]-->`, `<!--[if !mso]><!--><b></b><!--<![endif]-->`},
		{`<!--[if gt IE 6]><!--> <b> </b> <![endif]-->`, `<!--[if gt IE 6]><!--><b></b><![endif]-->`},
		{`<!--[if IE]foo<![endif]-->`, `<!--[if IE]foo<![endif]-->`}, // #596
		{`<!--# SSI-->`, `<!--# SSI-->`},                             // #657
	}

	m := minify.New()
	htmlMinifier := &Minifier{KeepSpecialComments: true}
	for _, tt := range htmlTests {
		t.Run(tt.html, func(t *testing.T) {
			r := bytes.NewBufferString(tt.html)
			w := &bytes.Buffer{}
			err := htmlMinifier.Minify(m, w, r, nil)
			test.Minify(t, tt.html, err, w.String(), tt.expected)
		})
	}
}

func TestHTMLKeepWhitespace(t *testing.T) {
	htmlTests := []struct {
		html     string
		expected string
	}{
		{`cats  and 	dogs `, `cats and dogs`},
		{` <div> <i> test </i> <b> test </b> </div> `, `<div> <i> test </i> <b> test </b> </div>`},
		{`<strong>x </strong>y`, `<strong>x </strong>y`},
		{`<strong>x </strong> y`, `<strong>x </strong> y`},
		{"<strong>x </strong>\ny", "<strong>x </strong>\ny"},
		{`<p>x </p>y`, `<p>x </p>y`},
		{`x <p>y</p>`, `x <p>y`},
		{` <!doctype html> <!--comment--> <html> <body><p></p></body></html> `, `<!doctype html><p>`}, // spaces before html and at the start of html are dropped
		{`<p>x<br> y`, `<p>x<br> y`},
		{`<p>x </b> <b> y`, `<p>x </b> <b> y`},
		{`a <code>code</code> b`, `a <code>code</code> b`},
		{`a <code></code> b`, `a <code></code> b`},
		{`a <script>script</script> b`, `a <script>script</script> b`},
		{"text\n<!--comment-->\ntext", "text\ntext"},
		{"text\n<!--comment-->text<!--comment--> text", "text\ntext text"},
		{"abc\n</body>\ndef", "abc\ndef"},
		{"<x>\n<!--y-->\n</x>", "<x>\n</x>"},
		{"<style>lala{color:red}</style>", "<style>lala{color:red}</style>"},
		{"<li>one\n</li>\n<li>two", "<li>one\n<li>two"}, // #442
	}

	m := minify.New()
	htmlMinifier := &Minifier{KeepWhitespace: true}
	for _, tt := range htmlTests {
		t.Run(tt.html, func(t *testing.T) {
			r := bytes.NewBufferString(tt.html)
			w := &bytes.Buffer{}
			err := htmlMinifier.Minify(m, w, r, nil)
			test.Minify(t, tt.html, err, w.String(), tt.expected)
		})
	}
}

func TestHTMLKeepQuotes(t *testing.T) {
	htmlTests := []struct {
		html     string
		expected string
	}{
		{`<p attr="test">`, `<p attr="test">`},
		{`<p attr='test'>`, `<p attr='test'>`},
		{`<p attr=test>`, `<p attr=test>`},
		{`<meta name='viewport' content='width=device-width, initial-scale=1'>`, `<meta name='viewport' content='width=device-width,initial-scale=1'>`},
	}

	m := minify.New()
	htmlMinifier := &Minifier{KeepQuotes: true}
	for _, tt := range htmlTests {
		t.Run(tt.html, func(t *testing.T) {
			r := bytes.NewBufferString(tt.html)
			w := &bytes.Buffer{}
			err := htmlMinifier.Minify(m, w, r, nil)
			test.Minify(t, tt.html, err, w.String(), tt.expected)
		})
	}
}

func TestHTMLURL(t *testing.T) {
	htmlTests := []struct {
		url      string
		html     string
		expected string
	}{
		{`http://example.com/`, `<a href=http://example.com/>link</a>`, `<a href=//example.com/>link</a>`},
		{`https://example.com/`, `<a href=http://example.com/>link</a>`, `<a href=http://example.com/>link</a>`},
		{`http://example.com/`, `<a href=https://example.com/>link</a>`, `<a href=https://example.com/>link</a>`},
		{`https://example.com/`, `<a href=https://example.com/>link</a>`, `<a href=//example.com/>link</a>`},
		{`http://example.com/`, `<a href="   http://example.com  ">x</a>`, `<a href=//example.com>x</a>`},
		{`http://example.com/`, `<link rel="stylesheet" type="text/css" href="http://example.com">`, `<link rel=stylesheet href=//example.com>`},
		{`http://example.com/`, `<!doctype html> <html xmlns="http://www.w3.org/1999/xhtml" xml:lang="en"> <head profile="http://dublincore.org/documents/dcq-html/"> <!-- Barlesque 2.75.0 --> <meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />`,
			`<!doctype html><html xmlns=//www.w3.org/1999/xhtml xml:lang=en><head profile=//dublincore.org/documents/dcq-html/><meta charset=utf-8>`},
		{`http://example.com/`, `<html xmlns="http://www.w3.org/1999/xhtml"></html>`, `<html xmlns=//www.w3.org/1999/xhtml>`},
		{`https://example.com/`, `<html xmlns="http://www.w3.org/1999/xhtml"></html>`, `<html xmlns=http://www.w3.org/1999/xhtml>`},
		{`http://example.com/`, `<html xmlns="https://www.w3.org/1999/xhtml"></html>`, `<html xmlns=https://www.w3.org/1999/xhtml>`},
		{`https://example.com/`, `<html xmlns="https://www.w3.org/1999/xhtml"></html>`, `<html xmlns=//www.w3.org/1999/xhtml>`},
	}

	m := minify.New()
	m.AddFunc("text/html", Minify)
	for _, tt := range htmlTests {
		t.Run(tt.url, func(t *testing.T) {
			r := bytes.NewBufferString(tt.html)
			w := &bytes.Buffer{}
			m.URL, _ = url.Parse(tt.url)
			err := Minify(m, w, r, nil)
			test.Minify(t, tt.html, err, w.String(), tt.expected)
		})
	}
}

func TestHTMLGoTemplates(t *testing.T) {
	htmlTests := []struct {
		html     string
		expected string
	}{
		{`<a href={{ .Link }} />`, `<a href={{ .Link }}>`},
		{`<input type="file" accept="{{ .Accept }}, image/jpeg">`, `<input type=file accept="{{ .Accept }}, image/jpeg">`},
		{`<option value="0" {{ if eq .Type 0 }}selected{{ end }}>Foo</option>`, `<option value=0 {{ if eq .Type 0 }}selected{{ end }}>Foo`},
		{`<style>a { color: {{.Color}} }</style>`, `<style>a { color: {{.Color}} }</style>`},
		{`<div style=" color: {{.Color}} ">`, `<div style=" color: {{.Color}} ">`},
		{`<script>alert( {{.Alert}} )</script>`, `<script>alert( {{.Alert}} )</script>`},
		{`<button onclick=" alert( {{.Alert}} ) ">`, `<button onclick=" alert( {{.Alert}} ) ">`},
		{`<select>{{ range . }}<option>{{ . }}{{ end }}</select>`, `<select>{{ range . }}<option>{{ . }}{{ end }}</select>`},
		{`<p>Hello <code>{{""}}</code> there</p>`, `<p>Hello <code>{{""}}</code> there`},
		{`<select><option>Default</option>{{range $i, $lang := .Languages}}<option>{{$lang}}</option>{{end}}</select>`, `<select><option>Default{{range $i, $lang := .Languages}}<option>{{$lang}}{{end}}</select>`},
		{`<tr{{if .Deleted}} class="is-disabled"{{end}}>`, `<tr{{if .Deleted}} class="is-disabled"{{end}}>`},

		// whitespace
		{`{{ printf "  !  " }}`, `{{ printf "  !  " }}`},
		{`a<p>  {{a}}  </p>b`, `a<p>{{a}}</p>b`},
		{`a<span>  {{a}}  </span>b`, `a<span> {{a}} </span>b`},
		{`a <span>  {{a}}  </span> b`, `a <span>{{a}} </span>b`},
		{`  {{a}}  {{b}}  `, `{{a}} {{b}}`},
	}

	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("application/javascript", js.Minify)
	htmlMinifier := &Minifier{TemplateDelims: GoTemplateDelims}
	for _, tt := range htmlTests {
		t.Run(tt.html, func(t *testing.T) {
			r := bytes.NewBufferString(tt.html)
			w := &bytes.Buffer{}
			err := htmlMinifier.Minify(m, w, r, nil)
			test.Minify(t, tt.html, err, w.String(), tt.expected)
		})
	}
}

func TestHTMLPHPTemplates(t *testing.T) {
	htmlTests := []struct {
		html     string
		expected string
	}{
		{"\n<div class=\"foo\">\n<?php\necho \"hello\";\n?>\n</div>\n", "<div class=foo><?php\necho \"hello\";\n?></div>"},
	}

	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	htmlMinifier := &Minifier{TemplateDelims: PHPTemplateDelims}
	for _, tt := range htmlTests {
		t.Run(tt.html, func(t *testing.T) {
			r := bytes.NewBufferString(tt.html)
			w := &bytes.Buffer{}
			err := htmlMinifier.Minify(m, w, r, nil)
			test.Minify(t, tt.html, err, w.String(), tt.expected)
		})
	}
}

func TestSpecialTagClosing(t *testing.T) {
	m := minify.New()
	m.AddFunc("text/html", Minify)
	m.AddFunc("text/css", func(_ *minify.M, w io.Writer, r io.Reader, _ map[string]string) error {
		b, err := io.ReadAll(r)
		test.Error(t, err)
		test.String(t, string(b), "</script>")
		_, err = w.Write(b)
		return err
	})

	html := `<style></script></style>`
	r := bytes.NewBufferString(html)
	w := &bytes.Buffer{}
	err := Minify(m, w, r, nil)
	test.Minify(t, html, err, w.String(), html)
}

func TestReaderErrors(t *testing.T) {
	r := test.NewErrorReader(0)
	w := &bytes.Buffer{}
	m := minify.New()
	err := Minify(m, w, r, nil)
	test.T(t, err, test.ErrPlain, "return error at first read")
}

func TestWriterErrors(t *testing.T) {
	errorTests := []struct {
		html string
		n    []int
	}{
		{`<!doctype>`, []int{0}},
		{`text`, []int{0}},
		{`<foo attr=val>`, []int{0, 1, 2, 3, 4, 5}},
		{`</foo>`, []int{0}},
		{`<style>x</style>`, []int{2}},
		{`<textarea>x</textarea>`, []int{2}},
		{`<code>x</code>`, []int{2}},
		{`<pre>x</pre>`, []int{2}},
		{`<svg>x</svg>`, []int{0}},
		{`<math>x</math>`, []int{0}},
		{`<!--[if IE 6]> text <![endif]-->`, []int{0, 1, 2}},
		{`<![if IE 6]> text <![endif]>`, []int{0}},
	}

	m := minify.New()
	m.Add("text/html", &Minifier{
		KeepSpecialComments: true,
	})

	for _, tt := range errorTests {
		for _, n := range tt.n {
			t.Run(fmt.Sprint(tt.html, " ", tt.n), func(t *testing.T) {
				r := bytes.NewBufferString(tt.html)
				w := test.NewErrorWriter(n)
				err := m.Minify("text/html", w, r)
				test.T(t, err, test.ErrPlain)
			})
		}
	}
}

func TestMinifyErrors(t *testing.T) {
	errorTests := []struct {
		html string
		err  error
	}{
		{`<style>abc</style>`, test.ErrPlain},
		{`<p style="abc"/>`, test.ErrPlain},
		{`<p onclick="abc"/>`, test.ErrPlain},
		{`<svg></svg>`, test.ErrPlain},
		{`<math></math>`, test.ErrPlain},
	}

	m := minify.New()
	m.AddFunc("text/css", func(_ *minify.M, w io.Writer, r io.Reader, _ map[string]string) error {
		return test.ErrPlain
	})
	m.AddFunc("application/javascript", func(_ *minify.M, w io.Writer, r io.Reader, _ map[string]string) error {
		return test.ErrPlain
	})
	m.AddFunc("image/svg+xml", func(_ *minify.M, w io.Writer, r io.Reader, _ map[string]string) error {
		return test.ErrPlain
	})
	m.AddFunc("application/mathml+xml", func(_ *minify.M, w io.Writer, r io.Reader, _ map[string]string) error {
		return test.ErrPlain
	})
	for _, tt := range errorTests {
		t.Run(tt.html, func(t *testing.T) {
			r := bytes.NewBufferString(tt.html)
			w := &bytes.Buffer{}
			err := Minify(m, w, r, nil)
			test.T(t, err, tt.err)
		})
	}
}

func TestMinifyErrorPropagation(t *testing.T) {
	errorTests := []struct {
		html string
		err  string
	}{
		{"line\n<script><</script>", "unexpected < in expression on line 2 and column 9"},
		{"line\n<script> <</script>", "unexpected < in expression on line 2 and column 10"},
		{"<div onclick='<'>", "unexpected < in expression on line 1 and column 15"},
	}

	m := minify.New()
	m.AddFunc("application/javascript", js.Minify)
	for _, tt := range errorTests {
		t.Run(tt.html, func(t *testing.T) {
			r := bytes.NewBufferString(tt.html)
			w := &bytes.Buffer{}
			err := Minify(m, w, r, nil)
			if err == nil {
				test.Fail(t)
			}
			errMessage := err.Error()
			if i := strings.IndexByte(errMessage, '\n'); i != -1 {
				errMessage = errMessage[:i]
			}
			test.String(t, errMessage, tt.err)
		})
	}
}

////////////////////////////////////////////////////////////////

func ExampleMinify() {
	m := minify.New()
	m.AddFunc("text/html", Minify)
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("image/svg+xml", svg.Minify)
	m.AddFuncRegexp(regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$"), js.Minify)
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
