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
