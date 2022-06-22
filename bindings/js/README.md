JavaScript bindings for the Go minifiers for web formats `minify`, see [github.com/tdewolff/minify](https://github.com/tdewolff/minify).

## Requisites
Make sure to have [Go](https://go.dev/doc/install) installed.

## Usage
There are three functions available in JavaScript: configure the minifiers, minify a string, and minify a file. Below an example of their usage:

```js
const minify = require('@tdewolff/minify');

# default config option values
minify.config({
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
    'js-no-nullish-operator': false,
    'json-precision': 0,
    'json-keep-numbers': false,
    'svg-keep-comments': false,
    'svg-precision': 0,
    'xml-keep-whitespace': false,
})

s = minify.string('text/html', '<span style="color:#ff0000;" class="text">Some  text</span>')
console.log(s)  // <span style=color:red class=text>Some text</span>

minify.file('text/html', 'example.html', 'example.min.html')  // creates example.min.html
```

## Mediatypes
The first argument is the mediatype of the content. The following mediatypes correspond to the configured minifiers:

- `text/css`: CSS
- `text/html`: HTML
- `image/svg+xml`: SVG
- `(application|text)/(x-)?(java|ecma)script`: JS
- `*/json */*-json`: JSON
- `*/xml */*-xml`: XML
