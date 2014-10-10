package minify

import (
	"errors"
	"io"
	"io/ioutil"
	"bytes"
	"os/exec"
)

func (minify Minify) Js(r io.ReadCloser) (io.ReadCloser, error) {
	defer func() {
		r.Close()
	}()

	var cmd *exec.Cmd
	if len(minify.JsMinifier) == 0 {
		return nil, errors.New("JS minifier not set")
	} else if len(minify.JsMinifier) == 1 {
		cmd = exec.Command(minify.JsMinifier[0])
	} else {
		cmd = exec.Command(minify.JsMinifier[0], minify.JsMinifier[1:]...)
	}

	stdOut, err := cmd.StdoutPipe()
	if err != nil { return r, err }

	stdIn, err := cmd.StdinPipe()
	if err != nil { return r, err }

	if err = cmd.Start(); err != nil { return r, err }
	if _, err := io.Copy(stdIn, r); err != nil { return nil, err }
	stdIn.Close()

	buffer := new(bytes.Buffer)
	if _, err = io.Copy(buffer, stdOut); err != nil { return nil, err }
	stdOut.Close()

	return ioutil.NopCloser(buffer), cmd.Wait()
}