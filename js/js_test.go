package js

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/buffer"
	"github.com/tdewolff/test"
)

func TestJS(t *testing.T) {
	jsTests := []struct {
		js       string
		expected string
	}{
		{`/*comment*/`, ``},
		//{`/*!comment*/`, `/*!comment*/`},
		{`debugger`, ``},
		{`"use strict"`, `"use strict"`},
		{`1.0`, `1`},
		{`1000`, `1e3`},
		{`0b1001`, `9`},
		{`0o11`, `9`},
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
		{`new RegExp()`, `new RegExp`},
		{`new new a()()`, `new(new a)`},
		{`switch (a) { case b: 5; default: 6}`, `switch(a){case b:5;default:6}`},
		{`switch (a) { case b: {var c;return c}; default: 6}`, `switch(a){case b:{var c;return c}default:6}`},
		{`with (a = b) x`, `with(a=b)x`},
		{`with (a = b) {x}`, `with(a=b)x`},
		{`import 'path'`, `import'path'`},
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
		{`async function f(){for await (var a of b){a}}`, `async function f(){for await(var a of b)a}`},
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

		// rename true, false, undefined, Infinity
		{`x=true`, `x=!0`},
		{`x=false`, `x=!1`},
		{`x=false()`, `x=(!1)()`},
		{`false`, `!1`},
		{`x=undefined`, `x=void 0`},
		{`x=undefined()`, `x=(void 0)()`},
		{`x=undefined.a`, `x=(void 0).a`},
		//{`undefined=5;x=undefined`, `undefined=5;x=undefined`},
		{`x=Infinity`, `x=1/0`},
		{`x=Infinity()`, `x=(1/0)()`},
		{`x=2**Infinity`, `x=2**(1/0)`},
		//{`Infinity=5;x=Infinity`, `Infinity=5;x=Infinity`},
		{`class a extends undefined {}`, `class a extends(void 0){}`},
		{`new true`, `new(!0)`},
		{`function*a(){yield undefined}`, `function*a(){yield}`},
		{`function*a(){yield*undefined}`, `function*a(){yield*void 0}`},

		// if/else statements
		{`if(a){return b}`, `if(a)return b`},
		{`if(a){b = 5;return b}`, `if(a)return b=5,b`},
		{`if(a)`, `a`},
		{`if(a){}`, `a`},
		{`if(a) b`, `a&&b`},
		{`if(a,b) c`, `a,b&&c`},
		{`if(a){}else;`, `a`},
		{`if(a){}else{}`, `a`},
		{`if(a){}else{;}`, `a`},
		{`if(a){}else{b}`, `a||b`},
		{`if(a)a;else b`, `a||b`},
		{`if(a)b;else b`, `a,b`},
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
		{`if(a)`, `a`},
		{`if(a)b=c`, `a&&(b=c)`},
		{`if(!a)b=c`, `a||(b=c)`},
		{`if(a||d)b=c`, `(a||d)&&(b=c)`},
		{`if(a);else b=c`, `a||(b=c)`},
		{`if(!a);else b=c`, `a&&(b=c)`},
		{`if(a)b=c;else e`, `a?b=c:e`},
		{`if(a)b=c,f;else e`, `a?(b=c,f):e`},
		{`if(a){b=c}else{if(d){e=f}else{g=h}}`, `a?b=c:d?e=f:g=h`},
		{`if(a){b=c}else if(d){e=f}else if(g){h=i}`, `a?b=c:d?e=f:g&&(h=i)`},
		{`if(a){if(b)c;else d}else{e}`, `a?b?c:d:e`},
		//{`if(a){if(b)c;else d}else{d}`, `a&&b?c:d`},
		{`b=5;return a+b`, `return b=5,a+b`},
		{`b=5;throw a+b`, `throw b=5,a+b`},
		{`if(a)return a;else return b`, `return a||b`},
		{`if(a)throw a;else throw b`, `throw a||b`},
		{`if(a)return a;else a=b`, `if(a)return a;a=b`},
		{`if(a){a++;return a}else a=b`, `if(a)return a++,a;a=b`},
		{`if(a){a++;return a}else if(b)a=b`, `if(a)return a++,a;b&&(a=b)`},
		{`if(a){a++;return}else a=b`, `if(a){a++;return}a=b`},
		//{`if(a){a++;return}else return`, `return a?void a++:void 0`},
		{`if(a){return}else {a=b;while(c){}}`, `if(a)return;for(a=b;c;){}`},
		{`if(a){a++;return a}else return`, `return a?(a++,a):void 0`},
		{`a=b;if(a){return a}else return b`, `return a=b,a||b`},
		{`if(a){return a}return b`, `return a||b`},
		{`if(a);else return a;return b`, `return a?b:a`},
		{`if(a){return a}b=c;return b`, `return a||(b=c,b)`},
		{`if(a){return}b=c;return b`, `if(a)return;return b=c,b`},
		{`if(a){return a}b=c;return`, `if(a)return a;b=c;return`},
		//{`if(a){if(b)return b;return}`, `if(a)return b||void 0`},
		//{`if(a)b=5;else b=6`, `b=a?5:6`},          // not used by Uglify? only for non-global
		//{`if(a)b[4]=5;else b[4]=6`, `b[4]=a?5:6`}, // not used by Uglify?
		//{`if(a)fun(x);else fun(y)`, `fun(a?x:y)`}, // not used by Uglify? unsafe?
		{`if(a){throw a}b=c;throw b`, `throw a||(b=c,b)`},
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
		//{`function f(){if(a){a=5;return}a=6;return a}`, `function f(){if(!a){a=6;return a}a=5}`},

		// var declarations
		{`var a;var b`, `var a,b`},
		//{`var a;var a`, `var a`},
		{`const a=1;const b=2`, `const a=1,b=2`},
		{`let a=1;let b=2`, `let a=1,b=2`},
		//{`var a;if(a)var b;else b`, `var a,b;a||b`},
		//{`var a;if(a)var b=5`, `var a;if(a)var b=5`},
		{`var a;for(var b=0;b;b++){}`, `for(var a,b=0;b;b++){}`},
		{`const a=3;for(const b=0;b;b++){}`, `const a=3;for(const b=0;b;b++){}`},
		{`var a;for(let b=0;b;b++){}`, `var a;for(let b=0;b;b++){}`},
		{`var a;while(b){}`, `for(var a;b;){}`},
		{`var [a,]=[b,]`, `var[a]=[b]`},
		{`var [a,z]=[b,c]`, `var[a,z]=[b,c]`},
		{`var [a,,]=[b,,]`, `var[a,,]=[b,,]`},
		{`var {a,}=b`, `var{a}=b`},
		{`{let a}`, `{let a}`}, // TODO: remove entire block
		{`for(var [a] in b){}`, `for(var[a]in b){}`},
		{`for(var {a} of b){}`, `for(var{a}of b){}`},

		// function and method declarations
		//{`function g(){return}`, `function g(){}`},
		//{`function g(){return undefined}`, `function g(){}`},
		//{`function g(){return void 0}`, `function g(){}`},
		//{`function g(){return;var a;a=b}`, `function g(){var a;}`},
		//{`function g(){return 5;function f(){}}`, `function g(){return 5;function f(){}}`},
		// TODO: minify for/while with if/continue or if/break constructions? Not if continue/break has label
		//{`for (var a of b){continue;a=5}`, `for(var a of b){}`},
		//{`for (var a of b){break;a=5}`, `for(var a of b){break}`},
		//{`function g(){if(a)return a;else return b;var c;c=d}`, `function g(){var c;return a||b}`},
		{`class a{static g(){}}`, `class a{static g(){}}`},
		{`class a{static [1](){}}`, `class a{static[1](){}}`},
		{`class a{static*g(){}}`, `class a{static*g(){}}`},
		{`class a{static*[1](){}}`, `class a{static*[1](){}}`},
		{`class a{get g(){}}`, `class a{get g(){}}`},
		{`class a{get [1](){}}`, `class a{get[1](){}}`},
		{`class a{set g(){}}`, `class a{set g(){}}`},
		{`class a{set [1](){}}`, `class a{set[1](){}}`},
		{`class a{static async g(){}}`, `class a{static async g(){}}`},
		{`class a{static async [1](){}}`, `class a{static async[1](){}}`},
		{`class a{static async*g(){}}`, `class a{static async*g(){}}`},
		{`class a{static async*[1](){}}`, `class a{static async*[1](){}}`},
		{`class a{"f"(){}}`, `class a{f(){}}`},
		{`class a{f(){};g(){}}`, `class a{f(){}g(){}}`},

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
		{`async a => await b`, `async a=>await b`},
		//{`var a=function(){return 5}`, `var a=()=>5`},
		//{`var a=async function(b){b=6;return 5}`, `var a=async b=>(b=6,5)`},
		//{`(function(){return 5})()`, `(()=>5)()`},
		//{`class c{a(){return 5}}`, `class c{a:()=>5}`},
		//{`export default{a(){return 5}}`, `export default{a:()=>5}`},
		//{`var v={async [[1]](a){return a}}`, `var v={[[1]]:async a=>a}`},
		//{`var a={b:()=>c=5}`, `var a={b(){c=5}}`},
		//{`var a={b:function(){c=5}}`, `var a={b(){c=5}}`},
		//{`var a={b:async function(){c=5}}`, `var a={async b(){c=5}}`},
		//{`var a={b:function*(){c=5}}`, `var a={*b(){c=5}}`},

		// remove groups
		{`a=(b+c)+d`, `a=b+c+d`},
		{`a=b+(c+d)`, `a=b+(c+d)`},
		{`a=b*(c+d)`, `a=b*(c+d)`},
		{`a=(b*c)+d`, `a=b*c+d`},
		{`a=(b.c)++`, `a=b.c++`},
		{`a=(b++).c`, `a=(b++).c`},
		{`a=!(b++)`, `a=!b++`},
		{`a=(b+c)(d)`, `a=(b+c)(d)`},
		{`a=b**(c**d)`, `a=b**c**d`},
		{`a=(b**c)**d`, `a=(b**c)**d`},
		{`a=false**2`, `a=(!1)**2`},
		{`a=(++b)**2`, `a=++b**2`},
		{`a=(a||b)&&c`, `a=(a||b)&&c`},
		{`a=a||(b&&c)`, `a=a||b&&c`},
		{`a=(a&&b)||c`, `a=a&&b||c`},
		{`a=a&&(b||c)`, `a=a&&(b||c)`},
		{`a=c&&(a??b)`, `a=c&&(a??b)`},
		{`a=(a||b)||(c||d)`, `a=a||b||c||d`},
		{`a=!(!b)`, `a=!!b`},
		{`a=(b())`, `a=b()`},
		{`a=(b)?.(c,d)`, `a=b?.(c,d)`},
		{`a=(b,c)?.(d)`, `a=(b,c)?.(d)`},
		{`a=(b?c:e)?.(d)`, `a=(b?c:e)?.(d)`},
		{`a=b?c:c`, `a=(b,c)`},
		{`a=b?b:c=f`, `a=b?b:c=f`}, // don't write as a=b||(c=f)
		{`a=b||(c=f)`, `a=b||(c=f)`},
		{`a=(-5)**3`, `a=(-5)**3`},
		{`a=5**(-3)`, `a=5**(-3)`},
		{`a=(-(+5))**3`, `a=(-+5)**3`}, // could remove +
		{`a=(b,c)+3`, `a=(b,c)+3`},
		{`(a,b)&&c`, `a,b&&c`},
		{`function*x(){a=(yield b)}`, `function*x(){a=yield b}`},
		{`function*x(){a=yield (yield b)}`, `function*x(){a=yield yield b}`},
		{`if((a))while((b)){}`, `if(a)while(b){}`},
		{`({a}=5)`, `({a})=5`},
		{`(function(){})`, `!function(){}`},
		{`(function(){}())`, `!function(){}()`},
		{`(function(){})()`, `!function(){}()`},
		{`(function(){})();x=5;f=6`, `!function(){}(),x=5,f=6`},
		{`(async function(){})`, `!async function(){}`},
		{`(class a{})`, `!class a{}`},
		{`(let [a])`, `!let[a]`},
		{`x=(function(){})`, `x=function(){}`},
		{`x=(function(){}())`, `x=function(){}()`},
		{`x=(function(){})()`, `x=function(){}()`},
		{`x=(function(){}).a`, `x=function(){}.a`},
		{`await(x+y)`, `await(x+y)`},
		{`async function g(){await(x+y)}`, `async function g(){await(x+y)}`},
		{`await(fun()())`, `await(fun()())`},
		{`async function g(){await(fun()())}`, `async function g(){await fun()()}`},
		{`a=1+"2"+(3+4)`, `a=1+"2"+(3+4)`},
		{`(-1)()`, `(-1)()`},
		{`(-1)(-2)`, `(-1)(-2)`},
		{`(+new Date).toString(32)`, `(+new Date).toString(32)`},
		{`new(a.b)instanceof c`, `new a.b instanceof c`},
		{`(2).toFixed(0)`, `2..toFixed(0)`},
		{`(0.2).toFixed(0)`, `.2.toFixed(0)`},
		{`(2e-8).toFixed(0)`, `2e-8.toFixed(0)`},
		{`(-2).toFixed(0)`, `(-2).toFixed(0)`},
		{`(a)=>((b)=>c)`, `a=>b=>c`},
		{`function f(a=(3+2)){}`, `function f(a=3+2){}`},
		{`function*a(){yield a.b}`, `function*a(){yield a.b}`},
		{`function*a(){(yield a).b}`, `function*a(){(yield a).b}`},
		{`function*a(){yield a["-"]}`, `function*a(){yield a["-"]}`},
		{`function*a(){(yield a)["-"]}`, `function*a(){(yield a)["-"]}`},

		// other
		//{`a=a+5`, `a+=5`},
		//{`a=5+a`, `a+=5`},
		{`async function g(){await x+y}`, `async function g(){await x+y}`},
		//{`!a&&!b&&!c`, `!(a||b||c)`}, // can be unsafe if not all are truthy/falsy
		{`a?true:false`, `!!a`},
		{`a==b?true:false`, `a==b`},
		{`!a?true:false`, `!a`},
		{`a?false:true`, `!a`},
		{`!a?false:true`, `!!a`},
		{`a?!0:!1`, `!!a`},
		{`a?0:1`, `a?0:1`},
		{`!!a?0:1`, `!!a?0:1`},
		{`a&&b?!1:!0`, `!(a&&b)`},
		{`a&&b?!0:!1`, `!!(a&&b)`},
		{`a?true:5`, `!!a||5`},
		{`a?5:false`, `!!a&&5`},
		{`!a?true:5`, `!a||5`},
		{`!a?5:false`, `!a&&5`},
		{`a==b?true:5`, `a==b||5`},
		{`a!=b?true:5`, `a!=b||5`},
		{`a==b?false:5`, `a!=b&&5`},
		{`a!=b?false:5`, `a==b&&5`},
		{`a==b?5:true`, `a!=b||5`},
		{`a==b?5:false`, `a==b&&5`},
		{`a<b?5:true`, `!(a<b)||5`},
		{`!(a<b)?5:true`, `a<b||5`},
		{`!42`, `!1`},
		{`!"str"`, `!1`},
		{`!/regexp/`, `!1`},
		{`new a()`, `new a`},
		{`new a()()`, `(new a)()`},
		{`a={"property": val1, "2": val2, "3name": val3};`, `a={property:val1,2:val2,"3name":val3}`},
		{`a=obj["if"]`, `a=obj.if`},
		{`a=obj["2"]`, `a=obj[2]`},
		{`a=obj["3name"]`, `a=obj["3name"]`},

		// merge expressions
		{`a();b();return c()`, `return a(),b(),c()`},
		{`a();b();throw c()`, `throw a(),b(),c()`},
		{`a=5;if(b)while(c){}`, `if(a=5,b)while(c){}`},
		{`a=5;for(;b;)c()`, `for(a=5;b;)c()`},
		{`a=5;for(b=4;b;)c()`, `for(a=5,b=4;b;)c()`},
		{`a=5;for(var b=4;b;)c()`, `a=5;for(var b=4;b;)c()`},
		{`a=5;switch(b=4){}`, `switch(a=5,b=4){}`},
		{`a=5;with(b=4){}`, `with(a=5,b=4){}`},
		{`(function(){})();(function(){})()`, `!function(){}(),function(){}()`},

		// edge-cases
		{`let o=null;try{o=(o?.a).b||"FAIL"}catch(x){}console.log(o||"PASS")`, `let o=null;try{o=(o?.a).b||"FAIL"}catch(x){}console.log(o||"PASS")`},
		{"var a=/\\s?auto?\\s?/i\nvar b", "var a=/\\s?auto?\\s?/i,b"}, // #14
		{"false`string`", "(!1)`string`"},                // #181
		{"x / /\\d+/.exec(s)[0]", "x//\\d+/.exec(s)[0]"}, // #183
	}

	m := minify.New()
	o := Minifier{KeepVarNames: true}
	for _, tt := range jsTests {
		t.Run(tt.js, func(t *testing.T) {
			r := bytes.NewBufferString(tt.js)
			w := &bytes.Buffer{}
			err := o.Minify(m, w, r, nil)
			test.Minify(t, tt.js, err, w.String(), tt.expected)
		})
	}
}

