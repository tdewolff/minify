JavaScript bindings for the Go minifiers for web formats `minify`, see [github.com/tdewolff/minify](https://github.com/tdewolff/minify).

## Installation on Windows
THIS DOES NOT WORK UNFORTUNATELY

- Install [NPM](https://nodejs.org/en/download)
- Install [Python](https://www.python.org/downloads/) (optional?)
- Open Windows Command Prompt and run:
- `$ npm install @tdewolff/minify`

### Build from source
- Install [Git](https://git-scm.com/)
- Install [NPM](https://nodejs.org/en/download)
- Install [Python](https://www.python.org/downloads/)
- Install [Build Tools for Visual Studio](https://visualstudio.microsoft.com/downloads/#build-tools-for-visual-studio-2022) under "Tools for Visual Studio". Make sure to also enable the "Desktop development with C++", see [NodeJS - On Windows](https://github.com/nodejs/node-gyp#on-windows)
- Install [TDM-GCC](https://jmeubank.github.io/tdm-gcc/download/) and select the 64+32 bit version, this is only to provide the `mingw32-make` binary
- Install [Go](https://go.dev/doc/install)
- Open the Git Bash and run:
- `$ git clone https://github.com/tdewolff/minify`
- `$ cd minify/bindings/js`
- `$ npm install`

## Usage
There are three functions available in JavaScript: configure the minifiers, minify a string, and minify a file. Below an example of their usage:

```js
import { config, string, file } from '@tdewolff/minify';

# default config option values
config({
    'css-precision': 0,
    'html-keep-comments': false,
    'html-keep-conditional-comments': false,
    'html-keep-default-attr-vals': false,
    'html-keep-document-tags': false,
    'html-keep-end-tags': false,
    'html-keep-whitespace': false,
    'html-keep-quotes': false,
    'js-precision': 0,
    'js-keep-var-names': false,
    'js-version': 0,
    'json-precision': 0,
    'json-keep-numbers': false,
    'svg-keep-comments': false,
    'svg-precision': 0,
    'xml-keep-whitespace': false,
})

const s = string('text/html', '<span style="color:#ff0000;" class="text">Some  text</span>')
console.log(s)  // <span style=color:red class=text>Some text</span>

file('text/html', 'example.html', 'example.min.html')  // creates example.min.html
```

## Mediatypes
The first argument is the mediatype of the content. The following mediatypes correspond to the configured minifiers:

- `text/css`: CSS
- `text/html`: HTML
- `image/svg+xml`: SVG
- `(application|text)/(x-)?(java|ecma)script`: JS
- `*/json */*-json`: JSON
- `*/xml */*-xml`: XML
