// +build gofuzz
package fuzz

import (
	"github.com/alex-bacart/minify/v2"
	"github.com/tdewolff/parse/v2"
)

func Fuzz(data []byte) int {
	m := minify.New()
	data = minify.DataURI(m, data)
	return 1
}
