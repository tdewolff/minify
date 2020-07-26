package benchmarks

import (
	"io/ioutil"

	"github.com/alex-bacart/minify/v2"
	"github.com/tdewolff/parse/v2/buffer"
)

var m = minify.New()
var r = map[string]*buffer.Reader{}
var w = map[string]*buffer.Writer{}

func load(filename string) {
	sample, _ := ioutil.ReadFile(filename)
	r[filename] = buffer.NewReader(sample)
	w[filename] = buffer.NewWriter(make([]byte, 0, len(sample)))
}
