package minify

import (
	"io"
	"io/ioutil"
	"bytes"
)

type Minifier func(io.Reader) (io.Reader, error)

func inlineMinify(minifier Minifier, val []byte) []byte {
	buffer, err := minifier(bytes.NewBuffer(val))
	if err == nil {
		if newVal, err := ioutil.ReadAll(buffer); err == nil {
			return newVal
		}
	}
	return val
}

func inlineMinifyString(minifier Minifier, val string) string {
	return string(inlineMinify(minifier, []byte(val)))
}