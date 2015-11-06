// Package json minifies JSON following the specifications at http://json.org/.
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

////////////////////////////////////////////////////////////////

type Minifier struct{}

func Minify(m *minify.M, w io.Writer, r io.Reader, params interface{}) error {
	return (&Minifier{}).Minify(m, w, r, params)
}

// Minify minifies JSON data, it reads from r and writes to w.
func (o *Minifier) Minify(_ *minify.M, w io.Writer, r io.Reader, _ interface{}) error {
	skipComma := true
	p := json.NewParser(r)
	for {
		state := p.State()
		gt, text := p.Next()
		if gt == json.ErrorGrammar {
			if p.Err() != io.EOF {
				return p.Err()
			}
			return nil
		}

		if !skipComma && gt != json.EndObjectGrammar && gt != json.EndArrayGrammar {
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
		skipComma = gt == json.StartObjectGrammar || gt == json.StartArrayGrammar

		if gt == json.StringGrammar {
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
