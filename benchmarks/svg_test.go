package benchmarks

import (
	"testing"

	"github.com/tdewolff/minify/v2/svg"
)

var svgSamples = []string{
	"sample_arctic.svg",
	"sample_gopher.svg",
	"sample_usa.svg",
	"sample_car.svg",
	"sample_tiger.svg",
}

func init() {
	for _, sample := range svgSamples {
		load(sample)
	}
}

func BenchmarkSVG(b *testing.B) {
	for _, sample := range svgSamples {
		b.Run(sample, func(b *testing.B) {
			b.SetBytes(int64(r[sample].Len()))

			for i := 0; i < b.N; i++ {
				r[sample].Reset()
				w[sample].Reset()
				svg.Minify(m, w[sample], r[sample], nil)
			}
		})
	}
}
