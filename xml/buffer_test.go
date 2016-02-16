package xml // import "github.com/tdewolff/minify/xml"

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tdewolff/parse/xml"
)

func TestBuffer(t *testing.T) {
	//    0 12  3           45   6   7   8             9   0
	s := `<p><a href="//url">text</a>text<!--comment--></p>`
	z := NewTokenBuffer(xml.NewLexer(bytes.NewBufferString(s)))

	tok := z.Shift()
	assert.Equal(t, "p", string(tok.Text), "first token must be <p>")
	assert.Equal(t, 0, z.pos, "must have shifted first token and restored position")
	assert.Equal(t, 0, len(z.buf), "must have shifted first token and restored length")

	assert.Equal(t, "href", string(z.Peek(2).Text), "third token must be href")
	assert.Equal(t, 0, z.pos, "must not have changed positon after peeking")
	assert.Equal(t, 3, len(z.buf), "must have two tokens after peeking")

	assert.Equal(t, "p", string(z.Peek(8).Text), "nineth token must be <p>")
	assert.Equal(t, 0, z.pos, "must not have changed positon after peeking")
	assert.Equal(t, 9, len(z.buf), "must have nine tokens after peeking")

	assert.Equal(t, xml.ErrorToken, z.Peek(9).TokenType, "tenth token must be error")
	assert.Equal(t, z.Peek(9), z.Peek(10), "tenth and eleventh token must both be EOF")
	assert.Equal(t, 10, len(z.buf), "must have ten tokens after peeking")

	tok = z.Shift()
	tok = z.Shift()
	assert.Equal(t, "a", string(tok.Text), "third token must be <a>")
	assert.Equal(t, 2, z.pos, "must not have changed positon after peeking")
}
