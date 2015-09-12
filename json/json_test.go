package json // import "github.com/tdewolff/minify/json"

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tdewolff/minify"
)

func assertJSON(t *testing.T, input, expected string) {
	m := minify.New()
	b := &bytes.Buffer{}
	assert.Nil(t, Minify(m, b, bytes.NewBufferString(input), "application/json", nil), "Minify must not return error in "+input)
	assert.Equal(t, expected, b.String(), "Minify must give expected result in "+input)
}

func TestJSON(t *testing.T) {
	assertJSON(t, "{ \"a\": [1, 2] }", "{\"a\":[1,2]}")
	assertJSON(t, "[{ \"a\": [{\"x\": null}, true] }]", "[{\"a\":[{\"x\":null},true]}]")
	assertJSON(t, "{ \"a\": 1, \"b\": 2 }", "{\"a\":1,\"b\":2}")
}

////////////////////////////////////////////////////////////////

func ExampleMinify() {
	m := minify.New()
	m.AddFuncRegexp(regexp.MustCompile("[/+]json$"), Minify)

	if err := m.Minify(os.Stdout, os.Stdin, "application/json", nil); err != nil {
		fmt.Println("minify.Minify:", err)
	}
}
