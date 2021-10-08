// +build gofuzz
package fuzz

import (
	"bytes"
	"io/ioutil"

	"github.com/ezoic/minify/v2"
	"github.com/ezoic/minify/v2/json"
)

func Fuzz(data []byte) int {
	r := bytes.NewBuffer(data)
	_ = json.Minify(minify.New(), ioutil.Discard, r, nil)
	return 1
}
