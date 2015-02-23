package minify // import "github.com/tdewolff/minify"

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"testing"
)

var errDummy = errors.New("dummy error")

// from os/exec/exec_test.go
func helperCommand(t *testing.T, s ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--"}
	cs = append(cs, s...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

////////////////////////////////////////////////////////////////

var m = &DefaultMinifier{map[string]Func{
	"copy": func(m Minifier, w io.Writer, r io.Reader) error {
		io.Copy(w, r)
		return nil
	},
	"nil": func(m Minifier, w io.Writer, r io.Reader) error {
		return nil
	},
	"err": func(m Minifier, w io.Writer, r io.Reader) error {
		return errDummy
	},
}, map[string]string{}}

func TestMinify(t *testing.T) {
	if err := m.Minify("?", nil, nil); err != ErrNotExist {
		t.Error(err, "!=", ErrNotExist)
	}
	if err := m.Minify("nil", nil, nil); err != nil {
		t.Error(err, "!=", nil)
	}
	if err := m.Minify("err", nil, nil); err != errDummy {
		t.Error(err, "!=", errDummy)
	}

	b := []byte("test")
	if out, err := m.MinifyBytes("nil", b); err != nil || !bytes.Equal(out, []byte{}) {
		t.Error(err, "!=", nil, "||", out, "!=", []byte{})
	}
	if out, err := m.MinifyBytes("?", b); err != ErrNotExist || !bytes.Equal(out, b) {
		t.Error(err, "!=", ErrNotExist, "||", out, "!=", b)
	}

	s := "test"
	if out, err := m.MinifyString("nil", s); err != nil || out != "" {
		t.Error(err, "!=", nil, "||", out, "!=", "")
	}
	if out, err := m.MinifyString("?", s); err != ErrNotExist || out != s {
		t.Error(err, "!=", ErrNotExist, "||", out, "!=", s)
	}
}

// func TestDefaultMinifiers(t *testing.T) {
// 	m := NewMinifier()
// 	w := &bytes.Buffer{}

// 	r := bytes.NewBufferString("html")
// 	if err := m.Minify("text/html", w, r); err != nil {
// 		t.Error(err)
// 	}
// 	r = bytes.NewBufferString("prop:val;")
// 	if err := m.Minify("text/css", w, r); err != nil {
// 		t.Error(err)
// 	}

// 	if w.String() != "htmlprop:val" {
// 		t.Error(w.String(), "!=", "htmlprop:val")
// 	}
// }

func TestAdd(t *testing.T) {
	m := NewMinifier()
	w := &bytes.Buffer{}
	r := bytes.NewBufferString("test")

	m.Add("err", func(m Minifier, w io.Writer, r io.Reader) error {
		return errDummy
	})
	if err := m.Minify("err", nil, nil); err != errDummy {
		t.Error(err, "!=", errDummy)
	}

	if err := m.AddCmd("copy", helperCommand(t, "copy")); err != nil {
		t.Error(err)
	}
	if err := m.Minify("copy", w, r); err != nil {
		t.Error(err)
	}
	if w.String() != "test" {
		t.Error(w.String(), "!= test")
	}

	if err := m.AddCmd("err", helperCommand(t, "err")); err != nil {
		t.Error(err)
	}
	if err := m.Minify("err", w, r); err.Error() != "exit status 1" {
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
