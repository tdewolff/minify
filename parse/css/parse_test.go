package css

import (
	"fmt"
	"io"
	"testing"

	"github.com/tdewolff/minify/v2/parse"
	"github.com/tdewolff/test"
)

////////////////////////////////////////////////////////////////

func TestParse(t *testing.T) {
	var parseTests = []struct {
		inline   bool
		css      string
		expected string
	}{
		{true, " x : y ; ", "x:y;"},
		{true, "color: red;", "color:red;"},
		{true, "color : red;", "color:red;"},
		{true, "color: red; border: 0;", "color:red;border:0;"},
		{true, "color: red !important;", "color:red!important;"},
		{true, "color: red ! important;", "color:red!important;"},
		{true, "white-space: -moz-pre-wrap;", "white-space:-moz-pre-wrap;"},
		{true, "display: -moz-inline-stack;", "display:-moz-inline-stack;"},
		{true, "x: 10px / 1em;", "x:10px/1em;"},
		{true, "x: 1em/1.5em \"Times New Roman\", Times, serif;", "x:1em/1.5em \"Times New Roman\",Times,serif;"},
		{true, "x: hsla(100,50%, 75%, 0.5);", "x:hsla(100,50%,75%,0.5);"},
		{true, "x: hsl(100,50%, 75%);", "x:hsl(100,50%,75%);"},
		{true, "x: rgba(255, 238 , 221, 0.3);", "x:rgba(255,238,221,0.3);"},
		{true, "x: 50vmax;", "x:50vmax;"},
		{true, "color: linear-gradient(to right, black, white);", "color:linear-gradient(to right,black,white);"},
		{true, "color: calc(100%/2 - 1em);", "color:calc(100%/2 - 1em);"},
		{true, "color: calc(100%/2--1em);", "color:calc(100%/2--1em);"},
		{false, "<!-- @charset; -->", "<!--@charset;-->"},
		{false, "@media print, screen { }", "@media print,screen{}"},
		{false, "@media { @viewport ; }", "@media{@viewport;}"},
		{false, "@keyframes 'diagonal-slide' {  from { left: 0; top: 0; } to { left: 100px; top: 100px; } }", "@keyframes 'diagonal-slide'{from{left:0;top:0;}to{left:100px;top:100px;}}"},
		{false, "@keyframes movingbox{0%{left:90%;}50%{left:10%;}100%{left:90%;}}", "@keyframes movingbox{0%{left:90%;}50%{left:10%;}100%{left:90%;}}"},
		{false, ".foo { color: #fff;}", ".foo{color:#fff;}"},
		{false, ".foo { ; _color: #fff;}", ".foo{_color:#fff;}"},
		{false, "a { color: red; border: 0; }", "a{color:red;border:0;}"},
		{false, "a { color: red; border: 0; } b { padding: 0; }", "a{color:red;border:0;}b{padding:0;}"},
		{false, "/* comment */", "/* comment */"},

		// extraordinary
		{true, "color: red;;", "color:red;"},
		{true, "margin: 10px/*comment*/50px;", "margin:10px 50px;"},
		{true, "color:#c0c0c0", "color:#c0c0c0;"},
		{true, "background:URL(x.png);", "background:URL(x.png);"},
		{true, "filter: progid : DXImageTransform.Microsoft.BasicImage(rotation=1);", "filter:progid:DXImageTransform.Microsoft.BasicImage(rotation=1);"},
		{true, "/*a*/\n/*c*/\nkey: value;", "key:value;"},
		{true, "@-moz-charset;", "@-moz-charset;"},
		{true, "--custom-variable:  (0;)  ;", "--custom-variable:  (0;)  ;"},
		{false, "@import;@import;", "@import;@import;"},
		{false, ".a .b#c, .d<.e { x:y; }", ".a .b#c,.d<.e{x:y;}"},
		{false, ".a[b~=c]d { x:y; }", ".a[b~=c]d{x:y;}"},
		// {false, "{x:y;}", "{x:y;}"},
		{false, "a{}", "a{}"},
		{false, "a,.b/*comment*/ {x:y;}", "a,.b{x:y;}"},
		{false, "a,.b/*comment*/.c {x:y;}", "a,.b.c{x:y;}"},
		{false, "a{x:; z:q;}", "a{x:;z:q;}"},
		{false, "@font-face { x:y; }", "@font-face{x:y;}"},
		{false, "a:not([controls]){x:y;}", "a:not([controls]){x:y;}"},
		{false, "@document regexp('https:.*') { p { color: red; } }", "@document regexp('https:.*'){p{color:red;}}"},
		{false, "@media all and ( max-width:400px ) { }", "@media all and (max-width:400px){}"},
		{false, "@media (max-width:400px) { }", "@media(max-width:400px){}"},
		{false, "@media (max-width:400px)", "@media(max-width:400px);"},
		{false, "@font-face { ; font:x; }", "@font-face{font:x;}"},
		{false, "@-moz-font-face { ; font:x; }", "@-moz-font-face{font:x;}"},
		{false, "@unknown abc { {} lala }", "@unknown abc{{} lala }"},
		{false, "a[x={}]{x:y;}", "a[x={}]{x:y;}"},
		{false, "a[x=,]{x:y;}", "a[x=,]{x:y;}"},
		{false, "a[x=+]{x:y;}", "a[x=+]{x:y;}"},
		{false, ".cla .ss > #id { x:y; }", ".cla .ss>#id{x:y;}"},
		{false, ".cla /*a*/ /*b*/ .ss{}", ".cla .ss{}"},
		{false, "a{x:f(a(),b);}", "a{x:f(a(),b);}"},
		{false, "a{x:y!z;}", "a{x:y!z;}"},
		{false, "[class*=\"column\"]+[class*=\"column\"]:last-child{a:b;}", "[class*=\"column\"]+[class*=\"column\"]:last-child{a:b;}"},
		{false, "@media { @viewport }", "@media{@viewport;}"},
		{false, "table { @unknown }", "table{@unknown;}"},

		// early endings
		{false, "selector{", "selector{"},
		{false, "@media{selector{", "@media{selector{"},

		// bad grammar
		{false, "}", "ERROR(})"},
		{true, "}", "ERROR(})"},
		{true, "~color:red", "ERROR(~color:red)"},
		{true, "(color;red)", "ERROR((color;red))"},
		{true, "color(;red)", "ERROR(color(;red))"},
		{false, ".foo { *color: #fff;}", ".foo{*color:#fff;}"},
		{true, "*color: red; font-size: 12pt;", "*color:red;font-size:12pt;"},
		{true, "*--custom: red;", "*--custom: red;"},
		{true, "_color: red; font-size: 12pt;", "_color:red;font-size:12pt;"},
		{false, ".foo { baddecl } .bar { color:red; }", ".foo{ERROR(baddecl)}.bar{color:red;}"},
		{false, ".foo { baddecl baddecl baddecl; height:100px } .bar { color:red; }", ".foo{ERROR(baddecl baddecl baddecl;)height:100px;}.bar{color:red;}"},
		{false, ".foo { visibility: hidden;” } .bar { color:red; }", ".foo{visibility:hidden;ERROR(”)}.bar{color:red;}"},
		{false, ".foo { baddecl (; color:red; }", ".foo{ERROR(baddecl (; color:red; })"},

		// issues
		{false, "@media print {.class{width:5px;}}", "@media print{.class{width:5px;}}"},                  // #6
		{false, ".class{width:calc((50% + 2em)/2 + 14px);}", ".class{width:calc((50% + 2em)/2 + 14px);}"}, // #7
		{false, ".class [c=y]{}", ".class [c=y]{}"},                                                       // tdewolff/minify#16
		{false, "table{font-family:Verdana}", "table{font-family:Verdana;}"},                              // tdewolff/minify#22

		// go-fuzz
		{false, "@-webkit-", "@-webkit-;"},
	}
	for _, tt := range parseTests {
		t.Run(tt.css, func(t *testing.T) {
			output := ""
			p := NewParser(parse.NewInputString(tt.css), tt.inline)
			for {
				grammar, _, data := p.Next()
				data = parse.Copy(data)
				if grammar == ErrorGrammar {
					if err := p.Err(); err != io.EOF {
						data = []byte("ERROR(")
						for _, val := range p.Values() {
							data = append(data, val.Data...)
						}
						data = append(data, ")"...)
					} else {
						break
					}
				} else if grammar == AtRuleGrammar || grammar == BeginAtRuleGrammar || grammar == QualifiedRuleGrammar || grammar == BeginRulesetGrammar || grammar == DeclarationGrammar || grammar == CustomPropertyGrammar {
					if grammar == DeclarationGrammar || grammar == CustomPropertyGrammar {
						data = append(data, ":"...)
					}
					for _, val := range p.Values() {
						data = append(data, val.Data...)
					}
					if grammar == BeginAtRuleGrammar || grammar == BeginRulesetGrammar {
						data = append(data, "{"...)
					} else if grammar == AtRuleGrammar || grammar == DeclarationGrammar || grammar == CustomPropertyGrammar {
						data = append(data, ";"...)
					} else if grammar == QualifiedRuleGrammar {
						data = append(data, ","...)
					}
				}
				output += string(data)
			}
			test.String(t, output, tt.expected)
		})
	}

	// coverage
	for i := 0; ; i++ {
		if GrammarType(i).String() == fmt.Sprintf("Invalid(%d)", i) {
			break
		}
	}
	test.T(t, Token{IdentToken, []byte("data")}.String(), "Ident('data')")
}

