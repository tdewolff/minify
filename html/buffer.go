package html // import "github.com/tdewolff/minify/html"

import "github.com/tdewolff/parse/html"

// Token is a single token unit with an attribute value (if given) and hash of the data.
type Token struct {
	html.TokenType
	Hash    html.Hash
	Data    []byte
	Text    []byte
	AttrVal []byte
	Traits  traits
}

// TokenBuffer is a buffer that allows for token look-ahead.
type TokenBuffer struct {
	l *html.Lexer

	buf []Token
	pos int

	prevN int
}

// NewTokenBuffer returns a new TokenBuffer.
func NewTokenBuffer(l *html.Lexer) *TokenBuffer {
	return &TokenBuffer{
		l:   l,
		buf: make([]Token, 0, 8),
	}
}

func (z *TokenBuffer) read(t *Token) {
	t.TokenType, t.Data = z.l.Next()
	t.Text = z.l.Text()
	if t.TokenType == html.AttributeToken {
		t.AttrVal = z.l.AttrVal()
		t.Hash = html.ToHash(t.Text)
		t.Traits = attrMap[t.Hash]
	} else if t.TokenType == html.StartTagToken || t.TokenType == html.EndTagToken {
		t.AttrVal = nil
		t.Hash = html.ToHash(t.Text)
		t.Traits = tagMap[t.Hash]
	} else {
		t.AttrVal = nil
		t.Hash = 0
		t.Traits = 0
	}
}

// Peek returns the ith element and possibly does an allocation.
// Peeking past an error will panic.
func (z *TokenBuffer) Peek(pos int) *Token {
	pos += z.pos
	if pos >= len(z.buf) {
		if len(z.buf) > 0 && z.buf[len(z.buf)-1].TokenType == html.ErrorToken {
			return &z.buf[len(z.buf)-1]
		}

		c := cap(z.buf)
		d := len(z.buf) - z.pos
		p := pos - z.pos + 1 // required peek length
		var buf []Token
		if 2*p > c {
			buf = make([]Token, 0, 2*c+p)
		} else {
			buf = z.buf
		}
		copy(buf[:d], z.buf[z.pos:])

		buf = buf[:p]
		pos -= z.pos
		for i := d; i < p; i++ {
			z.read(&buf[i])
			if buf[i].TokenType == html.ErrorToken {
				buf = buf[:i+1]
				pos = i
				break
			}
		}
		z.pos, z.buf = 0, buf
	}
	return &z.buf[pos]
}

// Shift returns the first element and advances position.
func (z *TokenBuffer) Shift() *Token {
	z.l.Free(z.prevN)
	if z.pos >= len(z.buf) {
		t := &z.buf[:1][0]
		z.read(t)
		z.prevN = len(t.Data)
		return t
	}
	t := &z.buf[z.pos]
	z.pos++
	z.prevN = len(t.Data)
	return t
}
