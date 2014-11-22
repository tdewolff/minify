package minify

import (
	"bytes"
	"errors"
	"io"
	"os/exec"
)

var ErrNotExist = errors.New("minifier does not exist for mime type")

////////////////////////////////////////////////////////////////

type MinifyFunc func(Minify, io.Writer, io.Reader) error

type Minify struct {
	Minifier map[string]MinifyFunc
}

func NewMinify() *Minify {
	return &Minify{
		map[string]MinifyFunc{
			"text/html":              (Minify).HTML,
			"text/css":               (Minify).CSS,
		},
	}
}

func (m *Minify) Implement(mime string, f MinifyFunc) {
	m.Minifier[mime] = f
}

func (m *Minify) ImplementCmd(mime string, cmd *exec.Cmd) error {
	m.Minifier[mime] = func(m Minify, w io.Writer, r io.Reader) error {
		stdOut, err := cmd.StdoutPipe()
		if err != nil {
			return err
		}
		defer stdOut.Close()
		stdIn, err := cmd.StdinPipe()
		if err != nil {
			return err
		}
		defer stdIn.Close()

		if err = cmd.Start(); err != nil {
			return err
		}
		if _, err := io.Copy(stdIn, r); err != nil {
			return err
		}
		stdIn.Close()
		if _, err = io.Copy(w, stdOut); err != nil {
			return err
		}

		return cmd.Wait()
	}
	return nil
}

func (m Minify) Filter(mime string, w io.Writer, r io.Reader) error {
	if f, ok := m.Minifier[mime]; ok {
		if err := f(m, w, r); err != nil {
			return err
		}
		return nil
	}
	return ErrNotExist
}

func (m Minify) FilterBytes(mime string, v []byte) []byte {
	b := &bytes.Buffer{}
	if err := m.Filter(mime, b, bytes.NewBuffer(v)); err != nil {
		return v
	}
	return b.Bytes()
}

func (m Minify) FilterString(mime string, v string) string {
	b := &bytes.Buffer{}
	if err := m.Filter(mime, b, bytes.NewBufferString(v)); err != nil {
		return v
	}
	return b.String()
}
