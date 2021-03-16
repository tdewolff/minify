// +build gofuzz

package fuzz

import (
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/parse/v2"
)

// Fuzz is a fuzz test.
func Fuzz(data []byte) int {
	data = parse.Copy(data) // ignore const-input error for OSS-Fuzz
	_ = minify.DataURI(minify.New(), data)
	return 1
}
