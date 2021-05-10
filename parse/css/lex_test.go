package css

import (
	"fmt"
	"io"
	"testing"

	"github.com/tdewolff/minify/v2/parse"
	"github.com/tdewolff/test"
)

type TTs []TokenType

func TestTokens(t *testing.T) {
	var tokenTests = []struct {
		css     string
		ttypes  []TokenType
		lexemes []string
	}{
		{" ", TTs{}, []string{}},
		{"5.2 .4", TTs{NumberToken, NumberToken}, []string{"5.2", ".4"}},
		{"color: red;", TTs{IdentToken, ColonToken, IdentToken, SemicolonToken}, []string{"color", ":", "red", ";"}},
		{"background: url(\"http://x\");", TTs{IdentToken, ColonToken, URLToken, SemicolonToken}, []string{"background", ":", `url("http://x")`, ";"}},
		{"background: URL(x.png);", TTs{IdentToken, ColonToken, URLToken, SemicolonToken}, []string{"background", ":", "URL(x.png)", ";"}},
		{"color: rgb(4, 0%, 5em);", TTs{IdentToken, ColonToken, FunctionToken, NumberToken, CommaToken, PercentageToken, CommaToken, DimensionToken, RightParenthesisToken, SemicolonToken}, []string{"color", ":", "rgb(", "4", ",", "0%", ",", "5em", ")", ";"}},
		{"body { \"string\" }", TTs{IdentToken, LeftBraceToken, StringToken, RightBraceToken}, []string{"body", "{", `"string"`, "}"}},
		{"body { \"str\\\"ing\" }", TTs{IdentToken, LeftBraceToken, StringToken, RightBraceToken}, []string{"body", "{", `"str\"ing"`, "}"}},
		{".class { }", TTs{DelimToken, IdentToken, LeftBraceToken, RightBraceToken}, []string{".", "class", "{", "}"}},
		{"#class { }", TTs{HashToken, LeftBraceToken, RightBraceToken}, []string{"#class", "{", "}"}},
		{"#class\\#withhash { }", TTs{HashToken, LeftBraceToken, RightBraceToken}, []string{`#class\#withhash`, "{", "}"}},
		{"@media print { }", TTs{AtKeywordToken, IdentToken, LeftBraceToken, RightBraceToken}, []string{"@media", "print", "{", "}"}},
		{"/*comment*/", TTs{CommentToken}, []string{"/*comment*/"}},
		{"/*com* /ment*/", TTs{CommentToken}, []string{"/*com* /ment*/"}},
		{"~= |= ^= $= *=", TTs{IncludeMatchToken, DashMatchToken, PrefixMatchToken, SuffixMatchToken, SubstringMatchToken}, []string{"~=", "|=", "^=", "$=", "*="}},
		{"||", TTs{ColumnToken}, []string{"||"}},
		{"<!-- -->", TTs{CDOToken, CDCToken}, []string{"<!--", "-->"}},
		{"U+1234", TTs{UnicodeRangeToken}, []string{"U+1234"}},
		{"5.2 .4 4e-22", TTs{NumberToken, NumberToken, NumberToken}, []string{"5.2", ".4", "4e-22"}},
		{"--custom-variable", TTs{CustomPropertyNameToken}, []string{"--custom-variable"}},

		// unexpected ending
		{"ident", TTs{IdentToken}, []string{"ident"}},
		{"123.", TTs{NumberToken, DelimToken}, []string{"123", "."}},
		{"\"string", TTs{StringToken}, []string{`"string`}},
		{"123/*comment", TTs{NumberToken, CommentToken}, []string{"123", "/*comment"}},
		{"U+1-", TTs{IdentToken, NumberToken, DelimToken}, []string{"U", "+1", "-"}},

		// unicode
		{"fooδbar􀀀", TTs{IdentToken}, []string{"fooδbar􀀀"}},
		{"foo\\æ\\†", TTs{IdentToken}, []string{"foo\\æ\\†"}},
		{"'foo\u554abar'", TTs{StringToken}, []string{"'foo\u554abar'"}},
		{"\\000026B", TTs{IdentToken}, []string{"\\000026B"}},
		{"\\26 B", TTs{IdentToken}, []string{"\\26 B"}},

		// hacks
		{`\-\mo\z\-b\i\nd\in\g:\url(//business\i\nfo.co.uk\/labs\/xbl\/xbl\.xml\#xss);`, TTs{IdentToken, ColonToken, URLToken, SemicolonToken}, []string{`\-\mo\z\-b\i\nd\in\g`, ":", `\url(//business\i\nfo.co.uk\/labs\/xbl\/xbl\.xml\#xss)`, ";"}},
		{"width/**/:/**/ 40em;", TTs{IdentToken, CommentToken, ColonToken, CommentToken, DimensionToken, SemicolonToken}, []string{"width", "/**/", ":", "/**/", "40em", ";"}},
		{":root *> #quince", TTs{ColonToken, IdentToken, DelimToken, DelimToken, HashToken}, []string{":", "root", "*", ">", "#quince"}},
		{"html[xmlns*=\"\"]:root", TTs{IdentToken, LeftBracketToken, IdentToken, SubstringMatchToken, StringToken, RightBracketToken, ColonToken, IdentToken}, []string{"html", "[", "xmlns", "*=", `""`, "]", ":", "root"}},
		{"body:nth-of-type(1)", TTs{IdentToken, ColonToken, FunctionToken, NumberToken, RightParenthesisToken}, []string{"body", ":", "nth-of-type(", "1", ")"}},
		{"color/*\\**/: blue\\9;", TTs{IdentToken, CommentToken, ColonToken, IdentToken, SemicolonToken}, []string{"color", `/*\**/`, ":", `blue\9`, ";"}},
		{"color: blue !ie;", TTs{IdentToken, ColonToken, IdentToken, DelimToken, IdentToken, SemicolonToken}, []string{"color", ":", "blue", "!", "ie", ";"}},

		// escapes, null and replacement character
		{"c\\\x00olor: white;", TTs{IdentToken, ColonToken, IdentToken, SemicolonToken}, []string{"c\\\x00olor", ":", "white", ";"}},
		{"null\\0", TTs{IdentToken}, []string{`null\0`}},
		{"\\", TTs{DelimToken}, []string{"\\"}},
		{"abc\\", TTs{IdentToken, DelimToken}, []string{"abc", "\\"}},
		{"#\\", TTs{DelimToken, DelimToken}, []string{"#", "\\"}},
		{"#abc\\", TTs{HashToken, DelimToken}, []string{"#abc", "\\"}},
		{"\"abc\\", TTs{StringToken}, []string{"\"abc\\"}}, // should officially not include backslash, but no biggie
		{"url(abc\\", TTs{BadURLToken}, []string{"url(abc\\"}},
		{"\"a\x00b\"", TTs{StringToken}, []string{"\"a\x00b\""}},
		{"a\\\x00b", TTs{IdentToken}, []string{"a\\\x00b"}},
		{"url(a\x00b)", TTs{BadURLToken}, []string{"url(a\x00b)"}}, // null character cannot be unquoted
		{"/*a\x00b*/", TTs{CommentToken}, []string{"/*a\x00b*/"}},

		// coverage
		{"  \n\r\n\r\"\\\r\n\\\r\"", TTs{StringToken}, []string{"\"\\\r\n\\\r\""}},
		{"U+?????? U+ABCD?? U+ABC-DEF", TTs{UnicodeRangeToken, UnicodeRangeToken, UnicodeRangeToken}, []string{"U+??????", "U+ABCD??", "U+ABC-DEF"}},
		{"U+? U+A?", TTs{UnicodeRangeToken, UnicodeRangeToken}, []string{"U+?", "U+A?"}},
		{"U+ U+ABCDEF?", TTs{IdentToken, DelimToken, IdentToken, DelimToken, IdentToken, DelimToken}, []string{"U", "+", "U", "+", "ABCDEF", "?"}},
		{"-5.23 -moz", TTs{NumberToken, IdentToken}, []string{"-5.23", "-moz"}},
		{"()", TTs{LeftParenthesisToken, RightParenthesisToken}, []string{"(", ")"}},
		{"url( //url\n  )", TTs{URLToken}, []string{"url( //url\n  )"}},
		{"url( ", TTs{URLToken}, []string{"url( "}},
		{"url( //url  ", TTs{URLToken}, []string{"url( //url  "}},
		{"url(\")a", TTs{URLToken}, []string{"url(\")a"}},
		{"url(a'\\\n)a", TTs{BadURLToken, IdentToken}, []string{"url(a'\\\n)", "a"}},
		{"url(\"\n)a", TTs{BadURLToken, IdentToken}, []string{"url(\"\n)", "a"}},
		{"url(a h)a", TTs{BadURLToken, IdentToken}, []string{"url(a h)", "a"}},
		{"<!- | @4 ## /2", TTs{DelimToken, DelimToken, DelimToken, DelimToken, DelimToken, NumberToken, DelimToken, DelimToken, DelimToken, NumberToken}, []string{"<", "!", "-", "|", "@", "4", "#", "#", "/", "2"}},
		{"\"s\\\n\"", TTs{StringToken}, []string{"\"s\\\n\""}},
		{"\"a\\\"b\"", TTs{StringToken}, []string{"\"a\\\"b\""}},
		{"\"s\n", TTs{BadStringToken}, []string{"\"s\n"}},

		// small
		{"\"abcd", TTs{StringToken}, []string{"\"abcd"}},
		{"/*comment", TTs{CommentToken}, []string{"/*comment"}},
		{"U+A-B", TTs{UnicodeRangeToken}, []string{"U+A-B"}},
		{"url((", TTs{BadURLToken}, []string{"url(("}},
		{"id\u554a", TTs{IdentToken}, []string{"id\u554a"}},
	}
	for _, tt := range tokenTests {
		t.Run(tt.css, func(t *testing.T) {
			l := NewLexer(parse.NewInputString(tt.css))
			i := 0
			tokens := []TokenType{}
			lexemes := []string{}
			for {
				token, lexeme := l.Next()
				if token == ErrorToken {
					test.T(t, l.Err(), io.EOF)
					break
				} else if token == WhitespaceToken {
					continue
				}
				tokens = append(tokens, token)
				lexemes = append(lexemes, string(lexeme))
				i++
			}
			test.T(t, tokens, tt.ttypes, "token types must match")
			test.T(t, lexemes, tt.lexemes, "token data must match")
		})
	}

	// coverage
	for i := 0; ; i++ {
		if TokenType(i).String() == fmt.Sprintf("Invalid(%d)", i) {
			break
		}
	}
	test.T(t, NewLexer(parse.NewInputString("x")).consumeBracket(), ErrorToken, "consumeBracket on 'x' must return error")
}

func TestOffset(t *testing.T) {
	z := parse.NewInputString(`div{background:url(link);}`)
	l := NewLexer(z)
	test.T(t, z.Offset(), 0)
	_, _ = l.Next()
	test.T(t, z.Offset(), 3) // div
	_, _ = l.Next()
	test.T(t, z.Offset(), 4) // {
	_, _ = l.Next()
	test.T(t, z.Offset(), 14) // background
	_, _ = l.Next()
	test.T(t, z.Offset(), 15) // :
	_, _ = l.Next()
	test.T(t, z.Offset(), 24) // url(link)
	_, _ = l.Next()
	test.T(t, z.Offset(), 25) // ;
	_, _ = l.Next()
	test.T(t, z.Offset(), 26) // }
}

////////////////////////////////////////////////////////////////

func ExampleNewLexer() {
	l := NewLexer(parse.NewInputString("color: red;"))
	out := ""
	for {
		tt, data := l.Next()
		if tt == ErrorToken {
			break
		} else if tt == WhitespaceToken || tt == CommentToken {
			continue
		}
		out += string(data)
	}
	fmt.Println(out)
	// Output: color:red;
}
