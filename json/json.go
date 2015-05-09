// Package json is a minifier written in Go that minifies JSON following the specifications at http://json.org/.
package json // import "github.com/tdewolff/minify/json"

import (
	"io"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/parse/json"
)

var (
	quoteBytes = []byte("\"")
	commaBytes = []byte(",")
	colonBytes = []byte(":")
)

// Minify minifies JSON files, it reads from r and writes to w.
func Minify(_ minify.Minifier, _ string, w io.Writer, r io.Reader) error {
	skipComma := true
	z := json.NewTokenizer(r)
	for {
		state := z.State()
		tt, text := z.Next()
		if tt == json.ErrorToken {
			if z.Err() != io.EOF {
				return z.Err()
			}
			return nil
		}

		if !skipComma && tt != json.EndObjectToken && tt != json.EndArrayToken {
			if state == json.ObjectKeyState || state == json.ArrayState {
				if _, err := w.Write(commaBytes); err != nil {
					return err
				}
			} else if state == json.ObjectValueState {
				if _, err := w.Write(colonBytes); err != nil {
					return err
				}
			}
		}
		skipComma = tt == json.StartObjectToken || tt == json.StartArrayToken

		if tt == json.StringToken {
			if _, err := w.Write(quoteBytes); err != nil {
				return err
			}
			if _, err := w.Write(text); err != nil {
				return err
			}
			if _, err := w.Write(quoteBytes); err != nil {
				return err
			}
		} else if _, err := w.Write(text); err != nil {
			return err
		}
	}
}
