# Minify
Minify is a CLI implemention of the minify [library package](https://github.com/tdewolff/minify/blob/master/README.md).

## Installation
Make sure you have [Go](http://golang.org/) and [Git](http://git-scm.com/) installed.

Run the following command

	go get github.com/tdewolff/minify/cmd/minify

and the `minify` command should be in your `$GOPATH/bin`.

## Usage

	Usage: minify [options] [file]
	Options:
	  -o: Output file (stdout when empty)
	  -x: File extension (css, html, js, json, svg or xml), optional for input files
	  -d: Directory to search for files
	  -r: Recursively minify everything

## Examples
The following commands are variations one can use to minify a file:

```sh
$ minify -o file.min.html file.html

$ minify -x css -o file.min.less file.less

$ minify -x js < file.js > file.min.js

$ cat file.html | minify -x html > file.min.html
```

It is also possible to overwrite the input file by the output file. However, this won't work with input/output redirection streams. Using the following command the input file will be loaded into memory first before writing to the output file:

```sh
$ minify -o file.html file.html
```

The following commands minify the files in a directory:
```sh
$ minify -d path/to/dir

$ minify -d path/to/dir -r
```
