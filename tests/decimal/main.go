// +build gofuzz

package fuzz

import (
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/parse/v2"
)

// Fuzz is a fuzz test.
func Fuzz(data []byte) int {
	prec := 0
	if len(data) > 0 {
		x := data[0]
		data = data[1:]
		prec = int(x) % 20
	}
	for _, c := range data {
		if (c < '0' || c > '9') && c != '.' {
			return 0
		}
	}
	data = parse.Copy(data) // ignore const-input error for OSS-Fuzz
	data = minify.Decimal(data, prec)
	return 1
}
