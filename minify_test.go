package minify // import "github.com/tdewolff/minify"

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
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

func helperMinifyString(t *testing.T, m *DefaultMinifier, mediatype string) string {
	s, err := m.MinifyString(mediatype, "")
	assert.Nil(t, err, "minifier must not return error")
	return s
}

////////////////////////////////////////////////////////////////

var m = &DefaultMinifier{map[string]Func{
	"dummy/copy": func(m Minifier, w io.Writer, r io.Reader) error {
		io.Copy(w, r)
		return nil
	},
	"dummy/nil": func(m Minifier, w io.Writer, r io.Reader) error {
		return nil
	},
	"dummy/err": func(m Minifier, w io.Writer, r io.Reader) error {
		return errDummy
	},
	"dummy/param": func(m Minifier, w io.Writer, r io.Reader) error {
		if cs := m.Param("charset"); cs != "" {
			w.Write([]byte(cs))
		} else {
			w.Write([]byte(m.Param("mediatype")))
		}
		return nil
	},
	"dummy/param2": func(m Minifier, w io.Writer, r io.Reader) error {
		return m.Minify(m.Param("type")+"/"+m.Param("sub"), w, r)
	},
	"type/sub": func(m Minifier, w io.Writer, r io.Reader) error {
		w.Write([]byte("type/sub"))
		return nil
	},
	"type/*": func(m Minifier, w io.Writer, r io.Reader) error {
		w.Write([]byte("type/*"))
		return nil
	},
	"*/*": func(m Minifier, w io.Writer, r io.Reader) error {
		w.Write([]byte("*/*"))
		return nil
	},
}, map[string]string{}}

func TestMinify(t *testing.T) {
	assert.Equal(t, ErrNotExist, m.Minify("?", nil, nil), "must return ErrNotExist when minifier doesn't exist")
	assert.Nil(t, m.Minify("dummy/nil", nil, nil), "must return nil for dummy/nil")
	assert.Equal(t, errDummy, m.Minify("dummy/err", nil, nil), "must return errDummy for dummy/err")

	b := []byte("test")
	out, err := m.MinifyBytes("dummy/nil", b)
	assert.Nil(t, err, "must not return error for dummy/nil")
	assert.Equal(t, []byte{}, out, "must return empty byte array for dummy/nil")
	out, err = m.MinifyBytes("?", b)
	assert.Equal(t, ErrNotExist, err, "must return ErrNotExist when minifier doesn't exist")
	assert.Equal(t, b, out, "must return input byte array when minifier doesn't exist")

	s := "test"
	out2, err := m.MinifyString("dummy/nil", s)
	assert.Nil(t, err, "must not return error for dummy/nil")
	assert.Equal(t, "", out2, "must return empty string for dummy/nil")
	out2, err = m.MinifyString("?", s)
	assert.Equal(t, ErrNotExist, err, "must return ErrNotExist when minifier doesn't exist")
	assert.Equal(t, s, out2, "must return input string when minifier doesn't exist")
}

func TestAdd(t *testing.T) {
	m := NewMinifier()
	w := &bytes.Buffer{}
	r := bytes.NewBufferString("test")
	m.Add("dummy/err", func(m Minifier, w io.Writer, r io.Reader) error {
		return errDummy
	})

	assert.Equal(t, errDummy, m.Minify("dummy/err", nil, nil), "must return errDummy for dummy/err")
	assert.Nil(t, m.AddCmd("dummy/copy", helperCommand(t, "dummy/copy")), "must return nil when adding command")
	assert.Nil(t, m.Minify("dummy/copy", w, r), "must return nil for dummy/copy command")
	assert.Equal(t, "test", w.String(), "must return input string for dummy/copy command")
	assert.Nil(t, m.AddCmd("dummy/err", helperCommand(t, "dummy/err")), "must return nil for dummy/err command")
	assert.Equal(t, "exit status 1", m.Minify("dummy/err", w, r).Error(), "must return proper exit status when command encounters error")
}

func TestWildcard(t *testing.T) {
	assert.Equal(t, "type/sub", helperMinifyString(t, m, "type/sub"), "must return type/sub for type/sub")
	assert.Equal(t, "type/*", helperMinifyString(t, m, "type/*"), "must return type/* for type/*")
	assert.Equal(t, "*/*", helperMinifyString(t, m, "*/*"), "must return */* for */*")
	assert.Equal(t, "type/*", helperMinifyString(t, m, "type/sub2"), "must return type/* for type/sub2")
	assert.Equal(t, "*/*", helperMinifyString(t, m, "type2/sub"), "must return */* for type2/sub")
	assert.Equal(t, "UTF-8", helperMinifyString(t, m, "dummy/param;charset=UTF-8"), "must return dummy/param with charset=UTF-8 for dummy/param;charset=UTF-8")
	assert.Equal(t, "UTF-8", helperMinifyString(t, m, " dummy/param ; charset = UTF-8 "), "must return dummy/param with charset=UTF-8 for ' dummy/param ; charset = UTF-8 '")
	assert.Equal(t, "dummy/param", helperMinifyString(t, m, "dummy/param2;type=dummy;sub=param"), "must return dummy/param inside dummy/param2 for dummy/param2;type=dummy;sub=param")
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
	case "dummy/copy":
		io.Copy(os.Stdout, os.Stdin)
	case "dummy/err":
		os.Exit(1)
	}
	os.Exit(0)
}
