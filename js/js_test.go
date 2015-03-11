package js // import "github.com/tdewolff/minify/js"

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tdewolff/minify"
)

func assertJS(t *testing.T, input, expected string) {
	m := minify.NewMinifier()
	b := &bytes.Buffer{}
	assert.Nil(t, Minify(m, "text/javascript", b, bytes.NewBufferString(input)), "Minify must not return error in "+input)
	assert.Equal(t, expected, b.String(), "Minify must give expected result in "+input)
}

func TestCSS(t *testing.T) {
	assertJS(t, "/*comment*/", "")
	assertJS(t, "// comment\na", "a")
	assertJS(t, "function x(){}", "function x(){}")
	assertJS(t, "function x(a, b){}", "function x(a,b){}")
	assertJS(t, "a  b", "a b")
	assertJS(t, "a\n\nb", "a\nb")
	assertJS(t, "a// comment\nb", "a\nb")
	assertJS(t, "''\na", "''\na")
	assertJS(t, "''\n''", "''''")
	assertJS(t, "]\n0", "]\n0")
	assertJS(t, "a\n{", "a\n{")
	assertJS(t, ";\na", ";a")
	assertJS(t, ",\na", ",a")
	assertJS(t, "a + ++b", "a+ ++b")                                          // JSMin caution
	assertJS(t, "var a=/\\s?auto?\\s?/i\nvar", "var a=/\\s?auto?\\s?/i\nvar") // #14
}
