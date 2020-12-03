// +build gofuzz
package fuzz

import (
	"bytes"
	"io/ioutil"

	"github.com/alex-bacart/minify/v2"
	"github.com/alex-bacart/minify/v2/js"
)

// Fuzz is a fuzz test.
func Fuzz(data []byte) int {
	r := bytes.NewBuffer(data)
	_ = js.Minify(minify.New(), ioutil.Discard, r, nil)
	return 1
}
