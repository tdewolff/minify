package css

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/parse/v2/css"
	"github.com/tdewolff/test"
)

func TestCSS(t *testing.T) {
	cssTests := []struct {
		css      string
		expected string
	}{
		{"/*comment*/", ""},
		{"/*! bang  comment */", "/*!bang comment*/"},
		{"a;", "a"},
		{"i{}/*! bang  comment */", "i{}/*!bang comment*/"},
		{"i { key: value; key2: value; }", "i{key:value;key2:value}"},
		{".cla .ss > #id { x:y; }", ".cla .ss>#id{x:y}"},
		{".cla[id ^= L] { x:y; }", ".cla[id^=L]{x:y}"},
		{"area:focus { outline : 0;}", "area:focus{outline:0}"},
		{"@import 'file';", "@import 'file'"},
		{"@import url('file');", "@import 'file'"},
		{"@import url(//url);", `@import "//url"`},
		{"@import url(\n//url\n);", `@import "//url"`},
		{"@import url();", `@import ""`},
		{"@font-face { x:y; }", "@font-face{x:y}"},

		{"input[type=\"radio\"]{x:y}", "input[type=radio]{x:y}"},
		{"input[type=\"radio\" i]{x:y}", "input[type=radio i]{x:y}"},
		{"DIV{margin:1em}", "div{margin:1em}"},
		{".CLASS{margin:1em}", ".CLASS{margin:1em}"},
		{"@MEDIA all{}", "@media all{}"},
		{"@media only screen and (max-width : 800px){}", "@media only screen and (max-width:800px){}"},
		{"@media (-webkit-min-device-pixel-ratio:1.5),(min-resolution:1.5dppx){}", "@media(-webkit-min-device-pixel-ratio:1.5),(min-resolution:1.5dppx){}"},
		{"[class^=icon-] i[class^=icon-],i[class*=\" icon-\"]{x:y}", "[class^=icon-] i[class^=icon-],i[class*=\" icon-\"]{x:y}"},
		{"html{line-height:1;}html{line-height:1;}", "html{line-height:1}html{line-height:1}"},
		{"a { b: 1", "a{b:1}"},
		{"@unknown { border:1px solid #000 }", "@unknown{border:1px solid #000 }"},
		{":root { --custom-variable:0px; }", ":root{--custom-variable:0px}"},

		// recurring property overwrites previous
		//{"a{color:blue;color:red}", "a{color:red}"},
		//{"a{unknownprop:blue;unknownprop:red}", "a{unknownprop:red}"},
		//{"a{unknownprop:blue;otherunknownprop:red}", "a{unknownprop:blue;otherunknownprop:red}"},
		//{"a{background-color:blue;background:0 0}", "a{background:0 0}"},
		//{"a{color:blue}a{color:red}", "a{color:blue}a{color:red}"}, // not supported

		// case sensitivity
		{"@counter-style Ident{}", "@counter-style Ident{}"},

		// coverage
		{"a, b + c { x:y; }", "a,b+c{x:y}"},

		// bad declaration
		{".clearfix { *zoom: 1px; }", ".clearfix{*zoom:1px}"},
		{".clearfix { *zoom: 1px }", ".clearfix{*zoom:1px}"},
		{".clearfix { order:4; *zoom: 1px; color:red; }", ".clearfix{order:4;*zoom:1px;color:red}"},

		// go-fuzz
		{"input[type=\"\x00\"] {  a: b\n}.a{}", "input[type=\"\x00\"]{a:b}.a{}"},
		{"a{a:)'''", "a{a:)'''}"},
		{"{T:l(", "{t:l()}"},
		{"{background:0 0 0", "{background:0 0}"},
		{"{d:url( \n  \n\t0", "{d:url()}"},
		{"{d:urL(     '0", `{d:url("'")}`},
		{`{-ms-filter:"`, `{-ms-filter:"}`},
	}

	m := minify.New()
	for _, tt := range cssTests {
		t.Run(tt.css, func(t *testing.T) {
			r := bytes.NewBufferString(tt.css)
			w := &bytes.Buffer{}
			err := Minify(m, w, r, nil)
			test.Minify(t, tt.css, err, w.String(), tt.expected)
		})
	}

	// coverage
	test.T(t, Token{css.IdentToken, []byte("data"), nil, 0, 0}.String(), "Ident(data)")
	test.T(t, Token{css.FunctionToken, nil, []Token{{css.IdentToken, []byte("data"), nil, 0, 0}}, 0, 0}.String(), "[Ident(data)]")
}

