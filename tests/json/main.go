//go:build gofuzz
// +build gofuzz

package fuzz

import (
	"bytes"
	"io"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/json"
)

// Fuzz is a fuzz test.
func Fuzz(data []byte) int {
	r := bytes.NewBuffer(data)
	_ = json.Minify(minify.New(), io.Discard, r, nil)
	return 1
}
