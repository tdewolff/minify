package svg // import "github.com/tdewolff/minify/svg"

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
)

func assertSVG(t *testing.T, input, expected string) {
	m := minify.New()
	m.AddFunc("image/svg+xml", Minify)
	b := &bytes.Buffer{}
	assert.Nil(t, m.Minify("image/svg+xml", b, bytes.NewBufferString(input)), "Minify must not return error in "+input)
	assert.Equal(t, expected, b.String(), "Minify must give expected result in "+input)
}

////////////////////////////////////////////////////////////////

func TestSVG(t *testing.T) {
	assertSVG(t, `<!-- comment -->`, ``)
	assertSVG(t, `<!DOCTYPE foo SYSTEM "Foo.dtd">`, ``)
	assertSVG(t, `<?xml version="1.0" ?>`, ``)
	assertSVG(t, `<style> <![CDATA[ x ]]> </style>`, `<style>x</style>`)
	assertSVG(t, `<svg version="1.0"></svg>`, ``)
	assertSVG(t, `<path x=" a "/>`, `<path x="a"/>`)
	assertSVG(t, "<path x=\" a \n b \"/>", "<path x=\"a b\"/>")
	assertSVG(t, `<path x="5.0px" y="0%"/>`, `<path x="5" y="0"/>`)
	assertSVG(t, `<svg viewBox="5.0px 5px 240 0.10"><path/></svg>`, `<svg viewBox="5 5 240 .1"><path/></svg>`)
	assertSVG(t, `<path d="M 100 100 L 300 100 L 200 100 z"/>`, `<path d="M100 100L300 100 200 100z"/>`)
	assertSVG(t, `<path d="M100 -100M200 300z"/>`, `<path d="M100-100 200 300z"/>`)
	assertSVG(t, `<path d="M0.5 0.6 M -100 0.5z"/>`, `<path d="M.5.6-100 .5z"/>`)
	assertSVG(t, `<path d="M01.0 0.6 z"/>`, `<path d="M1 .6z"/>`)
	assertSVG(t, `<?xml version="1.0" encoding="utf-8"?>`, ``)
	assertSVG(t, `<svg viewbox="0 0 16 16"><path/></svg>`, `<svg viewbox="0 0 16 16"><path/></svg>`)
	assertSVG(t, `<g></g>`, ``)
	assertSVG(t, `<path fill="#ffffff"/>`, `<path fill="#fff"/>`)
	//assertSVG(t, `<line x1="5" y1="10" x2="20" y2="40"/>`, `<path d="M5 10l15 30"/>`)
	//assertSVG(t, `<rect x="5" y="10" width="20" height="40"/>`, `<path d="M5 10h20v40H5z"/>`)
	assertSVG(t, `<svg contentStyleType="text/json ; charset=iso-8859-1"><style>{a : true}</style></svg>`, `<svg contentStyleType="text/json;charset=iso-8859-1"><style>{a : true}</style></svg>`)
	assertSVG(t, `<metadata><dc:title /></metadata>`, ``)

	// from SVGO
	assertSVG(t, `<!DOCTYPE bla><?xml?><!-- comment --><metadata/>`, ``)
}

////////////////////////////////////////////////////////////////

func ExampleMinify() {
	m := minify.New()
	m.AddFunc("image/svg+xml", Minify)
	m.AddFunc("text/css", css.Minify)

	if err := m.Minify("image/svg+xml", os.Stdout, os.Stdin); err != nil {
		fmt.Println("minify.Minify:", err)
	}
}
