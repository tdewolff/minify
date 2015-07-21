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
	assertSVG(t, "<!-- comment -->", "")
	assertSVG(t, "<!DOCTYPE foo SYSTEM \"Foo.dtd\">", "")
	assertSVG(t, "<?xml version=\"1.0\" ?>", "")
	assertSVG(t, "<style> <![CDATA[ x ]]> </style>", "<style>x</style>")
	assertSVG(t, "<svg version=\"1.0\"></svg>", "<svg/>")
	assertSVG(t, "<svg x=\" a \"/>", "<svg x=\"a\"/>")
	assertSVG(t, "<svg x=\" a \n b \"/>", "<svg x=\"a b\"/>")
	assertSVG(t, "<svg x=\"5.0px\"/>", "<svg x=\"5px\"/>")
	assertSVG(t, "<path d=\"M 100 100 L 300 100 L 200 100 z\"/>", "<path d=\"M100 100L300 100 200 100z\"/>")
	assertSVG(t, "<path d=\"M100 -100M200 300z\"/>", "<path d=\"M100-100 200 300z\"/>")
	assertSVG(t, "<path d=\"M0.5 0.6 M -100 0.5z\"/>", "<path d=\"M.5.6-100 .5z\"/>")
	assertSVG(t, "<path d=\"M01.0 0.6 z\"/>", "<path d=\"M1 .6z\"/>")
	assertSVG(t, `<?xml version="1.0" encoding="utf-8"?>`, "")
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
