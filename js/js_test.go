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
		{`/*comment*/`, ``},
		//{`/*!comment*/`, `/*!comment*/`},
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
		{`x = function (){};y=z`, `x=function(){},y=z`},
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
		{`a=b;c=d`, `a=b,c=d`},

		// strings
		{`"string\'string"`, `"string'string"`},
		{`'string\"string'`, `'string"string'`},
		{`'string\t\f\v\bstring'`, "'string\t\f\v\bstring'"},
		{`"string\a\c\'string"`, `"stringac'string"`},
		{`"string\∀string"`, `"string∀string"`},
		{`"string\0\uFFFFstring"`, `"string\0\uFFFFstring"`},
		{`"string\x00\x55\x0A\x0D\x22\x27string"`, `"string\0U\n\r\"'string"`},
		{`"string\000\12\015\042\47\411string"`, `"string\0\n\r\"'!1string"`},
		{"'string\\n\\rstring'", "'string\\n\\rstring'"},
		{"'string\\\r\nstring\\\nstring\\\rstring\\\u2028string\\\u2029string'", "'stringstringstringstringstringstring'"},
		//{`"string" + "string"`, `"stringstring"`},

		// rename true, false, undefined
		{`x=true`, `x=!0`},
		{`x=false`, `x=!1`},
		{`x=false()`, `x=(!1)()`},
		{`false`, `!1`},
		{`x=undefined`, `x=void 0`},
		{`x=undefined()`, `x=(void 0)()`},
		{`var undefined=5;x=undefined`, `var undefined=5;x=undefined`},
		{`class a extends undefined {}`, `class a extends(void 0){}`},

		// if/else statements
		{`if(a){return b}`, `if(a)return b`},
		{`if(a){b = 5;return b}`, `if(a)return b=5,b`},
		{`if(a)`, `a`},
		{`if(a){}`, `a`},
		{`if(a) b`, `a&&b`},
		{`if(a){}else;`, `a`},
		{`if(a){}else{}`, `a`},
		{`if(a){}else{;}`, `a`},
		{`if(a){}else{b}`, `a||b`},
		//{`if(a)a;else b`, `a||b`},
		//{`if(a)b;else b`, `a,b`},
		{`if(a){b=c}else if(d){e=f}`, `a?b=c:d&&(e=f)`},
		{`if(a){b=c;y=z}else if(d){e=f}`, `a?(b=c,y=z):d&&(e=f)`},
		{`if(a)while(b){c;d}else e`, `if(a)while(b)c,d;else e`},
		{`if(a)while(b){c}else e`, `if(a)while(b)c;else e`},
		{`if(a){ if(b) c }`, `a&&b&&c`},
		{`if(a){ if(b) c } else e`, `a?b&&c:e`},
		{`if(a){ if(b) c; else d} else e`, `a?b?c:d:e`},
		{`if(a){ if(b) c; else for(x;y;z){f=g}} else e`, `if(a)if(b)c;else for(x;y;z)f=g;else e`},
		{`if(a){ if(b) c; else {for(x;y;z){f=g}}} else e`, `if(a)if(b)c;else for(x;y;z)f=g;else e`},
		{`if(a)a={b};else e`, `a?a={b}:e`},
		{`if(a) a; else [e]=4`, `a?a:[e]=4`},
		{`if(a){ a = b?c:function(d){f} } else e`, `a?a=b?c:function(d){f}:e`},
		{`if(a)while(b){if(c)d; else e}else f`, `if(a)while(b)c?d:e;else f`},
		{`if(a)b=c`, `a&&(b=c)`},
		{`if(!a)b=c`, `a||(b=c)`},
		{`if(a||d)b=c`, `(a||d)&&(b=c)`},
		{`if(a);else b=c`, `a||(b=c)`},
		{`if(!a);else b=c`, `a&&(b=c)`},
		{`if(a)b=c;else e`, `a?b=c:e`},
		{`if(a)b=c,f;else e`, `a?(b=c,f):e`},
		{`if(a){b=c}else{if(d){e=f}else{g=h}}`, `a?b=c:d?e=f:g=h`},
		{`b=5;return a+b`, `return b=5,a+b`},
		{`b=5;throw a+b`, `throw b=5,a+b`},
		{`if(a)return a;else return b`, `return a?a:b`},
		{`if(a)throw a;else throw b`, `throw a?a:b`},
		{`if(a)return a;else a=b`, `if(a)return a;a=b`},
		{`if(a){a++;return a}else a=b`, `if(a)return a++,a;a=b`},
		{`if(a){a++;return a}else if(b)a=b`, `if(a)return a++,a;b&&(a=b)`},
		{`if(a){a++;return}else a=b`, `if(a){a++;return}a=b`},
		//{`if(a){a++;return}else return`, `return a?void a++:void 0`},
		{`if(a){return}else {a=b;while(c){}}`, `if(a)return;a=b;while(c){}`},
		{`if(a){a++;return a}else return`, `return a?(a++,a):void 0`},
		{`a=b;if(a){return a}else return b`, `return a=b,a?a:b`},
		{`if(a){return a}return b`, `return a?a:b`},
		{`if(a);else return a;return b`, `return a?b:a`},
		{`if(a){return a}b=c;return b`, `return a?a:(b=c,b)`},
		{`if(a){return}b=c;return b`, `if(a)return;return b=c,b`},
		{`if(a){return a}b=c;return`, `if(a)return a;b=c;return`},
		{`if(a){throw a}b=c;throw b`, `throw a?a:(b=c,b)`},
		{`if(a)a++;else b;if(b)b++;else c`, `a?a++:b,b?b++:c`},
		{`if(false)a++;else b`, `b`},
		//{`if(false){var a;a++}else b`, `var a;b`},
		//{`if(false){function a(c){return d};a++}else b`, `var a;b`},
		{`if(!1)a++;else b`, `b`},
		{`if(null)a++;else b`, `b`},
		//{`var a;if(false)var b`, `var a,b`},
		//{`var a;if(false)var b=5`, `var a,b`},
		//{`var a;if(false)const b`, `var a`},
		//{`var a;if(false)function f(){}`, `var a;if(false)function f(){}`},

		// var declarations
		{`var a;var b`, `var a,b`},
		{`const a=1;const b=2`, `const a=1,b=2`},
		{`let a=1;let b=2`, `let a=1,b=2`},
		//{`var a;if(a)var b;else b`, `var a,b;a||b`},
		//{`var a;if(a)var b=5`, `var a;if(a)var b=5`},
		{`var a;for(var b=0;b;b++){}`, `for(var a,b=0;b;b++){}`},
		{`const a=3;for(const b=0;b;b++){}`, `const a=3;for(const b=0;b;b++){}`},
		{`var a;while(b){}`, `for(var a;b;){}`},

		// function declarations
		{`function g(){return}`, `function g(){}`},
		{`function g(){return undefined}`, `function g(){}`},
		{`function g(){return void 0}`, `function g(){}`},
		{`function g(){return;var a;a=b}`, `function g(){}`},
		//{`function g(){if(a)return a;else return b;var c;c=d}`, `function g(){return a||b}`},

		// arrow functions
		{`() => {}`, `()=>{}`},
		{`(a) => {}`, `a=>{}`},
		{`(...a) => {}`, `(...a)=>{}`},
		{`(a = 0) => {}`, `(a=0)=>{}`},
		{`(a, b) => {}`, `(a,b)=>{}`},
		{`a => {a++}`, `a=>{a++}`},
		{`x = (a) => {}`, `x=a=>{}`},
		{`x = (a) => {return}`, `x=a=>{}`},
		{`x = (a) => {return a}`, `x=a=>a`},
		{`x = (a) => {a++;return a}`, `x=a=>a++,a`},
		{`x = (a) => {a++}`, `x=a=>{a++}`},
		{`x = (a,b) => a+b`, `x=(a,b)=>a+b`},

		// remove groups
		{`a=(b+c)+d`, `a=b+c+d`},
		{`a=b+(c+d)`, `a=b+c+d`},
		{`a=b*(c+d)`, `a=b*(c+d)`},
		{`a=(b*c)+d`, `a=b*c+d`},
		{`a=(b.c)++`, `a=b.c++`},
		{`a=(b++).c`, `a=b++.c`},
		{`a=!(b++)`, `a=!b++`},
		{`a=(b+c)(d)`, `a=(b+c)(d)`},
		{`a=b**(c**d)`, `a=b**c**d`},
		{`a=(b**c)**d`, `a=(b**c)**d`},
		{`a=false**2`, `a=(!1)**2`},
		{`a=(a||b)&&c`, `a=(a||b)&&c`},
		{`a=a||(b&&c)`, `a=a||b&&c`},
		{`a=(a&&b)||c`, `a=a&&b||c`},
		{`a=a&&(b||c)`, `a=a&&(b||c)`},
		{`a=c&&(a??b)`, `a=c&&(a??b)`},
		{`a=!(!b)`, `a=!!b`},
		{`a=(b())`, `a=b()`},
		{`a=(b)?.(c,d)`, `a=b?.(c,d)`},
		{`a=(b,c)?.(d)`, `a=(b,c)?.(d)`},
		{`a=(b?c:e)?.(d)`, `a=(b?c:e)?.(d)`},
		{`function*x(){a=(yield b)}`, `function*x(){a=yield b}`},
		{`function*x(){a=yield (yield b)}`, `function*x(){a=yield yield b}`},
		{`if((a))while((b)){}`, `if(a)while(b){}`},
		{`(function(){})`, `!function(){}`},
		{`(function(){}())`, `!function(){}()`},
		{`(function(){})()`, `!function(){}()`},
		{`(function(){})();x=5;f=6`, `!function(){}(),x=5,f=6`},
		{`x=(function(){})`, `x=function(){}`},
		{`x=(function(){}())`, `x=function(){}()`},
		{`x=(function(){})()`, `x=function(){}()`},
		{`(class a{})`, `!class a{}`},

		// variable renaming
		{`x=function(){var name}`, `x=function(){var a}`},
		{`x=function(){var name; name++}`, `x=function(){var a;a++}`},
		{`x=function(){function name(){}}`, `x=function(){function a(){}}`},
		{`x=function(){function name(arg1, arg2){return arg1, arg2}}`, `x=function(){function a(b,c){return b,c}}`},
		{`x=function(){function name(arg1, arg2){return arg1, arg2} return arg1}`, `x=function(){function a(b,c){return b,c}return arg1}`},
		{`x=function(){function name(arg1, arg2){return arg1, arg2} return a}`, `x=function(){function b(c,d){return c,d}return a}`},
		{`x=function(){function add(l,r){return add(l,r)}function nadd(l,r){return-add(l,r)}}`, `x=function(){function a(b,c){return a(b,c)}function b(c,d){return-a(c,d)}}`},
		{`function a(){var b}`, `function a(){var b}`},
		{`!function(){x=function(){return fun()};var fun=function(){return 0}}`, `!function(){x=function(){return a()};var a=function(){return 0}}`},
		{`!function(){var x=function(){return y};const y=5}`, `!function(){var a=function(){return b};const b=5}`},
		{`!function(){if(1)const y=5;var x=function(){return y}}`, `!function(){if(1)const a=5;var a=function(){return y}}`},
		{`!function(){var x=function(){return y};if(1)const y=5}`, `!function(){var a=function(){return x};if(1)const b=5}`},
		{`!function(){var x=function(){return y};if(1)var y=5}`, `!function(){var a=function(){return b};if(1)var b=5}`},
		{`!function(){var x,y,z=(x,y)=>x+y}`, `!function(){var a,b,c=(a,b)=>a+b}`},
		//{`import name from 'file'; name('str')`, `import a from'file';a('str')`},
		{`name=function(){var a1,a2,a3,a4,a5,a6,a7,a8,a9,a10,a11,a12,a13,a14,a15,a16,a17,a18,a19,a20,a21,a22,a23,a24,a25,a26,a27,a28,a29,a30,a31,a32,a33,a34,a35,a36,a37,a38,a39,a40,a41,a42,a43,a44,a45,a46,a47,a48,a49,a50,a51,a52,a53,a54,a55,a56,a57,a58,a59,a60,a61,a62,a63,a64,a65,a66,a67,a68,a69,a70,a71,a72,a73,a74,a75,a76,a77,a78,a79,a80,a81,a82,a83,a84,a85,a86,a87,a88,a89,a90,a91,a92,a93,a94,a95,a96,a97,a98,a99,a100,a101,a102,a103,a104,a105,a106,a107,a108,a109}`, `name=function(){var a,b,c,d,e,f,g,h,i,j,k,l,m,n,o,p,q,r,s,t,u,v,w,x,y,z,A,B,C,D,E,F,G,H,I,J,K,L,M,N,O,P,Q,R,S,T,U,V,W,X,Y,Z,_,$,aa,ab,ac,ad,ae,af,ag,ah,ai,aj,ak,al,am,an,ao,ap,aq,ar,as,at,au,av,aw,ax,ay,az,aA,aB,aC,aD,aE,aF,aG,aH,aI,aJ,aK,aL,aM,aN,aO,aP,aQ,aR,aS,aT,aU,aV,aW,aX,aY,aZ,a_,a$,ba}`},

		// edge-cases
		{`let o=null;try{o=(o?.a).b||"FAIL"}catch(x){}console.log(o||"PASS")`, `let o=null;try{o=(o?.a).b||"FAIL"}catch(x){}console.log(o||"PASS")`},
		{"x / /\\d+/.exec(s)[0]", "x//\\d+/.exec(s)[0]"},              // #183
		{"false`string`", "(!1)`string`"},                             // #181
		{"var a=/\\s?auto?\\s?/i\nvar b", "var a=/\\s?auto?\\s?/i,b"}, // #14
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
