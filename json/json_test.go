package json // import "github.com/tdewolff/minify/json"

import (
	"bytes"
	"fmt"
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
		t.Run(tt.json, func(t *testing.T) {
			w := &bytes.Buffer{}
			err := Minify(m, w, []byte(tt.json), nil)
			test.Minify(t, tt.json, err, w.String(), tt.expected)
		})
	}
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
			t.Run(fmt.Sprint(tt.json, " ", tt.n), func(t *testing.T) {
				w := test.NewErrorWriter(n)
				err := Minify(m, w, []byte(tt.json), nil)
				test.T(t, err, test.ErrPlain)
			})
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
