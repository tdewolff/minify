// +build gofuzz

package fuzz

import (
	"github.com/tdewolff/minify/v2/parse"
	"github.com/tdewolff/minify/v2/svg"
)

// Fuzz is a fuzz test.
func Fuzz(data []byte) int {
	pathDataBuffer := svg.NewPathData(&svg.Minifier{})
	data = parse.Copy(data) // ignore const-input error for OSS-Fuzz
	_ = pathDataBuffer.ShortenPathData(data)
	return 1
}
