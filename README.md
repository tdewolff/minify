# Minify [![GoDoc](http://godoc.org/github.com/tdewolff/minify?status.svg)](http://godoc.org/github.com/tdewolff/minify) [![GoCover](http://gocover.io/_badge/github.com/tdewolff/minify)](http://gocover.io/github.com/tdewolff/minify)

**Update: [online live demo](http://pi.tacodewolff.nl:8080/minify) running on a Raspberry Pi 2.**

**Update: [command-line-interface](https://github.com/tdewolff/minify/tree/master/cmd/minify) executable `minify` provided for tooling.**

Minify is a minifier package written in [Go][1]. It has build-in HTML5, CSS3 and JS minifiers and provides an interface to implement any minifier. The implemented minifiers are very high performance and streaming (which implies O(n)).

It associates minification functions with mime types, allowing embedded resources (like CSS or JS in HTML files) to be minified too. The user can add any mime-based implementation. Users can also implement a mime type using an external command (like the ClosureCompiler, UglifyCSS, ...). It is possible to pass parameters through the mimetype to specify the charset for example.

Bottleneck for minification is mainly io and can be significantly sped up by having the file loaded into memory and providing `Bytes() []byte` like `bytes.Buffer` does.

**Table of Contents**

- [Minify](#minify)
	- [Comparison](#comparison)
		- [Alternatives](#alternatives)
	- [HTML](#html)
		- [Beware](#beware)
	- [CSS](#css)
	- [JS](#js)
	- [Installation](#installation)
	- [Usage](#usage)
		- [New](#new)
		- [From reader](#from-reader)
		- [From bytes](#from-bytes)
		- [From string](#from-string)
		- [Custom minifier](#custom-minifier)
		- [Mediatypes](#mediatypes)
	- [Examples](#examples)
	- [License](#license)

## Comparison
HTML (with JS and CSS) minification typically runs at about 20-30MB/s ~= 70-100GB/h, depending on the composition of the file.

Website | Original | Minified | Ratio | Time<sup>&#42;</sup>
------- | -------- | -------- | ----- | -----------------------
[Amazon](http://www.amazon.com/) | 463kB | **414kB** | 89% | 15ms
[BBC](http://www.bbc.com/) | 113kB | **96kB** | 85% | 4ms
[StackOverflow](http://stackoverflow.com/) | 201kB | **182kB** | 91% | 8ms
[Wikipedia](http://en.wikipedia.org/wiki/President_of_the_United_States) | 435kB | **410kB** | 94%<sup>&#42;&#42;</sup> | 17ms

<sup>&#42;</sup>These times are measured on my home computer which is an average development computer. The duration varies a lot but it's important to see it's in the 10ms range! The benchmark uses the HTML, CSS and JS minifiers and excludes the time reading from and writing to a file from the measurement.

<sup>&#42;&#42;</sup>Is already somewhat minified, so this doesn't reflect the full potential of this minifier.

[HTML Compressor](https://code.google.com/p/htmlcompressor/) performs worse in output size (for HTML and CSS) and speed; it is a magnitude slower. Its whitespace removal is not precise or the user must provide the tags around which can be trimmed. According to HTML Compressor, it produces smaller files than a couple of other libraries. With HTML and CSS minification this package is better, but JS minification it is still too basic.

### Alternatives
An alternative library written in Go is [https://github.com/dchest/htmlmin](https://github.com/dchest/htmlmin). It is written using regular expressions and is therefore a lot simpler (and thus less bugs, not handling edge-cases) but about twice as slow. Also [https://github.com/omeid/jsmin](https://github.com/omeid/jsmin) contains a port of JSMin, just like this JS minifier, but is slower.

Other alternatives are bindings for existing minifiers written in other languages. These are inevitably more robust and tested but will often be slower.

## HTML
[![GoDoc](http://godoc.org/github.com/tdewolff/minify/html?status.svg)](http://godoc.org/github.com/tdewolff/minify/html) [![GoCover](http://gocover.io/_badge/github.com/tdewolff/minify/html)](http://gocover.io/github.com/tdewolff/minify/html)

The HTML5 minifier uses these minifications:

- strip unnecessary whitespace and otherwise collapse it to one space
- strip superfluous quotes, or uses single/double quotes whichever requires fewer escapes
- strip default attribute values and attribute boolean values
- strip unrequired tags (`html`, `head`, `body`, ...)
- strip unrequired end tags (`tr`, `td`, `li`, ... and often `p`)
- strip default protocols (`http:` and `javascript:`)
- strip comments (except conditional comments)
- ahorten `doctype` and `meta` charset
- lowercase tags, attributes and some values to enhance gzip compression

After recent benchmarking and profiling it became really fast and minifies pages in the 10ms range, making it viable for on-the-fly minification.

However, be careful when doing on-the-fly minification. Minification typically trims off 10% and does this at worst around about 20MB/s. This means the users has to download slower than 2MB/s to make on-the-fly minification worthwhile. This may or may not apply in your situation. Rather use caching!

### Beware
Make sure your HTML doesn't depend on whitespace between `block` elements that have been changed to `inline` or `inline-block` elements using CSS. Your layout *should not* depend on those whitespaces as the minifier will remove them. An example is a menu consisting of multiple `<li>` that have `display:inline-block` applied and have whitespace in between them. It is bad practise to rely on whitespace for element positioning anyways!

## CSS
[![GoDoc](http://godoc.org/github.com/tdewolff/minify/css?status.svg)](http://godoc.org/github.com/tdewolff/minify/css) [![GoCover](http://gocover.io/_badge/github.com/tdewolff/minify/css)](http://gocover.io/github.com/tdewolff/minify/css)

Minification typically runs at about 10MB/s ~= 35GB/h.

Library | Original | Minified | Ratio | Time<sup>&#42;</sup>
------- | -------- | -------- | ----- | -----------------------
[Bootstrap](http://getbootstrap.com/) | 134kB | **111kB** | 83% | 13ms
[Gumby](http://gumbyframework.com/) | 182kB | **167kB** | 92% | 18ms

<sup>&#42;</sup>The benchmark excludes the time reading from and writing to a file from the measurement.

The CSS minifier will only use safe minifications:

- remove comments and (most) whitespace
- remove trailing semicolons
- optimize `margin`, `padding` and `border-width` number of sides
- remove unnecessary decimal zeros and the `+` sign
- remove dimension and percentage for zero values
- remove quotes for URLs
- remove quotes for font families and make lowercase
- rewrite hex colors to/from color names, or to 3 digit hex
- rewrite `rgb(` and `rgba(` colors to hex/name when possible
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
[![GoDoc](http://godoc.org/github.com/tdewolff/minify/js?status.svg)](http://godoc.org/github.com/tdewolff/minify/js) [![GoCover](http://gocover.io/_badge/github.com/tdewolff/minify/js)](http://gocover.io/github.com/tdewolff/minify/js)

The JS minifier is pretty basic. It removes comments, whitespace and line breaks whenever it can. It follows the rules by [JSMin](http://www.crockford.com/javascript/jsmin.html) but additionally fixes the error in the 'caution' section.

Minification typically runs at about 45MB/s ~= 160GB/h.

Library | Original | Minified | Ratio | Time<sup>&#42;</sup>
------- | -------- | -------- | ----- | -----------------------
[ACE](https://github.com/ajaxorg/ace-builds) | 616kB | **433kB** | 70% | 13ms
[jQuery](http://jquery.com/download/) | 242kB | **130kB** | 54% | 5ms
[jQuery UI](http://jqueryui.com/download/) | 459kB | **300kB** | 65% | 11ms
[Moment](http://momentjs.com/) | 97kB | **51kB** | 52% | 2ms

<sup>&#42;</sup>The benchmark excludes the time reading from and writing to a file from the measurement.

## Installation
Run the following command

	go get github.com/tdewolff/minify

or add the following imports and run the project with `go get`
``` go
import (
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/js"
)
```

## Usage
### New
Retrieve a minifier struct which holds a map of mediatype &#8594; minifier functions.
``` go
m := minify.New()
```

The following loads all provided minifiers.
``` go
m := minify.New()
m.AddFunc("text/html", html.Minify)
m.AddFunc("text/css", css.Minify)
m.AddFunc("text/javascript", js.Minify)
```

### From reader
Minify from an `io.Reader` to an `io.Writer` for a specific mediatype.
``` go
if err := m.Minify(mediatype, w, r); err != nil {
	log.Fatal("Minify:", err)
}
```

Minify HTML, CSS or JS directly from an `io.Reader` to an `io.Writer`. The passed mediatype is not required for these functions, but are filled out for clarity.
``` go
if err := html.Minify(m, "text/html", w, r); err != nil {
	log.Fatal("Minify:", err)
}

if err := css.Minify(m, "text/css", w, r); err != nil {
	log.Fatal("Minify:", err)
}

if err := js.Minify(m, "text/javascript", w, r); err != nil {
	log.Fatal("Minify:", err)
}
```

### From bytes
Minify from and to a `[]byte` for a specific mediatype.
``` go
b, err = minify.Bytes(m, mediatype, b)
if err != nil {
	log.Fatal("Bytes:", err)
}
```

### From string
Minify from and to a `string` for a specific mediatype.
``` go
s, err = minify.String(m, mediatype, s)
if err != nil {
	log.Fatal("String:", err)
}
```

### Custom minifier
Add a function for a specific mediatype.
``` go
m.AddFunc(mediatype, func(m minify.Minifier, mediatype string, w io.Writer, r io.Reader) error {
	// ...
	return nil
})
```

Add a command `cmd` with arguments `args` for a specific mediatype.
``` go
m.AddCmd(mediatype, exec.Command(cmd, args...))
```

### Mediatypes
Mediatypes can contain wildcards (`*`) and parameters (`; key1=val2; key2=val2`). For example a minifier with `image/*` will match any image mime.

Mediatypes such as `text/plain; charset=UTF-8` will be processed by `text/plain` or `text/*` or `*/*` whichever exists. The mediatype string is passed to the minifier function which can retrieve the parameters using `mime.ParseMediaType`.

## Examples
Basic example that minifies from stdin to stdout and loads the default HTML, CSS and JS minifiers. Optionally, one can enable `java -jar build/compiler.jar` to run for JS (for example the [ClosureCompiler](https://code.google.com/p/closure-compiler/)). Note that reading the file into a buffer first and writing to a pre-allocated buffer would be faster (but would disable streaming).
``` go
package main

import (
	"log"
	"os"
	"os/exec"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/js"
)

func main() {
	m := minify.New()
	m.AddFunc("text/html", html.Minify)
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/javascript", js.Minify)
	// Or use the following for better minification of JS but lower speed:
	// m.AddCmd("text/javascript", exec.Command("java", "-jar", "build/compiler.jar"))

	if err := m.Minify("text/html", os.Stdout, os.Stdin); err != nil {
		log.Fatal("Minify:", err)
	}
}
```

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

// Outputs "Becausemycoffeewastoocold,Iheateditinthemicrowave."
func main() {
	m := minify.New()

	// remove newline and space bytes
	m.AddFunc("text/plain", func(m minify.Minifier, mediatype string, w io.Writer, r io.Reader) error {
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

	out, err := minify.String(m, "text/plain", "Because my coffee was too cold, I heated it in the microwave.")
	if err != nil {
		log.Fatal("Minify:", err)
	}
	fmt.Println(out)
}
```

## License
Released under the [MIT license](LICENSE.md).

[1]: http://golang.org/ "Go Language"
