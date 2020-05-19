// +build gofuzz
package fuzz

import "github.com/tdewolff/minify/v2/svg"

func Fuzz(data []byte) int {
	pathDataBuffer := svg.NewPathData(&svg.Minifier{Decimals: -1})
	data = pathDataBuffer.ShortenPathData(data)
	return 1
}
