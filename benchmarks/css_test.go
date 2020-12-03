package benchmarks

import (
	"testing"

	"github.com/alex-bacart/minify/v2/css"
)

var cssSamples = []string{
	"sample_bootstrap.css",
	"sample_gumby.css",
	"sample_fontawesome.css",
	"sample_normalize.css",
}

func init() {
	for _, sample := range cssSamples {
		load(sample)
	}
}

func BenchmarkCSS(b *testing.B) {
	for _, sample := range cssSamples {
		b.Run(sample, func(b *testing.B) {
			b.SetBytes(int64(r[sample].Len()))

			for i := 0; i < b.N; i++ {
				r[sample].Reset()
				w[sample].Reset()
				css.Minify(m, w[sample], r[sample], nil)
			}
		})
	}
}
