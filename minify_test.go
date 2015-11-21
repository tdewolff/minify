package minify // import "github.com/tdewolff/minify"

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
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

func helperMinifyString(t *testing.T, m *M, mediatype string) string {
	s, err := m.String(mediatype, "")
	assert.Nil(t, err, "minifier must not return error for '"+mediatype+"'")
	return s
}

////////////////////////////////////////////////////////////////

var m *M

func init() {
	m = New()
	m.AddFunc("dummy/copy", func(m *M, w io.Writer, r io.Reader, _ map[string]string) error {
		io.Copy(w, r)
		return nil
	})
	m.AddFunc("dummy/nil", func(m *M, w io.Writer, r io.Reader, _ map[string]string) error {
		return nil
	})
	m.AddFunc("dummy/err", func(m *M, w io.Writer, r io.Reader, _ map[string]string) error {
		return errDummy
	})
	m.AddFunc("dummy/charset", func(m *M, w io.Writer, r io.Reader, params map[string]string) error {
		w.Write([]byte(params["charset"]))
		return nil
	})
	m.AddFunc("dummy/params", func(m *M, w io.Writer, r io.Reader, params map[string]string) error {
		return m.Minify(params["type"]+"/"+params["sub"], w, r)
	})
	m.AddFunc("type/sub", func(m *M, w io.Writer, r io.Reader, _ map[string]string) error {
		w.Write([]byte("type/sub"))
		return nil
	})
	m.AddFuncRegexp(regexp.MustCompile("^type/.+$"), func(m *M, w io.Writer, r io.Reader, _ map[string]string) error {
		w.Write([]byte("type/*"))
		return nil
	})
	m.AddFuncRegexp(regexp.MustCompile("^.+/.+$"), func(m *M, w io.Writer, r io.Reader, _ map[string]string) error {
		w.Write([]byte("*/*"))
		return nil
	})
}

func TestMinify(t *testing.T) {
	assert.Equal(t, ErrNotExist, m.Minify("?", nil, nil), "must return ErrNotExist when minifier doesn't exist")
	assert.Nil(t, m.Minify("dummy/nil", nil, nil), "must return nil for dummy/nil")
	assert.Equal(t, errDummy, m.Minify("dummy/err", nil, nil), "must return errDummy for dummy/err")

	b := []byte("test")
	out, err := m.Bytes("dummy/nil", b)
	assert.Nil(t, err, "must not return error for dummy/nil")
	assert.Equal(t, []byte{}, out, "must return empty byte array for dummy/nil")
	out, err = m.Bytes("?", b)
	assert.Equal(t, ErrNotExist, err, "must return ErrNotExist when minifier doesn't exist")
	assert.Equal(t, b, out, "must return input byte array when minifier doesn't exist")

	s := "test"
	out2, err := m.String("dummy/nil", s)
	assert.Nil(t, err, "must not return error for dummy/nil")
	assert.Equal(t, "", out2, "must return empty string for dummy/nil")
	out2, err = m.String("?", s)
	assert.Equal(t, ErrNotExist, err, "must return ErrNotExist when minifier doesn't exist")
	assert.Equal(t, s, out2, "must return input string when minifier doesn't exist")
}

func TestAdd(t *testing.T) {
	m := New()
	w := &bytes.Buffer{}
	r := bytes.NewBufferString("test")
	m.AddFunc("dummy/err", func(m *M, w io.Writer, r io.Reader, _ map[string]string) error {
		return errDummy
	})
	assert.Equal(t, errDummy, m.Minify("dummy/err", nil, nil), "must return errDummy for dummy/err")

	m.AddCmd("dummy/copy", helperCommand(t, "dummy/copy"))
	m.AddCmd("dummy/err", helperCommand(t, "dummy/err"))
	m.AddCmdRegexp(regexp.MustCompile("err$"), helperCommand(t, "werr"))
	assert.Nil(t, m.Minify("dummy/copy", w, r), "must return nil for dummy/copy command")
	assert.Equal(t, "test", w.String(), "must return input string for dummy/copy command")
	assert.Equal(t, "exit status 1", m.Minify("dummy/err", w, r).Error(), "must return proper exit status when command encounters error")
	assert.Equal(t, "exit status 2", m.Minify("werr", w, r).Error(), "must return proper exit status when command encounters error")
	assert.Equal(t, "exit status 2", m.Minify("stderr", w, r).Error(), "must return proper exit status when command encounters error")
}

