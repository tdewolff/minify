package css // import "github.com/tdewolff/minify/css"

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tdewolff/minify"
)

func assertCSS2(t *testing.T, input, expected string) {
	m := minify.NewMinifier()
	b := &bytes.Buffer{}
	assert.Nil(t, Minify2(m, b, bytes.NewBufferString(input)), "Minify must not return error in "+input)
	assert.Equal(t, expected, b.String(), "Minify must give expected result in "+input)
}

func TestCSS2(t *testing.T) {
	assertCSS2(t, "/*comment*/", "")
	assertCSS2(t, "css{}", "")
	assertCSS2(t, "key: value;", "key:value")
	assertCSS2(t, "margin: 0 1; padding: 0 1;", "margin:0 1;padding:0 1")
	assertCSS2(t, "i { key: value; key2: value; }", "i{key:value;key2:value}")
	assertCSS2(t, "color: #FF0000;", "color:red")
	assertCSS2(t, "color: #000000;", "color:#000")
	assertCSS2(t, "color: black;", "color:#000")
	assertCSS2(t, "color: rgb(255,255,255);", "color:#fff")
	assertCSS2(t, "color: rgb(100%,100%,100%);", "color:#fff")
	assertCSS2(t, "color: rgba(255,0,0,1);", "color:red")
	assertCSS2(t, "font-weight: bold; font-weight: normal;", "font-weight:700;font-weight:400")
	assertCSS2(t, "font: bold \"Times new Roman\",\"Sans-Serif\";", "font:700 times new roman,\"sans-serif\"")
	assertCSS2(t, "outline: none;", "outline:0")
	assertCSS2(t, "border-left: none;", "border-left:0")
	assertCSS2(t, "margin: 1 1 1 1;", "margin:1")
	assertCSS2(t, "margin: 1 2 1 2;", "margin:1 2")
	assertCSS2(t, "margin: 1 2 3 2;", "margin:1 2 3")
	assertCSS2(t, "margin: 1 2 3 4;", "margin:1 2 3 4")
	assertCSS2(t, "margin: 0em;", "margin:0")
	assertCSS2(t, ".cla .ss > #id { x:y; }", ".cla .ss>#id{x:y}")
	assertCSS2(t, ".cla[id ^= L] { x:y; }", ".cla[id^=L]{x:y}")
	assertCSS2(t, "area:focus { outline : 0;}", "area:focus{outline:0}")
	assertCSS2(t, "@import 'file';", "@import 'file'")
	assertCSS2(t, "@import 'file' { x:y; };", "@import 'file'{x:y}")
	assertCSS2(t, "<!-- x:y; -->", "<!--x:y-->")
	assertCSS2(t, "font-family:'Arial', 'Times New Roman';", "font-family:arial,times new roman")
	assertCSS2(t, "background:url('http://domain.com/image.png');", "background:url(http://domain.com/image.png)")
	assertCSS2(t, "filter: progid : DXImageTransform.Microsoft.BasicImage(rotation=1);", "filter:progid:DXImageTransform.Microsoft.BasicImage(rotation=1)")
	assertCSS2(t, "filter: progid:DXImageTransform.Microsoft.Alpha(Opacity=0);", "filter:alpha(opacity=0)")
	assertCSS2(t, "content: \"a\\\nb\";", "content:\"ab\"")
	// assertCSS2(t, "color:#fff;@charset x;", "color:#fff")
	// assertCSS2(t, "color:#fff;@import x;", "color:#fff")
	// assertCSS2(t, "@charset x;@import x;", "@charset x;@import x")
	// assertCSS2(t, "@charset;@import;", "")

	assertCSS2(t, "font:27px/13px arial,sans-serif", "font:27px/13px arial,sans-serif")
	assertCSS2(t, "text-decoration: none !important", "text-decoration:none!important")
	assertCSS2(t, "color:#fff", "color:#fff")
	assertCSS2(t, "border:2px rgb(255,255,255);", "border:2px #fff")
	assertCSS2(t, "margin:-1px", "margin:-1px")
	assertCSS2(t, "margin:+1px", "margin:1px")
	assertCSS2(t, "margin:0.5em", "margin:.5em")
	assertCSS2(t, "margin:-0.5em", "margin:-.5em")
	assertCSS2(t, "margin:05em", "margin:5em")
	assertCSS2(t, "margin:.50em", "margin:.5em")
	assertCSS2(t, "margin:5.0em", "margin:5em")
	assertCSS2(t, "color:#c0c0c0", "color:silver")
	assertCSS2(t, "input[type=\"radio\"]{x:y}", "input[type=radio]{x:y}")
	assertCSS2(t, "-ms-filter: \"progid:DXImageTransform.Microsoft.Alpha(Opacity=80)\";", "-ms-filter:\"alpha(opacity=80)\"")
	assertCSS2(t, "filter: progid:DXImageTransform.Microsoft.Alpha(Opacity=80);", "filter:alpha(opacity=80)")
	assertCSS2(t, "MARGIN:1EM", "margin:1em")
	assertCSS2(t, "color:CYAN", "color:cyan")
	assertCSS2(t, "background:URL(x.PNG);", "background:url(x.PNG)")
	assertCSS2(t, "DIV{margin:1em}", "div{margin:1em}")
	assertCSS2(t, ".CLASS{margin:1em}", ".CLASS{margin:1em}")
	assertCSS2(t, "@media only screen and (max-width:800px)", "@media only screen and (max-width:800px)")
	assertCSS2(t, "@media (-webkit-min-device-pixel-ratio:1.5),(min-resolution:1.5dppx)", "@media (-webkit-min-device-pixel-ratio:1.5),(min-resolution:1.5dppx)")
	assertCSS2(t, "[class^=icon-] i[class^=icon-],i[class*=\" icon-\"]{x:y}", "[class^=icon-] i[class^=icon-],i[class*=\" icon-\"]{x:y}")
	assertCSS2(t, "html{line-height:1;}html{line-height:1;}", "html{line-height:1}html{line-height:1}")
	assertCSS2(t, ".clearfix { *zoom: 1; }", ".clearfix{*zoom:1}")

	// coverage
	assertCSS2(t, "margin: 1 1;", "margin:1")
	assertCSS2(t, "margin: 1 2;", "margin:1 2")
	assertCSS2(t, "margin: 1 1 1;", "margin:1")
	assertCSS2(t, "margin: 1 2 1;", "margin:1 2")
	assertCSS2(t, "margin: 1 2 3;", "margin:1 2 3")
	assertCSS2(t, "margin: 0%;", "margin:0")
	assertCSS2(t, "color: rgb(255,64,64);", "color:#ff4040")
	assertCSS2(t, "color: rgb(256,-34,2342435);", "color:#f0f")
	assertCSS2(t, "color: rgb(120%,-45%,234234234%);", "color:#f0f")
	assertCSS2(t, "color: rgb(0, 1, ident);", "color:rgb(0,1,ident)")
	assertCSS2(t, "a, b + c { x:y; }", "a,b+c{x:y}")
	assertCSS2(t, "color: rgb(ident);", "color:rgb(ident)")
}
