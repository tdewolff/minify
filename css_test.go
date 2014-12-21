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
	helperCSS(t, "color: rgb(255,255,255);", "color:#FFF")
	helperCSS(t, "color: rgb(100%,100%,100%);", "color:#FFF")
	helperCSS(t, "color: rgba(255,0,0,1);", "color:red")
	helperCSS(t, "font-weight: bold; font-weight: normal;", "font-weight:700;font-weight:400")
	helperCSS(t, "outline: none;", "outline:0")
	helperCSS(t, "margin: 1 1 1 1;", "margin:1")
	helperCSS(t, "margin: 1 2 1 2;", "margin:1 2")
	helperCSS(t, "margin: 1 2 3 2;", "margin:1 2 3")
	helperCSS(t, "margin: 1 2 3 4;", "margin:1 2 3 4")
	helperCSS(t, "margin: 0em;", "margin:0")
	helperCSS(t, ".cla .ss > #id { x:y; }", ".cla .ss>#id{x:y}")
	helperCSS(t, ".cla[id ^= L] { x:y; }", ".cla[id^=L]{x:y}")

	helperCSS(t, "font:27px/13px arial,sans-serif", "font:27px/13px arial,sans-serif")
	helperCSS(t, "text-decoration:none!important", "text-decoration:none!important")
	helperCSS(t, "margin:-1px", "margin:-1px")
	helperCSS(t, "color:#fff", "color:#FFF")
	helperCSS(t, "border:2px rgb(255,255,255);", "border:2px #FFF")

	// coverage
	helperCSS(t, "margin: 1 1;", "margin:1")
	helperCSS(t, "margin: 1 2;", "margin:1 2")
	helperCSS(t, "margin: 1 1 1;", "margin:1")
	helperCSS(t, "margin: 1 2 1;", "margin:1 2")
	helperCSS(t, "margin: 1 2 3;", "margin:1 2 3")
	helperCSS(t, "color: rgb(255,64,64);", "color:#FF4040")
	helperCSS(t, "color: rgb(256,-34,2342435);", "color:#F0F")
	helperCSS(t, "color: rgb(120%,-56%,234234234%);", "color:#F0F")
	helperCSS(t, "color: rgb(ident);", "color:rgb(ident)")
}