func TestParseError(t *testing.T) {
	var parseErrorTests = []struct {
		inline bool
		css    string
		col    int
	}{
		{false, "}", 2},
		{true, "}", 1},
		{false, "selector", 9},
		{true, "color 0", 7},
		{true, "--color 0", 9},
		{true, "--custom-variable:0", 0},
	}
	for _, tt := range parseErrorTests {
		t.Run(tt.css, func(t *testing.T) {
			p := NewParser(parse.NewInputString(tt.css), tt.inline)
			for {
				grammar, _, _ := p.Next()
				if grammar == ErrorGrammar {
					if tt.col == 0 {
						test.T(t, p.Err(), io.EOF)
					} else if perr, ok := p.Err().(*parse.Error); ok {
						test.That(t, p.HasParseError())
						_, col, _ := perr.Position()
						test.T(t, col, tt.col)
					} else {
						test.Fail(t, "bad error:", p.Err())
					}
					break
				}
			}
		})
	}
}

func TestParseOffset(t *testing.T) {
	z := parse.NewInputString(`div{background:url(link);}`)
	p := NewParser(z, false)
	test.T(t, z.Offset(), 0)
	_, _, _ = p.Next()
	test.T(t, z.Offset(), 4) // div{
	_, _, _ = p.Next()
	test.T(t, z.Offset(), 25) // background:url(link);
	_, _, _ = p.Next()
	test.T(t, z.Offset(), 26) // }
}

