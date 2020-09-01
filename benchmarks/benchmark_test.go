package benchmarks

import (
	"io/ioutil"
	"runtime"
	"testing"

	"github.com/tdewolff/minify/v2/minify"
	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/buffer"
)

func benchmark(b *testing.B, mediatype string, sample string) {
	m := minify.Default
	in, err := ioutil.ReadFile(sample)
	if err != nil {
		panic(err)
	}
	b.Run(sample, func(b *testing.B) {
		out := make([]byte, 0, len(in))
		b.SetBytes(int64(len(in)))
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			runtime.GC()
			r := buffer.NewReader(parse.Copy(in))
			w := buffer.NewWriter(out[:0])
			b.StartTimer()

			if err := m.Minify(mediatype, w, r); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkCSS(b *testing.B) {
	var samples = []string{
		"sample_bootstrap.css",
		"sample_gumby.css",
		"sample_fontawesome.css",
		"sample_normalize.css",
	}
	for _, sample := range samples {
		benchmark(b, "text/css", sample)
	}
}

func BenchmarkHTML(b *testing.B) {
	var samples = []string{
		"sample_amazon.html",
		"sample_bbc.html",
		"sample_blogpost.html",
		"sample_es6.html",
		"sample_stackoverflow.html",
		"sample_wikipedia.html",
	}
	for _, sample := range samples {
		benchmark(b, "text/html", sample)
	}
}

func BenchmarkJS(b *testing.B) {
	var samples = []string{
		"sample_ace.js",
		"sample_dot.js",
		"sample_jquery.js",
		"sample_jqueryui.js",
		"sample_moment.js",
	}
	for _, sample := range samples {
		benchmark(b, "application/javascript", sample)
	}
}

func BenchmarkJSON(b *testing.B) {
	var samples = []string{
		"sample_large.json",
		"sample_testsuite.json",
		"sample_twitter.json",
	}
	for _, sample := range samples {
		benchmark(b, "application/json", sample)
	}
}

func BenchmarkSVG(b *testing.B) {
	var samples = []string{
		"sample_arctic.svg",
		"sample_gopher.svg",
		"sample_usa.svg",
		"sample_car.svg",
		"sample_tiger.svg",
	}
	for _, sample := range samples {
		benchmark(b, "image/svg+xml", sample)
	}
}

func BenchmarkXML(b *testing.B) {
	var samples = []string{
		"sample_books.xml",
		"sample_catalog.xml",
		"sample_omg.xml",
	}
	for _, sample := range samples {
		benchmark(b, "application/xml", sample)
	}
}
