package benchmarks

import (
	"io"
	"io/ioutil"

	"github.com/tdewolff/buffer"
	"github.com/tdewolff/minify"
)

var m = minify.New()
var r = map[string]*buffer.Reader{}
var w = map[string]*buffer.Writer{}

func load(filename string) {
	sample, _ := ioutil.ReadFile(filename)
	r[filename] = buffer.NewReader(sample)
	w[filename] = buffer.NewWriter(make([]byte, 0, len(sample)))
}

func readerToBytes(r io.Reader) []byte {
	var b []byte
	if r != nil {
		if buffer, ok := r.(interface {
			Bytes() []byte
		}); ok {
			b = buffer.Bytes()
		} else {
			var err error
			b, err = ioutil.ReadAll(r)
			if err != nil {
				panic(err)
			}
		}
	}
	return b
}
