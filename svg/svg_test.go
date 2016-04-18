package svg // import "github.com/tdewolff/minify/svg"

import (
	"bytes"
	"os"
	"strconv"
	"testing"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/parse/svg"
	"github.com/tdewolff/parse/xml"
	"github.com/tdewolff/test"
)

func TestSVG(t *testing.T) {
	svgTests := []struct {
		svg      string
		expected string
	}{
		{`<!-- comment -->`, ``},
		{`<!DOCTYPE foo SYSTEM "Foo.dtd">`, ``},
		{`<?xml version="1.0" ?>`, ``},
		{`<style> <![CDATA[ x ]]> </style>`, `<style>x</style>`},
		{`<style> <![CDATA[ <<<< ]]> </style>`, `<style>&lt;&lt;&lt;&lt;</style>`},
		{`<style> <![CDATA[ <<<<< ]]> </style>`, `<style><![CDATA[ <<<<< ]]></style>`},
		{`<svg version="1.0"></svg>`, ``},
		{`<path x=" a "/>`, `<path x="a"/>`},
		{"<path x=\" a \n b \"/>", `<path x="a b"/>`},
		{`<path x="5.0px" y="0%"/>`, `<path x="5" y="0"/>`},
		{`<svg viewBox="5.0px 5px 240 0.10"><path/></svg>`, `<svg viewBox="5 5 240 .1"><path/></svg>`},
		{`<path d="M 100 100 L 300 100 L 200 100 z"/>`, `<path d="M100 100L300 100 200 100z"/>`},
		{`<path d="M100 -100M200 300z"/>`, `<path d="M100-100 200 300z"/>`},
		{`<path d="M0.5 0.6 M -100 0.5z"/>`, `<path d="M.5.6-100 .5z"/>`},
		{`<path d="M01.0 0.6 z"/>`, `<path d="M1 .6z"/>`},
		{`<?xml version="1.0" encoding="utf-8"?>`, ``},
		{`<svg viewbox="0 0 16 16"><path/></svg>`, `<svg viewbox="0 0 16 16"><path/></svg>`},
		{`<g></g>`, ``},
		{`<path fill="#ffffff"/>`, `<path fill="#fff"/>`},
		//{`<line x1="5" y1="10" x2="20" y2="40"/>`, `<path d="M5 10l15 30"/>`},
		//{`<rect x="5" y="10" width="20" height="40"/>`, `<path d="M5 10h20v40H5z"/>`},
		{`<svg contentStyleType="text/json ; charset=iso-8859-1"><style>{a : true}</style></svg>`, `<svg contentStyleType="text/json;charset=iso-8859-1"><style>{a : true}</style></svg>`},
		{`<metadata><dc:title /></metadata>`, ``},

		// from SVGO
		{`<!DOCTYPE bla><?xml?><!-- comment --><metadata/>`, ``},

		{`<polygon fill="none" stroke="#000" points="-0.1,"/>`, `<polygon fill="none" stroke="#000" points="-0.1,"/>`}, // #45

		// go fuzz
		{`<0 d=09e9.6e-9e0`, `<0 d="9e9 6e-10e0"`}, // TODO: fix this with the new ShortenPathdata functions
	}

	m := minify.New()
	for _, tt := range svgTests {
		r := bytes.NewBufferString(tt.svg)
		w := &bytes.Buffer{}
		test.Minify(t, tt.svg, Minify(m, w, r, nil), w.String(), tt.expected)
	}
}

func TestSVGStyle(t *testing.T) {
	svgTests := []struct {
		svg      string
		expected string
	}{
		{`<style> <![CDATA[ @media x < y {} ]]> </style>`, `<style>@media x &lt; y{}</style>`},
		{`<style> <![CDATA[ * { content: '<<<<<'; } ]]> </style>`, `<style><![CDATA[*{content:'<<<<<'}]]></style>`},
		{`<style/><![CDATA[ * { content: '<<<<<'; ]]>`, `<style/><![CDATA[ * { content: '<<<<<'; ]]>`},
	}

	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	for _, tt := range svgTests {
		r := bytes.NewBufferString(tt.svg)
		w := &bytes.Buffer{}
		test.Minify(t, tt.svg, Minify(m, w, r, nil), w.String(), tt.expected)
	}
}

func TestGetAttribute(t *testing.T) {
	r := bytes.NewBufferString(`<rect x="0" y="1" width="2" height="3" rx="4" ry="5"/>`)
	l := xml.NewLexer(r)
	tb := NewTokenBuffer(l)
	tb.Shift()
	attrs, _ := tb.Attributes(svg.X, svg.Y, svg.Width, svg.Height, svg.Rx, svg.Ry)
	for i := 0; i < 6; i++ {
		test.That(t, attrs[i] != nil, "attr must not be nil")
		val := string(attrs[i].AttrVal)
		j, _ := strconv.ParseInt(val, 10, 32)
		test.That(t, int(j) == i, "attr data is bad at position", i)
	}
}

////////////////////////////////////////////////////////////////

func ExampleMinify() {
	m := minify.New()
	m.AddFunc("image/svg+xml", Minify)
	m.AddFunc("text/css", css.Minify)

	if err := m.Minify("image/svg+xml", os.Stdout, os.Stdin); err != nil {
		panic(err)
	}
}
