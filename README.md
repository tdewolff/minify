# Minify [![Build Status](https://travis-ci.org/tdewolff/minify.svg?branch=master)](https://travis-ci.org/tdewolff/minify) [![GoDoc](http://godoc.org/github.com/tdewolff/minify?status.svg)](http://godoc.org/github.com/tdewolff/minify) [![Join the chat at https://gitter.im/tdewolff/minify](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/tdewolff/minify?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

**I will be away for 5 months, starting May. v2 should be the preferred stable release to use. Master has some new changes for SVG that haven't yet endured the test of time, bug reports are appreciated.**

Minify is a minifier package written in [Go][1]. It has build-in HTML5, CSS3, JS, JSON, SVG and XML minifiers and provides an interface to implement any minifier. Minification is the process of removing bytes from a file (such as whitespace) without changing its output and therefore speeding up transmission over the internet. The implemented minifiers are high performance and streaming (which implies O(n)).

It associates minification functions with mimetypes, allowing embedded resources (like CSS or JS in HTML files) to be minified too. The user can add any mime-based implementation. Users can also implement a mimetype using an external command (like the ClosureCompiler, UglifyCSS, ...). It is possible to pass parameters through the mediatype to specify the charset for example.

[Demo](http://go.tacodewolff.nl/) · [CLI](https://github.com/tdewolff/minify/tree/master/cmd/minify) · [Download](https://dl.equinox.io/tdewolff/minify/stable)

#### Table of Contents

- [Minify](#minify---)
	- [Prologue](#prologue)
	- [Installation](#installation)
	- [API stability](#api-stability)
	- [Testing](#testing)
	- [HTML](#html--)
		- [Whitespace removal](#whitespace-removal)
	- [CSS](#css--)
	- [JS](#js--)
	- [JSON](#json--)
	- [SVG](#svg--)
	- [XML](#xml--)
	- [Usage](#usage)
		- [New](#new)
		- [From reader](#from-reader)
		- [From bytes](#from-bytes)
		- [From string](#from-string)
		- [Custom minifier](#custom-minifier)
		- [Mediatypes](#mediatypes)
	- [Examples](#examples)
		- [Common minifiers](#common-minifiers)
		- [Custom minifier](#custom-minifier-1)
		- [ResponseWriter](#responsewriter)
		- [Templates](#templates)
	- [License](#license)

#### Status

* CSS: **fully implemented**
* HTML: **fully implemented**
* JS: basic JSmin-like implementation
* JSON: **fully implemented**
* SVG: partially implemented; in development
* XML: **fully implemented**

## Prologue
Minifiers or bindings to minifiers exist in almost all programming languages. Some implementations are merely using several regular-expressions to trim whitespace and comments (even though regex for parsing HTML/XML is ill-advised, for a good read see [Regular Expressions: Now You Have Two Problems](http://blog.codinghorror.com/regular-expressions-now-you-have-two-problems/)). Some implementations are much more profound, such as the [YUI Compressor](http://yui.github.io/yuicompressor/) and [Google Closure Compiler](https://github.com/google/closure-compiler) for JS. As most existing implementations either use Java or JavaScript and don't focus on performance, they are pretty slow. And loading the whole file into memory is bad for really large files (or impossible for infinite streams).

This minifier proves to be that fast and extensive minifier that can handle HTML and any other filetype it may contain (CSS, JS, ...). It streams the input and output and can minify files concurrently.

## Installation
Run the following command

	go get github.com/tdewolff/minify

or add the following imports and run the project with `go get`
``` go
import (
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/js"
	"github.com/tdewolff/minify/json"
	"github.com/tdewolff/minify/svg"
	"github.com/tdewolff/minify/xml"
)
```

## API stability
There is no guarantee for absolute stability, but I take issues and bugs seriously and don't take API changes lightly. The library will be maintained in a compatible way unless vital bugs prevent me from doing so. There has been one API change after v1 which added options support and I took the opportunity to push through some more API clean up as well. There are no plans whatsoever for future API changes.

- minify-v1.0.0 depends on parse-v1.0.0
- minify-v1.1.0 depends on parse-v1.1.0
- minify-v2.0.0 depends on parse-v2.0.0
- minify-tip will always compile with my other packages on tip

The API differences between v1 and v2 are listed below. If `m := minify.New()` and `w` and `r` are your writer and reader respectfully, then **v1** &#8594; **v2**:
 - `minify.Bytes(m, ...)` &#8594; `m.Bytes(...)`
 - `minify.String(m, ...)` &#8594; `m.String(...)`
 - `html.Minify(m, "text/html", w, r)` &#8594; `html.Minify(m, w, r, nil)` also for `css`, `js`, ...
 - `css.Minify(m, "text/css;inline=1", w, r)` &#8594; `css.Minify(m, w, r, map[string]string{"inline":"1"})`

## Testing
For all subpackages and the imported `parse` and `buffer` packages, test coverage of 100% is pursued. Besides full coverage, the minifiers are [fuzz tested](https://github.com/tdewolff/fuzz) using [github.com/dvyukov/go-fuzz](http://www.github.com/dvyukov/go-fuzz), see [the wiki](https://github.com/tdewolff/minify/wiki) for the most important bugs found by fuzz testing. Furthermore am I working on adding visual testing to ensure that minification doesn't change anything visually. By using the WebKit browser to render the original and minified pages we can check whether any pixel is different.

These tests ensure that everything works as intended, the code does not crash (whatever the input) and that it doesn't change the final result visually. If you still encounter a bug, please report [here](https://github.com/tdewolff/minify/issues)!

## HTML

HTML (with JS and CSS) minification typically runs at about 40MB/s ~= 140GB/h, depending on the composition of the file.

Website | Original | Minified | Ratio | Time<sup>&#42;</sup>
------- | -------- | -------- | ----- | -----------------------
[Amazon](http://www.amazon.com/) | 463kB | **414kB** | 90% | 11ms
[BBC](http://www.bbc.com/) | 113kB | **96kB** | 85% | 3ms
[StackOverflow](http://stackoverflow.com/) | 201kB | **182kB** | 91% | 5ms
[Wikipedia](http://en.wikipedia.org/wiki/President_of_the_United_States) | 435kB | **410kB** | 94%<sup>&#42;&#42;</sup> | 10ms

<sup>&#42;</sup>These times are measured on my home computer which is an average development computer. The duration varies a lot but it's important to see it's in the 10ms range! The benchmark uses all the minifiers and excludes reading from and writing to the file from the measurement.

<sup>&#42;&#42;</sup>Is already somewhat minified, so this doesn't reflect the full potential of this minifier.

The HTML5 minifier uses these minifications:

- strip unnecessary whitespace and otherwise collapse it to one space (or newline if it originally contained a newline)
- strip superfluous quotes, or uses single/double quotes whichever requires fewer escapes
- strip default attribute values and attribute boolean values
- strip some empty attributes
- strip unrequired tags (`html`, `head`, `body`, ...)
- strip unrequired end tags (`tr`, `td`, `li`, ... and often `p`)
- strip default protocols (`http:`, `https:` and `javascript:`)
- strip comments (except conditional comments)
- shorten `doctype` and `meta` charset
- lowercase tags, attributes and some values to enhance gzip compression

Options:

- `KeepDefaultAttrVals` do not remove default attribute value such as `<script type="text/javascript">`
- `KeepWhitespace` do not remove whitespace between inline tags but still collapse multiple whitespace characters into one

After recent benchmarking and profiling it became really fast and minifies pages in the 10ms range, making it viable for on-the-fly minification.

However, be careful when doing on-the-fly minification. Minification typically trims off 10% and does this at worst around about 20MB/s. This means users have to download slower than 2MB/s to make on-the-fly minification worthwhile. This may or may not apply in your situation. Rather use caching!

### Whitespace removal
The whitespace removal mechanism collapses all sequences of whitespace (spaces, newlines, tabs) to a single space. If the sequence contained a newline or carriage return it will collapse into a newline character instead. It trims all text parts (in between tags) depending on whether it was preceded by a space from a previous piece of text and whether it is followed up by a block element or an inline element. In the former case we can omit spaces while for inline elements whitespace has significance.

Make sure your HTML doesn't depend on whitespace between `block` elements that have been changed to `inline` or `inline-block` elements using CSS. Your layout *should not* depend on those whitespaces as the minifier will remove them. An example is a menu consisting of multiple `<li>` that have `display:inline-block` applied and have whitespace in between them. It is bad practise to rely on whitespace for element positioning anyways!

## CSS

Minification typically runs at about 25MB/s ~= 90GB/h.

Library | Original | Minified | Ratio | Time<sup>&#42;</sup>
------- | -------- | -------- | ----- | -----------------------
[Bootstrap](http://getbootstrap.com/) | 134kB | **111kB** | 83% | 5ms
[Gumby](http://gumbyframework.com/) | 182kB | **167kB** | 90% | 7ms

<sup>&#42;</sup>The benchmark excludes the time reading from and writing to a file from the measurement.

The CSS minifier will only use safe minifications:

- remove comments and unnecessary whitespace
- remove trailing semicolons
- optimize `margin`, `padding` and `border-width` number of sides
- shorten numbers by removing unnecessary `+` and zeros and rewriting with/without exponent
- remove dimension and percentage for zero values
- remove quotes for URLs
- remove quotes for font families and make lowercase
- rewrite hex colors to/from color names, or to 3 digit hex
- rewrite `rgb(`, `rgba(`, `hsl(` and `hsla(` colors to hex or name
- replace `normal` and `bold` by numbers for `font-weight` and `font`
- replace `none` &#8594; `0` for `border`, `background` and `outline`
- lowercase all identifiers except classes, IDs and URLs to enhance gzip compression
- shorten MS alpha function
- rewrite data URIs with base64 or ASCII whichever is shorter
- calls minifier for data URI mediatypes, thus you can compress embedded SVG files if you have that minifier attached

It does purposely not use the following techniques:

- (partially) merge rulesets
- (partially) split rulesets
- collapse multiple declarations when main declaration is defined within a ruleset (don't put `font-weight` within an already existing `font`, too complex)
- remove overwritten properties in ruleset (this not always overwrites it, for example with `!important`)
- rewrite properties into one ruleset if possible (like `margin-top`, `margin-right`, `margin-bottom` and `margin-left` &#8594; `margin`)
- put nested ID selector at the front (`body > div#elem p` &#8594; `#elem p`)
- rewrite attribute selectors for IDs and classes (`div[id=a]` &#8594; `div#a`)
- put space after pseudo-selectors (IE6 is old, move on!)

It's great that so many other tools make comparison tables: [CSS Minifier Comparison](http://www.codenothing.com/benchmarks/css-compressor-3.0/full.html), [CSS minifiers comparison](http://www.phpied.com/css-minifiers-comparison/) and [CleanCSS tests](http://goalsmashers.github.io/css-minification-benchmark/). From the last link, this CSS minifier is almost without doubt the fastest and has near-perfect minification rates. It falls short with the purposely not implemented and often unsafe techniques, so that's fine.

## JS

The JS minifier is pretty basic. It removes comments, whitespace and line breaks whenever it can. It employs all the rules that [JSMin](http://www.crockford.com/javascript/jsmin.html) does too, but has additional improvements. For example the prefix-postfix bug is fixed.

Minification typically runs at about 45MB/s ~= 160GB/h. Common speeds of PHP and JS implementations are about 100-300kB/s (see [Uglify2](http://lisperator.net/uglifyjs/), [Adventures in PHP web asset minimization](https://www.happyassassin.net/2014/12/29/adventures-in-php-web-asset-minimization/)).

Library | Original | Minified | Ratio | Time<sup>&#42;</sup>
------- | -------- | -------- | ----- | -----------------------
[ACE](https://github.com/ajaxorg/ace-builds) | 630kB | **442kB** | 70% | 14ms
[jQuery](http://jquery.com/download/) | 242kB | **130kB** | 54% | 5ms
[jQuery UI](http://jqueryui.com/download/) | 459kB | **300kB** | 65% | 11ms
[Moment](http://momentjs.com/) | 97kB | **51kB** | 52% | 2ms

<sup>&#42;</sup>The benchmark excludes the time reading from and writing to a file from the measurement.

TODO:
- shorten local variables / function parameters names
- precise semicolon and newline omission

## JSON

Minification typically runs at about 95MB/s ~= 340GB/h. It shaves off about 15% of filesize for common indented JSON such as generated by [JSON Generator](http://www.json-generator.com/).

The JSON minifier only removes whitespace, which is the only thing that can be left out.

## SVG

Minification typically runs at about 15MB/s ~= 55GB/h. Performance improvement are due.

The SVG minifier uses these minifications:

- trim and collapse whitespace between all tags
- strip comments, empty `doctype`, XML prelude, `metadata`
- strip SVG version
- strip CDATA sections wherever possible
- collapse tags with no content to a void tag
- collapse empty container tags (`g`, `svg`, ...)
- minify style tag and attributes with the CSS minifier
- minify colors
- shorten lengths and numbers and remove default `px` unit
- shorten `path` data
- convert `rect`, `line`, `polygon`, `polyline` to `path`
- use relative or absolute positions in path data whichever is shorter

TODO:
- convert attributes to style attribute whenever shorter
- merge path data? (same style and no intersection -- the latter is difficult)
- truncate decimals

## XML

Minification typically runs at about 70MB/s ~= 250GB/h.

The XML minifier uses these minifications:

- strip unnecessary whitespace and otherwise collapse it to one space (or newline if it originally contained a newline)
- strip comments
- collapse tags with no content to a void tag
- strip CDATA sections wherever possible

Options:

- `KeepWhitespace` do not remove whitespace between inline tags but still collapse multiple whitespace characters into one

## Usage
Any input stream is being buffered by the minification functions. This is how the underlying buffer package inherently works to ensure high performance. The output stream however is not buffer. It is wise to preallocate a buffer as big as the input to which the output is written, or otherwise use `bufio` to buffer to a streaming writer.

### New
Retrieve a minifier struct which holds a map of mediatype &#8594; minifier functions.
``` go
m := minify.New()
```

The following loads all provided minifiers.
``` go
m := minify.New()
m.AddFunc("text/css", css.Minify)
m.AddFunc("text/html", html.Minify)
m.AddFunc("text/javascript", js.Minify)
m.AddFunc("image/svg+xml", svg.Minify)
m.AddFuncRegexp(regexp.MustCompile("[/+]json$"), json.Minify)
m.AddFuncRegexp(regexp.MustCompile("[/+]xml$"), xml.Minify)
```

You can set options to several minifiers.
``` go
m.Add("text/html", &html.Minifier{
	KeepDefaultAttrVals: true,
	KeepWhitespace: true,
})
```

### From reader
Minify from an `io.Reader` to an `io.Writer` for a specific mediatype.
``` go
if err := m.Minify(mediatype, w, r); err != nil {
	panic(err)
}
```

Minify formats directly from an `io.Reader` to an `io.Writer`. The `params map[string]string` would contain the mediatype parameters, pass `nil` if non-existent.
``` go
if err := css.Minify(m, w, r, params); err != nil {
	panic(err)
}

if err := html.Minify(m, w, r, params); err != nil {
	panic(err)
}

if err := js.Minify(m, w, r, params); err != nil {
	panic(err)
}

if err := json.Minify(m, w, r, params); err != nil {
	panic(err)
}

if err := svg.Minify(m, w, r, params); err != nil {
	panic(err)
}

if err := xml.Minify(m, w, r, params); err != nil {
	panic(err)
}
```

### From bytes
Minify from and to a `[]byte` for a specific mediatype.
``` go
b, err = m.Bytes(mediatype, b)
if err != nil {
	panic(err)
}
```

### From string
Minify from and to a `string` for a specific mediatype.
``` go
s, err = m.String(mediatype, s)
if err != nil {
	panic(err)
}
```

### From reader
Get a minifying reader for a specific mediatype.
``` go
mr := m.Reader(mediatype, r)
if _, err := mr.Read(b); err != nil {
	panic(err)
}
```

### From writer
Get a minifying writer for a specific mediatype. Must be explicitly closed because it uses an `io.Pipe` underneath.
``` go
mw := m.Writer(mediatype, w)
mw.Write([]byte("input"))
if err := mw.Close(); err != nil {
	panic(err)
}
```

### Custom minifier
Add a minifier for a specific mimetype.
``` go
type CustomMinifier struct {
	KeepLineBreaks bool
}

func (c *CustomMinifier) Minify(m *minify.M, w io.Writer, r io.Reader, params map[string]string) error {
	// ...
	return nil
}

m.Add(mimetype, &CustomMinifier{KeepLineBreaks: true})
// or
m.AddRegexp(regexp.MustCompile("/x-custom$"), &CustomMinifier{KeepLineBreaks: true})
```

Add a minify function for a specific mimetype.
``` go
m.AddFunc(mimetype, func(m *minify.M, w io.Writer, r io.Reader, params map[string]string) error {
	// ...
	return nil
})
m.AddFuncRegexp(regexp.MustCompile("/x-custom$"), func(m *minify.M, w io.Writer, r io.Reader, params map[string]string) error {
	// ...
	return nil
})
```

Add a command `cmd` with arguments `args` for a specific mimetype.
``` go
m.AddCmd(mimetype, exec.Command(cmd, args...))
m.AddCmdRegexp(regexp.MustCompile("/x-custom$"), exec.Command(cmd, args...))
```

### Mediatypes
Using the `params map[string]string` argument one can pass parameters to the minifier such as seen in mediatypes (`type/subtype; key1=val2; key2=val2`). Examples are the encoding or charset of the data. Calling `Minify` will split the mimetype and parameters for the minifiers for you, but `MinifyMimetype` can be used if you already have them split up.

Minifiers can also be added using a regular expression. For example a minifier with `image/.*` will match any image mime.

## Examples
### Common minifiers
Basic example that minifies from stdin to stdout and loads the default HTML, CSS and JS minifiers. Optionally, one can enable `java -jar build/compiler.jar` to run for JS (for example the [ClosureCompiler](https://code.google.com/p/closure-compiler/)). Note that reading the file into a buffer first and writing to a pre-allocated buffer would be faster (but would disable streaming).
``` go
package main

import (
	"log"
	"os"
	"os/exec"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/js"
	"github.com/tdewolff/minify/json"
	"github.com/tdewolff/minify/svg"
	"github.com/tdewolff/minify/xml"
)

func main() {
	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/html", html.Minify)
	m.AddFunc("text/javascript", js.Minify)
	m.AddFunc("image/svg+xml", svg.Minify)
	m.AddFuncRegexp(regexp.MustCompile("[/+]json$"), json.Minify)
	m.AddFuncRegexp(regexp.MustCompile("[/+]xml$"), xml.Minify)

	// Or use the following for better minification of JS but lower speed:
	// m.AddCmd("text/javascript", exec.Command("java", "-jar", "build/compiler.jar"))

	if err := m.Minify("text/html", os.Stdout, os.Stdin); err != nil {
		panic(err)
	}
}
```

### Custom minifier
Custom minifier showing an example that implements the minifier function interface. Within a custom minifier, it is possible to call any minifier function (through `m minify.Minifier`) recursively when dealing with embedded resources.
``` go
package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/tdewolff/minify"
)

func main() {
	m := minify.New()
	m.AddFunc("text/plain", func(m *minify.M, w io.Writer, r io.Reader, _ map[string]string) error {
		// remove newlines and spaces
		rb := bufio.NewReader(r)
		for {
			line, err := rb.ReadString('\n')
			if err != nil && err != io.EOF {
				return err
			}
			if _, errws := io.WriteString(w, strings.Replace(line, " ", "", -1)); errws != nil {
				return errws
			}
			if err == io.EOF {
				break
			}
		}
		return nil
	})

	in := "Because my coffee was too cold, I heated it in the microwave."
	out, err := m.String("text/plain", in)
	if err != nil {
		panic(err)
	}
	fmt.Println(out)
	// Output: Becausemycoffeewastoocold,Iheateditinthemicrowave.
}
```

### ResponseWriter
ResponseWriter example which returns a ResponseWriter that minifies the content and then writes to the original ResponseWriter. Any write after applying this filter will be minified.
``` go
type MinifyResponseWriter struct {
	http.ResponseWriter
	io.WriteCloser
}

func (m MinifyResponseWriter) Write(b []byte) (int, error) {
	return m.WriteCloser.Write(b)
}

// MinifyResponseWriter must be closed explicitly by calling site.
func MinifyFilter(mediatype string, res http.ResponseWriter) MinifyResponseWriter {
	m := minify.New()
	// add minfiers

	mw := m.Writer(mediatype, res)
	return MinifyResponseWriter{res, mw}
}
```

``` go
// Usage
func(w http.ResponseWriter, req *http.Request) {
	w = MinifyFilter("text/html", w)
	io.WriteString(w, "<p class="message"> This HTTP response will be minified. </p>")
	// Output: <p class=message>This HTTP response will be minified.
}
```

### Templates

Here's an example of a replacement for `template.ParseFiles` from `template/html`, which automatically minifies each template before parsing it.

Be aware that minifying templates will work in most cases but not all. Because the HTML minifier only works for valid HTML5, your template must be valid HTML5 of itself. Template tags are parsed as regular text by the minifier.

``` go
func compileTemplates(filenames ...string) (*template.Template, error) {
	m := minify.New()
	m.AddFunc("text/html", html.Minify)

	var tmpl *template.Template
	for _, filename := range filenames {
		name := filepath.Base(filename)
		if tmpl == nil {
			tmpl = template.New(name)
		} else {
			tmpl = tmpl.New(name)
		}

		b, err := ioutil.ReadFile(filename)
		if err != nil {
			return nil, err
		}

		mb, err := m.Bytes("text/html", b)
		if err != nil {
			return nil, err
		}
		tmpl.Parse(string(mb))
	}
	return tmpl, nil
}
```

Example usage:

``` go
templates := template.MustCompile(compileTemplates("view.html", "home.html"))
```

## License
Released under the [MIT license](LICENSE.md).

[1]: http://golang.org/ "Go Language"