////////////////////////////////////////////////////////////////

type Obj struct{}

func (*Obj) F() {}

var f1 func(*Obj)

func BenchmarkFuncPtr(b *testing.B) {
	for i := 0; i < b.N; i++ {
		f1 = (*Obj).F
	}
}

var f2 func()

func BenchmarkMemFuncPtr(b *testing.B) {
	obj := &Obj{}
	for i := 0; i < b.N; i++ {
		f2 = obj.F
	}
}

func ExampleNewParser() {
	p := NewParser(parse.NewInputString("color: red;"), true) // false because this is the content of an inline style attribute
	out := ""
	for {
		gt, _, data := p.Next()
		if gt == ErrorGrammar {
			break
		} else if gt == AtRuleGrammar || gt == BeginAtRuleGrammar || gt == BeginRulesetGrammar || gt == DeclarationGrammar {
			out += string(data)
			if gt == DeclarationGrammar {
				out += ":"
			}
			for _, val := range p.Values() {
				out += string(val.Data)
			}
			if gt == BeginAtRuleGrammar || gt == BeginRulesetGrammar {
				out += "{"
			} else if gt == AtRuleGrammar || gt == DeclarationGrammar {
				out += ";"
			}
		} else {
			out += string(data)
		}
	}
	fmt.Println(out)
	// Output: color:red;
}
