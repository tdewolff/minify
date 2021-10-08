// +build gofuzz
package fuzz

import (
	"bytes"
	"io/ioutil"

	"github.com/ezoic/minify/v2"
	"github.com/ezoic/minify/v2/css"
)

func Fuzz(data []byte) int {
	r := bytes.NewBuffer(data)
	_ = css.Minify(minify.New(), ioutil.Discard, r, nil)
	return 1
}
