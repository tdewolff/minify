package js // import "github.com/tdewolff/minify/js"

import (
	"bytes"
	"os"
	"testing"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/test"
)

func TestJS(t *testing.T) {
	var jsTests = []struct {
		js       string
		expected string
	}{
		{"/*comment*/", ""},
		{"// comment\na", "a"},
		{"/*! bang  comment */", "/*!bang comment*/"},
		{"function x(){}", "function x(){}"},
		{"function x(a, b){}", "function x(a,b){}"},
		{"a  b", "a b"},
		{"a\n\nb", "a\nb"},
		{"a// comment\nb", "a\nb"},
		{"''\na", "''\na"},
		{"''\n''", "''''"},
		{"]\n0", "]\n0"},
		{"a\n{", "a\n{"},
		{";\na", ";a"},
		{",\na", ",a"},
		{"}\na", "}\na"},
		{"+\na", "+\na"},
		{"+\n(", "+\n("},
		{"+\n\"\"", "+\"\""},
		{"a + ++b", "a+ ++b"},                                          // JSMin caution
		{"var a=/\\s?auto?\\s?/i\nvar", "var a=/\\s?auto?\\s?/i\nvar"}, // #14
		{"`\n", "`"}, // go fuzz
	}

	m := minify.New()
	for _, tt := range jsTests {
		r := bytes.NewBufferString(tt.js)
		w := &bytes.Buffer{}
		test.Minify(t, tt.js, Minify(m, w, r, nil), w.String(), tt.expected)
	}
}

func TestReaderErrors(t *testing.T) {
	m := minify.New()
	r := test.NewErrorReader(0)
	w := &bytes.Buffer{}
	test.Error(t, Minify(m, w, r, nil), test.ErrPlain, "return error at first read")
}

func TestWriterErrors(t *testing.T) {
	var errorTests = []struct {
		js string
		n  []int
	}{
		//01 2345
		{"a\n{5 5", []int{0, 1, 4}},
		{`/*!comment*/`, []int{0, 1, 2}},
	}

	m := minify.New()
	for _, tt := range errorTests {
		for _, n := range tt.n {
			r := bytes.NewBufferString(tt.js)
			w := test.NewErrorWriter(n)
			test.Error(t, Minify(m, w, r, nil), test.ErrPlain, "return error at write ", n)
		}
	}
}

////////////////////////////////////////////////////////////////

func ExampleMinify() {
	m := minify.New()
	m.AddFunc("text/javascript", Minify)

	if err := m.Minify("text/javascript", os.Stdout, os.Stdin); err != nil {
		panic(err)
	}
}
