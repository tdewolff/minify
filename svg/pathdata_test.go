package svg // import "github.com/tdewolff/minify/svg"

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathData(t *testing.T) {
	pathDataBuffer := &PathData{}
	var pathDataTests = []struct {
		pathData string
		expected string
	}{
		// fuzz
		{"", ""},
		{"ML", ""},
		{".8.00c0", ""},
		{".1.04h0e6.0e6.0e0.0", "h0 0 0 0"},
	}

	for _, tt := range pathDataTests {
		out := ShortenPathData([]byte(tt.pathData), pathDataBuffer)
		assert.Equal(t, tt.expected, string(out), "Path data must match")
	}
}

////////////////////////////////////////////////////////////////

func BenchmarkShortenPathData(b *testing.B) {
	pathDataBuffer := &PathData{}
	r := []byte("M8.64,223.948c0,0,143.468,3.431,185.777-181.808c2.673-11.702-1.23-20.154,1.316-33.146h16.287c0,0-3.14,17.248,1.095,30.848c21.392,68.692-4.179,242.343-204.227,196.59L8.64,223.948z")
	for i := 0; i < b.N; i++ {
		ShortenPathData(r, pathDataBuffer)
	}
}
