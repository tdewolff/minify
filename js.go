package minify

import (
	"io"
	"os"
	"bytes"
	"os/exec"
)

func Js(r io.Reader) (io.Reader, error) {
	if _, err := exec.LookPath("java"); err != nil {
		return r, err
	}

	if _, err := os.Stat("compiler.jar"); err != nil {
		return r, err
	}

	args := []string{"-jar", "compiler.jar"}
	//args = append(args, "--compilation_level", "ADVANCED_OPTIMIZATIONS")

	cmd := exec.Command("java", args...)
	stdErr, err := cmd.StderrPipe()
	if err != nil {
		return r, err
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

	if _, err = io.Copy(os.Stderr, stdErr); err != nil {
		return nil, err
	}
	stdErr.Close()

	buffer := new(bytes.Buffer)
	if _, err = io.Copy(buffer, stdOut); err != nil {
		return nil, err
	}
	stdOut.Close()

	if buffer.Len() > 0 {
		buffer.Truncate(buffer.Len() - 1)
	}
	return buffer, cmd.Wait()
}