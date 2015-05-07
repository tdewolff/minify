package css // import "github.com/tdewolff/minify/css"

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tdewolff/minify"
)

func assertCSS(t *testing.T, m *minify.Minify, isStylesheet bool, input, expected string) {
	mediatype := "text/css"
	if !isStylesheet {
		if len(input)%2 == 0 {
			mediatype = "text/css;inline=1"
		} else {
			mediatype = "text/css; inline=1"
		}
	}

	b := &bytes.Buffer{}
	assert.Nil(t, Minify(m, mediatype, b, bytes.NewBufferString(input)), "Minify must not return error in "+input)
	assert.Equal(t, expected, b.String(), "Minify must give expected result in "+input)
}

////////////////////////////////////////////////////////////////

func TestCSS(t *testing.T) {
	m := minify.New()
	m.AddFunc("text/css", Minify)

	assertCSS(t, m, false, "/*comment*/", "")
	assertCSS(t, m, false, ";", "")
	assertCSS(t, m, false, "key: value;", "key:value")
	assertCSS(t, m, false, "margin: 0 1; padding: 0 1;", "margin:0 1;padding:0 1")
	assertCSS(t, m, false, "color: #FF0000;", "color:red")
	assertCSS(t, m, false, "color: #000000;", "color:#000")
	assertCSS(t, m, false, "color: black;", "color:#000")
	assertCSS(t, m, false, "color: rgb(255,255,255);", "color:#fff")
	assertCSS(t, m, false, "color: rgb(100%,100%,100%);", "color:#fff")
	assertCSS(t, m, false, "color: rgba(255,0,0,1);", "color:red")
	assertCSS(t, m, false, "font-weight: bold; font-weight: normal;", "font-weight:700;font-weight:400")
	assertCSS(t, m, false, "font: bold \"Times new Roman\",\"Sans-Serif\";", "font:700 times new roman,\"sans-serif\"")
	assertCSS(t, m, false, "outline: none;", "outline:0")
	assertCSS(t, m, false, "outline: none !important;", "outline:0!important")
	assertCSS(t, m, false, "border-left: none;", "border-left:0")
	assertCSS(t, m, false, "margin: 1 1 1 1;", "margin:1")
	assertCSS(t, m, false, "margin: 1 2 1 2;", "margin:1 2")
	assertCSS(t, m, false, "margin: 1 2 3 2;", "margin:1 2 3")
	assertCSS(t, m, false, "margin: 1 2 3 4;", "margin:1 2 3 4")
	assertCSS(t, m, false, "margin: 1 1 1 a;", "margin:1 1 1 a")
	assertCSS(t, m, false, "margin: 1 1 1 1 !important;", "margin:1!important")
	assertCSS(t, m, false, "padding:.2em .4em .2em", "padding:.2em .4em")
	assertCSS(t, m, false, "margin: 0em;", "margin:0")
	assertCSS(t, m, false, "font-family:'Arial', 'Times New Roman';", "font-family:arial,times new roman")
	assertCSS(t, m, false, "background:url('http://domain.com/image.png');", "background:url(http://domain.com/image.png)")
	assertCSS(t, m, false, "filter: progid : DXImageTransform.Microsoft.BasicImage(rotation=1);", "filter:progid:DXImageTransform.Microsoft.BasicImage(rotation=1)")
	assertCSS(t, m, false, "filter: progid:DXImageTransform.Microsoft.Alpha(Opacity=0);", "filter:alpha(opacity=0)")
	assertCSS(t, m, false, "content: \"a\\\nb\";", "content:\"ab\"")
	assertCSS(t, m, false, "content: \"a\\\r\nb\\\r\nc\";", "content:\"abc\"")
	assertCSS(t, m, true, "i { key: value; key2: value; }", "i{key:value;key2:value}")
	assertCSS(t, m, true, ".cla .ss > #id { x:y; }", ".cla .ss>#id{x:y}")
	assertCSS(t, m, true, ".cla[id ^= L] { x:y; }", ".cla[id^=L]{x:y}")
	assertCSS(t, m, true, "area:focus { outline : 0;}", "area:focus{outline:0}")
	assertCSS(t, m, true, "@import 'file';", "@import 'file'")
	assertCSS(t, m, true, "@font-face { x:y; }", "@font-face{x:y}")

	assertCSS(t, m, false, "font:27px/13px arial,sans-serif", "font:27px/13px arial,sans-serif")
	assertCSS(t, m, false, "text-decoration: none !important", "text-decoration:none!important")
	assertCSS(t, m, false, "color:#fff", "color:#fff")
	assertCSS(t, m, false, "border:2px rgb(255,255,255);", "border:2px #fff")
	assertCSS(t, m, false, "margin:-1px", "margin:-1px")
	assertCSS(t, m, false, "margin:+1px", "margin:1px")
	assertCSS(t, m, false, "margin:0.5em", "margin:.5em")
	assertCSS(t, m, false, "margin:-0.5em", "margin:-.5em")
	assertCSS(t, m, false, "margin:05em", "margin:5em")
	assertCSS(t, m, false, "margin:.50em", "margin:.5em")
	assertCSS(t, m, false, "margin:5.0em", "margin:5em")
	assertCSS(t, m, false, "color:#c0c0c0", "color:silver")
	assertCSS(t, m, false, "-ms-filter: \"progid:DXImageTransform.Microsoft.Alpha(Opacity=80)\";", "-ms-filter:\"alpha(opacity=80)\"")
	assertCSS(t, m, false, "filter: progid:DXImageTransform.Microsoft.Alpha(Opacity = 80);", "filter:alpha(opacity=80)")
	assertCSS(t, m, false, "MARGIN:1EM", "margin:1em")
	assertCSS(t, m, false, "color:CYAN", "color:cyan")
	assertCSS(t, m, false, "background:URL(x.PNG);", "background:url(x.PNG)")
	assertCSS(t, m, false, "background:url(/*nocomment*/)", "background:url(/*nocomment*/)")
	assertCSS(t, m, false, "background:url(data:,text)", "background:url(data:,text)")
	assertCSS(t, m, false, "background:url('data:text/xml; version = 2.0,content')", "background:url(data:text/xml;version=2.0,content)")
	assertCSS(t, m, false, "background:url('data:\\'\",text')", "background:url('data:\\'\",text')")
	assertCSS(t, m, false, "margin:0 0 18px 0;", "margin:0 0 18px")
	assertCSS(t, m, true, "input[type=\"radio\"]{x:y}", "input[type=radio]{x:y}")
	assertCSS(t, m, true, "DIV{margin:1em}", "div{margin:1em}")
	assertCSS(t, m, true, ".CLASS{margin:1em}", ".CLASS{margin:1em}")
	assertCSS(t, m, true, "@MEDIA all{}", "@media all{}")
	assertCSS(t, m, true, "@media only screen and (max-width : 800px){}", "@media only screen and (max-width:800px){}")
	assertCSS(t, m, true, "@media (-webkit-min-device-pixel-ratio:1.5),(min-resolution:1.5dppx){}", "@media(-webkit-min-device-pixel-ratio:1.5),(min-resolution:1.5dppx){}")
	assertCSS(t, m, true, "[class^=icon-] i[class^=icon-],i[class*=\" icon-\"]{x:y}", "[class^=icon-] i[class^=icon-],i[class*=\" icon-\"]{x:y}")
	assertCSS(t, m, true, "html{line-height:1;}html{line-height:1;}", "html{line-height:1}html{line-height:1}")
	assertCSS(t, m, true, ".clearfix { *zoom: 1; }", ".clearfix{*zoom:1}")
	assertCSS(t, m, true, "a { b: 1", "a{b:1}")

	// coverage
	assertCSS(t, m, false, "margin: 1 1;", "margin:1")
	assertCSS(t, m, false, "margin: 1 2;", "margin:1 2")
	assertCSS(t, m, false, "margin: 1 1 1;", "margin:1")
	assertCSS(t, m, false, "margin: 1 2 1;", "margin:1 2")
	assertCSS(t, m, false, "margin: 1 2 3;", "margin:1 2 3")
	assertCSS(t, m, false, "margin: 0%;", "margin:0")
	assertCSS(t, m, false, "color: rgb(255,64,64);", "color:#ff4040")
	assertCSS(t, m, false, "color: rgb(256,-34,2342435);", "color:#f0f")
	assertCSS(t, m, false, "color: rgb(120%,-45%,234234234%);", "color:#f0f")
	assertCSS(t, m, false, "color: rgb(0, 1, ident);", "color:rgb(0,1,ident)")
	assertCSS(t, m, false, "color: rgb(ident);", "color:rgb(ident)")
	assertCSS(t, m, false, "margin: rgb(ident);", "margin:rgb(ident)")
	assertCSS(t, m, false, "filter: progid:b().c.Alpha(rgba(x));", "filter:progid:b().c.Alpha(rgba(x))")
	assertCSS(t, m, true, "a, b + c { x:y; }", "a,b+c{x:y}")
}