func TestCSSInline(t *testing.T) {
	cssTests := []struct {
		css      string
		expected string
	}{
		{"/*comment*/", ""},
		{"/*! bang  comment */", ""},
		{"}", "}"},
		{";", ""},
		{"empty:", "empty:"},
		{"key: value;", "key:value"},
		{"margin: 0 1; padding: 0 1;", "margin:0 1;padding:0 1"},
		{"color: #FF0000;", "color:red"},
		{"color: #000000;", "color:#000"},
		{"color: #aabbccdd;", "color:#abcd"},
		{"color: #aabbccff;", "color:#abc"},
		{"color: #aabbcc00;", "color:#0000"},
		{"color: black;", "color:#000"},
		{"color: rgb(255,255,255);", "color:#fff"},
		{"color: rgb(100%,100%,100%);", "color:#fff"},
		{"color: rgba(255,0,0,1);", "color:red"},
		{"color: rgba(255,0,0,2);", "color:red"},
		{"color: rgba(255,0,0,0.5);", "color:rgba(255,0,0,.5)"},             // {"color: rgba(255,0,0,0.5);", "color:#ff000080"},
		{"color: rgba(0,0,0,-1);", "color:transparent"},                     // {"color: rgba(255,0,0,-1);", "color:#0000"},
		{"color: rgba(0%,0%,0%,0.2);", "color:rgba(0,0,0,.2)"},              // {"color: rgba(0%,15%,25%,0.2);", "color:#00264033"},
		{"color: rgba(0,0,0,0.5);", "color:rgba(0,0,0,.5)"},                 // {"color: rgba(0,0,0,0.5);", "color:#00000080"},
		{"color: rgba(0,0,0,0.264705882);", "color:rgba(0,0,0,.264705882)"}, // {"color: rgba(0,0,0,0.264705882);", "color:#0004"},
		{"color: rgba(0,0,0,0.025);", "color:rgba(0,0,0,.025)"},
		{"color: rgba(255 0 0 / 1);", "color:red"},
		{"color: rgba(0 100% 50% / 100%);", "color:#00ff80"},
		{"color: rgba(0 100% 50% / 60%);", "color:rgba(0 100% 50%/.6)"},
		{"color: rgb(20%,40%,60%,50%);", "color:rgb(51,102,153,.5)"},
		{"color: rgb(0%,80%,100%,50%);", "color:rgb(0,204,255,.5)"},
		{"color: rgb(1,2,3,90%);", "color:rgb(1,2,3,.9)"},
		{"color: rgb(1,2,3,.01);", "color:rgb(1,2,3,1%)"},
		{"color: rgb(1,2,3,.0099);", "color:rgb(1,2,3,.99%)"},
		{"color: hsla(5,0%,10%,0.75);", "color:hsla(5,0%,10%,.75)"}, // {"color: hsla(5,0%,10%,0.75);", "color:#1a1a1abf"},
		{"color: hsl(0,100%,50%);", "color:red"},
		{"color: hsl(-360,100%,50%);", "color:red"},
		{"color: hsla(1,2%,3%,1);", "color:#080807"},
		{"color: hsla(0,0%,0%,0);", "color:transparent"}, // {"color: hsla(1,2%,3%,0);", "color:#0000"},
		{"color: hsl(48,100%,50%);", "color:#fc0"},
		{"color: hsla(0 100% 50% / 1);", "color:red"},
		{"color: hsla(0 100% 50% / 60%);", "color:hsla(0 100% 50%/.6)"},
		{"color: hsla(400, 150%, 150%, 2);", "color:#fff"},
		//{"color: hwb(0 0% 0%);", "color:red"}, TODO
		//{"color: hwb(120 20% 20%/50%);", "color:"}, TODO
		{"background-color:transparent", "background-color:initial"},
		{"background-position:top", "background-position:top"},
		{"background-position:bottom", "background-position:bottom"},
		{"background-position:center", "background-position:50%"},
		{"background-position:left 50%", "background-position:0"},
		{"background-position:center center", "background-position:50%"},
		{"background-position:center bottom", "background-position:50% 100%"},
		{"background-position:top right", "background-position:100% 0"},
		{"background-position:bottom left", "background-position:0 100%"},
		{"background-position:top center", "background-position:50% 0"},
		{"background-position:bottom center", "background-position:50% 100%"},
		{"background-position:bottom 5% right 0%", "background-position:100% 95%"},
		{"background-position:bottom 0 right 10%", "background-position:90% 100%"},
		{"background-position:top 10% left 5%", "background-position:5% 10%"},
		{"background-position:top 10% left", "background-position:0 10%"},
		{"background-position:left 10% top", "background-position:10% 0"},
		{"background-position:center left 5%", "background-position:5%"},
		{"background-position:center right 10%", "background-position:90%"},
		{"background-position:right .75rem center", "background-position:right .75rem center"},
		{"background-position:right 50% bottom 50%", "background-position:50%"},
		{"background-position:right 100% bottom 100%", "background-position:0 0"},
		{"background-position:left 1% center", "background-position:1%"},
		{"background-position:center top 1%", "background-position:50% 1%"},
		{"background-position:right 0 top 0", "background-position:100% 0"},
		{"background-position:center 0px center 0%", "background-position:50%"},
		{"background-position:center 0px center 0%,,right 0 top 0", "background-position:50%,,100% 0"},
		{"background-repeat:space space", "background-repeat:space"},
		{"background-repeat:round round", "background-repeat:round"},
		{"background-repeat:repeat repeat", "background-repeat:repeat"},
		{"background-repeat:no-repeat no-repeat", "background-repeat:no-repeat"},
		{"background-repeat:repeat no-repeat", "background-repeat:repeat-x"},
		{"background-repeat:no-repeat repeat", "background-repeat:repeat-y"},
		{"background-repeat:no-repeat repeat,,repeat no-repeat", "background-repeat:repeat-y,,repeat-x"},
		{"background-size:auto auto", "background-size:auto"},
		{"background-size:30% auto", "background-size:30%"},
		{"background-size:200px auto", "background-size:200px"},
		{"background-size:200px auto,,30% auto", "background-size:200px,,30%"},
		{"background:red", "background:red"},
		{"background:red none", "background:red"},
		{"background:red none 0 0", "background:red"},
		{"background:#ff0000 none 1 1", "background:red 1 1"},
		{"background:#0000 1 1", "background:1 1"},
		{"background:transparent", "background:0 0"},
		{"background:transparent no-repeat", "background:no-repeat"},
		{"background:#0000 none padding-box 0 0 / auto auto scroll border-box repeat repeat", "background:0 0"},
		{"background:0 0 / auto", "background:0 0"},
		{"background:0 0 / auto 10%", "background:0 0/auto 10%"},
		{"background:0 / auto 10%", "background:0/auto 10%"},
		{"background:0 0/200px auto", "background:0 0/200px"},
		{"background:0 0 no-repeat", "background:no-repeat"},
		{"background:0% 0%", "background:0 0"},
		{"background:left top", "background:0 0"},
		{"background:no-repeat repeat", "background:repeat-y"},
		{"background:top right", "background:100% 0"},
		{"background:bottom left", "background:0 100%"},
		{"background:#fff url(foo.svg) no-repeat right .75rem center / auto calc(100% - 1.5rem)", "background:#fff url(foo.svg)no-repeat right .75rem center/auto calc(100% - 1.5rem)"},
		{"background:#fff / 5% auto", "background:#fff/5%"},
		{"background:#fff / auto 5%", "background:#fff/auto 5%"},
		{"background:#fff / auto 78px", "background:#fff/auto 78px"},
		{"background:calc(5%-2%) center", "background:calc(5%-2%)"},
		{"background:0 0 / 80% no-repeat url('firefox-logo.svg'), white 0 0 url('lizard.png');", "background:0 0/80% no-repeat url(firefox-logo.svg),#fff url(lizard.png)"},
		{"background:rgba(255,0,0,1) url(foo.svg) no-repeat right .75rem center / auto calc(100% - 1.5rem)", "background:red url(foo.svg)no-repeat right .75rem center/auto calc(100% - 1.5rem)"},
		{"background:left top,#fff / 5% auto", "background:0 0,#fff/5%"},
		{"box-shadow:0 0 0 0", "box-shadow:0 0"},
		{"box-shadow:0 0 0 0,0 0 0 0", "box-shadow:0 0,0 0"},
		{"box-shadow:0 inset 0 0 blue 0", "box-shadow:0 inset 0 blue"},
		{"box-shadow:rgba(0,0,0,0) 0 8px", "box-shadow:transparent 0 8px"},
		{"box-shadow: inset .5em 0,, .39em 0", "box-shadow:inset .5em 0,,.39em 0"},
		{"box-shadow:initial", "box-shadow:none"},
		{"font-weight: normal;", "font-weight:400"},
		{"font-weight: bold;", "font-weight:700"},
		{"font: ;", "font:"},
		{"font: 2em;", "font:2em"},
		{"font: caption;", "font:caption"},
		{"font: bold 5px \"Times new Roman\",\"Sans-Serif\";", "font:700 5px times new roman,sans-serif"},
		{"font: bold xx-small times new roman;", "font:700 xx-small times new roman"},
		{"font: normal normal normal normal 20px normal", "font:20px normal"},
		{"font:27px/13px arial,sans-serif", "font:27px/13px arial,sans-serif"},
		{"font:normal normal bold normal medium/normal arial,sans-serif", "font:700 medium arial,sans-serif"},
		{"font:400 medium/normal 'Arial'", "font:medium arial"},
		{"font:medium/normal 'Arial'", "font:medium arial"},
		{"font-family:'Arial', 'Times New Roman';", "font-family:arial,times new roman"},
		{"font-family:'a  b';", "font-family:'a  b'"},
		{"font-family:' a b ';", "font-family:' a b '"},
		{"outline: none;", "outline:none"},
		{"outline: solid black 0;", "outline:solid #000 0"},
		{"outline: none black medium;", "outline:#000"},
		{"outline: none !important;", "outline:none!important"},
		{"border-left: none;", "border-left:none"},
		{"border-left: none 0;", "border-left:0"},
		{"border-left: none medium currentcolor;", "border-left:none"},
		{"border-left: 0 dashed red;", "border-left:0 dashed red"},
		{"border-color: currentcolor red currentcolor;", "border-color:initial red initial"},
		{"border-color: currentcolor currentcolor currentcolor;", "border-color:initial"},
		{"border-color: red red red;", "border-color:red"},
		{"border-left-color: currentcolor;", "border-left-color:initial"},
		{"border-left-color: black;", "border-left-color:#000"},
		{"border: medium none;", "border:none"}, // #294
		{"outline-color: white;", "outline-color:#fff"},
		{"column-rule: medium currentcolor none;", "column-rule:none"},
		{"column-rule: medium white none;", "column-rule:#fff"},
		{"text-shadow: white 5px 5px;", "text-shadow:#fff 5px 5px"},
		{"text-decoration: currentcolor none solid", "text-decoration:none"},
		{"text-decoration: white none solid", "text-decoration:#fff"},
		{"text-emphasis: none currentcolor", "text-emphasis:none"},
		{"text-emphasis: none white", "text-emphasis:#fff"},
		{"margin: 1 1 1 1;", "margin:1"},
		{"margin: 1 2 1 2;", "margin:1 2"},
		{"margin: 1 2 3 2;", "margin:1 2 3"},
		{"margin: 1 2 3 4;", "margin:1 2 3 4"},
		{"margin: 1 1 1 a;", "margin:1 1 1 a"},
		{"margin: 1 1 1 1 !important;", "margin:1!important"},
		{"padding:.2em .4em .2em", "padding:.2em .4em"},
		{"margin: 0em;", "margin:0"},
		{"filter: progid : DXImageTransform.Microsoft.BasicImage(rotation=1);", "filter:progid:DXImageTransform.Microsoft.BasicImage(rotation=1)"},
		{"filter: progid:DXImageTransform.Microsoft.Alpha(Opacity=0);", "filter:alpha(opacity=0)"},
		{"content: \"a\\\nb\";", "content:\"ab\""},
		{"content: \"a\\\r\nb\\\r\nc\";", "content:\"abc\""},
		{"content: \"\";", "content:\"\""},
		{"flex:5 1 0", "flex:5"},
		{"flex:5 1 0%", "flex:5"},
		{"flex:5 1 0px", "flex:5"},
		{"flex:5 0 0%", "flex:5 0"},
		{"flex:5 0 0px", "flex:5 0"},
		{"flex:0 1 auto", "flex:initial"},
		{"flex:1 1 auto", "flex:auto"},
		{"flex:0 0 auto", "flex:none"},
		{"flex:0 0 5000%", "flex:0 0 5e3%"},
		{"flex:initial", "flex:initial"},
		{"flex:auto", "flex:auto"},
		{"flex:none", "flex:none"},
		{"flex:5 auto", "flex:5 auto"},
		{"flex:5 0px", "flex:5"},
		{"flex:5 0%", "flex:5"},
		{"flex:5 0", "flex:5 0"},
		{"flex-basis:0px", "flex-basis:0"},
		{"flex-basis:0%", "flex-basis:0"},
		{"flex-basis:initial", "flex-basis:auto"},
		{"flex-grow:initial", "flex-grow:0"},
		{"flex-shrink:initial", "flex-shrink:1"},
		{"order:initial", "order:0"},

		{"text-decoration: none !important", "text-decoration:none!important"},
		{"color:#fff", "color:#fff"},
		{"border:2px rgb(255,255,255);", "border:2px #fff"},
		{"margin:-1px", "margin:-1px"},
		{"margin:+1px", "margin:1px"},
		{"margin:0.5em", "margin:.5em"},
		{"margin:-0.5em", "margin:-.5em"},
		{"margin:05em", "margin:5em"},
		{"margin:.50em", "margin:.5em"},
		{"margin:5.0em", "margin:5em"},
		{"margin:5000em", "margin:5e3em"}, // CR: CSS Values and Units Module Level 3
		{"color:#c0c0c0", "color:silver"},
		{"-ms-filter: \"progid:DXImageTransform.Microsoft.Alpha(Opacity=80)\";", "-ms-filter:\"alpha(opacity=80)\""},
		{"filter: progid:DXImageTransform.Microsoft.Alpha(Opacity = 80);", "filter:alpha(opacity=80)"},
		{"width:attr(Name em)", "width:attr(Name em)"},
		{"content:CounterName", "content:CounterName"},
		{"background:url('http://domain.com/image.png');", "background:url(http://domain.com/image.png)"},
		{"background:url( 'http://domain.com/image.png' );", "background:url(http://domain.com/image.png)"},
		{"background:url(/*nocomment*/)", "background:url(/*nocomment*/)"},
		{"background:url(data:,text)", "background:url(data:,text)"},
		{"background:url('data:text/xml; version = 2.0,content')", "background:url(data:text/xml;version=2.0,content)"},
		{"background:url('data:\\'\",text')", "background:url('data:\\'\",text')"},
		{"margin:0 0 18px 0;", "margin:0 0 18px"},
		{"z-index:1000", "z-index:1000"},
		//{"flex:0px", "flex:0q"}, // TODO
		{"g:url('abc\\\ndef')", "g:url(abcdef)"},
		{"url:local('abc\\\ndef')", "url:local(abcdef)"},
		{"url:local('abc def') , url('abc def') format('truetype')", "url:local('abc def'),url('abc def')format('truetype')"},

		// case
		{"MARGIN:1EM", "margin:1em"},
		{"color:CYAN", "color:CYAN"},
		{"background:URL(x.PNG);", "background:URL(x.PNG)"},
		{"background:url(url) TOP RIGHT REPEAT-Y", "background:url(url)100% 0 REPEAT-Y"},
		{"background:url(url)TOP RIGHT REPEAT-Y", "background:url(url)100% 0 REPEAT-Y"},

		{"margin:calc(10px) calc(20px)", "margin:calc(10px)calc(20px)"},
		{"border-left:0 none", "border-left:0"},
		{"--custom-variable:0px;", "--custom-variable:0px"},
		{"--foo: 0px ;", "--foo:0px"},
		{"--foo: if(x > 5) this.width = 10", "--foo:if(x > 5) this.width = 10"},
		{"--foo: ;", "--foo: "},               // whitespace value
		{"--foo:;", "--foo:"},                 // empty value
		{"--foo: initial ;", "--foo:initial"}, // invalid value, serializes to empty
		{"color=blue;", "color=blue"},
		{"x: white , white", "x:white,white"},

		// TODO: functions
		{"width:calc(0%-0px)", "width:calc(0%0)"}, // invalid
		{"width:calc(0% - 0px)", "width:calc(0% - 0)"},
		{"width:calc(calc(0% - 0px) + 1em)", "width:calc(calc(0% - 0) + 1em)"},
		//{"width:calc(5px);", "width:5px"},
		//{"width:calc(5px - 3px);", "width:2px"},
		//{"width:calc(5px + -3px);", "width:2px"},
		//{"width:calc(5px - 3%);", "width:calc(5px - 3%)"},
		//{"width:calc(2*5px);", "width:10px"},
		//{"width:calc(10px/2);", "width:5px"},
		//{"width:calc(calc(5px));", "width:5px"},
		//{"width:calc(calc(5px - 1em)*3);", "width:calc((5px - 1em)*3)"},
		//{"width:calc(calc(5px - 1em) - 3%);", "width:calc(5px - 1em - 3%)"},
		//{"width:calc(3% - calc(5px - 1em));", "width:calc(3% - 5px + 1em)"},
		//{"width:calc(5px-3px);", "width:calc(5px-3px)"}, // invalid
		//{"width:calc(5px*3px);", "width:calc(5px*3px)"}, // invalid
		//{"width:calc(5px/3px);", "width:calc(5px/3px)"}, // invalid

		// TODO: dimensions
		//{"any:0deg 0s 0ms 0dpi 0dpcm 0dppx 0hz 0khz", "any:0 0s 0s 0dpi 0dpi 0dpi 0hz 0hz"},
		//{"width:96px", "width:1in"},
		//{"width:72pt", "width:1in"},
		//{"width:12pc", "width:2in"},
		//{"width:0.166666666666667in", "width:1pc"},
		//{"width:0.0625pc", "width:1px"},
		//{"width:40Q", "width:40q"},
		//{"width:120Q", "width:3cm"},
		//{"width:10mm", "width:1cm"},
		//{"width:120Q", "width:3cm"},
		//{"transform:rotate(360deg)", "transform:rotate(1turn)"},
		//{"transform:rotate(180deg)", "transform:rotate(180deg)"},
		//{"transform:rotate(100grad)", "transform:rotate(90deg)"},
		//{"transform:rotate(.25turn)", "transform:rotate(90deg)"},
		//{"transform:rotate(6.28318530717959rad)", "transform:rotate(1turn)"}, // TODO

		// case sensitivity
		{"animation:Ident", "animation:Ident"},
		{"animation-name:Ident", "animation-name:Ident"},

		// coverage
		{"margin: 1 1;", "margin:1"},
		{"margin: 1 2;", "margin:1 2"},
		{"margin: 1 1 1;", "margin:1"},
		{"margin: 1 2 1;", "margin:1 2"},
		{"margin: 1 2 3;", "margin:1 2 3"},
		// {"margin: 0%;", "margin:0"}, // TODO
		{"color: rgb(255,64,64);", "color:#ff4040"},
		{"color: rgb(256,-34,2342435);", "color:#f0f"},
		{"color: rgb(120%,-45%,234234234%);", "color:#f0f"},
		{"color: rgb(0, 1, ident);", "color:rgb(0,1,ident)"},
		{"color: rgb(ident);", "color:rgb(ident)"},
		{"color: hsl(0,-1%,-1%);", "color:#000"},
		{"color: hsl(-180,100%,50%);", "color:#0ff"},
		{"margin: rgb(ident);", "margin:rgb(ident)"},
		{"filter: progid:b().c.Alpha(rgba(x));", "filter:progid:b().c.Alpha(rgba(x))"},
		{"margin: rgb((brackets));", "margin:rgb((brackets))"},
		//{`background-color:transparent`, `background-color:#0000`}, // TODO for CSS3

		// bugs
		{"background: linear-gradient(-180deg, #355FFF 0%, #1F52FF 100%) 0% 0% / cover", "background:linear-gradient(-180deg,#355FFF 0%,#1F52FF 100%)0 0/cover"}, // #263
		{"font:1em -apple-system", "font:1em '-apple-system'"},     // support for IE9, IE10, IE11, fixes #251
		{"font:1em -", "font:1em '-'"},                             // support for IE9, IE10, IE11, fixes #251
		{"color:rgba(255,255,255,0)", "color:rgba(255,255,255,0)"}, // #327
		{"box-shadow:none", "box-shadow:none"},                     // #332

		// go-fuzz
		{"FONT-FAMILY: ru\"", "font-family:ru\""},
		{`d:hsl(0033333333333333333333333333333333333333333333333333333333333333333333333333333200,040000199823736,2444)`, `d:hsl(33333333333333333333333333333333333333333333333333333333333333333333333333333200,40000199823736,2444)`},
		{`d:hsl(0033333333333333333333333333333333333333333333333333333333333333333333333333333200%,040000199823736,2444)`, `d:hsl(33333333333333333333333333333333333333333333333333333333333333333333333333333200%,40000199823736,2444)`},
		{`d:hsl(-360000000000000000000000000000,50%,50%)`, `d:#bf4040`},
		{`d:hsl(360000000000000000000000000000,50%,50%)`, `d:#bf4040`},
		{`background:none,,,,,,,,,,,,,,,#00ff00`, `background:0 0,,,,,,,,,,,,,,,#0f0`},
		{`background:rgba(100%)`, `background:rgba(100%)`},
	}

	m := minify.New()
	params := map[string]string{"inline": "1"}
	for _, tt := range cssTests {
		t.Run(tt.css, func(t *testing.T) {
			r := bytes.NewBufferString(tt.css)
			w := &bytes.Buffer{}
			err := Minify(m, w, r, params)
			test.Minify(t, tt.css, err, w.String(), tt.expected)
		})
	}
}

