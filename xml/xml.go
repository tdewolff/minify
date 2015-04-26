package xml // import "github.com/tdewolff/minify/xml"

import (
	"io"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/parse"
	"github.com/tdewolff/parse/xml"
)

var (
	ltBytes                 = []byte("<")
	gtBytes                 = []byte(">")
	voidBytes               = []byte("/>")
	piBytes                 = []byte("?>")
	isBytes                 = []byte("=")
	spaceBytes              = []byte(" ")
	emptyBytes              = []byte("\"\"")
	endBytes                = []byte("</")
	escapedSingleQuoteBytes = []byte("&#39;")
	escapedDoubleQuoteBytes = []byte("&#34;")
)

type token struct {
	tt      xml.TokenType
	data    []byte
	attrVal []byte
}

////////////////////////////////////////////////////////////////

// Minify minifies XML files, it reads from r and writes to w.
// Removes unnecessary whitespace, tags, attributes, quotes and comments and typically saves 10% in size.
func Minify(m minify.Minifier, _ string, w io.Writer, r io.Reader) error {
	precededBySpace := true // on true the next text token must not start with a space
	attrEscapeBuffer := make([]byte, 0, 64)

	z := xml.NewTokenizer(r)
	tb := newTokenBuffer(z)
	for {
		t := *tb.Shift()
		switch t.tt {
		case xml.ErrorToken:
			if z.Err() == io.EOF {
				return nil
			}
			return z.Err()
		case xml.TextToken:
			if t.data = replaceMultipleWhitespace(t.data); len(t.data) > 0 {
				// whitespace removal; trim left
				if t.data[0] == ' ' && precededBySpace {
					t.data = t.data[1:]
				}

				// whitespace removal; trim right
				precededBySpace = false
				if len(t.data) == 0 {
					precededBySpace = true
				} else if t.data[len(t.data)-1] == ' ' {
					precededBySpace = true
					trim := false
					i := 0
					for {
						next := tb.Peek(i)
						// trim if EOF, text token with whitespace begin or block token
						if next.tt == xml.StartTagToken || next.tt == xml.EndTagToken || next.tt == xml.ErrorToken {
							trim = true
							break
						} else if next.tt == xml.TextToken {
							// remove if the text token starts with a whitespace
							trim = (len(next.data) > 0 && isWhitespace(next.data[0]))
							break
						}
						i++
					}
					if trim {
						t.data = t.data[:len(t.data)-1]
						precededBySpace = false
					}
				}
				if _, err := w.Write(t.data); err != nil {
					return err
				}
			}
		case xml.StartTagToken:
			if _, err := w.Write(ltBytes); err != nil {
				return err
			}
			if _, err := w.Write(t.data); err != nil {
				return err
			}
			// collapse empty tags to single void tag
			if next := tb.Peek(0); next.tt == xml.StartTagCloseToken {
				i := 1
				for {
					next = tb.Peek(i)
					i++
					// continue if text token is empty or whitespace
					if next.tt == xml.TextToken && isAllWhitespace(next.data) {
						continue
					} else if next.tt != xml.EndTagToken {
						break
					}
					tb.Shift()
					tb.Shift()
					if _, err := w.Write(voidBytes); err != nil {
						return err
					}
					break
				}
			}
		case xml.AttributeToken:
			val := t.attrVal
			if len(val) < 2 {
				if _, err := w.Write(emptyBytes); err != nil {
					return err
				}
				continue
			}

			if _, err := w.Write(spaceBytes); err != nil {
				return err
			}
			if _, err := w.Write(t.data); err != nil {
				return err
			}
			if _, err := w.Write(isBytes); err != nil {
				return err
			}

			// prefer single or double quotes depending on what occurs more often in value
			val = val[1 : len(val)-1]
			val = escapeAttrVal(&attrEscapeBuffer, val)
			if _, err := w.Write(val); err != nil {
				return err
			}
		case xml.StartTagCloseToken:
			if _, err := w.Write(gtBytes); err != nil {
				return err
			}
		case xml.StartTagCloseVoidToken:
			if _, err := w.Write(voidBytes); err != nil {
				return err
			}
		case xml.StartTagClosePIToken:
			if _, err := w.Write(piBytes); err != nil {
				return err
			}
		case xml.EndTagToken:
			if _, err := w.Write(endBytes); err != nil {
				return err
			}
			if _, err := w.Write(t.data); err != nil {
				return err
			}
			if _, err := w.Write(gtBytes); err != nil {
				return err
			}
		}
	}
}

