package benchmarks

import (
	"testing"

	"github.com/tdewolff/parse/v2/js"
)

func init() {
	for _, sample := range jsSamples {
		load(sample)
	}
}

func BenchmarkJSParse(b *testing.B) {
	for _, sample := range jsSamples {
		b.Run(sample, func(b *testing.B) {
			b.SetBytes(int64(r[sample].Len()))

			for i := 0; i < b.N; i++ {
				r[sample].Reset()
				_, _ = js.Parse(r[sample])
			}
		})
	}
}
