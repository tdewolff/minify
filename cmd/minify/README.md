# Minify [![Join the chat at https://gitter.im/tdewolff/minify](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/tdewolff/minify?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

**[Download binaries](https://dl.equinox.io/tdewolff/minify/stable) for Windows, Linux and Mac OS X**

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
	  --concat
	        Concatenate inputs to single output file
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
	  -u, --update
	        Update binary
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
$ minify file.html

$ minify --type=css -o file_minified.ext file.ext

$ minify --mime=text/javascript < file.js > file.min.js

$ cat file.html | minify --type=html > file.min.html
```

### Directories
You can also give directories as input, and these directories can be minified recursively:
```sh
$ minify . # minify files in current working directory (no subdirectories)

$ minify -r dir # minify files in dir recursively

$ minify -r --match=\.js dir # minify only javascript files in dir
```

### Concatenate
```sh
$ minify --concat -o style.css one.css two.css three.css

$ cat one.css two.css three.css | minify --type=css > style.css
```

### Watching
To watch file changes and automatically re-minify you can use the `--watch` option. Watching doesn't work (yet) when overwriting files by themselves, it also works only for one input directory and doesn't go together with concatenation.
```sh
$ minify -r --watch dir -o dir-min
```

