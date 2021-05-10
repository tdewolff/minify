// +build gofuzz

package fuzz

import "github.com/tdewolff/minify/v2/parse"

// Fuzz is a fuzz test.
func Fuzz(data []byte) int {
	_ = parse.Number(data)
	return 1
}
