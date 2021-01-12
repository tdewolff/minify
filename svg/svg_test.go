package svg

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/test"
)

func TestSVG(t *testing.T) {
	svgTests := []struct {
		svg      string
		expected string
	}{
		{`<!-- comment -->`, ``},
		{`<!DOCTYPE svg SYSTEM "foo.dtd">`, ``},
		{`<!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 1.1//EN" "foo.dtd" [ <!ENTITY x "bar"> ]>`, `<!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 1.1//EN" "foo.dtd" [ <!ENTITY x "bar"> ]>`},
		{`<!DOCTYPE svg SYSTEM "foo.dtd">`, ``},
		{`<?xml version="1.0" ?>`, ``},
		{`<style> <![CDATA[ x ]]> </style>`, `<style>x</style>`},
		{`<style> <![CDATA[ <<<< ]]> </style>`, `<style>&lt;&lt;&lt;&lt;</style>`},
		{`<style> <![CDATA[ <<<<< ]]> </style>`, `<style><![CDATA[ <<<<< ]]></style>`},
		{`<style/><![CDATA[ <<<<< ]]>`, `<style/><![CDATA[ <<<<< ]]>`},
		{`<svg version="1.0"></svg>`, `<svg version="1.0"/>`},
		{`<svg version="1.1" x="0" y="0px" width="100%" height="100%"><path/></svg>`, `<svg width="100%" height="100%"><path/></svg>`},
		{`<svg width="auto" height="auto"><path/></svg>`, `<svg width="auto" height="auto"><path/></svg>`},
		// TODO: what abour x="" y="" for viewBox?
		//{`<svg width="24" height="24" viewBox="0 0 24 24"></svg>`, `<svg width="24" height="24"/>`},
		{`<path x="a"> </path>`, `<path x="a"/>`},
		{`<path x=""> </path>`, `<path/>`},
		{`<path x=" a "/>`, `<path x="a"/>`},
		{"<path x=\" a \n b \"/>", `<path x="a b"/>`},
		{`<path x="5.0px" y="0%"/>`, `<path x="5" y="0"/>`},
		{`<svg viewBox="5.0px 5px 240IN px"><path/></svg>`, `<svg viewBox="5 5 240in px"><path/></svg>`},
		{`<svg viewBox="5.0!5px"><path/></svg>`, `<svg viewBox="5!5px"><path/></svg>`},
		{`<path d="M 100 100 L 300 100 L 200 100 z"/>`, `<path d="M1e2 1e2H3e2 2e2z"/>`},
		{`<path d="M100 -100M200 300z"/>`, `<path d="M1e2-1e2M2e2 3e2z"/>`},
		{`<path d="M0.5 0.6 M -100 0.5z"/>`, `<path d="M.5.6M-1e2.5z"/>`},
		{`<path d="M01.0 0.6 z"/>`, `<path d="M1 .6z"/>`},
		{`<path d="M20 20l-10-10z"/>`, `<path d="M20 20 10 10z"/>`},
		{`<?xml version="1.0" encoding="utf-8"?>`, ``},
		{`<svg viewbox="0 0 16 16"><path/></svg>`, `<svg viewbox="0 0 16 16"><path/></svg>`},
		{`<g></g>`, `<g/>`},
		{`<g><path/></g>`, `<g><path/></g>`},
		{`<g id="a"><g><path/></g></g>`, `<g id="a"><g><path/></g></g>`},
		{`<path fill="#ffffff"/>`, `<path fill="#fff"/>`},
		{`<path fill="#fff"/>`, `<path fill="#fff"/>`},
		{`<path fill="white"/>`, `<path fill="#fff"/>`},
		{`<path fill="#ff0000"/>`, `<path fill="red"/>`},
		{`<rect x="5" y="10" rx="2" ry="3">`, ``},
		{`<rect x="5" y="10" height="40"/>`, ``},
		{`<rect x="5" y="10" width="30" height="0%"/>`, ``},
		{`<rect x="5" y="10" width="30%" height="100%"/>`, `<rect x="5" y="10" width="30%" height="100%"/>`},
		{`<svg contentStyleType="text/json ; charset=iso-8859-1"><style>{a : true}</style></svg>`, `<svg contentStyleType="text/json;charset=iso-8859-1"><style>{a : true}</style></svg>`},
		{`<metadata><dc:title /></metadata>`, ``},
		{`<metadata><dc:title />`, ``},
		{`<foreignObject><foreignObject></foreignObject></foreignObject>`, `<foreignObject><foreignObject></foreignObject></foreignObject>`},
		{`<foreignObject>`, `<foreignObject>`},
		{`<foreignObject/>  text`, `<foreignObject/>text`},
		{`<foreignObject><foreignObject/></foreignObject>  text`, `<foreignObject><foreignObject/></foreignObject>text`},

		// from SVGO
		{`<!DOCTYPE bla><?xml?><!-- comment --><metadata/>`, ``},

		{`<polygon points="-0.1,"/>`, `<polygon points="-0.1,"/>`},                                   // #45
		{`<path stroke="url(#UPPERCASE)"/>`, `<path stroke="url(#UPPERCASE)"/>`},                     // #117
		{`<rect height="10"/><path/>`, `<path/>`},                                                    // #244
		{`<rect height="10"><path/></rect>`, ``},                                                     // #244
		{`<foreignObject><div></div></foreignObject>`, `<foreignObject><div></div></foreignObject>`}, // #291

		// go fuzz
		{`<0 d=09e9.6e-9e0`, `<0 d="09e9.6e-9e0"`},
		{`<line`, `<line`},
	}

	m := minify.New()
	for _, tt := range svgTests {
		t.Run(tt.svg, func(t *testing.T) {
			r := bytes.NewBufferString(tt.svg)
			w := &bytes.Buffer{}
			err := Minify(m, w, r, nil)
			test.Minify(t, tt.svg, err, w.String(), tt.expected)
		})
	}
}

