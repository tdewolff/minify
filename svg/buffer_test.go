package svg // import "github.com/tdewolff/minify/svg"

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tdewolff/parse/svg"
	"github.com/tdewolff/parse/xml"
)

func TestBuffer(t *testing.T) {
	//    0   12     3            4 5   6   7 8   9    01
	s := `<svg><path d="M0 0L1 1z"/>text<tag/>text</svg>`
	z := NewTokenBuffer(xml.NewLexer(bytes.NewBufferString(s)))

	tok := z.Shift()
	assert.Equal(t, svg.Svg, tok.Hash, "first token must be <svg>")
	assert.Equal(t, 0, z.pos, "must have shifted first token and restored position")
	assert.Equal(t, 0, len(z.buf), "must have shifted first token and restored position")

	assert.Equal(t, svg.D, z.Peek(2).Hash, "third token must be d")
	assert.Equal(t, 0, z.pos, "must not have changed positon after peeking")
	assert.Equal(t, 3, len(z.buf), "must have two tokens after peeking")

	assert.Equal(t, svg.Svg, z.Peek(8).Hash, "nineth token must be <svg>")
	assert.Equal(t, 0, z.pos, "must not have changed positon after peeking")
	assert.Equal(t, 9, len(z.buf), "must have nine tokens after peeking")

	assert.Equal(t, xml.ErrorToken, z.Peek(9).TokenType, "tenth token must be error")
	assert.Equal(t, z.Peek(9), z.Peek(10), "tenth and eleventh token must both be EOF")
	assert.Equal(t, 10, len(z.buf), "must have ten tokens after peeking")

	tok = z.Shift()
	tok = z.Shift()
	assert.Equal(t, svg.Path, tok.Hash, "third token must be <path>")
	assert.Equal(t, 2, z.pos, "must not have changed positon after peeking")
}
