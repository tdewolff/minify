package css // import "github.com/tdewolff/minify/css"

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tdewolff/minify"
)

func assertCSS(t *testing.T, isStylesheet bool, input, expected string) {
	var params map[string]string
	if !isStylesheet {
		params = map[string]string{"inline": "1"}
	}

	m := minify.New()
	b := &bytes.Buffer{}
	assert.Nil(t, Minify(m, b, bytes.NewBufferString(input), "text/css", params), "Minify must not return error in "+input)
	assert.Equal(t, expected, b.String(), "Minify must give expected result in "+input)
}

////////////////////////////////////////////////////////////////

func TestCSS(t *testing.T) {
	assertCSS(t, false, "/*comment*/", "")
	assertCSS(t, false, ";", "")
	assertCSS(t, false, "key: value;", "key:value")
	assertCSS(t, false, "margin: 0 1; padding: 0 1;", "margin:0 1;padding:0 1")
	assertCSS(t, false, "color: #FF0000;", "color:red")
	assertCSS(t, false, "color: #000000;", "color:#000")
	assertCSS(t, false, "color: black;", "color:#000")
	assertCSS(t, false, "color: rgb(255,255,255);", "color:#fff")
	assertCSS(t, false, "color: rgb(100%,100%,100%);", "color:#fff")
	assertCSS(t, false, "color: rgba(255,0,0,1);", "color:red")
	assertCSS(t, false, "color: rgba(255,0,0,2);", "color:red")
	assertCSS(t, false, "color: rgba(255,0,0,0.5);", "color:rgba(255,0,0,.5)")
	assertCSS(t, false, "color: rgba(255,0,0,-1);", "color:transparent")
	assertCSS(t, false, "color: hsl(0,100%,50%);", "color:red")
	assertCSS(t, false, "color: hsla(1,2%,3%,1);", "color:#080807")
	assertCSS(t, false, "color: hsla(1,2%,3%,0);", "color:transparent")
	assertCSS(t, false, "font-weight: bold; font-weight: normal;", "font-weight:700;font-weight:400")
	assertCSS(t, false, "font: bold \"Times new Roman\",\"Sans-Serif\";", "font:700 times new roman,\"sans-serif\"")
	assertCSS(t, false, "outline: none;", "outline:0")
	assertCSS(t, false, "outline: none !important;", "outline:0!important")
	assertCSS(t, false, "border-left: none;", "border-left:0")
	assertCSS(t, false, "margin: 1 1 1 1;", "margin:1")
	assertCSS(t, false, "margin: 1 2 1 2;", "margin:1 2")
	assertCSS(t, false, "margin: 1 2 3 2;", "margin:1 2 3")
	assertCSS(t, false, "margin: 1 2 3 4;", "margin:1 2 3 4")
	assertCSS(t, false, "margin: 1 1 1 a;", "margin:1 1 1 a")
	assertCSS(t, false, "margin: 1 1 1 1 !important;", "margin:1!important")
	assertCSS(t, false, "padding:.2em .4em .2em", "padding:.2em .4em")
	assertCSS(t, false, "margin: 0em;", "margin:0")
	assertCSS(t, false, "font-family:'Arial', 'Times New Roman';", "font-family:arial,times new roman")
	assertCSS(t, false, "background:url('http://domain.com/image.png');", "background:url(http://domain.com/image.png)")
	assertCSS(t, false, "filter: progid : DXImageTransform.Microsoft.BasicImage(rotation=1);", "filter:progid:DXImageTransform.Microsoft.BasicImage(rotation=1)")
	assertCSS(t, false, "filter: progid:DXImageTransform.Microsoft.Alpha(Opacity=0);", "filter:alpha(opacity=0)")
	assertCSS(t, false, "content: \"a\\\nb\";", "content:\"ab\"")
	assertCSS(t, false, "content: \"a\\\r\nb\\\r\nc\";", "content:\"abc\"")
	assertCSS(t, false, "content: \"\";", "content:\"\"")
	assertCSS(t, true, "i { key: value; key2: value; }", "i{key:value;key2:value}")
	assertCSS(t, true, ".cla .ss > #id { x:y; }", ".cla .ss>#id{x:y}")
	assertCSS(t, true, ".cla[id ^= L] { x:y; }", ".cla[id^=L]{x:y}")
	assertCSS(t, true, "area:focus { outline : 0;}", "area:focus{outline:0}")
	assertCSS(t, true, "@import 'file';", "@import 'file'")
	assertCSS(t, true, "@font-face { x:y; }", "@font-face{x:y}")

	assertCSS(t, false, "font:27px/13px arial,sans-serif", "font:27px/13px arial,sans-serif")
	assertCSS(t, false, "text-decoration: none !important", "text-decoration:none!important")
	assertCSS(t, false, "color:#fff", "color:#fff")
	assertCSS(t, false, "border:2px rgb(255,255,255);", "border:2px #fff")
	assertCSS(t, false, "margin:-1px", "margin:-1px")
	assertCSS(t, false, "margin:+1px", "margin:1px")
	assertCSS(t, false, "margin:0.5em", "margin:.5em")
	assertCSS(t, false, "margin:-0.5em", "margin:-.5em")
	assertCSS(t, false, "margin:05em", "margin:5em")
	assertCSS(t, false, "margin:.50em", "margin:.5em")
	assertCSS(t, false, "margin:5.0em", "margin:5em")
	assertCSS(t, false, "color:#c0c0c0", "color:silver")
	assertCSS(t, false, "-ms-filter: \"progid:DXImageTransform.Microsoft.Alpha(Opacity=80)\";", "-ms-filter:\"alpha(opacity=80)\"")
	assertCSS(t, false, "filter: progid:DXImageTransform.Microsoft.Alpha(Opacity = 80);", "filter:alpha(opacity=80)")
	assertCSS(t, false, "MARGIN:1EM", "margin:1em")
	assertCSS(t, false, "color:CYAN", "color:cyan")
	assertCSS(t, false, "background:URL(x.PNG);", "background:url(x.PNG)")
	assertCSS(t, false, "background:url(/*nocomment*/)", "background:url(/*nocomment*/)")
	assertCSS(t, false, "background:url(data:,text)", "background:url(data:,text)")
	assertCSS(t, false, "background:url('data:text/xml; version = 2.0,content')", "background:url(data:text/xml;version=2.0,content)")
	assertCSS(t, false, "background:url('data:\\'\",text')", "background:url('data:\\'\",text')")
	assertCSS(t, false, "margin:0 0 18px 0;", "margin:0 0 18px")
	assertCSS(t, true, "input[type=\"radio\"]{x:y}", "input[type=radio]{x:y}")
	assertCSS(t, true, "DIV{margin:1em}", "div{margin:1em}")
	assertCSS(t, true, ".CLASS{margin:1em}", ".CLASS{margin:1em}")
	assertCSS(t, true, "@MEDIA all{}", "@media all{}")
	assertCSS(t, true, "@media only screen and (max-width : 800px){}", "@media only screen and (max-width:800px){}")
	assertCSS(t, true, "@media (-webkit-min-device-pixel-ratio:1.5),(min-resolution:1.5dppx){}", "@media(-webkit-min-device-pixel-ratio:1.5),(min-resolution:1.5dppx){}")
	assertCSS(t, true, "[class^=icon-] i[class^=icon-],i[class*=\" icon-\"]{x:y}", "[class^=icon-] i[class^=icon-],i[class*=\" icon-\"]{x:y}")
	assertCSS(t, true, "html{line-height:1;}html{line-height:1;}", "html{line-height:1}html{line-height:1}")
	assertCSS(t, true, ".clearfix { *zoom: 1; }", ".clearfix{*zoom:1}")
	assertCSS(t, true, "a { b: 1", "a{b:1}")
	assertCSS(t, false, "background:none", "background:0 0")
	assertCSS(t, false, "background:none 1 1", "background:none 1 1")
	assertCSS(t, false, "z-index:1000", "z-index:1000")

	// coverage
	assertCSS(t, false, "margin: 1 1;", "margin:1")
	assertCSS(t, false, "margin: 1 2;", "margin:1 2")
	assertCSS(t, false, "margin: 1 1 1;", "margin:1")
	assertCSS(t, false, "margin: 1 2 1;", "margin:1 2")
	assertCSS(t, false, "margin: 1 2 3;", "margin:1 2 3")
	assertCSS(t, false, "margin: 0%;", "margin:0")
	assertCSS(t, false, "color: rgb(255,64,64);", "color:#ff4040")
	assertCSS(t, false, "color: rgb(256,-34,2342435);", "color:#f0f")
	assertCSS(t, false, "color: rgb(120%,-45%,234234234%);", "color:#f0f")
	assertCSS(t, false, "color: rgb(0, 1, ident);", "color:rgb(0,1,ident)")
	assertCSS(t, false, "color: rgb(ident);", "color:rgb(ident)")
	assertCSS(t, false, "margin: rgb(ident);", "margin:rgb(ident)")
	assertCSS(t, false, "filter: progid:b().c.Alpha(rgba(x));", "filter:progid:b().c.Alpha(rgba(x))")
	assertCSS(t, true, "a, b + c { x:y; }", "a,b+c{x:y}")

	// go-fuzz
	assertCSS(t, false, "FONT-FAMILY: ru\"", "font-family:ru\"")
	assertCSS(t, true, "input[type=\"\x00\"] {  a: b\n}.a{}", "input[type=\"\x00\"] {  a: b\n}.a{}")
	assertCSS(t, true, "a{a:)'''", "a{a:)'''}")
}

////////////////////////////////////////////////////////////////

func ExampleMinify() {
	m := minify.New()
	m.AddFunc("text/css", Minify)

	if err := m.Minify(os.Stdout, os.Stdin, "text/css", nil); err != nil {
		fmt.Println("minify.Minify:", err)
	}
}
