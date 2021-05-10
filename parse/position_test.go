package parse

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/tdewolff/test"
)

func TestPosition(t *testing.T) {
	var newlineTests = []struct {
		offset int
		buf    string
		line   int
		col    int
	}{
		// \u2028, \u2029, and \u2318 are three bytes long
		{0, "x", 1, 1},
		{1, "xx", 1, 2},
		{2, "x\nx", 2, 1},
		{2, "\n\nx", 3, 1},
		{3, "\nxxx", 2, 3},
		{2, "\r\nx", 2, 1},
		{1, "\rx", 2, 1},
		{3, "\u2028x", 2, 1},
		{3, "\u2029x", 2, 1},

		// edge cases
		{0, "", 1, 1},
		{2, "x", 1, 2},
		{0, "\nx", 1, 1},
		{1, "\r\ny", 1, 1},
		{-1, "x", 1, 1},
		{0, "\x00a", 1, 1},
		{2, "a\x00\n", 1, 3},

		// unicode
		{1, "x\u2028x", 1, 2},
		{2, "x\u2028x", 1, 2},
		{3, "x\u2028x", 1, 2},
		{0, "x\u2318x", 1, 1},
		{1, "x\u2318x", 1, 2},
		{2, "x\u2318x", 1, 2},
		{3, "x\u2318x", 1, 2},
		{4, "x\u2318x", 1, 3},
	}
	for _, tt := range newlineTests {
		t.Run(fmt.Sprint(tt.buf, " ", tt.offset), func(t *testing.T) {
			r := bytes.NewBufferString(tt.buf)
			line, col, _ := Position(r, tt.offset)
			test.T(t, line, tt.line, "line")
			test.T(t, col, tt.col, "column")
		})
	}
}

func TestPositionContext(t *testing.T) {
	var newlineTests = []struct {
		offset  int
		buf     string
		context string
	}{
		{10, "0123456789@123456789012345678901234567890123456789012345678901234567890123456789", "0123456789@1234567890123456789012345678901234567890123456..."}, // 80 characters -> 60 characters
		{40, "0123456789012345678901234567890123456789@123456789012345678901234567890123456789", "...01234567890123456789@12345678901234567890..."},
		{60, "012345678901234567890123456789012345678901234567890123456789@12345678901234567890", "...78901234567890123456789@12345678901234567890"},
		{60, "012345678901234567890123456789012345678901234567890123456789@12345678901234567890123", "...01234567890123456789@12345678901234567890123"},
		{60, "012345678901234567890123456789012345678901234567890123456789@123456789012345678901234", "...01234567890123456789@12345678901234567890..."},
		{60, "0123456789012345678901234567890123456789ÎÎÎÎÎÎÎÎÎÎ@123456789012345678901234567890", "...0123456789ÎÎÎÎÎÎÎÎÎÎ@12345678901234567890..."},
		{60, "012345678901234567890123456789012345678912456780123456789@12345678901234567890", "...789·12·45678·0123456789@12345678901234567890"},
	}
	for _, tt := range newlineTests {
		t.Run(fmt.Sprint(tt.buf, " ", tt.offset), func(t *testing.T) {
			r := bytes.NewBufferString(tt.buf)
			_, _, context := Position(r, tt.offset)
			i := strings.IndexByte(context, '\n')
			pointer := context[i+1:]
			context = context[:i]
			test.T(t, context[7:], tt.context)

			// check if @ and ^ are at the same position
			j := strings.IndexByte(context, '@')
			k := strings.IndexByte(pointer, '^')
			test.T(t, len([]rune(pointer[:k])), len([]rune(context[:j])))
		})
	}
}
