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

It also rewrites the doctype and meta charset into a shorter format according to [Google's HTML5 performance][https://developers.google.com/speed/articles/html5-performance].

TODO: comparison to other minifiers

## CSS
The CSS minifier is immature and needs more work. It features:

- removes unnecessary whitespace
- shortens color codes (by using hexadecimal color codes or color words)
- shortens a few other values

It is in need of a CSS tokenizer, preferably from another package, in future.

## Usage

TODO: examples

[1]: http://golang.org/ "Go Language"