func TestCSSKeepCSS2(t *testing.T) {
	tests := []struct {
		css      string
		expected string
	}{
		{`margin:5000`, `margin:5000`},
		{`margin:5000%`, `margin:5000%`},
		{`margin:5000em`, `margin:5000em`},
		{`color:transparent`, `color:transparent`},
	}

	m := minify.New()
	params := map[string]string{"inline": "1"}
	cssMinifier := &Minifier{KeepCSS2: true}
	for _, tt := range tests {
		t.Run(tt.css, func(t *testing.T) {
			r := bytes.NewBufferString(tt.css)
			w := &bytes.Buffer{}
			err := cssMinifier.Minify(m, w, r, params)
			test.Minify(t, tt.css, err, w.String(), tt.expected)
		})
	}
}

func TestReaderErrors(t *testing.T) {
	r := test.NewErrorReader(0)
	w := &bytes.Buffer{}
	m := minify.New()
	err := Minify(m, w, r, nil)
	test.T(t, err, test.ErrPlain, "return error at first read")
}

func TestWriterErrors(t *testing.T) {
	errorTests := []struct {
		css string
		n   []int
	}{
		{`@import 'file'`, []int{0, 2}},
		{`@media all{}`, []int{0, 2, 3, 4}},
		{`a[id^="L"]{margin:2in!important;color:red}`, []int{0, 4, 6, 7, 8, 9, 10, 11}},
		{`a{color:rgb(255,0,0)}`, []int{4}},
		{`a{color:rgb(255,255,255)}`, []int{4}},
		{`a{color:hsl(0,100%,50%)}`, []int{4}},
		{`a{color:hsl(360,100%,100%)}`, []int{4}},
		{`a{color:f(arg)}`, []int{4}},
		{`<!--`, []int{0}},
		{`/*!comment*/`, []int{0, 1, 2}},
		{`a{--var:val}`, []int{2, 3, 4}},
		{`a{*color:0}`, []int{2, 3}},
		{`a{color:0;baddecl 5}`, []int{5}},
		{`a[id="x" i],b{color:0}`, []int{5, 8}},
		{`a{color:()!important}`, []int{4, 6}},
		{`a{margin:5 4}`, []int{5}},
		{`a{margin=5}`, []int{2, 3}},
		{`a;`, []int{0}},
		{`a{000}`, []int{2}},
	}

	m := minify.New()
	for _, tt := range errorTests {
		for _, n := range tt.n {
			t.Run(fmt.Sprint(tt.css, " ", tt.n), func(t *testing.T) {
				r := bytes.NewBufferString(tt.css)
				w := test.NewErrorWriter(n)
				err := Minify(m, w, r, nil)
				test.T(t, err, test.ErrPlain)
			})
		}
	}
}

////////////////////////////////////////////////////////////////

func ExampleMinify() {
	m := minify.New()
	m.AddFunc("text/css", Minify)

	if err := m.Minify("text/css", os.Stdout, os.Stdin); err != nil {
		panic(err)
	}
}
