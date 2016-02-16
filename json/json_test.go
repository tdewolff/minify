package json // import "github.com/tdewolff/minify/json"

import (
	"bytes"
	"os"
	"regexp"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/test"
)

func TestJSON(t *testing.T) {
	var jsonTests = []struct {
		json     string
		expected string
	}{
		{"{ \"a\": [1, 2] }", "{\"a\":[1,2]}"},
		{"[{ \"a\": [{\"x\": null}, true] }]", "[{\"a\":[{\"x\":null},true]}]"},
		{"{ \"a\": 1, \"b\": 2 }", "{\"a\":1,\"b\":2}"},
	}

	m := minify.New()
	for _, tt := range jsonTests {
		b := &bytes.Buffer{}
		assert.Nil(t, Minify(m, b, bytes.NewBufferString(tt.json), nil), "Minify must not return error in "+tt.json)
		assert.Equal(t, tt.expected, b.String(), "Minify must give expected result in "+tt.json)
	}
}

func TestReaderErrors(t *testing.T) {
	m := minify.New()
	r := test.NewErrorReader(0)
	w := &bytes.Buffer{}
	assert.Equal(t, test.ErrPlain, Minify(m, w, r, nil), "Minify must return error at first read")
}

func TestWriterErrors(t *testing.T) {
	var errorTests = []int{0, 1, 2, 3, 4, 5, 7, 8}

	m := minify.New()
	for _, n := range errorTests {
		// writes:                  01    234  56  78
		r := bytes.NewBufferString(`{"key":[100,200]}`)
		w := test.NewErrorWriter(n)
		assert.Equal(t, test.ErrPlain, Minify(m, w, r, nil), "Minify must return error at write "+strconv.FormatInt(int64(n), 10))
	}
}

////////////////////////////////////////////////////////////////

func ExampleMinify() {
	m := minify.New()
	m.AddFuncRegexp(regexp.MustCompile("[/+]json$"), Minify)

	if err := m.Minify("application/json", os.Stdout, os.Stdin); err != nil {
		panic(err)
	}
}
