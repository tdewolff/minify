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
	        Output file or directory (must have trailing slash), leave blank to use stdout
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
Minify **index.html** to **index-min.html**:
```sh
$ minify -o index-min.html index.html
```

Minify **index.html** to standard output (leave `-o` blank):
```sh
$ minify index.html
```

Normally the mimetype is inferred from the extension, to set the mimetype explicitly:
```sh
$ minify --type=html -o index-min.xhtml index.xhtml
```

You need to set the type or the mimetype option when using standard input:
```sh
$ minify --mime=text/javascript < script.js > script-min.js

$ cat script.js | minify --type=html > script-min.js
```

### Directories
You can also give directories as input, and these directories can be minified recursively.

Minify files in the current working directory to **out/** (no subdirectories):
```sh
$ minify -o min/ .
```

Minify files recursively in **src**:
```sh
$ minify -r -o min/ src
```

Minify only javascript files in **src**:
```sh
$ minify -r -o min/ --match=\.js src
```

### Concatenate
When multiple inputs are given and either standard output or a single output file, it will concatenate the files together.

Concatenate **one.css** and **two.css** into **style.css**:
```sh
$ minify -o style.css one.css two.css
```

Concatenate all files in **styles** into **style.css**:
```sh
$ minify -o style.css styles/
```

You can also use `cat` as standard input to concatenate files and use gzip for example:
```sh
$ cat one.css two.css three.css | minify --type=css | gzip -9 -c > style.css.gz
```

### Watching
To watch file changes and automatically re-minify you can use the `-w` or `--watch` option.

Minify **style.css** to itself and watch changes:
```sh
$ minify -w -o style.css style.css
```

Minify and concatenate **one.css** and **two.css** to **style.css** and watch changes:
```sh
$ minify -w -o style.css one.css two.css
```

Minify files in **src** and subdirectories to **min** and watch changes:
```sh
$ minify -w -r -o min/ src
```
