package minify

import (
	"bytes"
	"testing"
)

func helperCSS(t *testing.T, input, expected string) {
	m := &Minifier{}
	b := &bytes.Buffer{}
	if err := m.CSS(b, bytes.NewBufferString(input)); err != nil {
		t.Error(err)
	}

	if b.String() != expected {
		t.Error(b.String(), "!=", expected)
	}
}

func TestCSS(t *testing.T) {
	helperCSS(t, "/*comment*/", "")
	helperCSS(t, "css{}", "")
	helperCSS(t, "key: value;", "key:value")
	helperCSS(t, "margin: 0 1; padding: 0 1;", "margin:0 1;padding:0 1")
	helperCSS(t, "i { key: value; key2: value; }", "i{key:value;key2:value}")
	helperCSS(t, "color: #FF0000;", "color:red")
	helperCSS(t, "color: #000000;", "color:#000")
	helperCSS(t, "color: black;", "color:#000")
	helperCSS(t, "color: rgb(255,255,255);", "color:#fff")
	helperCSS(t, "color: rgb(100%,100%,100%);", "color:#fff")
	helperCSS(t, "color: rgba(255,0,0,1);", "color:red")
	helperCSS(t, "font-weight: bold; font-weight: normal;", "font-weight:700;font-weight:400")
	helperCSS(t, "font: bold \"Times new Roman\",\"Sans-Serif\";", "font:700 times new roman,\"sans-serif\"")
	helperCSS(t, "outline: none;", "outline:0")
	helperCSS(t, "border-left: none;", "border-left:0")
	helperCSS(t, "margin: 1 1 1 1;", "margin:1")
	helperCSS(t, "margin: 1 2 1 2;", "margin:1 2")
	helperCSS(t, "margin: 1 2 3 2;", "margin:1 2 3")
	helperCSS(t, "margin: 1 2 3 4;", "margin:1 2 3 4")
	helperCSS(t, "margin: 0em;", "margin:0")
	helperCSS(t, ".cla .ss > #id { x:y; }", ".cla .ss>#id{x:y}")
	helperCSS(t, ".cla[id ^= L] { x:y; }", ".cla[id^=L]{x:y}")
	helperCSS(t, "area:focus { outline : 0;}", "area:focus{outline:0}")
	helperCSS(t, "@import 'file';", "@import 'file'")
	helperCSS(t, "@import 'file' { x:y; };", "@import 'file'{x:y}")
	helperCSS(t, "<!-- x:y; -->", "<!--x:y-->")
	helperCSS(t, "font-family:'Arial', 'Times New Roman';", "font-family:arial,times new roman")
	helperCSS(t, "background:url('http://domain.com/image.png');", "background:url(http://domain.com/image.png)")
	helperCSS(t, "filter: progid : DXImageTransform.Microsoft.BasicImage(rotation=1);", "filter:progid:DXImageTransform.Microsoft.BasicImage(rotation=1)")
	helperCSS(t, "content: \"a\\\nb\";", "content:\"ab\"")
	helperCSS(t, "color:#fff;@charset x;", "color:#fff")
	helperCSS(t, "color:#fff;@import x;", "color:#fff")
	helperCSS(t, "@charset x;@import x;", "@charset x;@import x")
	helperCSS(t, "@charset;@import;", "")

	helperCSS(t, "font:27px/13px arial,sans-serif", "font:27px/13px arial,sans-serif")
	helperCSS(t, "text-decoration: none !important", "text-decoration:none!important")
	helperCSS(t, "color:#fff", "color:#fff")
	helperCSS(t, "border:2px rgb(255,255,255);", "border:2px #fff")
	helperCSS(t, "margin:-1px", "margin:-1px")
	helperCSS(t, "margin:+1px", "margin:1px")
	helperCSS(t, "margin:0.5em", "margin:.5em")
	helperCSS(t, "margin:-0.5em", "margin:-.5em")
	helperCSS(t, "margin:05em", "margin:5em")
	helperCSS(t, "margin:.50em", "margin:.5em")
	helperCSS(t, "color:#c0c0c0", "color:silver")
	helperCSS(t, "input[type=\"radio\"]{x:y}", "input[type=radio]{x:y}")
	helperCSS(t, "-ms-filter: \"progid:DXImageTransform.Microsoft.Alpha(Opacity=80)\";", "-ms-filter:\"alpha(opacity=80)\"")
	helperCSS(t, "filter: progid:DXImageTransform.Microsoft.Alpha(Opacity=80);", "filter:alpha(opacity=80)")
	helperCSS(t, "MARGIN:1EM", "margin:1em")
	helperCSS(t, "color:CYAN", "color:cyan")
	helperCSS(t, "background:URL(x.PNG);", "background:url(x.PNG)")
	helperCSS(t, "DIV{margin:1em}", "div{margin:1em}")
	helperCSS(t, ".CLASS{margin:1em}", ".CLASS{margin:1em}")

	// advanced
	helperCSS(t, "test,test2,test { x:y; }", "test,test2{x:y}")
	helperCSS(t, "test{ x:y; x:z; }", "test{x:z}")
	helperCSS(t, "test[id=a]{x:y;}", "test#a{x:y}")
	helperCSS(t, "test[class='b']{x:y;}", "test.b{x:y}")
	helperCSS(t, "test{ font-style: italic; font-variant: small-caps; font-weight: bold; font-size: 12pt; line-height:110%; font-family:serif; }", "test{font:italic small-caps 700 12pt/110% serif}")

	// coverage
	helperCSS(t, "margin: 1 1;", "margin:1")
	helperCSS(t, "margin: 1 2;", "margin:1 2")
	helperCSS(t, "margin: 1 1 1;", "margin:1")
	helperCSS(t, "margin: 1 2 1;", "margin:1 2")
	helperCSS(t, "margin: 1 2 3;", "margin:1 2 3")
	helperCSS(t, "margin: 0%;", "margin:0")
	helperCSS(t, "color: rgb(255,64,64);", "color:#ff4040")
	helperCSS(t, "color: rgb(256,-34,2342435);", "color:#f0f")
	helperCSS(t, "color: rgb(120%,-45%,234234234%);", "color:#f0f")
	helperCSS(t, "color: rgb(0, 1, ident);", "color:rgb(0,1,ident)")
	helperCSS(t, "a, b + c { x:y; }", "a,b+c{x:y}")
	helperCSS(t, "color: rgb(ident);", "color:rgb(ident)")
}
