package minify // import "github.com/tdewolff/minify"

import (
	"io"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func assertDataURI(t *testing.T, m Minifier, s, e string) {
	assert.Equal(t, e, string(DataURI(m, []byte(s))), "data URIs must match")
}

func assertNumber(t *testing.T, x, e string) {
	assert.Equal(t, e, string(Number([]byte(x))), "numbers must match in "+x)
}

////////////////////////////////////////////////////////////////

func TestContentType(t *testing.T) {
	assert.Equal(t, "text/html", string(ContentType([]byte("text/html"))))
	assert.Equal(t, "text/html;charset=utf-8", string(ContentType([]byte("text/html; charset=UTF-8"))))
	assert.Equal(t, "text/html;charset=utf-8;param=\" ; \"", string(ContentType([]byte("text/html; charset=UTF-8 ; param = \" ; \""))))
	assert.Equal(t, "text/html,text/css", string(ContentType([]byte("text/html, text/css"))))
}

func TestDataURI(t *testing.T) {
	m := New()
	m.AddFunc("text/x", func(m Minifier, mediatype string, w io.Writer, r io.Reader) error {
		b, _ := ioutil.ReadAll(r)
		assert.Equal(t, "<?x?>", string(b))
		w.Write(b)
		return nil
	})
	assertDataURI(t, m, "data:text/x,<?x?>", "data:text/x,%3C%3Fx%3F%3E")
	assertDataURI(t, m, "data:,text", "data:,text")
	assertDataURI(t, m, "data:;base64,dGV4dA==", "data:,text")
	assertDataURI(t, m, "data:text/svg+xml;base64,PT09PT09", "data:text/svg+xml;base64,PT09PT09")
	assertDataURI(t, m, "data:text/xml;version=2.0,content", "data:text/xml;version=2.0,content")
	assertDataURI(t, m, "data:text/xml; version = 2.0,content", "data:text/xml;version=2.0,content")
	assertDataURI(t, m, "data:,=====", "data:,%3D%3D%3D%3D%3D")
	assertDataURI(t, m, "data:,======", "data:;base64,PT09PT09")
}

func TestNumber(t *testing.T) {
	assertNumber(t, "0", "0")
	assertNumber(t, "1.0", "1")
	assertNumber(t, "0.1", ".1")
	assertNumber(t, "+1", "1")
	assertNumber(t, "-0.1", "-.1")
	assertNumber(t, "100", "100")
	// assertNumber(t, "1000px", "1e3px")
	// assertNumber(t, "0.001px", "1e-3px")
	// assertNumber(t, "96px", "1in")
}
