package parse

import (
	"bytes"
	"testing"

	"github.com/tdewolff/test"
)

func TestError(t *testing.T) {
	err := NewError(bytes.NewBufferString("buffer"), 3, "message")

	line, column, context := err.Position()
	test.T(t, line, 1, "line")
	test.T(t, column, 4, "column")
	test.T(t, "\n"+context, "\n    1: buffer\n          ^", "context")

	test.T(t, err.Error(), "message on line 1 and column 4\n    1: buffer\n          ^", "error")
}

func TestErrorLexer(t *testing.T) {
	l := NewInputString("buffer")
	l.Move(3)
	err := NewErrorLexer(l, "message")

	line, column, context := err.Position()
	test.T(t, line, 1, "line")
	test.T(t, column, 4, "column")
	test.T(t, "\n"+context, "\n    1: buffer\n          ^", "context")

	test.T(t, err.Error(), "message on line 1 and column 4\n    1: buffer\n          ^", "error")
}

func TestErrorMessages(t *testing.T) {
	err := NewError(bytes.NewBufferString("buffer"), 3, "message %d", 5)
	test.T(t, err.Error(), "message 5 on line 1 and column 4\n    1: buffer\n          ^", "error")
}
