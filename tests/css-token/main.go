// +build gofuzz
package fuzz

import (
	"github.com/tdewolff/parse/v2/css"
)

func Fuzz(data []byte) int {
	_ = css.IsIdent(data)
	return 1
}
