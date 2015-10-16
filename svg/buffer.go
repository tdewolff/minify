package svg // import "github.com/tdewolff/minify/svg"

import (
	"github.com/tdewolff/parse/svg"
	"github.com/tdewolff/parse/xml"
)

// Token is a single token unit with an attribute value (if given) and hash of the data.
type Token struct {
	xml.TokenType
	Hash    svg.Hash
	Data    []byte
	AttrVal []byte
	n       int
}

// TokenBuffer is a buffer that allows for token look-ahead.
type TokenBuffer struct {
	l *xml.Lexer

	buf []Token
	pos int
}

// NewTokenBuffer returns a new TokenBuffer.
func NewTokenBuffer(l *xml.Lexer) *TokenBuffer {
	return &TokenBuffer{
		l:   l,
		buf: make([]Token, 0, 8),
	}
}

func (z *TokenBuffer) read(t *Token) {
	t.TokenType, t.Data, t.n = z.l.Next()
	if t.TokenType == xml.AttributeToken {
		t.AttrVal = z.l.AttrVal()
		t.Hash = svg.ToHash(t.Data)
	} else if t.TokenType == xml.StartTagToken || t.TokenType == xml.EndTagToken {
		t.AttrVal = nil
		t.Hash = svg.ToHash(t.Data)
	} else {
		t.AttrVal = nil
		t.Hash = 0
	}
}

// Peek returns the ith element and possibly does an allocation.
// Peeking past an error will panic.
func (z *TokenBuffer) Peek(pos int) *Token {
	pos += z.pos
	if pos >= len(z.buf) {
		if len(z.buf) > 0 && z.buf[len(z.buf)-1].TokenType == xml.ErrorToken {
			return &z.buf[len(z.buf)-1]
		}

		c := cap(z.buf)
		p := pos - z.pos + 1
		var buf []Token
		if 2*p > c {
			buf = make([]Token, 0, 2*c+p)
		} else {
			buf = z.buf
		}
		d := len(z.buf) - z.pos
		copy(buf[:d], z.buf[z.pos:])

		buf = buf[:p]
		for i := d; i < p; i++ {
			z.read(&buf[i])
			if buf[i].TokenType == xml.ErrorToken {
				p = i + 1
				break
			}
		}
		pos = p - 1
		z.pos, z.buf = 0, buf[:p]
	}
	return &z.buf[pos]
}

// Shift returns the first element and advances position.
func (z *TokenBuffer) Shift() *Token {
	if z.pos >= len(z.buf) {
		t := &z.buf[:1][0]
		z.read(t)
		z.l.Free(t.n)
		return t
	}
	t := &z.buf[z.pos]
	z.l.Free(t.n)
	z.pos++
	return t
}
