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
		// fuzz
		{"", ""},
		{"ML", ""},
		{".8.00c0", ""},
		{".1.04h0e6.0e6.0e0.0", "h0 0 0 0"},

		{"M10 10L11 10 11 11", "M10 10h1v1"},
		{"M10 10L11 11 0 0", "M10 10l1 1L0 0"},
		{"M246.614 51.028L246.614-5.665 189.922-5.665", "M246.614 51.028V-5.665H189.922"},

		{"M10 10 20 10", "M10 10H20"},
		{"M50 50 100 100", "M50 50l50 50"},
		{"m50 50 40 40m50 50", "m50 50 40 40m50 50"},
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
