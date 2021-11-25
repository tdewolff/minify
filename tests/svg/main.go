// +build gofuzz

package fuzz

import (
	"bytes"
	"io/ioutil"

	"github.com/lpha/minify/v2"
	"github.com/lpha/minify/v2/svg"
)

// Fuzz is a fuzz test.
func Fuzz(data []byte) int {
	r := bytes.NewBuffer(data)
	_ = svg.Minify(minify.New(), ioutil.Discard, r, nil)
	return 1
}
