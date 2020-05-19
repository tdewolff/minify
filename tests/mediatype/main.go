// +build gofuzz
package fuzz

import "github.com/tdewolff/minify/v2"

func Fuzz(data []byte) int {
	data = minify.Mediatype(data)
	return 1
}
