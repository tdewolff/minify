Python bindings for the Go minifiers for web formats `minify`, see [github.com/tdewolff/minify](https://github.com/tdewolff/minify).

## Usage
There are two functions available in Python: minify a string and minify a file. Below an example of both:

```python
import minify

s = minify.string('text/html', '<span style="color:#ff0000;" class="text">Some  text</span>')
print(s)  # <span style=color:red class=text>Some text</span>

minify.file('text/html', 'test.html', 'test.min.html')  # creates test.min.html from test.html
```

## Mediatypes
The first argument is the mediatype of the content. The following mediatypes correspond to the configured minifiers:

- `text/css`: CSS
- `text/html`: HTML
- `image/svg+xml`: SVG
- `(application|text)/(x-)?(java|ecma)script`: JS
- `*/json */*-json`: JSON
- `*/xml */*-xml`: XML
