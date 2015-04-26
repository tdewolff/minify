package css // import "github.com/tdewolff/minify/css"

import (
	"bytes"
	"io"
	"io/ioutil"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/xml"
)

func assertCSS(t *testing.T, m *minify.Minify, input, expected string) {
	b := &bytes.Buffer{}
	assert.Nil(t, Minify(m, "text/css", b, bytes.NewBufferString(input)), "Minify must not return error in "+input)
	assert.Equal(t, expected, b.String(), "Minify must give expected result in "+input)
}

func TestCSS(t *testing.T) {
	m := minify.New()
	m.AddFunc("text/css", Minify)

	assertCSS(t, m, "/*comment*/", "")
	assertCSS(t, m, "key: value;", "key:value")
	assertCSS(t, m, "margin: 0 1; padding: 0 1;", "margin:0 1;padding:0 1")
	assertCSS(t, m, "i { key: value; key2: value; }", "i{key:value;key2:value}")
	assertCSS(t, m, "color: #FF0000;", "color:red")
	assertCSS(t, m, "color: #000000;", "color:#000")
	assertCSS(t, m, "color: black;", "color:#000")
	assertCSS(t, m, "color: rgb(255,255,255);", "color:#fff")
	assertCSS(t, m, "color: rgb(100%,100%,100%);", "color:#fff")
	assertCSS(t, m, "color: rgba(255,0,0,1);", "color:red")
	assertCSS(t, m, "font-weight: bold; font-weight: normal;", "font-weight:700;font-weight:400")
	assertCSS(t, m, "font: bold \"Times new Roman\",\"Sans-Serif\";", "font:700 times new roman,\"sans-serif\"")
	assertCSS(t, m, "outline: none;", "outline:0")
	assertCSS(t, m, "border-left: none;", "border-left:0")
	assertCSS(t, m, "margin: 1 1 1 1;", "margin:1")
	assertCSS(t, m, "margin: 1 2 1 2;", "margin:1 2")
	assertCSS(t, m, "margin: 1 2 3 2;", "margin:1 2 3")
	assertCSS(t, m, "margin: 1 2 3 4;", "margin:1 2 3 4")
	assertCSS(t, m, "margin: 0em;", "margin:0")
	assertCSS(t, m, ".cla .ss > #id { x:y; }", ".cla .ss>#id{x:y}")
	assertCSS(t, m, ".cla[id ^= L] { x:y; }", ".cla[id^=L]{x:y}")
	assertCSS(t, m, "area:focus { outline : 0;}", "area:focus{outline:0}")
	assertCSS(t, m, "@import 'file';", "@import 'file'")
	assertCSS(t, m, "@import 'file' { x:y; };", "@import 'file'{x:y}")
	assertCSS(t, m, "<!-- x:y; -->", "<!--x:y-->")
	assertCSS(t, m, "font-family:'Arial', 'Times New Roman';", "font-family:arial,times new roman")
	assertCSS(t, m, "background:url('http://domain.com/image.png');", "background:url(http://domain.com/image.png)")
	assertCSS(t, m, "filter: progid : DXImageTransform.Microsoft.BasicImage(rotation=1);", "filter:progid:DXImageTransform.Microsoft.BasicImage(rotation=1)")
	assertCSS(t, m, "filter: progid:DXImageTransform.Microsoft.Alpha(Opacity=0);", "filter:alpha(opacity=0)")
	assertCSS(t, m, "content: \"a\\\nb\";", "content:\"ab\"")
	assertCSS(t, m, "content: \"a\\\r\nb\\\r\nc\";", "content:\"abc\"")

	assertCSS(t, m, "font:27px/13px arial,sans-serif", "font:27px/13px arial,sans-serif")
	assertCSS(t, m, "text-decoration: none !important", "text-decoration:none!important")
	assertCSS(t, m, "color:#fff", "color:#fff")
	assertCSS(t, m, "border:2px rgb(255,255,255);", "border:2px #fff")
	assertCSS(t, m, "margin:-1px", "margin:-1px")
	assertCSS(t, m, "margin:+1px", "margin:1px")
	assertCSS(t, m, "margin:0.5em", "margin:.5em")
	assertCSS(t, m, "margin:-0.5em", "margin:-.5em")
	assertCSS(t, m, "margin:05em", "margin:5em")
	assertCSS(t, m, "margin:.50em", "margin:.5em")
	assertCSS(t, m, "margin:5.0em", "margin:5em")
	assertCSS(t, m, "color:#c0c0c0", "color:silver")
	assertCSS(t, m, "input[type=\"radio\"]{x:y}", "input[type=radio]{x:y}")
	assertCSS(t, m, "-ms-filter: \"progid:DXImageTransform.Microsoft.Alpha(Opacity=80)\";", "-ms-filter:\"alpha(opacity=80)\"")
	assertCSS(t, m, "filter: progid:DXImageTransform.Microsoft.Alpha(Opacity=80);", "filter:alpha(opacity=80)")
	assertCSS(t, m, "MARGIN:1EM", "margin:1em")
	assertCSS(t, m, "color:CYAN", "color:cyan")
	assertCSS(t, m, "background:URL(x.PNG);", "background:url(x.PNG)")
	assertCSS(t, m, "DIV{margin:1em}", "div{margin:1em}")
	assertCSS(t, m, ".CLASS{margin:1em}", ".CLASS{margin:1em}")
	assertCSS(t, m, "@media only screen and (max-width:800px)", "@media only screen and (max-width:800px)")
	assertCSS(t, m, "@media (-webkit-min-device-pixel-ratio:1.5),(min-resolution:1.5dppx)", "@media (-webkit-min-device-pixel-ratio:1.5),(min-resolution:1.5dppx)")
	assertCSS(t, m, "[class^=icon-] i[class^=icon-],i[class*=\" icon-\"]{x:y}", "[class^=icon-] i[class^=icon-],i[class*=\" icon-\"]{x:y}")
	assertCSS(t, m, "html{line-height:1;}html{line-height:1;}", "html{line-height:1}html{line-height:1}")
	assertCSS(t, m, ".clearfix { *zoom: 1; }", ".clearfix{*zoom:1}")
	assertCSS(t, m, "background:url(/*nocomment*/)", "background:url(/*nocomment*/)")
	assertCSS(t, m, "background:url(data:,text)", "background:url(data:,text)")
	assertCSS(t, m, "background:url(data:;base64,dGV4dA==)", "background:url(data:,text)")
	assertCSS(t, m, "background:url(data:text/svg+xml;base64,PT09PT09)", "background:url(data:text/svg+xml;base64,PT09PT09)")
	assertCSS(t, m, "background:url(data:text/xml;version=2.0,content)", "background:url(data:text/xml;version=2.0,content)")
	assertCSS(t, m, "background:url('data:text/xml; version = 2.0,content')", "background:url(data:text/xml;version=2.0,content)")
	assertCSS(t, m, "background:url(data:text/css,color%3A#ff0000;)", "background:url(data:text/css,color%3Ared)")
	assertCSS(t, m, "background:url(data:,=====)", "background:url(data:,%3D%3D%3D%3D%3D)")
	assertCSS(t, m, "background:url(data:,======)", "background:url(data:;base64,PT09PT09)")

	// coverage
	assertCSS(t, m, "margin: 1 1;", "margin:1")
	assertCSS(t, m, "margin: 1 2;", "margin:1 2")
	assertCSS(t, m, "margin: 1 1 1;", "margin:1")
	assertCSS(t, m, "margin: 1 2 1;", "margin:1 2")
	assertCSS(t, m, "margin: 1 2 3;", "margin:1 2 3")
	assertCSS(t, m, "margin: 0%;", "margin:0")
	assertCSS(t, m, "color: rgb(255,64,64);", "color:#ff4040")
	assertCSS(t, m, "color: rgb(256,-34,2342435);", "color:#f0f")
	assertCSS(t, m, "color: rgb(120%,-45%,234234234%);", "color:#f0f")
	assertCSS(t, m, "color: rgb(0, 1, ident);", "color:rgb(0,1,ident)")
	assertCSS(t, m, "a, b + c { x:y; }", "a,b+c{x:y}")
	assertCSS(t, m, "color: rgb(ident);", "color:rgb(ident)")
	assertCSS(t, m, "margin: rgb(ident);", "margin:rgb(ident)")
	assertCSS(t, m, "filter: progid:b().c.Alpha(rgba(x));", "filter:progid:b().c.Alpha(rgba(x))")
}

func TestDataURI(t *testing.T) {
	m := minify.New()
	m.AddFunc("text/css", Minify)
	m.AddFunc("text/xml", func(m minify.Minifier, mediatype string, w io.Writer, r io.Reader) error {
		b, _ := ioutil.ReadAll(r)
		assert.Equal(t, "<?xml?>", string(b))
		w.Write(b)
		return nil
	})
	m.AddFuncRegexp(regexp.MustCompile("^.+[/+]xml$"), xml.Minify)
	assertCSS(t, m, "a{background:url('data:text/xml,<?xml?>')}", "a{background:url(data:text/xml,%3C%3Fxml%3F%3E)}")
}
