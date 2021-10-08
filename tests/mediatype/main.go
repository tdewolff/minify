// +build gofuzz
package fuzz

import (
	"github.com/ezoic/minify/v2"
	"github.com/tdewolff/parse/v2"
)

func Fuzz(data []byte) int {
	data = parse.Copy(data) // ignore const-input error for OSS-Fuzz
	data = minify.Mediatype(data)
	return 1
}
