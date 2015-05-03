package svg // import "github.com/tdewolff/minify/svg"

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tdewolff/minify"
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
}
