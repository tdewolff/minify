package minify

import (
	"io"
	"log"
	"io/ioutil"
	"bytes"
)

type Minify struct {
	MimeMinifier map[string]Minifier
	JsMinifier []string
	TemplateDelims []string
}

type Minifier func(Minify, io.ReadCloser) (io.ReadCloser, error)

func (minify Minify) inline(mime string, val []byte) []byte {
	if minifier, ok := minify.MimeMinifier[mime]; ok {
		buffer, err := minifier(minify, ioutil.NopCloser(bytes.NewBuffer(val)))
		if err == nil {
			if newVal, err := ioutil.ReadAll(buffer); err == nil {
				return newVal
			} else {
				log.Println("ioutil.ReadAll:", err)
			}
		} else {
			log.Println("minify.Minifier:", err)
		}
	}
	return val
}

func (minify Minify) inlineString(mime string, val string) string {
	return string(minify.inline(mime, []byte(val)))
}