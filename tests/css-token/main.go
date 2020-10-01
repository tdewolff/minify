// +build gofuzz
package fuzz

import (
	"github.com/tdewolff/parse/v2/css"
)

// Fuzz is a fuzz test.
func Fuzz(data []byte) int {
	_ = css.IsIdent(data)
	return 1
}
