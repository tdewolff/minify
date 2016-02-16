package minify // import "github.com/tdewolff/minify"

import (
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContentType(t *testing.T) {
	var contentTypeTests = []struct {
		contentType string
		expected    string
	}{
		{"text/html", "text/html"},
		{"text/html; charset=UTF-8", "text/html;charset=utf-8"},
		{"text/html; charset=UTF-8 ; param = \" ; \"", "text/html;charset=utf-8;param=\" ; \""},
		{"text/html, text/css", "text/html,text/css"},
	}
	for _, tt := range contentTypeTests {
		assert.Equal(t, tt.expected, string(ContentType([]byte(tt.contentType))), "ContentType must give expected result in "+tt.contentType)
	}
}

func TestDataURI(t *testing.T) {
	var dataURITests = []struct {
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
		assert.Equal(t, "<?x?>", string(b))
		w.Write(b)
		return nil
	})
	for _, tt := range dataURITests {
		assert.Equal(t, tt.expected, string(DataURI(m, []byte(tt.dataURI))), "DataURI must give expected result in "+tt.dataURI)
	}
}

func TestNumber(t *testing.T) {
	var numberTests = []struct {
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
		{"E\x1f", ""}, // fuzz
		//{"96px", "1in"},
	}
	for _, tt := range numberTests {
		assert.Equal(t, tt.expected, string(Number([]byte(tt.number))), "Number must give expected result in "+tt.number)
	}
}

func TestLenInt(t *testing.T) {
	var lenIntTests = []struct {
		number   int64
		expected int
	}{
		{0, 1},
		{1, 1},
		{10, 2},
		{99, 2},

		// coverage
		{100, 3},
		{1000, 4},
		{10000, 5},
		{100000, 6},
		{1000000, 7},
		{10000000, 8},
		{100000000, 9},
		{1000000000, 10},
		{10000000000, 11},
		{100000000000, 12},
		{1000000000000, 13},
		{10000000000000, 14},
		{100000000000000, 15},
		{1000000000000000, 16},
		{10000000000000000, 17},
		{100000000000000000, 18},
		{1000000000000000000, 19},
	}
	for _, tt := range lenIntTests {
		assert.Equal(t, tt.expected, lenInt64(tt.number), "lenInt must give expected result in "+strconv.FormatInt(tt.number, 10))
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

func BenchmarkLenIntLog(b *testing.B) {
	n := 0
	for i := 0; i < b.N; i++ {
		for j := 0; j < 1000; j++ {
			n += int(math.Log10(math.Abs(float64(num[j])))) + 1
		}
	}
}

func BenchmarkLenIntSwitch(b *testing.B) {
	n := 0
	for i := 0; i < b.N; i++ {
		for j := 0; j < 1000; j++ {
			n += lenInt64(num[j])
		}
	}
}

func BenchmarkNumber(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Number(RandNumBytes())
	}
}
