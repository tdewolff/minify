[![GoDoc](http://godoc.org/github.com/tdewolff/GoMinify?status.svg)](http://godoc.org/github.com/tdewolff/GoMinify)

# GoMinify

GoMinify is a minifier package written in [Go][1]. It has a build-in HTML and CSS minifier and provides an interface to implement any minifier.

It associates minification functions with mime types, allowing embedded resources (like CSS or JS in HTML files) to be minified too. The user can add any mime-based implementation. User can also implement a mime type using an external command.

## HTML
The HTML minifier is rather complete, it strips away:

- unnecessary whitespace
- superfluous quotes, or single/double quotes depending on whichever requires fewer escapes
- default attribute values or attribute booleans
- unrequired tags (html, head, body, ...)
- default URL protocol (http:)
- comments

It also rewrites the doctype and meta charset into a shorter format according to [Google's HTML5 performance](https://developers.google.com/speed/articles/html5-performance).

### Comparison

Website | Original size (kB) | GoMinify (kB) | Ratio | Time
------- | ------------------ | ------------- | ----- | ----
[Amazon](http://www.amazon.com/) | 1684 | 1386 | 82% | 2.5s
[StackOverflow](http://stackoverflow.com/) | 1248 | 1050 | 84% | 2.2s
[Wikipedia](http://en.wikipedia.org/wiki/President_of_the_United_States) | 1874 | 1673 | 89% | 3.1s

TODO: compare to other minifiers

## CSS
The CSS minifier is immature and needs more work. It features:

- removes unnecessary whitespace
- shortens color codes (by using hexadecimal color codes or color words)
- shortens a few other values

It is in need of a CSS tokenizer, preferably from another package, in future.

## Installation

Run the following command

	go get github.com/tdewolff/GoMinify

or add the following import and run project with `go get`

	import "github.com/tdewolff/GoMinify"

## Usage
Retrieve a minifier struct which holds a map of mime -> minifier function. The following loads the default HTML and CSS minifier:

	m := minify.NewMinifier()

To minify a generic stream, byte array or stream with mime type `mime`:
``` go
// stream
if err := m.Minify(mime, w, r); err != nil {
	fmt.Println("minify.Minify:", err)
}

// byte array
b, err := m.MinifyBytes(mime, b)
if err != nil {
	fmt.Println("minify.Minify:", err)
}

// string
s, err := m.MinifyString(mime, s)
if err != nil {
	fmt.Println("minify.Minify:", err)
}
```

Add function or command for specific mime type `mime`:
``` go
// function
m.Add(mime, func(m minify.Minifier, w io.Writer, r io.Reader) error {
	io.Copy(w, r)
	return nil
})

// external command
m.AddCmd(mime, exec.Command(cmd, args...))
```

### Examples
Basic example:
``` go
package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/tdewolff/GoMinify"
)

// Minifies HTML code from stdin to stdout
func main() {
	m := minify.NewMinifier()
	m.AddCmd("text/javascript", exec.Command("java", "-jar", "path/to/compiler.jar"))

	if err := m.Minify("text/html", os.Stdout, os.Stdin); err != nil {
		fmt.Println("minify.Minify:", err)
	}
}
```

Custom minifier:
``` go
package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/tdewolff/GoMinify"
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

			_, errws := io.WriteString(w, strings.Replace(line, " ", "", -1))
			if errws != nil {
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
		fmt.Println("minify.Minify:", err)
	}
	fmt.Println(out)
}
```

Within a custom minifier, one can call `m.MinifyBytes("mime", w, byteArray)` when dealing with embedded resources.

[1]: http://golang.org/ "Go Language"