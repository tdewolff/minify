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
The CSS minifier is sufficient but could use more work. It currently:

- removes most unnecessary whitespace
- shortens color codes (by using hexadecimal color codes or color identifiers)
- shortens zero values (`0em` &#8594; `0`)
- shortens single `margin`/`padding` values (`margin:1px 1px` &#8594; `margin:1px`)
- shortens a few other values (`outline:none` &#8594; `outline:0`)

In the future it needs to be able to collapse blocks with the same identifier, multiple `margin`/`padding`/`background`/... declarations into one, etc.

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
