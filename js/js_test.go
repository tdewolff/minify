package js

import (
	"bytes"
	"os"
	"testing"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/test"
)

func TestJS(t *testing.T) {
	jsTests := []struct {
		js       string
		expected string
	}{
		//{"/*comment*/", ""},
		//{"// comment\na", "a"},
		////{"/*! bang  comment */", "/*!bang comment*/"},
		//{"function x(){}", "function x(){}"},
		//{"function x(a, b){}", "function x(a,b){}"},
		//{"a  b", "a b"},
		//{"a\n\nb", "a\nb"},
		//{"a// comment\nb", "a\nb"},
		//{"''\na", "''\na"},
		//{"''\n''", "''\n''"},
		//{"]\n0", "]\n0"},
		//{"a\n{", "a\n{"},
		//{";\na", ";a"},
		//{",\na", ",a"},
		//{"}\na", "}\na"},
		//{"+\na", "+\na"},
		//{"+\n(", "+\n("},
		//{"+\n\"\"", "+\n\"\""},
		//{"a + ++b", "a+ ++b"}, // JSMin caution
		//{"var a=/\\s?auto?\\s?/i\nvar", "var a=/\\s?auto?\\s?/i\nvar"}, // #14
		//{"var a=0\n!function(){}", "var a=0\n!function(){}"},           // #107
		//{"function(){}\n\"string\"", "function(){}\n\"string\""},       // #109
		//{"false\n\"string\"", "false\n\"string\""},                     // #109
		//{"`\n", "`"},       // go fuzz
		//{"a\n~b", "a\n~b"}, // #132
		//{"x / /\\d+/.exec(s)[0]", "x/ /\\d+/.exec(s)[0]"}, // #183

		//{"function(){}\n`string`", "function(){}\n`string`"}, // #181
		//{"false\n`string`", "false\n`string`"},               // #181
		//{"`string`\nwhatever()", "`string`\nwhatever()"},     // #181

		//{"x+/**/++y", "x+ ++y"},                          // #185
		//{"x+\n++y", "x+\n++y"},                           // #185
		//{"f()/*!com\nment*/g()", "f()/*!com\nment*/g()"}, // #185
		//{"f()/*com\nment*/g()", "f()\ng()"},              // #185
		//{"f()/*!\n*/g()", "f()/*!\n*/g()"},               // #185

		//// go-fuzz
		//{`/\`, `/\`},
		{`1.0`, `1`},
		{`1000`, `1e3`},
		{`+ +x`, `+ +x`},
		{`- +x`, `-+x`},
		{`+ ++x`, `+ ++x`},
		{`- ++x`, `-++x`},
		{`a + ++b`, `a+ ++b`},
		{`a - ++b`, `a-++b`},
		{`a-- > b`, `a-- >b`},
		{`a-- < b`, `a--<b`},
		{`a < !--b`, `a< !--b`},
		{`a > !--b`, `a>!--b`},
		{`!--b`, `!--b`},
		{`/a/ + b`, `/a/+b`},
		{`/a/ instanceof b`, `/a/ instanceof b`},
		{`[a] instanceof b`, `[a]instanceof b`},
		{`let a = 5`, `let a=5`},
		{`let a = 5,b`, `let a=5,b`},
		{`let a,b = 5`, `let a,b=5`},
		{`function a(){}`, `function a(){}`},
		{`function a(b){}`, `function a(b){}`},
		{`function a(b, c){}`, `function a(b,c){}`},
		{`function * a(){}`, `function*a(){}`},
		{`function a(){}; return 5`, `function a(){}return 5`},
		{`x = function (){}`, `x=function(){}`},
		{`x = function a(){}`, `x=function a(){}`},
		{`x = function (a){}`, `x=function(a){}`},
		{`x = function (a,b){}`, `x=function(a,b){}`},
		{`x = function (){};y=z`, `x=function(){};y=z`},
		{`() => {}`, `()=>{}`},
		{`(a) => {}`, `a=>{}`},
		{`(...a) => {}`, `(...a)=>{}`},
		{`(a = 0) => {}`, `(a=0)=>{}`},
		{`(a, b) => {}`, `(a,b)=>{}`},
		{`a => {a++}`, `a=>a++`},
		{`x = (a) => {}`, `x=a=>{}`},
		{`return 5`, `return 5`},
		{`return .5`, `return.5`},
		{`return-5`, `return-5`},
		{`break a`, `break a`},
		{`continue a`, `continue a`},
		{`typeof a`, `typeof a`},
		{`new RegExp()`, `new RegExp()`},
		{`new new a()()`, `new new a()()`},
		{`switch (a) { case b: 5; default: 6}`, `switch(a){case b:5;default:6}`},
		{`switch (a) { case b: {var c;return c}; default: 6}`, `switch(a){case b:{var c;return c}default:6}`},
		{`with (a = b) x`, `with(a=b)x`},
		{`with (a = b) {x}`, `with(a=b)x`},
		{`import 'path'`, `import'path'`},
		{`import * as b from 'path'`, `import*as b from'path'`},
		{`import x from 'path'`, `import x from'path'`},
		{`import * as b from 'path'`, `import*as b from'path'`},
		{`import {a as b, c} from 'path'`, `import{a as b,c}from'path'`},
		{`import x, * as b from 'path'`, `import x,*as b from'path'`},
		{`import x, {a as b, c} from 'path'`, `import x,{a as b,c}from'path'`},
		{`export * from 'path'`, `export*from'path'`},
		{`export * as ns from 'path'`, `export*as ns from'path'`},
		{`export {a as b, c} from 'path'`, `export{a as b,c}from'path'`},
		{`export {a as b, c}`, `export{a as b,c}`},
		{`export var a = b`, `export var a=b`},
		{`export default a = b`, `export default a=b`},
		{`export default a = b;c=d`, `export default a=b;c=d`},
		{`export default function a(){};c=d`, `export default function a(){}c=d`},
		{`class {}`, `class{}`},
		{`class a {}`, `class a{}`},
		{`class a extends b {}`, `class a extends b{}`},
		{`class a extends(!b){}`, `class a extends(!b){}`},
		{`class { f(a) {} }`, `class{f(a){}}`},
		{`class { f(a) {}; static g(b) {} }`, `class{f(a){}static g(b){}}`},
		{`return;a`, `return;a`},
		{`break;a`, `break;a`},
		{`if(a){return b}`, `if(a)return b`},
		{`if(a){b = 5;return b}`, `if(a){b=5;return b}`},
		{`if(a);`, `if(a)`},
		{`if(a){}`, `if(a)`},
		{`if(a) b`, `if(a)b`},
		{`if(a){}else;`, `if(a)`},
		{`if(a){}else{}`, `if(a)`},
		{`if(a){}else{;}`, `if(a)`},
		{`if(a){}else{a}`, `if(a);else a`},
		{`if(a){b=c}else if(d){e=f}`, `if(a)b=c;else if(d)e=f`},
		{`if(a){b=c;y=z}else if(d){e=f}`, `if(a){b=c;y=z}else if(d)e=f`},
		{`if(a)while(b){c;d}else e`, `if(a)while(b){c;d}else e`},
		{`if(a)while(b){c}else e`, `if(a)while(b)c;else e`},
		{`if(a){ if(b) c } else e`, `if(a){if(b)c}else e`},
		{`if(a){ if(b) c; else d} else e`, `if(a)if(b)c;else d;else e`},
		{`if(a){ if(b) c; else for(x;y;z){f=g}} else e`, `if(a)if(b)c;else for(x;y;z)f=g;else e`},
		{`if(a){ if(b) c; else {for(x;y;z){f=g}}} else e`, `if(a)if(b)c;else for(x;y;z)f=g;else e`},
		{`if(a)a={b};else e`, `if(a)a={b};else e`},
		{`if(a) a; else [e]=4`, `if(a)a;else[e]=4`},
		{`if(a){ a = b?c:function(d){f} } else e`, `if(a)a=b?c:function(d){f};else e`},
		{`if(a)while(b){if(c)d; else e}else f`, `if(a)while(b)if(c)d;else e;else f`},
		{`for (var a = 5; a < 10; a++){a}`, `for(var a=5;a<10;a++)a`},
		{`for (a,b = 5; a < 10; a++){a}`, `for(a,b=5;a<10;a++)a`},
		{`for await (var a of b){a}`, `for await(var a of b)a`},
		{`for (var a in b){a}`, `for(var a in b)a`},
		{`for (var a of b){a}`, `for(var a of b)a`},
		{`while(a < 10){a}`, `while(a<10)a`},
		{`do {a} while(a < 10)`, `do a;while(a<10)`},
		{`do [a]=5; while(a < 10)`, `do[a]=5;while(a<10)`},
		{`do [a]=5; while(a < 10);return a`, `do[a]=5;while(a<10);return a`},
		{`throw a`, `throw a`},
		{`throw [a]`, `throw[a]`},
		{`try {a}`, `try{a}`},
		{`try {a} catch {b}`, `try{a}catch{b}`},
		{`try {a} catch(e) {b}`, `try{a}catch(e){b}`},
		{`try {a} catch(e) {b} finally {c}`, `try{a}catch(e){b}finally{c}`},
		{`try {a} finally {c}`, `try{a}finally{c}`},
		{`x=true`, `x=!0`},
		{`x=false`, `x=!1`},
		{`x=false()`, `x=(!1)()`},
		{`x=undefined`, `x=void 0`},
		{`x=undefined()`, `x=(void 0)()`},
		{`var undefined=5;x=undefined`, `var undefined=5;x=undefined`},
		{`class a extends undefined {}`, `class a extends(void 0){}`},

		// variable renaming
		{`x=function(){var name}`, `x=function(){var a}`},
		{`x=function(){var name; name++}`, `x=function(){var a;a++}`},
		{`x=function(){function name(){}}`, `x=function(){function a(){}}`},
		{`x=function(){function name(arg1, arg2){return arg1, arg2}}`, `x=function(){function a(b,c){return b,c}}`},
		{`x=function(){function name(arg1, arg2){return arg1, arg2} return arg1}`, `x=function(){function a(b,c){return b,c}return arg1}`},
		{`x=function(){function name(arg1, arg2){return arg1, arg2} return a}`, `x=function(){function b(c,d){return c,d}return a}`},
		{`x=function(){function add(l,r){return add(l,r)}function nadd(l,r){return-add(l,r)}}`, `x=function(){function a(b,c){return a(b,c)}function b(c,d){return-a(c,d)}}`},
		{`function a(){var b}`, `function a(){var b}`},
		//{`import name from 'file'; name('str')`, `import a from'file';a('str')`},

		// edge-cases
		{`let o=null;try{o=(o?.a).b||"FAIL"}catch(x){}console.log(o||"PASS")`, `let o=null;try{o=(o?.a).b||"FAIL"}catch(x){}console.log(o||"PASS")`},
	}

	m := minify.New()
	for _, tt := range jsTests {
		t.Run(tt.js, func(t *testing.T) {
			r := bytes.NewBufferString(tt.js)
			w := &bytes.Buffer{}
			err := Minify(m, w, r, nil)
			test.Minify(t, tt.js, err, w.String(), tt.expected)
		})
	}
}

////////////////////////////////////////////////////////////////

func ExampleMinify() {
	m := minify.New()
	m.AddFunc("application/javascript", Minify)

	if err := m.Minify("application/javascript", os.Stdout, os.Stdin); err != nil {
		panic(err)
	}
}
