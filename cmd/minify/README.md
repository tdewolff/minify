# Minify

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

Optionally, you can run `go install github.com/tdewolff/minify/v2/cmd/minify@latest` to install the latest version.

### Arch Linux
Using yay, see [AUR](https://aur.archlinux.org/packages/minify/)
```
yay -S minify
```

### FreeBSD
```
pkg install minify
```

### Alpine Linux
Enable the [community repo](https://wiki.alpinelinux.org/wiki/Enable_Community_Repository)
```
apk add minify
```

### MacOS
Using Homebrew, see [Brew tap](https://github.com/tdewolff/homebrew-tap/)
```
brew install tdewolff/tap/minify
```

### Ubuntu
```
sudo apt-get update
sudo apt-get install minify
```

Note: may be outdated

### Docker
Pull the image:

```
docker pull tdewolff/minify
```

> The `ENTRYPOINT` of the container is the `minify` command

and run the image, for example in interactive mode:

```bash
docker run -i tdewolff/minify sh -c 'echo "(function(){ if (a == false) { return 0; } else { return 1; } })();" | minify --type js'
```

which will output

```
(function(){return a==!1?0:1})()
```

## Usage
    Usage: minify [options] [input]

    Options:
      -a, --all                              Minify all files, including hidden files and files in hidden directories
      -b, --bundle                           Bundle files by concatenation into a single file
          --cpuprofile string                Export CPU profile
          --css-precision int                Number of significant digits to preserve in numbers, 0 is all
          --exclude string                   Filename exclusion pattern, excludes files from being processed
      -h, --help                             Show usage
          --html-keep-comments               Preserve all comments
          --html-keep-conditional-comments   Preserve all IE conditional comments
          --html-keep-default-attrvals       Preserve default attribute values
          --html-keep-document-tags          Preserve html, head and body tags
          --html-keep-end-tags               Preserve all end tags
          --html-keep-quotes                 Preserve quotes around attribute values
          --html-keep-whitespace             Preserve whitespace characters but still collapse multiple into one
          --include string                   Filename inclusion pattern, includes files previously excluded
          --js-keep-var-names                Preserve original variable names
          --js-precision int                 Number of significant digits to preserve in numbers, 0 is all
          --js-version int                   ECMAScript version to toggle supported optimizations (e.g. 2019, 2020), by default 0 is the latest version
          --json-keep-numbers                Preserve original numbers instead of minifying them
          --json-precision int               Number of significant digits to preserve in numbers, 0 is all
      -l, --list                             List all accepted filetypes
          --match string                     Filename matching pattern, only matching files are processed
          --memprofile string                Export memory profile
          --mime string                      Mimetype (eg. text/css), optional for input filenames, has precedence over --type
      -o, --output string                    Output file or directory (must have trailing slash), leave blank to use stdout
      -p, --preserve strings[=mode,ownership,timestamps]   Preserve options (mode, ownership, timestamps, links, all)
      -q, --quiet                            Quiet mode to suppress all output
      -r, --recursive                        Recursively minify directories
          --svg-keep-comments                Preserve all comments
          --svg-precision int                Number of significant digits to preserve in numbers, 0 is all
      -s, --sync                             Copy all files to destination directory and minify when filetype matches
          --type string                      Filetype (eg. css), optional for input filenames
          --url string                       URL of file to enable URL minification
      -v, --verbose count                    Verbose mode, set twice for more verbosity
          --version                          Version
      -w, --watch                            Watch files and minify upon changes
          --xml-keep-whitespace              Preserve whitespace characters but still collapse multiple into one

    Input:
      Files or directories, leave blank to use stdin. Specify --mime or --type to use stdin and stdout.


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
$ minify -r -o out/ --match="\.js$" src
```

A trailing slash in the source path will copy all files inside the directory, while omitting the trainling slash will copy the directory as well. Both `src/` and `src/*` are equivalent, except that the second case uses input expansion from bash and ignores hidden files starting with a dot.

A trailing slash in the destination path will write a single file into a directory instead of to a file of that name.

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
