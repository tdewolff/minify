package minify

import (
	"testing"

	"github.com/tdewolff/test"
)

func TestMinify(t *testing.T) {
	css, err := CSS(`a { color: blue; }`)
	test.Error(t, err)
	test.String(t, css, `a{color:blue}`)

	html, err := HTML(`<!doctype html><html><head><title>Title</title></head><body><p style="color: #ff0000;"> Text </p></body></html>`)
	test.Error(t, err)
	test.String(t, html, `<!doctype html><title>Title</title><p style=color:red>Text`)

	svg, err := SVG(`<svg xmlns="http://www.w3.org/2000/svg" width="200" height="100"><path d="M 0,0 L 10, 0 z" style="color: #00ff00;"/></svg>`)
	test.Error(t, err)
	test.String(t, svg, `<svg xmlns="http://www.w3.org/2000/svg" width="200" height="100"><path d="M0 0H10z" style="color:#0f0"/></svg>`)

	js, err := JS(`var a = 5.0;`)
	test.Error(t, err)
	test.String(t, js, `var a=5`)

	json, err := JSON(`{"key" : 5.00}`)
	test.Error(t, err)
	test.String(t, json, `{"key":5}`)

	xml, err := XML(`<note> text </note>`)
	test.Error(t, err)
	test.String(t, xml, `<note>text</note>`)
}
