package xml // import "github.com/tdewolff/minify/xml"

import (
	"github.com/tdewolff/parse"
	"github.com/tdewolff/parse/xml"
)

type Token struct {
	xml.TokenType
	Data    []byte
	AttrVal []byte
}

type TokenBuffer struct {
	tokenizer *xml.Tokenizer

	buf []Token
	pos int
}

func NewTokenBuffer(tokenizer *xml.Tokenizer) *TokenBuffer {
	return &TokenBuffer{
		tokenizer: tokenizer,
		buf:       make([]Token, 0, 8),
	}
}

func (z *TokenBuffer) Read(p []Token) int {
	for i := 0; i < len(p); i++ {
		tt, data := z.tokenizer.Next()
		if !z.tokenizer.IsEOF() {
			data = parse.Copy(data)
		}

		var attrVal []byte
		if tt == xml.AttributeToken {
			attrVal = z.tokenizer.AttrVal()
			if !z.tokenizer.IsEOF() {
				attrVal = parse.Copy(attrVal)
			}
		}
		p[i] = Token{tt, data, attrVal}
		if tt == xml.ErrorToken {
			return i + 1
		}
	}
	return len(p)
}

// Peek returns the ith element and possibly does an allocation.
// Peeking past an error will panic.
func (z *TokenBuffer) Peek(i int) *Token {
	end := z.pos + i
	if end >= len(z.buf) {
		c := cap(z.buf)
		d := len(z.buf) - z.pos
		var buf []Token
		if 2*d > c {
			buf = make([]Token, d, 2*c)
		} else {
			buf = z.buf[:d]
		}
		copy(buf, z.buf[z.pos:])

		n := z.Read(buf[d:cap(buf)])
		end -= z.pos
		z.pos, z.buf = 0, buf[:d+n]
	}
	return &z.buf[end]
}

// Shift returns the first element and advances position.
func (z *TokenBuffer) Shift() *Token {
	t := z.Peek(0)
	z.pos++
	return t
}
