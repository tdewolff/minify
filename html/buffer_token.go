package html

import (
	"github.com/tdewolff/parse"
	"github.com/tdewolff/parse/html"
)

type tokenBuffer struct {
	tokenizer *html.Tokenizer

	buf []token
	pos int
}

func newTokenBuffer(tokenizer *html.Tokenizer) *tokenBuffer {
	return &tokenBuffer{
		tokenizer: tokenizer,
		buf:       make([]token, 0, 8),
	}
}

func (z *tokenBuffer) Read(p []token) int {
	for i := 0; i < len(p); i++ {
		tt, data := z.tokenizer.Next()
		if !z.tokenizer.IsEOF() {
			data = parse.Copy(data)
		}

		var attrVal []byte
		var hash html.Hash
		if tt == html.AttributeToken {
			attrVal = z.tokenizer.AttrVal()
			if !z.tokenizer.IsEOF() {
				attrVal = parse.Copy(attrVal)
			}
			hash = html.ToHash(data)
		} else if tt == html.StartTagToken || tt == html.EndTagToken {
			hash = z.tokenizer.RawTag()
			if hash == 0 {
				hash = html.ToHash(data)
			}
		}
		p[i] = token{tt, data, attrVal, hash}

		if tt == html.ErrorToken {
			return i + 1
		}
	}
	return len(p)
}

// Peek returns the ith element and possibly does an allocation.
// Peeking past an error will panic.
func (z *tokenBuffer) Peek(i int) *token {
	end := z.pos + i
	if end >= len(z.buf) {
		c := cap(z.buf)
		d := len(z.buf) - z.pos
		var buf []token
		if 2*d > c {
			buf = make([]token, d, 2*c)
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
func (z *tokenBuffer) Shift() *token {
	t := z.Peek(0)
	z.pos++
	return t
}
