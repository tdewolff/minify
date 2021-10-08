// +build gofuzz
package fuzz

import (
	"bytes"
	"io/ioutil"

	"github.com/ezoic/minify/v2"
	"github.com/ezoic/minify/v2/svg"
)

func Fuzz(data []byte) int {
	r := bytes.NewBuffer(data)
	_ = svg.Minify(minify.New(), ioutil.Discard, r, nil)
	return 1
}
