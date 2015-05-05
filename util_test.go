package minify // import "github.com/tdewolff/minify"

import (
	"io"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func assertMinifyDataURI(t *testing.T, m Minifier, s, e string) {
	assert.Equal(t, e, string(MinifyDataURI(m, []byte(s))), "data URIs must match")
}

func assertMinifyNumber(t *testing.T, x, e string) {
	assert.Equal(t, e, string(MinifyNumber([]byte(x))), "numbers must match in "+x)
}

func TestDataURI(t *testing.T) {
	m := New()
	m.AddFunc("text/x", func(m Minifier, mediatype string, w io.Writer, r io.Reader) error {
		b, _ := ioutil.ReadAll(r)
		assert.Equal(t, "<?x?>", string(b))
		w.Write(b)
		return nil
	})
	assertMinifyDataURI(t, m, "data:text/x,<?x?>", "data:text/x,%3C%3Fx%3F%3E")
	assertMinifyDataURI(t, m, "data:,text", "data:,text")
	assertMinifyDataURI(t, m, "data:;base64,dGV4dA==", "data:,text")
	assertMinifyDataURI(t, m, "data:text/svg+xml;base64,PT09PT09", "data:text/svg+xml;base64,PT09PT09")
	assertMinifyDataURI(t, m, "data:text/xml;version=2.0,content", "data:text/xml;version=2.0,content")
	assertMinifyDataURI(t, m, "data:text/xml; version = 2.0,content", "data:text/xml;version=2.0,content")
	assertMinifyDataURI(t, m, "data:,=====", "data:,%3D%3D%3D%3D%3D")
	assertMinifyDataURI(t, m, "data:,======", "data:;base64,PT09PT09")
}

func TestNumber(t *testing.T) {
	assertMinifyNumber(t, "0", "0")
	assertMinifyNumber(t, "1.0", "1")
	assertMinifyNumber(t, "0.1", ".1")
	assertMinifyNumber(t, "+1", "1")
	assertMinifyNumber(t, "-0.1", "-.1")
	assertMinifyNumber(t, "100", "100")
	// assertMinifyNumber(t, "1000px", "1e3px")
	// assertMinifyNumber(t, "0.001px", "1e-3px")
	// assertMinifyNumber(t, "96px", "1in")
}
