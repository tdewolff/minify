package html // import "github.com/tdewolff/minify/html"

import (
	"github.com/tdewolff/parse"
	"github.com/tdewolff/parse/html"
)

type Token struct {
	html.TokenType
	Data    []byte
	AttrVal []byte
	Hash    html.Hash
}

type TokenBuffer struct {
	Tokenizer *html.Tokenizer

	buf []Token
	pos int
}

func NewTokenBuffer(Tokenizer *html.Tokenizer) *TokenBuffer {
	return &TokenBuffer{
		Tokenizer: Tokenizer,
		buf:       make([]Token, 0, 8),
	}
}

func (z *TokenBuffer) Read(p []Token) int {
	for i := 0; i < len(p); i++ {
		tt, data := z.Tokenizer.Next()
		if !z.Tokenizer.IsEOF() {
			data = parse.Copy(data)
		}

		var attrVal []byte
		var hash html.Hash
		if tt == html.AttributeToken {
			attrVal = z.Tokenizer.AttrVal()
			if !z.Tokenizer.IsEOF() {
				attrVal = parse.Copy(attrVal)
			}
			hash = html.ToHash(data)
		} else if tt == html.StartTagToken || tt == html.EndTagToken {
			hash = z.Tokenizer.RawTag()
			if hash == 0 {
				hash = html.ToHash(data)
			}
		}
		p[i] = Token{tt, data, attrVal, hash}
		if tt == html.ErrorToken {
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
