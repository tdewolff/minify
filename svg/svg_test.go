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
		{`<style/><![CDATA[ <<<<< ]]>`, `<style/><![CDATA[ <<<<< ]]>`},
		{`<svg version="1.0"></svg>`, ``},
		{`<svg version="1.1" x="0" y="0px" width="100%" height="100%"><path/></svg>`, `<svg><path/></svg>`},
		{`<path x=" a "/>`, `<path x="a"/>`},
		{"<path x=\" a \n b \"/>", `<path x="a b"/>`},
		{`<path x="5.0px" y="0%"/>`, `<path x="5" y="0"/>`},
		{`<svg viewBox="5.0px 5px 240 0.10"><path/></svg>`, `<svg viewBox="5 5 240 .1"><path/></svg>`},
		{`<path d="M 100 100 L 300 100 L 200 100 z"/>`, `<path d="M100 100H300 200z"/>`},
		{`<path d="M100 -100M200 300z"/>`, `<path d="M100-100M200 300z"/>`},
		{`<path d="M0.5 0.6 M -100 0.5z"/>`, `<path d="M.5.6M-100 .5z"/>`},
		{`<path d="M01.0 0.6 z"/>`, `<path d="M1 .6z"/>`},
		{`<path d="M20 20l-10-10z"/>`, `<path d="M20 20 10 10z"/>`},
		{`<?xml version="1.0" encoding="utf-8"?>`, ``},
		{`<svg viewbox="0 0 16 16"><path/></svg>`, `<svg viewbox="0 0 16 16"><path/></svg>`},
		{`<g></g>`, ``},
		{`<g><path/></g>`, `<path/>`},
		{`<g id="a"><g><path/></g></g>`, `<g id="a"><path/></g>`},
		{`<path fill="#ffffff"/>`, `<path fill="#fff"/>`},
		{`<path fill='#fff'/>`, `<path fill="#fff"/>`},
		{`<line x1="5" y1="10" x2="20" y2="40"/>`, `<path d="M5 10 20 40z"/>`},
		{`<rect x="5" y="10" width="20" height="40"/>`, `<path d="M5 10h20v40H5z"/>`},
		{`<rect x="-5.669" y="147.402" fill="#843733" width="252.279" height="14.177"/>`, `<path fill="#843733" d="M-5.669 147.402H246.61v14.177H-5.669z"/>`},
		{`<polygon points="1,2 3,4"/>`, `<path d="M1 2 3 4z"/>`},
		{`<polyline points="1,2 3,4"/>`, `<path d="M1 2 3 4"/>`},
		{`<svg contentStyleType="text/json ; charset=iso-8859-1"><style>{a : true}</style></svg>`, `<svg contentStyleType="text/json;charset=iso-8859-1"><style>{a : true}</style></svg>`},
		{`<metadata><dc:title /></metadata>`, ``},

		// from SVGO
		{`<!DOCTYPE bla><?xml?><!-- comment --><metadata/>`, ``},

		{`<polygon fill="none" stroke="#000" points="-0.1,"/>`, `<polygon fill="none" stroke="#000" points="-0.1,"/>`}, // #45

		// go fuzz
		{`<0 d=09e9.6e-9e0`, `<0 d=""`},
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

// func TestSVGDecimals(t *testing.T) {
// 	var svgTests = []struct {
// 		svg      string
// 		expected string
// 	}{
// 		{`<svg x="1.234" y="0.001" width="1.001"><path/></svg>`, `<svg x="1.2" width="1"><path/></svg>`},
// 	}

// 	m := minify.New()
// 	o := &Minifier{Decimals: 1}
// 	for _, tt := range svgTests {
// 		b := &bytes.Buffer{}
// 		assert.Nil(t, o.Minify(m, b, bytes.NewBufferString(tt.svg), nil), "Minify must not return error in "+tt.svg)
// 		assert.Equal(t, tt.expected, b.String(), "Minify must give expected result in "+tt.svg)
// 	}
// }

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

////////////////////////////////////////////////////////////////

func BenchmarkGetAttributes(b *testing.B) {
	r := bytes.NewBufferString(`<rect x="0" y="1" width="2" height="3" rx="4" ry="5"/>`)
	l := xml.NewLexer(r)
	tb := NewTokenBuffer(l)
	tb.Shift()
	tb.Peek(6)
	for i := 0; i < b.N; i++ {
		tb.Attributes(svg.X, svg.Y, svg.Width, svg.Height, svg.Rx, svg.Ry)
	}
}
