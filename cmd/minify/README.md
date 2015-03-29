# Minify
Minify is a CLI implemention of the minify [library package](https://github.com/tdewolff/minify/blob/master/README.md).

## Installation
Make sure you have [Go](http://golang.org/) and [Git](http://git-scm.com/) installed.

Run the following command

	go get github.com/tdewolff/minify

and the `minify` command should be in your `$GOPATH/bin`.

## Usage

	Usage: minify [options] [file]
	Options:
	  -o="": Output file (stdout when empty)
	  -x="": File extension (html, css or js), optional for input files

## Examples

```sh
minify -o file.min.html file.html

minify -x css -o file.min.less file.less

minify -x js < file.js > file.min.js

cat file.html | minify -x html > file.min.html
```