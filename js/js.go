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
	lineTerminatorQueued := false
	whitespaceQueued := false
	prev := js.LineTerminatorToken
	prevLast := byte(' ')
	for {
		tt, text := l.Next()
		if tt == js.ErrorToken {
			if l.Err() != io.EOF {
				return l.Err()
			}
			return nil
		} else if tt == js.CommentToken {
			continue
		}

		if tt == js.LineTerminatorToken {
			lineTerminatorQueued = true
			continue
		} else if tt == js.WhitespaceToken {
			whitespaceQueued = true
			continue
		} else {
			first := text[0]
			if (prev == js.IdentifierToken || prev == js.NumericToken || prev == js.PunctuatorToken || prev == js.StringToken || prev == js.RegexpToken) && (tt == js.IdentifierToken || tt == js.NumericToken || tt == js.PunctuatorToken || tt == js.RegexpToken) {
				if lineTerminatorQueued && (tt != js.PunctuatorToken || first == '{' || first == '[' || first == '(' || first == '+' || first == '-') && (prev != js.PunctuatorToken || prevLast == '}' || prevLast == ']' || prevLast == ')' || prevLast == '+' || prevLast == '-' || prevLast == '"' || prevLast == '\'') {
					if _, err := w.Write(newlineBytes); err != nil {
						return err
					}
				} else if whitespaceQueued && (prev != js.StringToken && prev != js.PunctuatorToken && tt != js.PunctuatorToken || first == prevLast && (prevLast == '+' || prevLast == '-')) {
					if _, err := w.Write(spaceBytes); err != nil {
						return err
					}
				}
			}
			lineTerminatorQueued = false
			whitespaceQueued = false
			prev = tt
			prevLast = text[len(text)-1]
		}
		if _, err := w.Write(text); err != nil {
			return err
		}
	}
}
