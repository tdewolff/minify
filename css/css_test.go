package css // import "github.com/tdewolff/minify/css"

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/test"
)

func TestCSS(t *testing.T) {
	cssTests := []struct {
		css      string
		expected string
	}{
		{"/*comment*/", ""},
		{"/*! bang  comment */", "/*!bang comment*/"},
		{"i{}/*! bang  comment */", "i{}/*!bang comment*/"},
		{"i { key: value; key2: value; }", "i{key:value;key2:value}"},
		{".cla .ss > #id { x:y; }", ".cla .ss>#id{x:y}"},
		{".cla[id ^= L] { x:y; }", ".cla[id^=L]{x:y}"},
		{"area:focus { outline : 0;}", "area:focus{outline:0}"},
		{"@import 'file';", "@import 'file'"},
		{"@import url('file');", "@import 'file'"},
		{"@import url(//url);", `@import "//url"`},
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

		{":root { --custom-variable:0px; }", ":root{--custom-variable:0px}"},

		// case sensitivity
		{"@counter-style Ident{}", "@counter-style Ident{}"},

		// coverage
		{"a, b + c { x:y; }", "a,b+c{x:y}"},

		// bad declaration
		{".clearfix { *zoom: 1px; }", ".clearfix{*zoom:1px}"},
		{".clearfix { *zoom: 1px }", ".clearfix{*zoom:1px}"},
		{".clearfix { color:green; *zoom: 1px; color:red; }", ".clearfix{color:green;*zoom:1px;color:red}"},

		// go-fuzz
		{"input[type=\"\x00\"] {  a: b\n}.a{}", "input[type=\"\x00\"]{a:b}.a{}"},
		{"a{a:)'''", "a{a:)'''}"},
		{"{T:l(", "{t:l(}"},
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
}

func TestCSSInline(t *testing.T) {
	cssTests := []struct {
		css      string
		expected string
	}{
		{"/*comment*/", ""},
		{"/*! bang  comment */", ""},
		{";", ""},
		{"empty:", "empty:"},
		{"key: value;", "key:value"},
		{"margin: 0 1; padding: 0 1;", "margin:0 1;padding:0 1"},
		{"color: #FF0000;", "color:red"},
		{"color: #000000;", "color:#000"},
		{"color: black;", "color:#000"},
		{"color: rgb(255,255,255);", "color:#fff"},
		{"color: rgb(100%,100%,100%);", "color:#fff"},
		{"color: rgba(255,0,0,1);", "color:red"},
		{"color: rgba(255,0,0,2);", "color:red"},
		{"color: rgba(255,0,0,0.5);", "color:#ff000080"},
		{"color: rgba(255,0,0,-1);", "color:#0000"},
		{"color: rgba(0%,15%,25%,0.2);", "color:#00264033"},
		{"color: rgba(0,0,0,0.5);", "color:#00000080"},
		{"color: rgb(255 0 0 / 1);", "color:red"},
		{"color: hsla(5,0%,10%,0.75);", "color:#1a1a1abf"},
		{"color: hsl(0,100%,50%);", "color:red"},
		{"color: hsla(1,2%,3%,1);", "color:#080807"},
		{"color: hsla(1,2%,3%,0);", "color:#0000"},
		{"color: hsl(48,100%,50%);", "color:#fc0"},
		{"color: hsl(0 100% 50% / 1);", "color:red"},
		{"color: hsl(400, 150%, 150%, 2);", "color:#fff"},
		{"background-position:center", "background-position:50%"},
		{"background-position:left 50%", "background-position:0"},
		{"background-position:center center", "background-position:50%"},
		{"background-position:center bottom", "background-position:50% 100%"},
		{"background-position:bottom 5% right 0%", "background-position:bottom 5% right"},
		{"background-position:bottom 0 right 10%", "background-position:bottom right 10%"},
		{"background-position:center right 10%", "background-position:center right 10%"},
		{"background-repeat:space space", "background-repeat:space"},
		{"background-repeat:round round", "background-repeat:round"},
		{"background-repeat:repeat repeat", "background-repeat:repeat"},
		{"background-repeat:no-repeat no-repeat", "background-repeat:no-repeat"},
		{"background-repeat:repeat no-repeat", "background-repeat:repeat-x"},
		{"background-repeat:no-repeat repeat", "background-repeat:repeat-y"},
		{"background-size:auto auto", "background-size:auto"},
		{"background-size:30% auto", "background-size:30%"},
		{"background: hsla(0,0%,100%,.7);", "background:#ffffffb2"},
		{"background:red none", "background:red"},
		{"background:red none 0 0", "background:red"},
		{"background:red none 1 1", "background:red 1 1"},
		{"background:#0000 1 1", "background:1 1"},
		{"background:transparent", "background:0 0"},
		{"background:transparent no-repeat", "background:no-repeat"},
		{"background:#0000 none padding-box 0 0 / auto auto scroll border-box repeat repeat", "background:0 0"},
		{"background:0 0 / auto", "background:0 0"},
		{"background:0 0 / auto 10%", "background:0 0/auto 10%"},
		{"background:0% 0%", "background:0 0"},
		{"background:left top", "background:0 0"},
		{"background:no-repeat repeat", "background:repeat-y"},
		{"font-weight: bold; font-weight: normal;", "font-weight:700;font-weight:400"},
		{"font: bold 5px \"Times new Roman\",\"Sans-Serif\";", "font:700 5px times new roman,sans-serif"},
		{"font: normal normal normal normal 20px normal", "font:20px normal"},
		{"font:27px/13px arial,sans-serif", "font:27px/13px arial,sans-serif"},
		{"font:normal normal bold normal medium/normal arial,sans-serif", "font:700 medium arial,sans-serif"},
		{"font:400 medium/normal 'Arial'", "font:medium arial"},
		{"font:medium/normal 'Arial'", "font:medium arial"},
		{"outline: none;", "outline:none"},
		{"outline: solid black 0;", "outline:solid #000 0"},
		{"outline: none black medium;", "outline:#000"},
		{"outline: none !important;", "outline:none!important"},
		{"border-left: none;", "border-left:none"},
		{"border-left: none 0;", "border-left:0"},
		{"border-left: none medium currentcolor;", "border-left:none"},
		{"border-left: 0 dashed red;", "border-left:0 dashed red"},
		{"margin: 1 1 1 1;", "margin:1"},
		{"margin: 1 2 1 2;", "margin:1 2"},
		{"margin: 1 2 3 2;", "margin:1 2 3"},
		{"margin: 1 2 3 4;", "margin:1 2 3 4"},
		{"margin: 1 1 1 a;", "margin:1 1 1 a"},
		{"margin: 1 1 1 1 !important;", "margin:1!important"},
		{"padding:.2em .4em .2em", "padding:.2em .4em"},
		{"margin: 0em;", "margin:0"},
		{"font-family:'Arial', 'Times New Roman';", "font-family:arial,times new roman"},
		{"filter: progid : DXImageTransform.Microsoft.BasicImage(rotation=1);", "filter:progid:DXImageTransform.Microsoft.BasicImage(rotation=1)"},
		{"filter: progid:DXImageTransform.Microsoft.Alpha(Opacity=0);", "filter:alpha(opacity=0)"},
		{"content: \"a\\\nb\";", "content:\"ab\""},
		{"content: \"a\\\r\nb\\\r\nc\";", "content:\"abc\""},
		{"content: \"\";", "content:\"\""},
		{"x: white , white", "x:#fff,#fff"},

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
		{"margin:5000em", "margin:5e3em"},
		{"color:#c0c0c0", "color:silver"},
		{"-ms-filter: \"progid:DXImageTransform.Microsoft.Alpha(Opacity=80)\";", "-ms-filter:\"alpha(opacity=80)\""},
		{"filter: progid:DXImageTransform.Microsoft.Alpha(Opacity = 80);", "filter:alpha(opacity=80)"},
		{"MARGIN:1EM", "margin:1em"},
		//{"color:CYAN", "color:cyan"}, // TODO
		{"width:attr(Name em)", "width:attr(Name em)"},
		{"content:CounterName", "content:CounterName"},
		{"background:url('http://domain.com/image.png');", "background:url(http://domain.com/image.png)"},
		{"background:url( 'http://domain.com/image.png' );", "background:url(http://domain.com/image.png)"},
		{"background:URL(x.PNG);", "background:url(x.PNG)"},
		{"background:url(/*nocomment*/)", "background:url(/*nocomment*/)"},
		{"background:url(data:,text)", "background:url(data:,text)"},
		{"background:url('data:text/xml; version = 2.0,content')", "background:url(data:text/xml;version=2.0,content)"},
		{"background:url('data:\\'\",text')", "background:url('data:\\'\",text')"},
		{"margin:0 0 18px 0;", "margin:0 0 18px"},
		{"z-index:1000", "z-index:1000"},
		{"box-shadow:0 0 0 0", "box-shadow:0 0"},
		{"flex:0px", "flex:0px"},
		{"g:url('abc\\\ndef')", "g:url(abcdef)"},
		{"url:local('abc\\\ndef')", "url:local(abcdef)"},

		{"any:0deg 0s 0ms 0dpi 0dpcm 0dppx 0hz 0khz", "any:0 0s 0ms 0dpi 0dpcm 0dppx 0hz 0khz"},
		{"width:calc(0%-0px)", "width:calc(0%-0px)"},
		{"margin:calc(10px) calc(20px)", "margin:calc(10px) calc(20px)"},
		{"border-left:0 none", "border-left:0"},
		{"--custom-variable:0px;", "--custom-variable:0px"},
		{"--foo: if(x > 5) this.width = 10", "--foo: if(x > 5) this.width = 10"},
		{"--foo: ;", "--foo: "},

		// case sensitivity
		{"animation:Ident", "animation:Ident"},
		{"animation-name:Ident", "animation-name:Ident"},

		// coverage
		{"margin: 1 1;", "margin:1"},
		{"margin: 1 2;", "margin:1 2"},
		{"margin: 1 1 1;", "margin:1"},
		{"margin: 1 2 1;", "margin:1 2"},
		{"margin: 1 2 3;", "margin:1 2 3"},
		// {"margin: 0%;", "margin:0"},
		{"color: rgb(255,64,64);", "color:#ff4040"},
		{"color: rgb(256,-34,2342435);", "color:#f0f"},
		{"color: rgb(120%,-45%,234234234%);", "color:#f0f"},
		{"color: rgb(0, 1, ident);", "color:rgb(0,1,ident)"},
		{"color: rgb(ident);", "color:rgb(ident)"},
		{"margin: rgb(ident);", "margin:rgb(ident)"},
		{"filter: progid:b().c.Alpha(rgba(x));", "filter:progid:b().c.Alpha(rgba(x))"},

		// go-fuzz
		{"FONT-FAMILY: ru\"", "font-family:ru\""},
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
		{`margin:5000em`, `margin:5000em`},
		{"color: rgba(0%,15%,25%,0.2);", "color:rgba(0%,15%,25%,.2)"},
		{"color: hsla(5,0%,10%,0.75);", "color:hsla(5,0%,10%,.75)"},
		{"color: rgba(0%,15%,25%,0);", "color:rgba(0%,15%,25%,0)"},
		{"color: hsla(5,0%,10%,0);", "color:hsla(5,0%,10%,0)"},
		{"color: rgba(0%,15%,25%,1);", "color:#002640"},
		{"color: hsla(5,0%,10%,1);", "color:#1a1a1a"},
		{"color: transparent;", "color:transparent"},
	}

	m := minify.New()
	params := map[string]string{"inline": "1"}
	cssMinifier := &Minifier{Decimals: -1, KeepCSS2: true}
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
