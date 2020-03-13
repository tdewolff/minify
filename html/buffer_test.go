package html

import (
	"testing"

	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/html"
	"github.com/tdewolff/test"
)

func TestBuffer(t *testing.T) {
	//    0 12  3           45   6   7   8             9   0
	s := `<p><a href="//url">text</a>text<!--comment--></p>`
	z := NewTokenBuffer(html.NewLexer(parse.NewInputString(s)))

	tok := z.Shift()
	test.That(t, tok.Hash == P, "first token is <p>")
	test.That(t, z.pos == 0, "shift first token and restore position")
	test.That(t, len(z.buf) == 0, "shift first token and restore length")

	test.That(t, z.Peek(2).Hash == Href, "third token is href")
	test.That(t, z.pos == 0, "don't change position after peeking")
	test.That(t, len(z.buf) == 3, "two tokens after peeking")

	test.That(t, z.Peek(8).Hash == P, "ninth token is <p>")
	test.That(t, z.pos == 0, "don't change position after peeking")
	test.That(t, len(z.buf) == 9, "nine tokens after peeking")

	test.That(t, z.Peek(9).TokenType == html.ErrorToken, "tenth token is an error")
	test.That(t, z.Peek(9) == z.Peek(10), "tenth and eleventh tokens are EOF")
	test.That(t, len(z.buf) == 10, "ten tokens after peeking")

	_ = z.Shift()
	tok = z.Shift()
	test.That(t, tok.Hash == A, "third token is <a>")
	test.That(t, z.pos == 2, "don't change position after peeking")
}
