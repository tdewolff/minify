package minify

import (
	"io"
	"log"
	"io/ioutil"
	"bytes"
)

type Minify struct {
	UglifyjsPath string
}

type Minifier func(io.Reader) (io.Reader, error)

func inline(minifier Minifier, val []byte) []byte {
	buffer, err := minifier(bytes.NewBuffer(val))
	if err == nil {
		if newVal, err := ioutil.ReadAll(buffer); err == nil {
			return newVal
		} else {
			log.Println(err)
		}
	} else {
		log.Println(err)
	}
	return val
}

func inlineString(minifier Minifier, val string) string {
	return string(inline(minifier, []byte(val)))
}