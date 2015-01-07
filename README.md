[![GoDoc](http://godoc.org/github.com/tdewolff/minify?status.svg)](http://godoc.org/github.com/tdewolff/minify)

~87% test coverage

# Minify

Minify is a minifier package written in [Go][1]. It has a build-in HTML5 and CSS3 minifier and provides an interface to implement any minifier.

It associates minification functions with mime types, allowing embedded resources (like CSS or JS in HTML files) to be minified too. The user can add any mime-based implementation. Users can also implement a mime type using an external command (like the ClosureCompiler, UglifyCSS, ...).

## HTML
The HTML5 minifier is rather complete and really fast, it strips away:

- unnecessary whitespace
- superfluous quotes, or uses single/double quotes whichever requires fewer escapes
- default attribute values and attribute boolean values
- unrequired tags (`html`, `head`, `body`, ...)
- protocols (`http:` and `javascript:`)
- comments (except conditional comments)
- long `doctype` or `meta` charset

After recent benchmarking and profiling it is really fast and minifies pages in the 10ms range, making it viable for on-the-fly minification.

However, be careful when doing on-the-fly minification. A simple site would typically have HTML pages of 5kB which ideally are compressed to say 4kB. If this would take about 10ms to minify, one has to download slower than 100kB/s to make minification effective. There is a lot of handwaving in this example but it's hardly effective to minify on-the-fly. Rather use caching!

### Comparison

Website | Original | [HTML Compressor](https://code.google.com/p/htmlcompressor/) | Minify | Ratio | Time<sup>&#42;</sup>
------- | -------- | ------------------------------------------------------------ | ------ | ----- | -----------------------
[Amazon](http://www.amazon.com/) | 463kB | 457kB | **443kB** | 96%<sup>&#42;&#42;</sup> | 15ms
[BBC](http://www.bbc.com/) | 113kB | 103kB | **101kB** | 89% | 8ms
[StackOverflow](http://stackoverflow.com/) | 201kB | 184kB | **184kB** | 92% | 18ms
[Wikipedia](http://en.wikipedia.org/wiki/President_of_the_United_States) | 435kB | 423kB | **414kB** | 95%<sup>&#42;&#42;&#42;</sup> | 31ms

<sup>&#42;</sup>These times are measured on my home computer which is an average development computer. The duration varies alot but it's important to see it's in the 20ms range! The benchmark uses only the HTML minifier and excludes the time reading from and writing to a file from the measurement.

<sup>&#42;&#42;</sup>Contains alot of internal CSS and JS blocks so this does not represent the ratio of HTML minification.

<sup>&#42;&#42;&#42;</sup>Is already somewhat minified, so this doesn't reflect the full potential of `minify.HTML`.

[HTML Compressor](https://code.google.com/p/htmlcompressor/) with all HTML options turned on performs worse in output size and speed. It does not omit the `html`, `head`, `body`, ... tags which explains much of the size difference. Furthermore, the whitespace removal is not precise or the user must provide the tags around which can be trimmed. HTML compressor is also an order of magnitude slower. According to HTML Compressor, it produces smaller files than a couple of other libraries, which means `minify.HTML` does better than all.

## CSS
The CSS minifier is very fast and complete, but will only use safe minifications:

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
- lowercase all identifiers except classes, IDs and URLs
- shorten MS alpha function
- remove empty rulesets
- remove repeated selectors
- remove overwritten properties in ruleset
- rewrite properties into one in ruleset if possible (like `margin-top`, `margin-right`, `margin-bottom` and `margin-left` &#8594; `margin`)
- rewrite attribute selectors for IDs and classes (`div[id=a]` &#8594; `div#a`)

It does purposely not use the following techniques:

- (partially) merge rulesets
- (partially) split rulesets
- collapse multiple declarations when main declaration is defined within a ruleset (don't put `font-weight` within an already existing `font`, too complex)
- put nested ID selector at the front (`body > div#elem p` &#8594; `#elem p`, unsafe)
- put space after pseudo-selectors (IE6 is old, move on!)

It's great that so many other tools make comparison tables: [CSS Minifier Comparison](http://www.codenothing.com/benchmarks/css-compressor-3.0/full.html), [CSS minifiers comparison](http://www.phpied.com/css-minifiers-comparison/) and [CleanCSS tests](http://goalsmashers.github.io/css-minification-benchmark/). From the last link, this CSS minifier is almost without doubt the fastest and has near-perfect minification rates. It falls short with the purposely not implemented and often unsafe techniques, so that's fine.

## Installation

Run the following command

	go get github.com/tdewolff/minify

or add the following import and run project with `go get`

``` go
import (
	"github.com/tdewolff/minify"
)
```

## Usage
### New
Retrieve a minifier struct which holds a map of mime &#8594; minifier functions.
``` go
m := minify.NewMinifier()
```

The following loads the default HTML and CSS minifiers.
``` go
m := minify.NewMinifierDefault()
```

### From reader
Minify from an `io.Reader` to an `io.Writer` with mime type `mime`.
``` go
if err := m.Minify(mime, w, r); err != nil {
	fmt.Println("Minify:", err)
}
```

Minify *HTML* directly from an `io.Reader` to an `io.Writer`.
``` go
if err := m.HTML(w, r); err != nil {
	fmt.Println("HTML:", err)
}
```

Minify *CSS* directly from an `io.Reader` to an `io.Writer`.
``` go
if err := m.CSS(w, r); err != nil {
	fmt.Println("CSS:", err)
}
```

### From bytes
Minify from and to a `[]byte` with mime type `mime`.
``` go
b, err := m.MinifyBytes(mime, b)
if err != nil {
	fmt.Println("MinifyBytes:", err)
}
```

### From string
Minify from and to a `string` with mime type `mime`.
``` go
s, err := m.MinifyString(mime, s)
if err != nil {
	fmt.Println("MinifyString:", err)
}
```

### Custom minifier
Add a function for a specific mime type `mime`.
``` go
m.Add(mime, func(m minify.Minifier, w io.Writer, r io.Reader) error {
	// ...
	return nil
})
```

Add a command `cmd` with arguments `args` for a specific mime type `mime`.
``` go
m.AddCmd(mime, exec.Command(cmd, args...))
```

## Examples
Basic example that minifies from stdin to stdout and loads the default HTML and CSS minifiers. Additionally, a JS minifier is set to run `java -jar build/compiler.jar` (for example the [ClosureCompiler](https://code.google.com/p/closure-compiler/)). Note that reading the file into a buffer first and writing to a buffer would be faster.
``` go
package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/tdewolff/minify"
)

func main() {
	m := minify.NewMinifierDefault()
	m.AddCmd("text/javascript", exec.Command("java", "-jar", "build/compiler.jar"))

	if err := m.Minify("text/html", os.Stdout, os.Stdin); err != nil {
		fmt.Println("Minify:", err)
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
		fmt.Println("Minify:", err)
	}
	fmt.Println(out)
}
```

[1]: http://golang.org/ "Go Language"
