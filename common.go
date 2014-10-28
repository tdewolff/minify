package minify

import (
	"bytes"
	"log"
	"io"
	"io/ioutil"

	"github.com/kballard/go-shellquote"
)

type Minifier func(Minify, io.ReadCloser) (io.ReadCloser, error)

type Minify struct {
	MimeMinifier map[string]Minifier
	JsMinifier []string
}

func NewMinify(jsMinifier string) *Minify {
	jsMinifierCmd, err := shellquote.Split(jsMinifier)
	if err != nil {
		jsMinifierCmd = []string{}
	}

	return &Minify{
		map[string]Minifier{
			"text/html":              (Minify).Html,
			"text/javascript":        (Minify).Js,
			"application/javascript": (Minify).Js,
			"text/css":               (Minify).Css,
		},
		jsMinifierCmd,
	}
}

func (minify Minify) Filter(mime string, r io.ReadCloser) (io.ReadCloser, error) {
	if minifier, ok := minify.MimeMinifier[mime]; ok {
		rm, err := minifier(minify, r)
		r.Close()
		if err != nil {
			return nil, err
		} else {
			return rm, nil
		}
	}
	return r, nil
}

func (minify Minify) inline(mime string, v []byte) []byte {
	if minifier, ok := minify.MimeMinifier[mime]; ok {
		b, err := minifier(minify, ioutil.NopCloser(bytes.NewBuffer(v)))
		if err == nil {
			if w, err := ioutil.ReadAll(b); err == nil {
				return w
			} else {
				log.Println("ioutil.ReadAll:", err)
			}
		} else {
			log.Println("minify.Minifier:", err)
		}
	}
	return v
}

func (minify Minify) inlineString(mime string, v string) string {
	return string(minify.inline(mime, []byte(v)))
}