////////////////////////////////////////////////////////////////

// replaceMultipleWhitespace replaces any series of whitespace characters by a single space.
func replaceMultipleWhitespace(b []byte) []byte {
	j := 0
	start := 0
	prevSpace := false
	for i, c := range b {
		if isWhitespace(c) {
			if !prevSpace {
				prevSpace = true
				b[i] = ' '
			} else {
				if start < i {
					if start != 0 {
						j += copy(b[j:], b[start:i])
					} else {
						j += i - start
					}
				}
				start = i + 1
			}
		} else {
			prevSpace = false
		}
	}
	if start != 0 {
		j += copy(b[j:], b[start:])
		return b[:j]
	}
	return b
}

// it is assumed that b[0] equals '&'
func isAtQuoteEntity(b []byte) (quote byte, n int, ok bool) {
	if len(b) < 5 {
		return 0, 0, false
	}
	if b[1] == '#' {
		if b[2] == 'x' {
			i := 3
			for i < len(b) && b[i] == '0' {
				i++
			}
			if i+2 < len(b) && b[i] == '2' && b[i+2] == ';' {
				if b[i+1] == '2' {
					return '"', i + 3, true // &#x22;
				} else if b[i+1] == '7' {
					return '\'', i + 3, true // &#x27;
				}
			}
		} else {
			i := 2
			for i < len(b) && b[i] == '0' {
				i++
			}
			if i+2 < len(b) && b[i] == '3' && b[i+2] == ';' {
				if b[i+1] == '4' {
					return '"', i + 3, true // &#34;
				} else if b[i+1] == '9' {
					return '\'', i + 3, true // &#39;
				}
			}
		}
	} else if len(b) >= 6 && b[5] == ';' {
		if parse.EqualCaseInsensitive(b[1:5], []byte{'q', 'u', 'o', 't'}) {
			return '"', 6, true // &quot;
		} else if parse.EqualCaseInsensitive(b[1:5], []byte{'a', 'p', 'o', 's'}) {
			return '\'', 6, true // &apos;
		}
	}
	return 0, 0, false
}

// escapeAttrVal returns the escape attribute value bytes without quotes.
func escapeAttrVal(buf *[]byte, b []byte) []byte {
	singles := 0
	doubles := 0
	for i, c := range b {
		if c == '&' {
			if quote, _, ok := isAtQuoteEntity(b[i:]); ok {
				if quote == '"' {
					doubles++
				} else {
					singles++
				}
			}
		} else if c == '"' {
			doubles++
		} else if c == '\'' {
			singles++
		}
	}

	var quote byte
	var escapedQuote []byte
	if doubles > singles {
		quote = '\''
		escapedQuote = escapedSingleQuoteBytes
	} else {
		quote = '"'
		escapedQuote = escapedDoubleQuoteBytes
	}

	// maximum size, not actual size
	if len(b)+2 > cap(*buf) {
		*buf = make([]byte, 0, len(b)+2)
	}

	t := (*buf)[:len(b)+2] // maximum size, not actual size
	t[0] = quote
	j := 1
	start := 0
	for i, c := range b {
		if c == '&' {
			if entityQuote, n, ok := isAtQuoteEntity(b[i:]); ok {
				j += copy(t[j:], b[start:i])
				if entityQuote != quote {
					j += copy(t[j:], []byte{entityQuote})
				} else {
					j += copy(t[j:], escapedQuote)
				}
				start = i + n
			}
		} else if c == quote {
			j += copy(t[j:], b[start:i])
			j += copy(t[j:], escapedQuote)
			start = i + 1
		} else if c == '\t' || c == '\n' || c == '\r' {
			b[i] = ' '
		}
	}
	j += copy(t[j:], b[start:])
	t[j] = quote
	return t[:j+1]
}

// isWhitespace returns true for space, \n, \t, \f, \r.
func isWhitespace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\f'
}

func isAllWhitespace(b []byte) bool {
	for _, c := range b {
		if !isWhitespace(c) {
			return false
		}
	}
	return true
}
