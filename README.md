[![GoDoc](http://godoc.org/github.com/tdewolff/minify?status.svg)](http://godoc.org/github.com/tdewolff/minify) [![GoCover](http://gocover.io/_badge/github.com/tdewolff/minify)](http://gocover.io/github.com/tdewolff/minify)

# Minify
Minify is a minifier package written in [Go][1]. It has a build-in HTML5, CSS3 and JS minifiers and provides an interface to implement any minifier. The implemented minifiers have high performance and are streaming, but having the buffer loaded into memory first (by having `Bytes() []byte` available) does speed up the process.

It associates minification functions with mime types, allowing embedded resources (like CSS or JS in HTML files) to be minified too. The user can add any mime-based implementation. Users can also implement a mime type using an external command (like the ClosureCompiler, UglifyCSS, ...). It is possible to pass parameters through the mimetype to specify the charset for example for future minifiers.

## Comparison
Minification typically runs at about 10-20MB/s ~= 35-70GB/h, depeding on the composition of the file.

Website | Original | Minified | Ratio | Time<sup>&#42;</sup>
------- | -------- | -------- | ----- | -----------------------
[Amazon](http://www.amazon.com/) | 463kB | **418kB** | 90% | 22ms
[BBC](http://www.bbc.com/) | 113kB | **96kB** | 85% | 7ms
[StackOverflow](http://stackoverflow.com/) | 201kB | **183kB** | 91% | 15ms
[Wikipedia](http://en.wikipedia.org/wiki/President_of_the_United_States) | 435kB | **413kB** | 95%<sup>&#42;&#42;</sup> | 31ms

<sup>&#42;</sup>These times are measured on my home computer which is an average development computer. The duration varies alot but it's important to see it's in the 20ms range! The benchmark uses the HTML, CSS and JS minifiers and excludes the time reading from and writing to a file from the measurement.

<sup>&#42;&#42;</sup>Is already somewhat minified, so this doesn't reflect the full potential of this minifier.

[HTML Compressor](https://code.google.com/p/htmlcompressor/) performs worse in output size (for HTML and CSS) and speed, it is a magnitude slower. Its whitespace removal is not precise or the user must provide the tags around which can be trimmed. According to HTML Compressor, it produces smaller files than a couple of other libraries. With HTML and CSS compression this package is better, but JS compression it is still too basic.

### Alternatives
An alternative library written in Go is [https://github.com/dchest/htmlmin](https://github.com/dchest/htmlmin). It is written using regular expressions and is therefore a lot simpler (and thus less bugs, not handling edge-cases) but about twice as slow. Other alternatives are bindings for existing minifiers written in other languages. These are inevitably more robust and tested but will often be slower.

## HTML
[![GoDoc](http://godoc.org/github.com/tdewolff/minify/html?status.svg)](http://godoc.org/github.com/tdewolff/minify/html) [![GoCover](http://gocover.io/_badge/github.com/tdewolff/minify/html)](http://gocover.io/github.com/tdewolff/minify/html)

The HTML5 minifier uses these minifications:

- strip unnecessary whitespace
- strip superfluous quotes, or uses single/double quotes whichever requires fewer escapes
- strip default attribute values and attribute boolean values
- strip unrequired tags (`html`, `head`, `body`, ...)
- strip default protocols (`http:` and `javascript:`)
- strip comments (except conditional comments)
- strip long `doctype` or `meta` charset
- lowercase tags, attributes and some values to enhance gzip compression

After recent benchmarking and profiling it became really fast and minifies pages in the 20ms range, making it viable for on-the-fly minification.

However, be careful when doing on-the-fly minification. A simple site would typically have HTML pages of 5kB which ideally are compressed to say 4kB. If this would take about 10ms to minify, one has to download slower than 100kB/s to make minification effective. There is a lot of handwaving in this example but it's hardly effective to minify on-the-fly. Rather use caching!

### Beware
Make sure your HTML doesn't depend on whitespace between `block` elements that have been changed to `inline` or `inline-block` elements using CSS. Your layout *should not* depend on those whitespaces as the minifier will remove them. An example is a list of `<li>`s which have `display:inline-block` applied and have whitespace in between them.

## CSS
[![GoDoc](http://godoc.org/github.com/tdewolff/minify/css?status.svg)](http://godoc.org/github.com/tdewolff/minify/css) [![GoCover](http://gocover.io/_badge/github.com/tdewolff/minify/css)](http://gocover.io/github.com/tdewolff/minify/css)

The CSS minifier will only use safe minifications:

- remove comments and (most) whitespace
- remove trailing semicolon(s)
- optimize `margin`, `padding` and `border-width` number of sides
- remove unnecessary decimal zeros and the `+` sign
- remove dimension and percentage for zero values
- remove quotes for URLs
- remove quotes for font families and make lowercase
- rewrite hex colors to/from color names, or to 3 digit hex
- rewrite `rgb(` and `rgba(` colors to hex/name when possible
- replace `normal` and `bold` by numbers for `font-weight` and `font`
- replace `none` &#8594; `0` for `border`, `background` and `outline`
- lowercase all identifiers except classes, IDs and URLs to enhace GZIP compression
- shorten MS alpha function
- remove empty rulesets
- remove repeated selectors
- rewrites data URI's with base64 or ASCII whichever is shorter
- calls minifier for data URI mediatypes, thus you can compress embedded SVG files if you have that minifier attached for example

It does purposely not use the following techniques:

- (partially) merge rulesets
- (partially) split rulesets
- collapse multiple declarations when main declaration is defined within a ruleset (don't put `font-weight` within an already existing `font`, too complex)
- remove overwritten properties in ruleset (this not always overwrites it, for example with `!important`)
- rewrite properties into one in ruleset if possible (like `margin-top`, `margin-right`, `margin-bottom` and `margin-left` &#8594; `margin`)
- put nested ID selector at the front (`body > div#elem p` &#8594; `#elem p`, unsafe)
- rewrite attribute selectors for IDs and classes (`div[id=a]` &#8594; `div#a`)
- put space after pseudo-selectors (IE6 is old, move on!)

It's great that so many other tools make comparison tables: [CSS Minifier Comparison](http://www.codenothing.com/benchmarks/css-compressor-3.0/full.html), [CSS minifiers comparison](http://www.phpied.com/css-minifiers-comparison/) and [CleanCSS tests](http://goalsmashers.github.io/css-minification-benchmark/). From the last link, this CSS minifier is almost without doubt the fastest and has near-perfect minification rates. It falls short with the purposely not implemented and often unsafe techniques, so that's fine.

Minification typically runs at about 8MB/s ~= 30GB/h.

Library | Original | Minified | Ratio | Time<sup>&#42;</sup>
------- | -------- | -------- | ----- | -----------------------
[Bootstrap](http://getbootstrap.com/) | 134kB | **111kB** | 83% | 17ms
[Gumby](http://gumbyframework.com/) | 182kB | **167kB** | 92% | 23ms

<sup>&#42;</sup>The benchmark excludes the time reading from and writing to a file from the measurement.

## JS
[![GoDoc](http://godoc.org/github.com/tdewolff/minify/js?status.svg)](http://godoc.org/github.com/tdewolff/minify/js) [![GoCover](http://gocover.io/_badge/github.com/tdewolff/minify/js)](http://gocover.io/github.com/tdewolff/minify/js)

The JS minifier is pretty basic. It removes comments, whitespace and line breaks whenever it can. It follows the rules by [JSMin](http://www.crockford.com/javascript/jsmin.html) but additionally fixes the error in the 'caution' section.

Minification typically runs at about 40MB/s ~= 145GB/h.

Library | Original | Minified | Ratio | Time<sup>&#42;</sup>
------- | -------- | -------- | ----- | -----------------------
[jQuery](http://jquery.com/download/) | 242kB | **130kB** | 54% | 6ms
[jQuery UI](http://jqueryui.com/download/) | 459kB | **300kB** | 65% | 12ms

<sup>&#42;</sup>The benchmark excludes the time reading from and writing to a file from the measurement.

## Trim
[![GoDoc](http://godoc.org/github.com/tdewolff/minify/trim?status.svg)](http://godoc.org/github.com/tdewolff/minify/trim) [![GoCover](http://gocover.io/_badge/github.com/tdewolff/minify/trim)](http://gocover.io/github.com/tdewolff/minify/trim)

The trim minifier strips the whitespace at the beginning and end of the stream, it can be used as a general purpose minifier for any mimetype.

## Installation
Run the following commands

	go get github.com/tdewolff/minify
	go get github.com/tdewolff/minify/html
	go get github.com/tdewolff/minify/css
	go get github.com/tdewolff/minify/js
	go get github.com/tdewolff/minify/trim

or add the following imports and run the project with `go get`
``` go
import (
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/js"
	"github.com/tdewolff/minify/trim"
)
```

## Usage
### New
Retrieve a minifier struct which holds a map of mediatype &#8594; minifier functions.
``` go
m := minify.NewMinifier()
```

The following loads all provided minifiers.
``` go
m := minify.NewMinifier()
m.Add("text/html", html.Minify)
m.Add("text/css", css.Minify)
m.Add("text/js", css.Minify)
m.Add("*/*", trim.Minify)
```

### From reader
Minify from an `io.Reader` to an `io.Writer` with mediatype `mediatype`.
``` go
if err := m.Minify(mediatype, w, r); err != nil {
	log.Fatal("Minify:", err)
}
```

Minify HTML, CSS or JS directly from an `io.Reader` to an `io.Writer`.
``` go
if err := html.Minify(m, w, r); err != nil {
	log.Fatal("Minify:", err)
}

if err := css.Minify(m, w, r); err != nil {
	log.Fatal("Minify:", err)
}

if err := js.Minify(m, w, r); err != nil {
	log.Fatal("Minify:", err)
}
```

### From bytes
Minify from and to a `[]byte` with mediatype type `mediatype`.
``` go
b, err := m.MinifyBytes(mediatype, b)
if err != nil {
	log.Fatal("MinifyBytes:", err)
}
```

### From string
Minify from and to a `string` with mediatype type `mediatype`.
``` go
s, err := m.MinifyString(mediatype, s)
if err != nil {
	log.Fatal("MinifyString:", err)
}
```

### Custom minifier
Add a function for a specific mediatype type `mediatype`.
``` go
m.Add(mediatype, func(m minify.Minifier, w io.Writer, r io.Reader) error {
	// ...
	return nil
})
```

Add a command `cmd` with arguments `args` for a specific mediatype type `mediatype`.
``` go
m.AddCmd(mediatype, exec.Command(cmd, args...))
```

### Mediatypes
Mediatypes can contain wildcards (`*`) and parameters (`; key1=val2; key2=val2`). For example a minifier with `image/*` will match any image mime.

Data such as `text/plain; charset=UTF-8` will be processed by `text/plain` or `text/*` or `*/*` whichever exists and will pass the parameter `"charset":"UTF-8"` which is retrievable by calling `Param(string) string` on the `Minifier` interface.

## Examples
Basic example that minifies from stdin to stdout and loads the default HTML and CSS minifiers. Additionally, a JS minifier is set to run `java -jar build/compiler.jar` (for example the [ClosureCompiler](https://code.google.com/p/closure-compiler/)). Note that reading the file into a buffer first and writing to a buffer would be faster.
``` go
package main

import (
	"log"
	"os"
	"os/exec"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/trim"
)

func main() {
	m := minify.NewMinifier()
	m.Add("text/html", html.Minify)
	m.Add("text/css", css.Minify)
	m.Add("*/*", trim.Minify)
	m.AddCmd("text/javascript", exec.Command("java", "-jar", "build/compiler.jar"))

	if err := m.Minify("text/html", os.Stdout, os.Stdin); err != nil {
		log.Fatal("Minify:", err)
	}
}
```

Custom minifier showing an example that implements the minifier function interface.  Within a custom minifier, it is possible to call any minifier function (through `m minify.Minifier`) recursively when dealing with embedded resources.
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

// Outputs "Becausemycoffeewastoocold,Iheateditinthemicrowave."
func main() {
	m := minify.NewMinifier()

	// remove newline and space bytes
	m.Add("text/plain", func(m minify.Minifier, w io.Writer, r io.Reader) error {
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

	out, err := m.MinifyString("text/plain", "Because my coffee was too cold, I heated it in the microwave.")
	if err != nil {
		log.Fatal("Minify:", err)
	}
	fmt.Println(out)
}
```

## License
Released under the [MIT license](LICENSE.md).

[1]: http://golang.org/ "Go Language"
