package js // import "github.com/tdewolff/minify/js"

import (
	"bytes"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
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
	        {"}\n'use strict';", "}\n'use strict;"},
	}

	m := minify.New()
	for _, tt := range jsTests {
		b := &bytes.Buffer{}
		assert.Nil(t, Minify(m, b, bytes.NewBufferString(tt.js), nil), "Minify must not return error in "+tt.js)
		assert.Equal(t, tt.expected, b.String(), "Minify must give expected result in "+tt.js)
	}
}

func TestReaderErrors(t *testing.T) {
	m := minify.New()
	r := test.NewErrorReader(0)
	w := &bytes.Buffer{}
	assert.Equal(t, test.ErrPlain, Minify(m, w, r, nil), "Minify must return error at first read")
}

func TestWriterErrors(t *testing.T) {
	var errorTests = []int{0, 1, 4}

	m := minify.New()
	for _, n := range errorTests {
		// writes:                  01 2345
		r := bytes.NewBufferString("a\n{5 5")
		w := test.NewErrorWriter(n)
		assert.Equal(t, test.ErrPlain, Minify(m, w, r, nil), "Minify must return error at write "+strconv.FormatInt(int64(n), 10))
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
