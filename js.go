package minify

import (
	"io"
	"io/ioutil"
	"os"
	"bytes"
	"os/exec"
)

func (minify Minify) Js(r io.ReadCloser) (io.ReadCloser, error) {
	if _, err := exec.LookPath("node"); err != nil { return r, err }
	if _, err := os.Stat(minify.UglifyjsPath); err != nil { return r, err }

	cmd := exec.Command("node", minify.UglifyjsPath)
	//stdErr, err := cmd.StderrPipe()
	//if err != nil { return r, err }

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

	//if _, err = io.Copy(os.Stderr, stdErr); err != nil { return nil, err }
	//stdErr.Close()

	if buffer.Len() > 0 {
		buffer.Truncate(buffer.Len() - 1)
	}
	return ioutil.NopCloser(buffer), cmd.Wait()
}