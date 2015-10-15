package html // import "github.com/tdewolff/parse/html"

// Token is a single token unit with an attribute value (if given) and hash of the data.
type Token struct {
	TokenType
	Hash    Hash
	Data    []byte
	AttrVal []byte
	n       int
}

// TokenBuffer is a buffer that allows for token look-ahead.
type TokenBuffer struct {
	l *Lexer

	buf []Token
	pos int
}

// NewTokenBuffer returns a new TokenBuffer.
func NewTokenBuffer(l *Lexer) *TokenBuffer {
	return &TokenBuffer{
		l:   l,
		buf: make([]Token, 0, 8),
	}
}

func (z *TokenBuffer) read(t *Token) {
	tt, data, n := z.l.Next()
	var attrVal []byte
	var hash Hash
	if tt == AttributeToken {
		attrVal = z.l.AttrVal()
		hash = ToHash(data)
	} else if tt == StartTagToken || tt == EndTagToken {
		hash = ToHash(data)
	}
	t.TokenType = tt
	t.Data = data
	t.AttrVal = attrVal
	t.Hash = hash
	t.n = n
}

// Peek returns the ith element and possibly does an allocation.
// Peeking past an error will panic.
func (z *TokenBuffer) Peek(pos int) *Token {
	pos += z.pos
	if pos >= len(z.buf) {
		if len(z.buf) > 0 && z.buf[len(z.buf)-1].TokenType == ErrorToken {
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
			if buf[i].TokenType == ErrorToken {
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
