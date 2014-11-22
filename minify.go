package minify

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os/exec"

	"github.com/kballard/go-shellquote"
)

var ErrNotExist = errors.New("minifier does not exist for mime type")

////////////////////////////////////////////////////////////////

type MinifyFunc func(Minify, io.Reader) (io.Reader, error)

type Minify struct {
	Minifier map[string]MinifyFunc
}

func NewMinify() *Minify {
	return &Minify{
		map[string]MinifyFunc{
			"text/html":              (Minify).Html,
			"text/css":               (Minify).Css,
		},
	}
}

func (m *Minify) Implement(mime string, f MinifyFunc) {
	m.Minifier[mime] = f
}

func (m *Minify) ImplementCmd(mime string, cmdString string) error {
	cmdSplit, err := shellquote.Split(cmdString)
	if err != nil {
		return err
	}

	m.Minifier[mime] = func (m Minify, r io.Reader) (io.Reader, error) {
		var cmd *exec.Cmd
		if len(cmdSplit) == 1 {
			cmd = exec.Command(cmdSplit[0])
		} else {
			cmd = exec.Command(cmdSplit[0], cmdSplit[1:]...)
		}

		stdOut, err := cmd.StdoutPipe()
		if err != nil {
			return r, err
		}

		stdIn, err := cmd.StdinPipe()
		if err != nil {
			return r, err
		}

		if err = cmd.Start(); err != nil {
			return r, err
		}
		if _, err := io.Copy(stdIn, r); err != nil {
			return nil, err
		}
		stdIn.Close()

		b := new(bytes.Buffer)
		if _, err = io.Copy(b, stdOut); err != nil {
			return nil, err
		}
		stdOut.Close()

		return b, cmd.Wait()
	}
	return nil
}

func (m Minify) Filter(mime string, r io.Reader) (io.Reader, error) {
	if f, ok := m.Minifier[mime]; ok {
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
