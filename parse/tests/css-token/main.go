// +build gofuzz

package fuzz

import (
	"github.com/tdewolff/minify/v2/parse/css"
)

// Fuzz is a fuzz test.
func Fuzz(data []byte) int {
	_ = css.IsIdent(data)
	return 1
}