func TestSVGStyle(t *testing.T) {
	svgTests := []struct {
		svg      string
		expected string
	}{
		{`<style> a > b {} </style>`, `<style>a>b{}</style>`},
		{`<style> <![CDATA[ @media x < y {} ]]> </style>`, `<style>@media x &lt; y{}</style>`},
		{`<style> <![CDATA[ * { content: '<<<<<'; } ]]> </style>`, `<style><![CDATA[*{content:'<<<<<'}]]></style>`},
		{`<style/><![CDATA[ * { content: '<<<<<'; ]]>`, `<style/><![CDATA[ * { content: '<<<<<'; ]]>`},
		{`<path style="fill: black; stroke: #ff0000;"/>`, `<path style="fill:#000;stroke:red"/>`},
	}

	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	for _, tt := range svgTests {
		t.Run(tt.svg, func(t *testing.T) {
			r := bytes.NewBufferString(tt.svg)
			w := &bytes.Buffer{}
			err := Minify(m, w, r, nil)
			test.Minify(t, tt.svg, err, w.String(), tt.expected)
		})
	}
}

func TestSVGPrecision(t *testing.T) {
	var svgTests = []struct {
		svg      string
		expected string
	}{
		{`<svg x="1.234" y="0.001" width="1.001"><path/></svg>`, `<svg x="1" y=".001" width="1"><path/></svg>`},
	}

	m := minify.New()
	o := &Minifier{Precision: 1}
	for _, tt := range svgTests {
		t.Run(tt.svg, func(t *testing.T) {
			r := bytes.NewBufferString(tt.svg)
			w := &bytes.Buffer{}
			err := o.Minify(m, w, r, nil)
			test.Minify(t, tt.svg, err, w.String(), tt.expected)
		})
	}
}

func TestReaderErrors(t *testing.T) {
	r := test.NewErrorReader(0)
	w := &bytes.Buffer{}
	m := minify.New()
	err := Minify(m, w, r, nil)
	test.T(t, err, test.ErrPlain, "return error at first read")
}

func TestWriterErrors(t *testing.T) {
	errorTests := []struct {
		svg string
		n   []int
	}{
		{`<!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 1.1//EN" "foo.dtd" [ <!ENTITY x "bar"> ]>`, []int{0}},
		{`abc`, []int{0}},
		{`<style>abc</style>`, []int{2}},
		{`<![CDATA[ <<<< ]]>`, []int{0}},
		{`<![CDATA[ <<<<< ]]>`, []int{0}},
		{`<path d="x"/>`, []int{0, 1, 2, 3, 4, 5}},
		{`<path></path>`, []int{1}},
		{`<svg>x</svg>`, []int{1, 3}},
		{`<svg>x</svg >`, []int{3}},
	}

	m := minify.New()
	for _, tt := range errorTests {
		for _, n := range tt.n {
			t.Run(fmt.Sprint(tt.svg, " ", tt.n), func(t *testing.T) {
				r := bytes.NewBufferString(tt.svg)
				w := test.NewErrorWriter(n)
				err := Minify(m, w, r, nil)
				test.T(t, err, test.ErrPlain)
			})
		}
	}
}

func TestMinifyErrors(t *testing.T) {
	errorTests := []struct {
		svg string
		err error
	}{
		{`<style>abc</style>`, test.ErrPlain},
		{`<style><![CDATA[abc]]></style>`, test.ErrPlain},
		{`<path style="abc"/>`, test.ErrPlain},
	}

	m := minify.New()
	m.AddFunc("text/css", func(_ *minify.M, w io.Writer, r io.Reader, _ map[string]string) error {
		return test.ErrPlain
	})
	for _, tt := range errorTests {
		t.Run(tt.svg, func(t *testing.T) {
			r := bytes.NewBufferString(tt.svg)
			w := &bytes.Buffer{}
			err := Minify(m, w, r, nil)
			test.T(t, err, tt.err)
		})
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
