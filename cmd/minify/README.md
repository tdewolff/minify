# Minify
Minify is a CLI implemention of the minify [library package](https://github.com/tdewolff/minify/blob/master/README.md).

## Installation
Make sure you have [Go](http://golang.org/) and [Git](http://git-scm.com/) installed.

Run the following command

	go get github.com/tdewolff/minify/cmd/minify

and the `minify` command will be in your `$GOPATH/bin`.

## Usage

	Usage: minify [options] [input]

	Options:
	  -a, --all
	        Minify all files, including hidden files and files in hidden directories
	  --html-keep-default-attrvals
	        Preserve default attribute values
	  --html-keep-whitespace
	        Preserve whitespace characters but still collapse multiple whitespace into one
	  -l, --list
	        List all accepted filetypes
	  --match string
	        Filename pattern matching using regular expressions, see https://github.com/google/re2/wiki/Syntax
	  --mime string
	        Mimetype (text/css, application/javascript, ...), optional for input filenames, has precendence over -type
	  -o, --output string
	        Output file or directory, leave blank to use stdout
	  -r, --recursive
	        Recursively minify directories
	  --type string
	        Filetype (css, html, js, ...), optional for input filenames
	  --url string
	        URL of file to enable URL minification
	  -v, --verbose
	        Verbose
	  -w, --watch
	        Watch files and minify upon changes
	  --xml-keep-whitespace
	        Preserve whitespace characters but still collapse multiple whitespace into one

	Input:
	  Files or directories, leave blank to use stdin

### Types

	css     text/css
	htm     text/html
	html    text/html
	js      text/javascript
	json    application/json
	svg     image/svg+xml
	xml     text/xml

## Examples
The following commands are variations one can use to minify files:

```sh
$ minify file.html # file.html &#8594; file.min.html

$ minify --type=css -o file_minified.ext file.ext # file.ext &#8594; file_minified.ext

$ minify --mime=text/javascript < file.js > file.min.js

$ cat file.html | minify --type=html > file.min.html
```

It is also possible to overwrite the input file by the output file. Overwriting existing files needs to happen forcefully. However, overwriting won't work with input/output redirection streams. Using the following command the input file will be loaded into memory first before writing to the output file:

```sh
$ minify file.html
```

You can also give directories as input, and these directories can be minified recursively:
```sh
$ minify . # minify files in current working directory (no subdirectories)

$ minify -r dir # minify files in dir recursively

$ minify -r --match=\.js dir # minify only javascript files in dir
```

Additionally, you can watch for file changes and automatically re-minify:
```sh
$ minify -r --watch dir -o dir-min
```
