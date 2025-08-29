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

### Debian / Ubuntu
```
sudo apt update
sudo apt install minify
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

    Usage: minify [options] inputs...
    
    Options:
      -a, --all                   Minify all files, including hidden files and files in hidden
                                  directories
      -b, --bundle                Bundle files by concatenation into a single file
          --css-precision int     Number of significant digits to preserve in numbers, 0 is all
          --css-version int       CSS version to toggle supported optimizations (e.g. 2), by default 0 is the latest version
          --exclude []string      Path exclusion pattern, excludes paths from being processed
          --ext map[string]string
                                  Filename extension mapping to filetype (eg. css or text/css)
      -h, --help                  Help
          --html-keep-comments    Preserve all comments
          --html-keep-conditional-comments
                                  Preserve all IE conditional comments
          --html-keep-default-attrvals
                                  Preserve default attribute values
          --html-keep-document-tags
                                  Preserve html, head and body tags
          --html-keep-end-tags    Preserve all end tags
          --html-keep-quotes      Preserve quotes around attribute values
          --html-keep-whitespace  Preserve whitespace characters but still collapse multiple into one
      -i, --inplace               Minify input files in-place instead of setting output
          --include []string      Path inclusion pattern, includes paths previously excluded
          --js-keep-var-names     Preserve original variable names
          --js-precision int      Number of significant digits to preserve in numbers, 0 is all
          --js-version int        ECMAScript version to toggle supported optimizations (e.g. 2019,
                                  2020), by default 0 is the latest version
          --json-keep-numbers     Preserve original numbers instead of minifying them
          --json-precision int    Number of significant digits to preserve in numbers, 0 is all
      -l, --list                  List all accepted filetypes
          --match []string        Filename matching pattern, only matching filenames are processed
          --mime string           Mimetype (eg. text/css), optional for input filenames (DEPRECATED, use                              --type)
      -o, --output string         Output file or directory, leave blank to use stdout
      -p, --preserve []string     Preserve options (mode, ownership, timestamps, links, all)
      -q, --quiet                 Quiet mode to suppress all output
      -r, --recursive             Recursively minify directories
      -s, --sync                  Copy all files to destination directory and minify when filetype
                                  matches
          --svg-keep-comments     Preserve all comments
          --svg-precision int     Number of significant digits to preserve in numbers, 0 is all
          --type string           Filetype (eg. css or text/css), optional when specifying inputs
          --url string            URL of file to enable URL minification
      -v, --verbose               Verbose mode, set twice for more verbosity
          --version               Version
      -w, --watch                 Watch files and minify upon changes
          --xml-keep-whitespace   Preserve whitespace characters but still collapse multiple into one
    
    Arguments:
      inputs    Input files or directories, leave blank to use stdin

### Types
Default extension mapping to mimetype (and thus minifier). Use `--ext` to add more mappings, see below for an example.

	asp          text/asp
	css          text/css
	ejs          text/x-ejs-template
	gohtml       text/x-go-template
	handlebars   text/x-handlebars-template
	htm          text/html
	html         text/html
	js           application/javascript
	json         application/json
	mjs          application/javascript
	mustache     text/x-mustache-template
	php          application/x-httpd-php
	rss          application/rss+xml
	svg          image/svg+xml
	tmpl         text/x-template
	webmanifest  application/manifest+json
	xhtml        application/xhtml+xml
	xml          text/xml

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
$ minify --type=application/javascript < script.js > script-min.js

$ cat script.js | minify --type=js > script-min.js
```

### Directories
You can also give directories as input, and these directories can be minified recursively.

Minify files in the current working directory to **out/...** (excluding subdirectories):
```sh
$ minify -o out/ *
```

Minify files recursively in **src/...** to **out/src/...**:
```sh
$ minify -r -o out/ src
```

Minify files recursively in **src/...** to **out/...**:
```sh
$ minify -r -o out/ src/
```

Minify only javascript files in **src/**:
```sh
$ minify -r -o out/ --match=*.js src/
```

A trailing slash in the source path will copy all files inside the directory, while omitting the trainling slash will copy the directory as well. Both `src/` and `src/.` are equivalent, however `src/*` uses input expansion from bash and ignores hidden files starting with a dot.

A trailing slash in the destination path forces writing into a directory. This removes ambiguity when minifying a single file which would otherwise write to a file.

#### Map custom extensions
You can map other extensions to a minifier by using the `--ext` option, which maps a filename extension to a filetype or mimetype, which is associated with a minifier.

```sh
$ minify -r -o out/ --ext.scss=text/css --ext.xjs=js src/
```
or
```sh
$ minify -r -o out/ --ext {scss:text/css xjs:js} src/
```

#### Matching and include/exclude patterns
The patterns for `--match`, `--include`, and `--exclude` can be either a glob or a regular expression. To use the latter, prefix the pattern with `~` (if you want to use a glob starting with `~`, escape the tilde `\~...`). Match only matches the base filename, while include/exclude match the full path. Be aware of bash expansion of glob patterns, which requires you to quote the pattern or escape asterisks.

Match will filters all files by the given pattern, eg. `--match '*.css'` will only minify CSS files. The `--include` and `--exclude` options allow to add or remove certain files or directories and is interpreted in the order given. For example, `minify -rvo out/ --exclude 'src/*/**' --include 'src/foo/**' src/` will minify the directory `src/`, except for `src/*/...` where `*` is not `foo`.

You may define multiple patterns within one option, such as: `--exclude '**/folder1/**' '**/folder2/**' '**/folder3/**'` Doing this might result in unexpected behaviour when it is followed immediately by the input files, as this would be interpreted as another pattern, and not as inputs. `--exclude dir_to_exclude folder_input` Instead format accordingly: `--exclude dir_to_exclude -- folder_input`.

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
