package benchmarks

import (
	"io"
	"testing"

	"github.com/tdewolff/parse/v2/js"
)

func init() {
	load("jquery50.js")
}

func BenchmarkJQueryParse(b *testing.B) {
	b.Run("jquery50.js", func(b *testing.B) {
		b.SetBytes(int64(r["jquery50.js"].Len()))

		for i := 0; i < b.N; i++ {
			r["jquery50.js"].Reset()
			_, _ = js.Parse(r["jquery50.js"])
		}
	})
}

func BenchmarkJQueryLex(b *testing.B) {
	b.Run("jquery50.js", func(b *testing.B) {
		b.SetBytes(int64(r["jquery50.js"].Len()))

		for i := 0; i < b.N; i++ {
			r["jquery50.js"].Reset()
			l := js.NewLexer(r["jquery50.js"])
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
