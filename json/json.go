// Package json minifies JSON following the specifications at http://json.org/.
package json

import (
	"io"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/parse/v2/json"
)

var (
	commaBytes = []byte(",")
	colonBytes = []byte(":")
)

////////////////////////////////////////////////////////////////

// DefaultMinifier is the default minifier.
var DefaultMinifier = &Minifier{}

// Minifier is a JSON minifier.
type Minifier struct{}

// Minify minifies JSON data, it reads from r and writes to w.
func Minify(m *minify.M, w io.Writer, r io.Reader, params map[string]string) error {
	return DefaultMinifier.Minify(m, w, r, params)
}

// Minify minifies JSON data, it reads from r and writes to w.
func (o *Minifier) Minify(_ *minify.M, w io.Writer, r io.Reader, _ map[string]string) error {
	skipComma := true

	p := json.NewParser(r)
	defer p.Restore()

	for {
		state := p.State()
		gt, text := p.Next()
		if gt == json.ErrorGrammar {
			if _, err := w.Write(nil); err != nil {
				return err
			}
			if p.Err() != io.EOF {
				return p.Err()
			}
			return nil
		}

		if !skipComma && gt != json.EndObjectGrammar && gt != json.EndArrayGrammar {
			if state == json.ObjectKeyState || state == json.ArrayState {
				w.Write(commaBytes)
			} else if state == json.ObjectValueState {
				w.Write(colonBytes)
			}
		}
		skipComma = gt == json.StartObjectGrammar || gt == json.StartArrayGrammar

		w.Write(text)
	}
}
