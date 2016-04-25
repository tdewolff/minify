package json // import "github.com/tdewolff/minify/json"

import (
	"bytes"
	"os"
	"regexp"
	"testing"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/test"
)

func TestJSON(t *testing.T) {
	jsonTests := []struct {
		json     string
		expected string
	}{
		{"{ \"a\": [1, 2] }", "{\"a\":[1,2]}"},
		{"[{ \"a\": [{\"x\": null}, true] }]", "[{\"a\":[{\"x\":null},true]}]"},
		{"{ \"a\": 1, \"b\": 2 }", "{\"a\":1,\"b\":2}"},
	}

	m := minify.New()
	for _, tt := range jsonTests {
		r := bytes.NewBufferString(tt.json)
		w := &bytes.Buffer{}
		test.Minify(t, tt.json, Minify(m, w, r, nil), w.String(), tt.expected)
	}
}

func TestReaderErrors(t *testing.T) {
	m := minify.New()
	r := test.NewErrorReader(0)
	w := &bytes.Buffer{}
	test.Error(t, Minify(m, w, r, nil), test.ErrPlain, "return error at first read")
}

func TestWriterErrors(t *testing.T) {
	errorTests := []struct {
		json string
		n    []int
	}{
		//01    234  56  78
		{`{"key":[100,200]}`, []int{0, 1, 2, 3, 4, 5, 7, 8}},
	}

	m := minify.New()
	for _, tt := range errorTests {
		for _, n := range tt.n {
			r := bytes.NewBufferString(tt.json)
			w := test.NewErrorWriter(n)
			test.Error(t, Minify(m, w, r, nil), test.ErrPlain, "return error at write", n, "in", tt.json)
		}
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
