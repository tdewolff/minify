package minify

import (
	"bytes"
	"io"
	"unicode"
)

// Default is a default minifier used to minify an unknown media type.
// It remove all whitespace at the beginning and end of the strea.
func (m Minifier) Default(w io.Writer, r io.Reader) error {
	var err error
	head := true
	buffer := make([]byte, 1024)
	for err == nil {
		var n int
		n, err = r.Read(buffer)
		if err != nil && err != io.EOF {
			return err
		}

		b := buffer[:n]
		if head {
			b = bytes.TrimLeftFunc(b, unicode.IsSpace)
			head = false
		}
		if n < len(buffer) {
			b = bytes.TrimRightFunc(b, unicode.IsSpace)
		}
		if _, errWrite := w.Write(b); errWrite != nil {
			return errWrite
		}
	}
	return nil
}