package minify

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"testing"
)

// from os/exec/exec_test.go
func helperCommand(t *testing.T, s ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--"}
	cs = append(cs, s...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func TestFilter(t *testing.T) {
	errDummy := errors.New("dummy error")

	m := &Minify{map[string]MinifyFunc{
		"nil": func(m Minify, w io.Writer, r io.Reader) error {
			return nil
		},
		"err": func(m Minify, w io.Writer, r io.Reader) error {
			return errDummy
		},
	}}

	if err := m.Filter("?", nil, nil); err != ErrNotExist {
		t.Error(err, "!=", ErrNotExist)
	}
	if err := m.Filter("nil", nil, nil); err != nil {
		t.Error(err, "!= nil")
	}
	if err := m.Filter("err", nil, nil); err != errDummy {
		t.Error(err, "!=", errDummy)
	}
}

func TestDefaultFilters(t *testing.T) {
	w := &bytes.Buffer{}

	r := bytes.NewBufferString("html")
	if err := NewMinify().Filter("text/html", w, r); err != nil {
		t.Error(err)
	}
	r = bytes.NewBufferString("key:value")
	if err := NewMinify().Filter("text/css", w, r); err != nil {
		t.Error(err)
	}

	if w.String() != "htmlkey:value" {
		t.Error(w.String(), "!= htmlkey:value")
	}
}

func TestFilterBytes(t *testing.T) {
	in := []byte("<html>test")
	exp := []byte("test")
	if out := NewMinify().FilterBytes("text/html", in); !bytes.Equal(out, exp) {
		t.Error(out, "!= test")
	}
	if out := NewMinify().FilterBytes("?", in); !bytes.Equal(out, in) {
		t.Error(out, "!= <html>test")
	}
}

func TestFilterString(t *testing.T) {
	in := "<html>test"
	exp := "test"
	if out := NewMinify().FilterString("text/html", in); out != exp {
		t.Error(out, "!=", exp)
	}
	if out := NewMinify().FilterString("?", in); out != in {
		t.Error(out, "!=", in)
	}
}

func TestImplement(t *testing.T) {
	errDummy := errors.New("dummy error")

	m := NewMinify()
	m.Implement("err", func(m Minify, w io.Writer, r io.Reader) error {
		return errDummy
	})

	if err := m.Filter("err", nil, nil); err != errDummy {
		t.Error(err, "!=", errDummy)
	}
}

func TestImplementCmd(t *testing.T) {
	m := NewMinify()
	w := &bytes.Buffer{}
	r := bytes.NewBufferString("test")

	if err := m.ImplementCmd("copy", helperCommand(t, "copy")); err != nil {
		t.Error(err)
	}
	if err := m.Filter("copy", w, r); err != nil {
		t.Error(err)
	}
	if w.String() != "test" {
		t.Error(w.String(), "!= test")
	}

	if err := m.ImplementCmd("err", helperCommand(t, "err")); err != nil {
		t.Error(err)
	}
	if err := m.Filter("err", w, r); err.Error() != "exit status 1" {
		t.Error(err)
	}
}

func TestHelperProcess(*testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)

	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "No command\n")
		os.Exit(2)
	}

	cmd, args := args[0], args[1:]
	switch cmd {
	case "copy":
		io.Copy(os.Stdout, os.Stdin)
	case "err":
		os.Exit(1)
	}
	os.Exit(0)
}