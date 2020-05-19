// +build gofuzz
package fuzz

import "github.com/tdewolff/minify/v2"

func Fuzz(data []byte) int {
	prec := 0
	if len(data) > 0 {
		x := data[0]
		data = data[1:]
		prec = int(x) % 32
	}
	data = minify.Number(data, prec)
	return 1
}
