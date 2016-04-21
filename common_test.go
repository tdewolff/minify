package minify // import "github.com/tdewolff/minify"

import (
	"io"
	"io/ioutil"
	"math/rand"
	"testing"

	"github.com/tdewolff/test"
)

func TestContentType(t *testing.T) {
	contentTypeTests := []struct {
		contentType string
		expected    string
	}{
		{"text/html", "text/html"},
		{"text/html; charset=UTF-8", "text/html;charset=utf-8"},
		{"text/html; charset=UTF-8 ; param = \" ; \"", "text/html;charset=utf-8;param=\" ; \""},
		{"text/html, text/css", "text/html,text/css"},
	}
	for _, tt := range contentTypeTests {
		test.Minify(t, tt.contentType, nil, string(ContentType([]byte(tt.contentType))), tt.expected)
	}
}

func TestDataURI(t *testing.T) {
	dataURITests := []struct {
		dataURI  string
		expected string
	}{
		{"data:,text", "data:,text"},
		{"data:text/plain;charset=us-ascii,text", "data:,text"},
		{"data:TEXT/PLAIN;CHARSET=US-ASCII,text", "data:,text"},
		{"data:text/plain;charset=us-asciiz,text", "data:;charset=us-asciiz,text"},
		{"data:;base64,dGV4dA==", "data:,text"},
		{"data:text/svg+xml;base64,PT09PT09", "data:text/svg+xml;base64,PT09PT09"},
		{"data:text/xml;version=2.0,content", "data:text/xml;version=2.0,content"},
		{"data:text/xml; version = 2.0,content", "data:text/xml;version=2.0,content"},
		{"data:,=====", "data:,%3D%3D%3D%3D%3D"},
		{"data:,======", "data:;base64,PT09PT09"},
		{"data:text/x,<?x?>", "data:text/x,%3C%3Fx%3F%3E"},
	}
	m := New()
	m.AddFunc("text/x", func(_ *M, w io.Writer, r io.Reader, _ map[string]string) error {
		b, _ := ioutil.ReadAll(r)
		test.String(t, string(b), "<?x?>")
		w.Write(b)
		return nil
	})
	for _, tt := range dataURITests {
		test.Minify(t, tt.dataURI, nil, string(DataURI(m, []byte(tt.dataURI))), tt.expected)
	}
}

func TestNumber(t *testing.T) {
	numberTests := []struct {
		number   string
		expected string
	}{
		{"0", "0"},
		{".0", "0"},
		{"1.0", "1"},
		{"0.1", ".1"},
		{"+1", "1"},
		{"-1", "-1"},
		{"-0.1", "-.1"},
		{"100", "100"},
		{"1000", "1e3"},
		{"0.001", ".001"},
		{"0.0001", "1e-4"},
		{"100e1", "1e3"},
		{"1.1e+1", "11"},
		{"0.252", ".252"},
		{"1.252", "1.252"},
		{"-1.252", "-1.252"},
		{"0.075", ".075"},
		{"789012345678901234567890123456789e9234567890123456789", "789012345678901234567890123456789e9234567890123456789"},
		{".000100009", "100009e-9"},
		{".0001000009", ".0001000009"},
		{".0001000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000009", ".0001000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000009"},
		{"E\x1f", "E\x1f"}, // fuzz
		//{"96px", "1in"},
	}
	for _, tt := range numberTests {
		test.Minify(t, tt.number, nil, string(Number([]byte(tt.number))), tt.expected)
	}
}

////////////////

var num []int64

func RandNumBytes() []byte {
	var b []byte
	n := rand.Int() % 10
	for i := 0; i < n; i++ {
		b = append(b, byte(rand.Int()%10)+'0')
	}
	if rand.Int()%2 == 0 {
		b = append(b, '.')
		n = rand.Int() % 10
		for i := 0; i < n; i++ {
			b = append(b, byte(rand.Int()%10)+'0')
		}
	}
	if rand.Int()%2 == 0 {
		b = append(b, 'e')
		if rand.Int()%2 == 0 {
			b = append(b, '-')
		}
		n = rand.Int() % 5
		for i := 0; i < n; i++ {
			b = append(b, byte(rand.Int()%10)+'0')
		}
	}
	return b
}

func TestMain(t *testing.T) {
	for j := 0; j < 1000; j++ {
		num = append(num, rand.Int63n(1000))
	}
}

func BenchmarkNumber(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Number(RandNumBytes())
	}
}
