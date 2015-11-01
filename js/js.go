// Package js minifies ECMAScript5.1 following the specifications at http://www.ecma-international.org/ecma-262/5.1/.
package js // import "github.com/tdewolff/minify/js"

import (
	"io"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/parse/js"
)

var (
	spaceBytes   = []byte(" ")
	newlineBytes = []byte("\n")
)

// Minify minifies JS data, it reads from r and writes to w.
func Minify(_ minify.Minifier, _ string, w io.Writer, r io.Reader) error {
	l := js.NewLexer(r)
	prev := js.LineTerminatorToken
	prevLast := byte(' ')
	lineTerminatorQueued := false
	whitespaceQueued := false
	for {
		tt, text, n := l.Next()
		l.Free(n)
		if tt == js.ErrorToken {
			if l.Err() != io.EOF {
				return l.Err()
			}
			return nil
		} else if tt == js.LineTerminatorToken {
			lineTerminatorQueued = true
		} else if tt == js.WhitespaceToken {
			whitespaceQueued = true
		} else if tt != js.CommentToken {
			first := text[0]
			if (prev == js.IdentifierToken || prev == js.NumericToken || prev == js.PunctuatorToken || prev == js.StringToken || prev == js.RegexpToken) && (tt == js.IdentifierToken || tt == js.NumericToken || tt == js.PunctuatorToken || tt == js.RegexpToken) {
				if lineTerminatorQueued && (prev != js.PunctuatorToken || prevLast == '}' || prevLast == ']' || prevLast == ')' || prevLast == '+' || prevLast == '-' || prevLast == '"' || prevLast == '\'') && (tt != js.PunctuatorToken || first == '{' || first == '[' || first == '(' || first == '+' || first == '-') {
					if _, err := w.Write(newlineBytes); err != nil {
						return err
					}
				} else if whitespaceQueued && (prev != js.StringToken && prev != js.PunctuatorToken && tt != js.PunctuatorToken || (prevLast == '+' || prevLast == '-') && first == prevLast) {
					if _, err := w.Write(spaceBytes); err != nil {
						return err
					}
				}
			}
			if _, err := w.Write(text); err != nil {
				return err
			}
			prev = tt
			prevLast = text[len(text)-1]
			lineTerminatorQueued = false
			whitespaceQueued = false
		}
	}
}
