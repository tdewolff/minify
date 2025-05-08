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
		// TODO: what about x="" y="" for viewBox?
		//{`<svg width="24" height="24" viewBox="0 0 24 24"></svg>`, `<svg width="24" height="24"/>`},
		{`<path x="a"> </path>`, `<path x="a"/>`},
		{`<path x=""> </path>`, `<path x=""/>`},
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
		{`<rect x="5" y="10" width="30" height="0%"/>`, `<rect x="5" y="10" width="30" height="0"/>`},
		{`<rect x="5" y="10" width="30%" height="100%"/>`, `<rect x="5" y="10" width="30%" height="100%"/>`},
		{`<svg contentStyleType="text/json ; charset=iso-8859-1"><style>{a : true}</style></svg>`, `<svg contentStyleType="text/json;charset=iso-8859-1"><style>{a : true}</style></svg>`},
		{`<metadata><dc:title /></metadata>`, ``},
		{`<metadata><dc:title />`, ``},
		{`<foreignObject><foreignObject></foreignObject></foreignObject>`, `<foreignObject><foreignObject></foreignObject></foreignObject>`},
		{`<foreignObject>`, `<foreignObject>`},
		{`<foreignObject/>  text`, `<foreignObject/>text`},
		{`<foreignObject><foreignObject/></foreignObject>  text`, `<foreignObject><foreignObject/></foreignObject>text`},
		{`<xyz:rect width="100%" height="100%" fill="green"/>`, ``},
		{`<svg:rect width="100%" height="100%" fill="green"/>`, `<rect width="100%" height="100%" fill="green"/>`},
		{`<xlink:rect xlink:width="100%"/>`, `<xlink:rect xlink:width="100%"/>`},

		// from SVGO
		{`<!DOCTYPE bla><?xml?><!-- comment --><metadata/>`, ``},

		{`<polygon points="-0.1,"/>`, `<polygon points="-0.1,"/>`},                                   // #45
		{`<path stroke="url(#UPPERCASE)"/>`, `<path stroke="url(#UPPERCASE)"/>`},                     // #117
		{`<rect height="10"/><path/>`, `<rect height="10"/><path/>`},                                 // #244
		{`<rect height="10"><path/></rect>`, `<rect height="10"><path/></rect>`},                     // #244
		{`<foreignObject><div></div></foreignObject>`, `<foreignObject><div></div></foreignObject>`}, // #291
		{`<svg x-foo=""/>`, `<svg x-foo=""/>`},                                                       // #576

		// go fuzz
		{`<0 d=09e9.6e-9e0`, `<0 d="09e9.6e-9e0"`},
		{`<line`, `<line`},

		// file
		{`<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<svg
   viewBox="0 0 16 16"
   version="1.1"
   id="svg1"
   sodipodi:docname="onlineupdate_16_inkscape.svg"
   inkscape:version="1.4 (e7c3feb100, 2024-10-09)"
   xmlns:inkscape="http://www.inkscape.org/namespaces/inkscape"
   xmlns:sodipodi="http://sodipodi.sourceforge.net/DTD/sodipodi-0.dtd"
   xmlns="http://www.w3.org/2000/svg"
   xmlns:svg="http://www.w3.org/2000/svg">
  <defs
     id="defs1" />
  <sodipodi:namedview
     id="namedview1"
     pagecolor="#ffffff"
     bordercolor="#666666"
     borderopacity="1.0"
     inkscape:showpageshadow="2"
     inkscape:pageopacity="0.0"
     inkscape:pagecheckerboard="0"
     inkscape:deskcolor="#d1d1d1"
     inkscape:zoom="49.4375"
     inkscape:cx="8.0101138"
     inkscape:cy="8"
     inkscape:window-width="1920"
     inkscape:window-height="1000"
     inkscape:window-x="0"
     inkscape:window-y="0"
     inkscape:window-maximized="1"
     inkscape:current-layer="svg1" />
  <path
     d="m3.5 2c-.82843 0-1.5.67157-1.5 1.5 0 .6558404.4135873 1.2024109 1 1.40625v6.1875c-.5864127.203839-1 .75041-1 1.40625 0 .82843.67157 1.5 1.5 1.5s1.5-.67157 1.5-1.5c0-.65584-.4135873-1.202411-1-1.40625v-6.1875c.0771596-.0268209.1479033-.0552639.21875-.09375l2.875 2.15625c-.0599582.1621025-.09375.348305-.09375.53125 0 .82843.67157 1.5 1.5 1.5s1.5-.67157 1.5-1.5-.67157-1.5-1.5-1.5c-.2847728 0-.5544546.0809461-.78125.21875l-2.84375-2.125c.0759039-.1794577.125-.3866425.125-.59375 0-.82843-.67157-1.5-1.5-1.5zm0 1c.27614 0 .5.22386.5.5s-.22386.5-.5.5-.5-.22386-.5-.5.22386-.5.5-.5zm5 4c.27614 0 .5.22386.5.5s-.22386.5-.5.5-.5-.22386-.5-.5.22386-.5.5-.5zm2.5.78125v4.3125l-2-2.03125v1.4375l2.28125 2.28125.21875.21875.21875-.21875 2.28125-2.28125v-1.4375l-2 2.03125v-4.3125zm-7.5 4.21875c.27614 0 .5.22386.5.5s-.22386.5-.5.5-.5-.22386-.5-.5.22386-.5.5-.5z"
     fill="#eff0f1"
     id="path1" />
</svg>`, `<svg viewBox="0 0 16 16" id="svg1" xmlns="http://www.w3.org/2000/svg"><path d="m3.5 2C2.67157 2 2 2.67157 2 3.5c0 .6558404.4135873 1.2024109 1 1.40625v6.1875c-.5864127.203839-1 .75041-1 1.40625.0.82843.67157 1.5 1.5 1.5S5 13.32843 5 12.5c0-.65584-.4135873-1.202411-1-1.40625v-6.1875c.0771596-.0268209.1479033-.0552639.21875-.09375l2.875 2.15625C7.0337918 7.1308525 7 7.317055 7 7.5 7 8.32843 7.67157 9 8.5 9S10 8.32843 10 7.5 9.32843 6 8.5 6c-.2847728.0-.5544546.0809461-.78125.21875L4.875 4.09375C4.9509039 3.9142923 5 3.7071075 5 3.5 5 2.67157 4.32843 2 3.5 2zm0 1c.27614.0.5.22386.5.5s-.22386.5-.5.5-.5-.22386-.5-.5.22386-.5.5-.5zm5 4c.27614.0.5.22386.5.5s-.22386.5-.5.5-.5-.22386-.5-.5.22386-.5.5-.5zm2.5.78125v4.3125L9 10.0625V11.5l2.28125 2.28125L11.5 14l.21875-.21875L14 11.5v-1.4375l-2 2.03125v-4.3125zM3.5 12c.27614.0.5.22386.5.5s-.22386.5-.5.5-.5-.22386-.5-.5.22386-.5.5-.5z" fill="#eff0f1" id="path1"/></svg>`},
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

func TestSVGInline(t *testing.T) {
	var svgTests = []struct {
		svg      string
		expected string
	}{
		{`<svg xmlns="http://www.w3.org/2000/svg"><path/></svg>`, `<svg><path/></svg>`},
	}

	m := minify.New()
	o := &Minifier{Inline: true}
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
