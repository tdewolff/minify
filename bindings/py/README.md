Python bindings for the Go minifiers for web formats `minify`, see [github.com/tdewolff/minify](https://github.com/tdewolff/minify).

## Requisites
Make sure to have [Go](https://go.dev/doc/install) installed.

## Usage
There are three functions available in Python: configure the minifiers, minify a string, and minify a file. Below an example of their usage:

```python
import minify

# default config option values
minify.config({
    'css-precision': 0,
    'html-keep-comments': False,
    'html-keep-conditional-comments': False,
    'html-keep-default-attr-vals': False,
    'html-keep-document-tags': False,
    'html-keep-end-tags': False,
    'html-keep-whitespace': False,
    'html-keep-quotes': False,
    'js-precision': 0,
    'js-keep-var-names': False,
    'js-version': 0,
    'json-precision': 0,
    'json-keep-numbers': False,
    'svg-keep-comments': False,
    'svg-precision': 0,
    'xml-keep-whitespace': False,
})

s = minify.string('text/html', '<span style="color:#ff0000;" class="text">Some  text</span>')
print(s)  # <span style=color:red class=text>Some text</span>

minify.file('text/html', 'example.html', 'example.min.html')  # creates example.min.html
```

## Mediatypes
The first argument is the mediatype of the content. The following mediatypes correspond to the configured minifiers:

- `text/css`: CSS
- `text/html`: HTML
- `image/svg+xml`: SVG
- `(application|text)/(x-)?(java|ecma)script`: JS
- `*/json */*-json`: JSON
- `*/xml */*-xml`: XML
