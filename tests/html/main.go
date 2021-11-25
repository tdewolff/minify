// +build gofuzz

package fuzz

import (
	"bytes"
	"io/ioutil"

	"github.com/lpha/minify/v2"
	"github.com/lpha/minify/v2/html"
)

// Fuzz is a fuzz test.
func Fuzz(data []byte) int {
	r := bytes.NewBuffer(data)
	_ = html.Minify(minify.New(), ioutil.Discard, r, nil)
	return 1
}
