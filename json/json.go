// Package json minifies JSON following the specifications at http://json.org/.
package json

import (
	"io"
	"strconv"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/json"
)

var (
	commaBytes = []byte(",")
	colonBytes = []byte(":")
)

////////////////////////////////////////////////////////////////

// Minifier is a JSON minifier.
type Minifier struct {
	Precision   int  // number of significant digits
	KeepNumbers bool // prevent numbers from being minified
}

// Minify minifies JSON data, it reads from r and writes to w.
func Minify(m *minify.M, w io.Writer, r io.Reader, params map[string]string) error {
	return (&Minifier{}).Minify(m, w, r, params)
}

// Minify minifies JSON data, it reads from r and writes to w.
func (o *Minifier) Minify(_ *minify.M, w io.Writer, r io.Reader, _ map[string]string) error {
	skipComma := true

	z := parse.NewInput(r)
	defer z.Restore()

	p := json.NewParser(z)
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

		if !o.KeepNumbers && 0 < len(text) && ('0' <= text[0] && text[0] <= '9' || text[0] == '-') {
			text = minify.Number(text, o.Precision)
			if text[0] == '.' {
				text = fraction(text, false)
			} else if 1 < len(text) && text[0] == '-' && text[1] == '.' {
				text = fraction(text[1:], true)
			}
		}
		w.Write(text)
	}
}

// fraction returns a JSON number for a minified number that starts with a dot, which JSON does not allow.
// It prepends a zero, or uses exponent notation when that is shorter: Number minifies 1e-3 to .001, which
// as 0.001 is longer than 1e-3.
func fraction(num []byte, neg bool) []byte {
	digits := 1
	for digits < len(num) && num[digits] == '0' {
		digits++
	}
	exp := strconv.AppendInt(nil, int64(len(num)-1), 10)
	var b []byte
	if len(num)-digits+2+len(exp) < len(num)+1 {
		b = make([]byte, 0, 1+len(num)-digits+2+len(exp))
		if neg {
			b = append(b, '-')
		}
		b = append(b, num[digits:]...)
		b = append(b, 'e', '-')
		b = append(b, exp...)
	} else {
		b = make([]byte, 0, 2+len(num))
		if neg {
			b = append(b, '-')
		}
		b = append(b, '0')
		b = append(b, num...)
	}
	return b
}
