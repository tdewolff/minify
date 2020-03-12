package benchmarks

import (
	"io"
	"testing"

	"github.com/tdewolff/parse/v2/js"
)

func init() {
	for _, sample := range jsSamples {
		load(sample)
	}
}

func BenchmarkJSLex(b *testing.B) {
	for _, sample := range jsSamples {
		b.Run(sample, func(b *testing.B) {
			b.SetBytes(int64(r[sample].Len()))

			for i := 0; i < b.N; i++ {
				r[sample].Reset()
				l := js.NewLexer(r[sample])
				for {
					tt, _ := l.Next()
					if tt == js.DivToken || tt == js.DivEqToken {
						tt, _ = l.RegExp()
					}
					if tt == js.ErrorToken {
						if l.Err() != io.EOF {
							panic(l.Err())
						}
						break
					}
				}
			}
		})
	}
}
