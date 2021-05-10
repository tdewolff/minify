package json

import (
	"fmt"
	"io"
	"testing"

	"github.com/tdewolff/minify/v2/parse"
	"github.com/tdewolff/test"
)

type GTs []GrammarType

func TestGrammars(t *testing.T) {
	var grammarTests = []struct {
		json     string
		expected []GrammarType
	}{
		{" \t\n\r", GTs{}}, // WhitespaceGrammar
		{"null", GTs{LiteralGrammar}},
		{"[]", GTs{StartArrayGrammar, EndArrayGrammar}},
		{"15.2", GTs{NumberGrammar}},
		{"0.4", GTs{NumberGrammar}},
		{"5e9", GTs{NumberGrammar}},
		{"-4E-3", GTs{NumberGrammar}},
		{"true", GTs{LiteralGrammar}},
		{"false", GTs{LiteralGrammar}},
		{"null", GTs{LiteralGrammar}},
		{`""`, GTs{StringGrammar}},
		{`"abc"`, GTs{StringGrammar}},
		{`"\""`, GTs{StringGrammar}},
		{`"\\"`, GTs{StringGrammar}},
		{"{}", GTs{StartObjectGrammar, EndObjectGrammar}},
		{`{"a": "b", "c": "d"}`, GTs{StartObjectGrammar, StringGrammar, StringGrammar, StringGrammar, StringGrammar, EndObjectGrammar}},
		{`{"a": [1, 2], "b": {"c": 3}}`, GTs{StartObjectGrammar, StringGrammar, StartArrayGrammar, NumberGrammar, NumberGrammar, EndArrayGrammar, StringGrammar, StartObjectGrammar, StringGrammar, NumberGrammar, EndObjectGrammar, EndObjectGrammar}},
		{"[null,]", GTs{StartArrayGrammar, LiteralGrammar, EndArrayGrammar}},
	}
	for _, tt := range grammarTests {
		t.Run(tt.json, func(t *testing.T) {
			p := NewParser(parse.NewInputString(tt.json))
			i := 0
			for {
				grammar, _ := p.Next()
				if grammar == ErrorGrammar {
					test.T(t, p.Err(), io.EOF)
					test.T(t, i, len(tt.expected), "when error occurred we must be at the end")
					break
				} else if grammar == WhitespaceGrammar {
					continue
				}
				test.That(t, i < len(tt.expected), "index", i, "must not exceed expected grammar types size", len(tt.expected))
				if i < len(tt.expected) {
					test.T(t, grammar, tt.expected[i], "grammar types must match")
				}
				i++
			}
		})
	}

	// coverage
	for i := 0; ; i++ {
		if GrammarType(i).String() == fmt.Sprintf("Invalid(%d)", i) {
			break
		}
	}
	for i := 0; ; i++ {
		if State(i).String() == fmt.Sprintf("Invalid(%d)", i) {
			break
		}
	}
}

func TestGrammarsErrorEOF(t *testing.T) {
	var grammarErrorTests = []struct {
		json string
	}{
		{`{"":"`},
		{"\"a\\"},
	}
	for _, tt := range grammarErrorTests {
		t.Run(tt.json, func(t *testing.T) {
			p := NewParser(parse.NewInputString(tt.json))
			for {
				grammar, _ := p.Next()
				if grammar == ErrorGrammar {
					test.T(t, p.Err(), io.EOF)
					break
				}
			}
		})
	}
}

func TestGrammarsError(t *testing.T) {
	var grammarErrorTests = []struct {
		json string
		col  int
	}{
		{"true, false", 5},
		{"[true false]", 7},
		{"]", 1},
		{"}", 1},
		{"{0: 1}", 2},
		{"{\"a\" 1}", 6},
		{"1.", 2},
		{"1e+", 2},
		{"[true, \x00]", 8},
		{"\"string\x00\"", 8},
		{"{\"id\": noquote}", 8},
		{"{\"id\"\x00: 5}", 6},
		{"{\"id: \x00}", 7},
		{"{\"id: 5\x00", 8},
	}
	for _, tt := range grammarErrorTests {
		t.Run(tt.json, func(t *testing.T) {
			p := NewParser(parse.NewInputString(tt.json))
			for {
				grammar, _ := p.Next()
				if grammar == ErrorGrammar {
					if perr, ok := p.Err().(*parse.Error); ok {
						_, col, _ := perr.Position()
						test.T(t, col, tt.col)
					} else {
						test.Fail(t, "not a parse error:", p.Err())
					}
					break
				}
			}
		})
	}
}

func TestStates(t *testing.T) {
	var stateTests = []struct {
		json     string
		expected []State
	}{
		{"null", []State{ValueState}},
		{"[null]", []State{ArrayState, ArrayState, ValueState}},
		{"{\"\":null}", []State{ObjectKeyState, ObjectValueState, ObjectKeyState, ValueState}},
	}
	for _, tt := range stateTests {
		t.Run(tt.json, func(t *testing.T) {
			p := NewParser(parse.NewInputString(tt.json))
			i := 0
			for {
				grammar, _ := p.Next()
				state := p.State()
				if grammar == ErrorGrammar {
					test.T(t, p.Err(), io.EOF)
					test.T(t, i, len(tt.expected), "when error occurred we must be at the end")
					break
				} else if grammar == WhitespaceGrammar {
					continue
				}
				test.That(t, i < len(tt.expected), "index", i, "must not exceed expected states size", len(tt.expected))
				if i < len(tt.expected) {
					test.T(t, state, tt.expected[i], "states must match")
				}
				i++
			}
		})
	}
}

func TestOffset(t *testing.T) {
	z := parse.NewInputString(`{"key": [5, "string", null, true]}`)
	p := NewParser(z)
	test.T(t, z.Offset(), 0)
	_, _ = p.Next()
	test.T(t, z.Offset(), 1) // {
	_, _ = p.Next()
	test.T(t, z.Offset(), 7) // "key":
	_, _ = p.Next()
	test.T(t, z.Offset(), 9) // [
	_, _ = p.Next()
	test.T(t, z.Offset(), 10) // 5
	_, _ = p.Next()
	test.T(t, z.Offset(), 20) // , "string"
	_, _ = p.Next()
	test.T(t, z.Offset(), 26) // , null
	_, _ = p.Next()
	test.T(t, z.Offset(), 32) // , true
	_, _ = p.Next()
	test.T(t, z.Offset(), 33) // ]
	_, _ = p.Next()
	test.T(t, z.Offset(), 34) // }
}

////////////////////////////////////////////////////////////////

func ExampleNewParser() {
	p := NewParser(parse.NewInputString(`{"key": 5}`))
	out := ""
	for {
		state := p.State()
		gt, data := p.Next()
		if gt == ErrorGrammar {
			break
		}
		out += string(data)
		if state == ObjectKeyState && gt != EndObjectGrammar {
			out += ":"
		}
		// not handling comma insertion
	}
	fmt.Println(out)
	// Output: {"key":5}
}
