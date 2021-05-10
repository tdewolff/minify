// +build gofuzz

package fuzz

import "github.com/tdewolff/minify/v2/parse"

// Fuzz is a fuzz test.
func Fuzz(data []byte) int {
	data = parse.Copy(data)
	_, _, _ = parse.DataURI(data)
	return 1
}
