package trim // import "github.com/tdewolff/minify/trim"

import (
	"bytes"
	"io"
	"unicode"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/parse"
)

const bufSize = 1024

// Minify is a default minifier used to trim an unknown media type.
// It remove all whitespace at the beginning and end of the stream.
func Minify(m minify.Minifier, w io.Writer, r io.Reader) error {
	if fr, ok := r.(interface {
		Bytes() []byte
	}); ok {
		b := bytes.TrimSpace(fr.Bytes())
		if _, errWrite := w.Write(b); errWrite != nil {
			return errWrite
		}
		return nil
	}

	head := true
	sb := parse.NewShiftBuffer(r)
	for {
		// cause a read
		if sb.Peek(0) == 0 && sb.Err() != nil {
			if sb.Err() == io.EOF {
				return nil
			}
			return sb.Err()
		}

		// consume whole buffer and unconsume trailing whitespace
		sb.MoveTo(sb.Len())
		trailingLen := sb.Len() - len(bytes.TrimRightFunc(sb.Buffered(), unicode.IsSpace))
		sb.Move(-trailingLen)

		b := sb.Shift()
		if head {
			b = bytes.TrimLeftFunc(b, unicode.IsSpace)
			head = len(b) == 0 // if it's all whitespace, we still need to trim leading whitespace next time
		}
		if _, errWrite := w.Write(b); errWrite != nil {
			return errWrite
		}
		sb.Move(trailingLen)
	}
}
