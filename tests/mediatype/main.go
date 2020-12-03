// +build gofuzz
package fuzz

import (
	"github.com/alex-bacart/minify/v2"
	"github.com/tdewolff/parse"
)

func Fuzz(data []byte) int {
	data = parse.Copy(data) // ignore const-input error for OSS-Fuzz
	data = minify.Mediatype(data)
	return 1
}