func TestJSVarRenaming(t *testing.T) {
	jsTests := []struct {
		js       string
		expected string
	}{
		{`x=function(){var name}`, `x=function(){var a}`},
		{`x=function(){var name; name++}`, `x=function(){var a;a++}`},
		{`x=function(){try{var x}catch(y){x}}`, `x=function(){try{var a}catch(b){a}}`},
		{`x=function(){try{var x}catch(x){x}}`, `x=function(){try{var a}catch(a){a}}`},
		{`x=function(){function name(){}}`, `x=function(){function a(){}}`},
		//{`x=function name(){}`, `x=function(){}`},
		{`x=function(){let a;{let b;a}}`, `x=function(){let a;{let b;a}}`},
		{`x=function({foo, bar}){}`, `x=function({foo:a,bar:b}){}`},
		{`x=function(){class Wheel{}}`, `x=function(){class a{}}`},
		{`x=function(){function name(arg1, arg2){return arg1, arg2}}`, `x=function(){function a(a,b){return a,b}}`},
		{`x=function(){function name(arg1, arg2){return arg1, arg2} return arg1}`, `x=function(){function a(a,b){return a,b}return arg1}`},
		{`x=function(){function name(arg1, arg2){return arg1, arg2} return a}`, `x=function(){function b(b,c){return b,c}return a}`},
		{`x=function(){function add(l,r){return add(l,r)}function nadd(l,r){return-add(l,r)}}`, `x=function(){function b(a,c){return b(a,c)}function a(a,c){return-b(a,c)}}`},
		{`function a(){var b}`, `function a(){var a}`},
		{`!function(){x=function(){return fun()};var fun=function(){return 0}}`, `!function(){x=function(){return a()};var a=function(){return 0}}`},
		{`!function(){var x=function(){return y};const y=5}`, `!function(){var a=function(){return b};const b=5}`},
		{`!function(){if(1)const x=5;var y=function(){return x}}`, `!function(){if(1)const b=5;var a=function(){return b}}`},
		{`!function(){if(1){const x=5;5}var y=function(){return x}}`, `!function(){if(1){const a=5;5}var a=function(){return x}}`},
		{`!function(){var x=function(){return y};if(1)const y=5}`, `!function(){var a=function(){return b};if(1)const b=5}`},
		{`!function(){var x=function(){return y};if(1){const y=5;5}}`, `!function(){var a=function(){return y};if(1){const a=5;5}}`},
		{`!function(){var x=function(){return y};if(1)var y=5}`, `!function(){var a=function(){return b};if(1)var b=5}`},
		{`!function(){var x=function(){return y};if(1){var y=5;5}}`, `!function(){var a=function(){return b};if(1){var b=5;5}}`},
		{`!function(){var x,y,z=(x,y)=>x+y}`, `!function(){var a,b,c=(a,b)=>a+b}`},
		{`!function(){var await;print({await});}`, `!function(){var a;print({await:a})}`},
		{`function a(){var name; return {name}}`, `function a(){var a;return{name:a}}`},
		{`function a(){try{}catch(arg){arg}}`, `function a(){try{}catch(a){a}}`},
		{`function a(){var name;try{}catch(name){var name}}`, `function a(){var a;try{}catch(b){var a}}`},
		{`function a(){var name;try{}catch(arg){var name}}`, `function a(){var a;try{}catch(b){var a}}`},
		{`name=function(){var a001,a002,a003,a004,a005,a006,a007,a008,a009,a010,a011,a012,a013,a014,a015,a016,a017,a018,a019,a020,a021,a022,a023,a024,a025,a026,a027,a028,a029,a030,a031,a032,a033,a034,a035,a036,a037,a038,a039,a040,a041,a042,a043,a044,a045,a046,a047,a048,a049,a050,a051,a052,a053,a054,a055,a056,a057,a058,a059,a060,a061,a062,a063,a064,a065,a066,a067,a068,a069,a070,a071,a072,a073,a074,a075,a076,a077,a078,a079,a080,a081,a082,a083,a084,a085,a086,a087,a088,a089,a090,a091,a092,a093,a094,a095,a096,a097,a098,a099,a100,a101,a102,a103,a104,a105,a106,a107,a108,a109}`, `name=function(){var a,b,c,d,e,f,g,h,i,j,k,l,m,n,o,p,q,r,s,t,u,v,w,x,y,z,A,B,C,D,E,F,G,H,I,J,K,L,M,N,O,P,Q,R,S,T,U,V,W,X,Y,Z,_,$,aa,ab,ac,ad,ae,af,ag,ah,ai,aj,ak,al,am,an,ao,ap,aq,ar,at,au,av,aw,ax,ay,az,aA,aB,aC,aD,aE,aF,aG,aH,aI,aJ,aK,aL,aM,aN,aO,aP,aQ,aR,aS,aT,aU,aV,aW,aX,aY,aZ,a_,a$,ba,bb}`}, // 'as' is a keyword
	}

	m := minify.New()
	o := Minifier{}
	for _, tt := range jsTests {
		t.Run(tt.js, func(t *testing.T) {
			r := bytes.NewBufferString(tt.js)
			w := &bytes.Buffer{}
			err := o.Minify(m, w, r, nil)
			test.Minify(t, tt.js, err, w.String(), tt.expected)
		})
	}
}

func BenchmarkJQuery(b *testing.B) {
	m := minify.New()
	buf, err := ioutil.ReadFile("../benchmarks/sample_jquery.js")
	if err != nil {
		panic(err)
	}
	for j := 0; j < 10; j++ {
		b.Run(fmt.Sprintf("%d", j), func(b *testing.B) {
			b.SetBytes(int64(len(buf)))
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				r := buffer.NewReader(parse.Copy(buf))
				w := buffer.NewWriter(make([]byte, 0, len(buf)))
				b.StartTimer()

				if err := Minify(m, w, r, nil); err != nil {
					b.Fatal(err)
				}
			}
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
