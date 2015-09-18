// Package js minifies ECMAScript5.1 following the specifications at http://www.ecma-international.org/ecma-262/5.1/.
package js // import "github.com/tdewolff/minify/js"

import (
	"fmt"
	"io"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/parse"
	"github.com/tdewolff/parse/js"
)

var (
	spaceBytes     = []byte(" ")
	newlineBytes   = []byte("\n")
	semicolonBytes = []byte(";")
)

var varNames = []byte{'a', 'g', 'h', 'j', 'k', 'l', 'm', 'o', 'p', 'q', 'x', 'y', 'z'}

const (
	DefaultState int = iota
	VarState
	FunctionState
	ArgumentState
)

// Minify minifies JS data, it reads from r and writes to w.
func Minify(_ minify.Minifier, _ string, w io.Writer, r io.Reader) error {
	l := js.NewLexer(r)
	prev := js.LineTerminatorToken
	prevLast := byte(' ')
	lineTerminatorQueued := false
	whitespaceQueued := false
	semicolonQueued := false

	state := DefaultState
	level := 0
	renames := 0
	varRename := []map[string][]byte{map[string][]byte{}}

	for {
		tt, text := l.Next()
		if tt == js.ErrorToken {
			if l.Err() != io.EOF {
				return l.Err()
			}
			return nil
		} else if tt == js.LineTerminatorToken {
			lineTerminatorQueued = true
			state = DefaultState
		} else if tt == js.WhitespaceToken {
			whitespaceQueued = true
		} else if tt == js.PunctuatorToken && text[0] == ';' {
			semicolonQueued = true
			state = DefaultState
		} else if tt != js.CommentToken {
			first := text[0]
			if tt == js.PunctuatorToken {
				if first == '{' {
					level++
					varRename = append(varRename, map[string][]byte{})
				} else if first == '}' && level > 0 {
					level--
					varRename = varRename[:len(varRename)-1]
				} else if first == '(' && state == FunctionState {
					state = ArgumentState
				} else if first == ')' && state == ArgumentState {
					state = DefaultState
				}
			} else if tt == js.IdentifierToken {
				if state == VarState || state == ArgumentState {
					if (level > 0 || state == ArgumentState) && len(text) > 1 {
						oldText := string(text)
						newText := []byte{varNames[renames]}
						fmt.Println("set:", oldText)
						varRename[level][oldText] = newText
						text = newText
						renames++
					}
				} else if len(text) == 3 && parse.Equal(text, []byte("var")) {
					state = VarState
				} else if len(text) == 8 && parse.Equal(text, []byte("function")) {
					state = FunctionState
				} else {
					oldText := string(text)
					fmt.Println("get:", oldText)
					for i := level; i > -1; i-- {
						fmt.Println(varRename[i])
						if newText, ok := varRename[i][oldText]; ok {
							text = newText
							break
						}
					}
				}
			}

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

			prev = tt
			prevLast = text[len(text)-1]
			lineTerminatorQueued = false
			whitespaceQueued = false
			if semicolonQueued && (tt != js.PunctuatorToken || first != '}') {
				if _, err := w.Write(semicolonBytes); err != nil {
					return err
				}
				semicolonQueued = false
			}

			if _, err := w.Write(text); err != nil {
				return err
			}
		}
	}
}
