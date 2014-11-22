package minify

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"

	"github.com/kballard/go-shellquote"
)

var ErrNotExist = errors.New("minifier does not exist for mime type")

////////////////////////////////////////////////////////////////

type Minifier func(Minify, io.Reader) (io.Reader, error)

type Minify struct {
	MimeMinifier map[string]Minifier
	JsMinifier   []string
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

func (m Minify) Filter(mime string, r io.Reader) (io.Reader, error) {
	if f, ok := m.MimeMinifier[mime]; ok {
		r, err := f(m, r)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	return nil, ErrNotExist
}

func (m Minify) FilterBytes(mime string, v []byte) []byte {
	r, err := m.Filter(mime, bytes.NewBuffer(v))
	if err != nil {
		return v
	}

	if w, err := ioutil.ReadAll(r); err == nil {
		return w
	}
	return v
}

func (m Minify) FilterString(mime string, v string) string {
	return string(m.FilterBytes(mime, []byte(v)))
}