func TestWildcard(t *testing.T) {
	assert.Equal(t, "type/sub", helperMinifyString(t, m, "type/sub"), "must return type/sub for type/sub")
	assert.Equal(t, "type/*", helperMinifyString(t, m, "type/*"), "must return type/* for type/*")
	assert.Equal(t, "*/*", helperMinifyString(t, m, "*/*"), "must return */* for */*")
	assert.Equal(t, "type/*", helperMinifyString(t, m, "type/sub2"), "must return type/* for type/sub2")
	assert.Equal(t, "*/*", helperMinifyString(t, m, "type2/sub"), "must return */* for type2/sub")
	assert.Equal(t, "UTF-8", helperMinifyString(t, m, "dummy/charset;charset=UTF-8"), "must return UTF-8 for dummy/charset;charset=UTF-8")
	assert.Equal(t, "UTF-8", helperMinifyString(t, m, "dummy/charset; charset = UTF-8 "), "must return UTF-8 for ' dummy/charset; charset = UTF-8 '")
	assert.Equal(t, "type/sub", helperMinifyString(t, m, "dummy/params;type=type;sub=sub"), "must return type/sub for dummy/params;type=type;sub=sub")
}

func TestHelperProcess(*testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
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
	default:
		os.Exit(2)
	}
	os.Exit(0)
}

////////////////////////////////////////////////////////////////

func ExampleM_Minify_custom() {
	m := New()
	m.AddFunc("text/plain", func(m *M, w io.Writer, r io.Reader, _ map[string]string) error {
		// remove all newlines and spaces
		rb := bufio.NewReader(r)
		for {
			line, err := rb.ReadString('\n')
			if err != nil && err != io.EOF {
				return err
			}
			if _, errws := io.WriteString(w, strings.Replace(line, " ", "", -1)); errws != nil {
				return errws
			}
			if err == io.EOF {
				break
			}
		}
		return nil
	})

	in := "Because my coffee was too cold, I heated it in the microwave."
	out, err := m.String("text/plain", in)
	if err != nil {
		panic(err)
	}
	fmt.Println(out)
	// Output: Becausemycoffeewastoocold,Iheateditinthemicrowave.
}

func ExampleM_Reader() {
	b := bytes.NewReader([]byte("input"))

	m := New()
	// add minfiers

	r := m.Reader("mime/type", b)
	if _, err := io.Copy(os.Stdout, r); err != nil {
		if _, err := io.Copy(os.Stdout, b); err != nil {
			panic(err)
		}
	}
}

func ExampleM_Writer() {
	m := New()
	// add minfiers

	w := m.Writer("mime/type", os.Stdout)
	if _, err := w.Write([]byte("input")); err != nil {
		panic(err)
	}
	if err := w.Close(); err != nil {
		panic(err)
	}
}

type MinifierResponseWriter struct {
	http.ResponseWriter
	io.Writer
}

func (m MinifierResponseWriter) Write(b []byte) (int, error) {
	return m.Writer.Write(b)
}

func ExampleM_Minify_responseWriter(res http.ResponseWriter) http.ResponseWriter {
	m := New()
	// add minfiers

	pr, pw := io.Pipe()
	go func(w io.Writer) {
		if err := m.Minify("mime/type", w, pr); err != nil {
			panic(err)
		}
	}(res)
	return MinifierResponseWriter{res, pw}
}
