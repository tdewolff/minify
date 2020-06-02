package benchmarks

import (
	"testing"

	"github.com/tdewolff/minify/v2/js"
)

func init() {
	load("jquery50.js")
}

func BenchmarkJQuery(b *testing.B) {
	b.Run("jquery50.js", func(b *testing.B) {
		b.SetBytes(int64(r["jquery50.js"].Len()))

		for i := 0; i < b.N; i++ {
			r["jquery50.js"].Reset()
			w["jquery50.js"].Reset()
			js.Minify(m, w["jquery50.js"], r["jquery50.js"], nil)
		}
	})
}
