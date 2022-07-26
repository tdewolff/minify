// +build gofuzz
package fuzz

import (
	"github.com/ezoic/parse/css"
)

func Fuzz(data []byte) int {
	_ = css.IsIdent(data)
	return 1
}
