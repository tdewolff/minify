// +build gofuzz
package fuzz

import "github.com/ezoic/minify/v2"

func Fuzz(data []byte) int {
	m := minify.New()
	data = minify.DataURI(m, data)
	return 1
}
