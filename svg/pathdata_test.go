package svg // import "github.com/tdewolff/minify/svg"

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathData(t *testing.T) {
	var pathDataTests = []struct {
		pathData string
		expected string
	}{
		{"M10 10 20 10", "M10 10H20"},
		{"M10 10 10 20", "M10 10V20"},
		{"M50 50 100 100", "M50 50l50 50"},
		{"m50 50 40 40m50 50", "m50 50 40 40m50 50"},
		{"M10 10zM15 15", "M10 10zm5 5"},
		{"M50 50H55V55", "M50 50h5v5"},
		{"M10 10L11 10 11 11", "M10 10h1v1"},
		{"M10 10l1 0 0 1", "M10 10h1v1"},
		{"M10 10L11 11 0 0", "M10 10l1 1L0 0"},
		{"M246.614 51.028L246.614-5.665 189.922-5.665", "M246.614 51.028V-5.665H189.922"},
		{"M100,200 C100,100 250,100 250,200 S400,300 400,200", "M100 200c0-100 150-100 150 0s150 100 150 0"},
		{"M200,300 Q400,50 600,300 T1000,300", "M200 300q200-250 400 0t400 0"},
		{"M300,200 h-150 a150,150 0 1,0 150,-150 z", "M300 200H150A150 150 0 1 0 300 50z"},
		{"x5 5L10 10", "L10 10"},

		// fuzz
		{"", ""},
		{"ML", ""},
		{".8.00c0", ""},
		{".1.04h0e6.0e6.0e0.0", "h0 0 0 0"},
		{"M.1.0.0.2Z", "M.1.2z"},
	}

	p := NewPathData(&Minifier{Decimals: -1})
	for _, tt := range pathDataTests {
		out := p.ShortenPathData([]byte(tt.pathData))
		assert.Equal(t, tt.expected, string(out), "Path data must match in "+tt.pathData)
	}
}

////////////////////////////////////////////////////////////////

func BenchmarkShortenPathData(b *testing.B) {
	p := NewPathData(&Minifier{})
	r := []byte("M8.64,223.948c0,0,143.468,3.431,185.777-181.808c2.673-11.702-1.23-20.154,1.316-33.146h16.287c0,0-3.14,17.248,1.095,30.848c21.392,68.692-4.179,242.343-204.227,196.59L8.64,223.948z")
	for i := 0; i < b.N; i++ {
		p.ShortenPathData(r)
	}
}
