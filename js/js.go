package js // import "github.com/tdewolff/minify/js"

import (
	"io"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/parse/js"
)

// Minify minifies CSS files, it reads from r and writes to w.
func Minify(m minify.Minifier, w io.Writer, r io.Reader) error {
	z := js.NewTokenizer(r)
	lineTerminatorQueued := false
	whitespaceQueued := false
	prev := js.LineTerminatorToken
	prevFirst := byte(' ')
	for {
		tt, text := z.Next()
		if tt == js.ErrorToken {
			if z.Err() != io.EOF {
				return z.Err()
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
			if (prev == js.IdentifierToken || prev == js.NumericToken) && (tt == js.IdentifierToken || tt == js.NumericToken) || prev == js.PunctuatorToken && tt == js.PunctuatorToken {
				if lineTerminatorQueued && (tt != js.PunctuatorToken || (prevFirst == '{' || prevFirst == '[' || prevFirst == '(' || prevFirst == '+' || prevFirst == '-') && (first == '}' || first == ']' || first == ')' || first == '+' || first == '-' || first == '"' || first == '\'')) {
					if _, err := w.Write([]byte("\n")); err != nil {
						return err
					}
				} else if whitespaceQueued && (tt != js.PunctuatorToken || first == prevFirst && (prevFirst == '+' || prevFirst == '-')) {
					if _, err := w.Write([]byte(" ")); err != nil {
						return err
					}
				}
			}
			lineTerminatorQueued = false
			whitespaceQueued = false
			prev = tt
			prevFirst = text[0]
		}
		if _, err := w.Write(text); err != nil {
			return err
		}
	}
}
