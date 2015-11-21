package css // import "github.com/tdewolff/minify/css"

import (
	"bytes"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/test"
)

func TestCSS(t *testing.T) {
	var cssTests = []struct {
		css      string
		expected string
	}{
		{"i { key: value; key2: value; }", "i{key:value;key2:value}"},
		{".cla .ss > #id { x:y; }", ".cla .ss>#id{x:y}"},
		{".cla[id ^= L] { x:y; }", ".cla[id^=L]{x:y}"},
		{"area:focus { outline : 0;}", "area:focus{outline:0}"},
		{"@import 'file';", "@import 'file'"},
		{"@font-face { x:y; }", "@font-face{x:y}"},

		{"input[type=\"radio\"]{x:y}", "input[type=radio]{x:y}"},
		{"DIV{margin:1em}", "div{margin:1em}"},
		{".CLASS{margin:1em}", ".CLASS{margin:1em}"},
		{"@MEDIA all{}", "@media all{}"},
		{"@media only screen and (max-width : 800px){}", "@media only screen and (max-width:800px){}"},
		{"@media (-webkit-min-device-pixel-ratio:1.5),(min-resolution:1.5dppx){}", "@media(-webkit-min-device-pixel-ratio:1.5),(min-resolution:1.5dppx){}"},
		{"[class^=icon-] i[class^=icon-],i[class*=\" icon-\"]{x:y}", "[class^=icon-] i[class^=icon-],i[class*=\" icon-\"]{x:y}"},
		{"html{line-height:1;}html{line-height:1;}", "html{line-height:1}html{line-height:1}"},
		{".clearfix { *zoom: 1; }", ".clearfix{*zoom:1}"},
		{"a { b: 1", "a{b:1}"},

		// coverage
		{"a, b + c { x:y; }", "a,b+c{x:y}"},

		// go-fuzz
		{"input[type=\"\x00\"] {  a: b\n}.a{}", "input[type=\"\x00\"]{a:b}.a{}"},
		{"a{a:)'''", "a{a:)'''}"},
	}

	m := minify.New()
	for _, tt := range cssTests {
		b := &bytes.Buffer{}
		assert.Nil(t, Minify(m, b, bytes.NewBufferString(tt.css), nil), "Minify must not return error in "+tt.css)
		assert.Equal(t, tt.expected, b.String(), "Minify must give expected result in "+tt.css)
	}
}

func TestCSSInline(t *testing.T) {
	var cssTests = []struct {
		css      string
		expected string
	}{
		{"/*comment*/", ""},
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
		{"color: rgba(255,0,0,0.5);", "color:rgba(255,0,0,.5)"},
		{"color: rgba(255,0,0,-1);", "color:transparent"},
		{"color: hsl(0,100%,50%);", "color:red"},
		{"color: hsla(1,2%,3%,1);", "color:#080807"},
		{"color: hsla(1,2%,3%,0);", "color:transparent"},
		{"color: hsl(48,100%,50%);", "color:#fc0"},
		{"font-weight: bold; font-weight: normal;", "font-weight:700;font-weight:400"},
		{"font: bold \"Times new Roman\",\"Sans-Serif\";", "font:700 times new roman,\"sans-serif\""},
		{"outline: none;", "outline:0"},
		{"outline: none !important;", "outline:0!important"},
		{"border-left: none;", "border-left:0"},
		{"margin: 1 1 1 1;", "margin:1"},
		{"margin: 1 2 1 2;", "margin:1 2"},
		{"margin: 1 2 3 2;", "margin:1 2 3"},
		{"margin: 1 2 3 4;", "margin:1 2 3 4"},
		{"margin: 1 1 1 a;", "margin:1 1 1 a"},
		{"margin: 1 1 1 1 !important;", "margin:1!important"},
		{"padding:.2em .4em .2em", "padding:.2em .4em"},
		{"margin: 0em;", "margin:0"},
		{"font-family:'Arial', 'Times New Roman';", "font-family:arial,times new roman"},
		{"background:url('http://domain.com/image.png');", "background:url(http://domain.com/image.png)"},
		{"filter: progid : DXImageTransform.Microsoft.BasicImage(rotation=1);", "filter:progid:DXImageTransform.Microsoft.BasicImage(rotation=1)"},
		{"filter: progid:DXImageTransform.Microsoft.Alpha(Opacity=0);", "filter:alpha(opacity=0)"},
		{"content: \"a\\\nb\";", "content:\"ab\""},
		{"content: \"a\\\r\nb\\\r\nc\";", "content:\"abc\""},
		{"content: \"\";", "content:\"\""},

		{"font:27px/13px arial,sans-serif", "font:27px/13px arial,sans-serif"},
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
		{"color:#c0c0c0", "color:silver"},
		{"-ms-filter: \"progid:DXImageTransform.Microsoft.Alpha(Opacity=80)\";", "-ms-filter:\"alpha(opacity=80)\""},
		{"filter: progid:DXImageTransform.Microsoft.Alpha(Opacity = 80);", "filter:alpha(opacity=80)"},
		{"MARGIN:1EM", "margin:1em"},
		{"color:CYAN", "color:cyan"},
		{"background:URL(x.PNG);", "background:url(x.PNG)"},
		{"background:url(/*nocomment*/)", "background:url(/*nocomment*/)"},
		{"background:url(data:,text)", "background:url(data:,text)"},
		{"background:url('data:text/xml; version = 2.0,content')", "background:url(data:text/xml;version=2.0,content)"},
		{"background:url('data:\\'\",text')", "background:url('data:\\'\",text')"},
		{"margin:0 0 18px 0;", "margin:0 0 18px"},
		{"background:none", "background:0 0"},
		{"background:none 1 1", "background:none 1 1"},
		{"z-index:1000", "z-index:1000"},

		// coverage
		{"margin: 1 1;", "margin:1"},
		{"margin: 1 2;", "margin:1 2"},
		{"margin: 1 1 1;", "margin:1"},
		{"margin: 1 2 1;", "margin:1 2"},
		{"margin: 1 2 3;", "margin:1 2 3"},
		{"margin: 0%;", "margin:0"},
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
	for _, tt := range cssTests {
		r := bytes.NewBufferString(tt.css)
		w := &bytes.Buffer{}
		assert.Nil(t, Minify(m, w, r, map[string]string{"inline": "1"}), "Minify must not return error in "+tt.css)
		assert.Equal(t, tt.expected, w.String(), "Minify must give expected result in "+tt.css)
	}
}

func TestReaderErrors(t *testing.T) {
	m := minify.New()
	r := test.NewErrorReader(0)
	w := &bytes.Buffer{}
	assert.Equal(t, test.ErrPlain, Minify(m, w, r, nil), "Minify must return error at first read")
}

func TestWriterErrors(t *testing.T) {
	var errorTests = []struct {
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
	}

	m := minify.New()
	for _, tt := range errorTests {
		for _, n := range tt.n {
			r := bytes.NewBufferString(tt.css)
			w := test.NewErrorWriter(n)
			assert.Equal(t, test.ErrPlain, Minify(m, w, r, nil), "Minify must return error in "+tt.css+" at write "+strconv.FormatInt(int64(n), 10))
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
