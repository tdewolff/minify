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

type Minifier func(io.ReadCloser) (io.ReadCloser, error)

func (minify Minify) MinifierByMime(mime string) Minifier {
	if mime == "text/html" {
		return minify.Html
	} else if mime == "text/javascript" {
		return minify.Js
	} else if mime == "text/css" {
		return minify.Css
	}
	return nil
}

func inline(minifier Minifier, val []byte) []byte {
	buffer, err := minifier(ioutil.NopCloser(bytes.NewBuffer(val)))
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