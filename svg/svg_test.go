package svg // import "github.com/tdewolff/minify/svg"

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tdewolff/minify"
)

func assertXML(t *testing.T, input, expected string) {
	m := minify.New()
	m.AddFunc("image/svg+xml", Minify)
	b := &bytes.Buffer{}
	assert.Nil(t, m.Minify("image/svg+xml", b, bytes.NewBufferString(input)), "Minify must not return error in "+input)
	assert.Equal(t, expected, b.String(), "Minify must give expected result in "+input)
}

////////////////////////////////////////////////////////////////

func TestXML(t *testing.T) {
	assertXML(t, "<!-- comment -->", "")
	assertXML(t, "<!DOCTYPE foo SYSTEM \"Foo.dtd\">", "")
	assertXML(t, "<?xml version=\"1.0\" ?>", "")
}
