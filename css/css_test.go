package css // import "github.com/tdewolff/minify/css"

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tdewolff/minify"
)

func assertCSS(t *testing.T, input, expected string) {
	m := minify.NewMinifier()
	b := &bytes.Buffer{}
	assert.Nil(t, Minify(m, b, bytes.NewBufferString(input)), "Minify must not return error in "+input)
	assert.Equal(t, expected, b.String(), "Minify must give expected result in "+input)
}

func TestCSS(t *testing.T) {
	assertCSS(t, "/*comment*/", "")
	assertCSS(t, "css{}", "")
	assertCSS(t, "key: value;", "key:value")
	assertCSS(t, "margin: 0 1; padding: 0 1;", "margin:0 1;padding:0 1")
	assertCSS(t, "i { key: value; key2: value; }", "i{key:value;key2:value}")
	assertCSS(t, "color: #FF0000;", "color:red")
	assertCSS(t, "color: #000000;", "color:#000")
	assertCSS(t, "color: black;", "color:#000")
	assertCSS(t, "color: rgb(255,255,255);", "color:#fff")
	assertCSS(t, "color: rgb(100%,100%,100%);", "color:#fff")
	assertCSS(t, "color: rgba(255,0,0,1);", "color:red")
	assertCSS(t, "font-weight: bold; font-weight: normal;", "font-weight:700;font-weight:400")
	assertCSS(t, "font: bold \"Times new Roman\",\"Sans-Serif\";", "font:700 times new roman,\"sans-serif\"")
	assertCSS(t, "outline: none;", "outline:0")
	assertCSS(t, "border-left: none;", "border-left:0")
	assertCSS(t, "margin: 1 1 1 1;", "margin:1")
	assertCSS(t, "margin: 1 2 1 2;", "margin:1 2")
	assertCSS(t, "margin: 1 2 3 2;", "margin:1 2 3")
	assertCSS(t, "margin: 1 2 3 4;", "margin:1 2 3 4")
	assertCSS(t, "margin: 0em;", "margin:0")
	assertCSS(t, ".cla .ss > #id { x:y; }", ".cla .ss>#id{x:y}")
	assertCSS(t, ".cla[id ^= L] { x:y; }", ".cla[id^=L]{x:y}")
	assertCSS(t, "area:focus { outline : 0;}", "area:focus{outline:0}")
	assertCSS(t, "@import 'file';", "@import 'file'")
	assertCSS(t, "@import 'file' { x:y; };", "@import 'file'{x:y}")
	assertCSS(t, "<!-- x:y; -->", "<!--x:y-->")
	assertCSS(t, "font-family:'Arial', 'Times New Roman';", "font-family:arial,times new roman")
	assertCSS(t, "background:url('http://domain.com/image.png');", "background:url(http://domain.com/image.png)")
	assertCSS(t, "filter: progid : DXImageTransform.Microsoft.BasicImage(rotation=1);", "filter:progid:DXImageTransform.Microsoft.BasicImage(rotation=1)")
	assertCSS(t, "content: \"a\\\nb\";", "content:\"ab\"")
	assertCSS(t, "color:#fff;@charset x;", "color:#fff")
	assertCSS(t, "color:#fff;@import x;", "color:#fff")
	assertCSS(t, "@charset x;@import x;", "@charset x;@import x")
	assertCSS(t, "@charset;@import;", "")

	assertCSS(t, "font:27px/13px arial,sans-serif", "font:27px/13px arial,sans-serif")
	assertCSS(t, "text-decoration: none !important", "text-decoration:none!important")
	assertCSS(t, "color:#fff", "color:#fff")
	assertCSS(t, "border:2px rgb(255,255,255);", "border:2px #fff")
	assertCSS(t, "margin:-1px", "margin:-1px")
	assertCSS(t, "margin:+1px", "margin:1px")
	assertCSS(t, "margin:0.5em", "margin:.5em")
	assertCSS(t, "margin:-0.5em", "margin:-.5em")
	assertCSS(t, "margin:05em", "margin:5em")
	assertCSS(t, "margin:.50em", "margin:.5em")
	assertCSS(t, "margin:5.0em", "margin:5em")
	assertCSS(t, "color:#c0c0c0", "color:silver")
	assertCSS(t, "input[type=\"radio\"]{x:y}", "input[type=radio]{x:y}")
	assertCSS(t, "-ms-filter: \"progid:DXImageTransform.Microsoft.Alpha(Opacity=80)\";", "-ms-filter:\"alpha(opacity=80)\"")
	assertCSS(t, "filter: progid:DXImageTransform.Microsoft.Alpha(Opacity=80);", "filter:alpha(opacity=80)")
	assertCSS(t, "MARGIN:1EM", "margin:1em")
	assertCSS(t, "color:CYAN", "color:cyan")
	assertCSS(t, "background:URL(x.PNG);", "background:url(x.PNG)")
	assertCSS(t, "DIV{margin:1em}", "div{margin:1em}")
	assertCSS(t, ".CLASS{margin:1em}", ".CLASS{margin:1em}")
	assertCSS(t, "@media only screen and (max-width:800px)", "@media only screen and (max-width:800px)")
	assertCSS(t, "@media (-webkit-min-device-pixel-ratio:1.5),(min-resolution:1.5dppx)", "@media (-webkit-min-device-pixel-ratio:1.5),(min-resolution:1.5dppx)")

	// coverage
	assertCSS(t, "margin: 1 1;", "margin:1")
	assertCSS(t, "margin: 1 2;", "margin:1 2")
	assertCSS(t, "margin: 1 1 1;", "margin:1")
	assertCSS(t, "margin: 1 2 1;", "margin:1 2")
	assertCSS(t, "margin: 1 2 3;", "margin:1 2 3")
	assertCSS(t, "margin: 0%;", "margin:0")
	assertCSS(t, "color: rgb(255,64,64);", "color:#ff4040")
	assertCSS(t, "color: rgb(256,-34,2342435);", "color:#f0f")
	assertCSS(t, "color: rgb(120%,-45%,234234234%);", "color:#f0f")
	assertCSS(t, "color: rgb(0, 1, ident);", "color:rgb(0,1,ident)")
	assertCSS(t, "a, b + c { x:y; }", "a,b+c{x:y}")
	assertCSS(t, "color: rgb(ident);", "color:rgb(ident)")
}
