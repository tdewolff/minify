# Minify [![Join the chat at https://gitter.im/tdewolff/minify](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/tdewolff/minify?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

**[Download binaries](https://github.com/tdewolff/minify/releases) for Windows, Linux and macOS**

Minify is a CLI implementation of the minify [library package](https://github.com/tdewolff/minify).

## Installation
Make sure you have [Go](http://golang.org/) and [Git](http://git-scm.com/) installed.

Run the following command

    mkdir $HOME/src
    cd $HOME/src
    git clone https://github.com/tdewolff/minify.git
    cd minify
    make install

and the `minify` command will be in `$GOPATH/bin` or `$HOME/go/bin`.

If you do not have `make`, instead run the following lines to install `minify` and enable bash tab completion:

    go install ./cmd/minify
    source minify_bash_tab_completion

### Arch Linux
Using `yay`:

```
yay -Syu minify
```

[Minify in AUR](https://aur.archlinux.org/packages/minify/)

### FreeBSD
```
pkg install minify
```

### MacOS
Using Homebrew:

```
brew install tdewolff/tap/minify
```

[Brew tap](https://github.com/tdewolff/homebrew-tap/)

### Ubuntu
Update the package index:

```
sudo apt-get update
```

Install minify deb package:
```
sudo apt-get install minify
```

## Usage
    Usage: minify [options] [input]

    Options:
      -a, --all                              Minify all files, including hidden files and files in hidden directories
      -b, --bundle                           Bundle files by concatenation into a single file
          --cpuprofile string                Export CPU profile
          --css-precision int                Number of significant digits to preserve in numbers, 0 is all (default 0)
      -h, --help                             Show usage
          --html-keep-comments               Preserve all comments
          --html-keep-conditional-comments   Preserve all IE conditional comments
          --html-keep-default-attrvals       Preserve default attribute values
          --html-keep-document-tags          Preserve html, head and body tags
          --html-keep-end-tags               Preserve all end tags
          --html-keep-quotes                 Preserve quotes around attribute values
          --html-keep-whitespace             Preserve whitespace characters but still collapse multiple into one
          --js-precision int                 Number of significant digits to preserve in numbers, 0 is all (default 0)
          --js-keep-var-names                Preserve original variable names
          --json-precision int               Number of significant digits to preserve in numbers, 0 is all (default 0)
      -l, --list                             List all accepted filetypes
          --match string                     Filename pattern matching using regular expressions
          --memprofile string                Export memory profile
          --mime string                      Mimetype (eg. text/css), optional for input filenames, has precedence over -type
      -o, --output string                    Output file or directory (must have trailing slash), leave blank to use stdout
      -r, --recursive                        Recursively minify directories
          --svg-precision int                Number of significant digits to preserve in numbers, 0 is all (default 0)
	  -s, --sync                             Copy all files to destination directory and minify when filetype matches
          --type string                      Filetype (eg. css), optional for input filenames
          --url string                       URL of file to enable URL minification
      -v, --verbose                          Verbose
          --version                          Version
      -w, --watch                            Watch files and minify upon changes
          --xml-keep-whitespace              Preserve whitespace characters but still collapse multiple into one

    Input:
      Files or directories, leave blank to use stdin

### Types

	css     text/css
	htm     text/html
	html    text/html
	js      application/javascript
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
$ minify --type=html -o index-min.tpl index.tpl
```

You need to set the type or the mimetype option when using standard input:
```sh
$ minify --mime=application/javascript < script.js > script-min.js

$ cat script.js | minify --type=js > script-min.js
```

### Directories
You can also give directories as input, and these directories can be minified recursively.

Minify files in the current working directory to **out/** (no subdirectories):
```sh
$ minify -o out/ *
```

Minify files recursively in **src/**:
```sh
$ minify -r -o out/ src
```

Minify only javascript files in **src/**:
```sh
$ minify -r -o out/ --match=\.js src
```

### Concatenate
When multiple inputs are given and the output is either standard output or a single file, it will concatenate the files together if you use the bundle option.

Concatenate **one.css** and **two.css** into **style.css**:
```sh
$ minify -b -o style.css one.css two.css
```

Concatenate all files in **styles/** into **style.css**:
```sh
$ minify -r -b -o style.css styles
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

Minify files in **src/** and subdirectories to **out/** and watch changes:
```sh
$ minify -w -r -o out/ src
```
