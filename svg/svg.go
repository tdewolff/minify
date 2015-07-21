// Package svg minifies SVG1.1 following the specifications at http://www.w3.org/TR/SVG11/.
package svg // import "github.com/tdewolff/minify/svg"

import (
	"fmt"
	"io"

	"github.com/tdewolff/buffer"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/parse"
	"github.com/tdewolff/parse/svg"
	"github.com/tdewolff/parse/xml"
)

var (
	ltBytes         = []byte("<")
	gtBytes         = []byte(">")
	voidBytes       = []byte("/>")
	isBytes         = []byte("=")
	spaceBytes      = []byte(" ")
	emptyBytes      = []byte("\"\"")
	endBytes        = []byte("</")
	cdataStartBytes = []byte("<![CDATA[")
	cdataEndBytes   = []byte("]]>")
)

////////////////////////////////////////////////////////////////

// Minify minifies SVG data, it reads from r and writes to w.
func Minify(m minify.Minifier, _ string, w io.Writer, r io.Reader) error {
	var tag svg.Hash

	attrMinifyBuffer := buffer.NewWriter(make([]byte, 0, 64))
	attrByteBuffer := make([]byte, 0, 64)

	l := xml.NewLexer(r)
	tb := xml.NewTokenBuffer(l)
	for {
		t := *tb.Shift()
		if t.TokenType == xml.CDATAToken {
			var useCDATA bool
			if t.Data, useCDATA = xml.EscapeCDATAVal(&attrByteBuffer, t.Data); !useCDATA {
				t.TokenType = xml.TextToken
			}
		}
		fmt.Println(t)
		switch t.TokenType {
		case xml.ErrorToken:
			if l.Err() == io.EOF {
				return nil
			}
			return l.Err()
		case xml.TextToken:
			t.Data = parse.ReplaceMultiple(parse.Trim(t.Data, parse.IsWhitespace), parse.IsWhitespace, ' ')
			if tag == svg.Style && len(t.Data) > 0 {
				if err := m.Minify("text/css", w, buffer.NewReader(t.Data)); err != nil {
					if err == minify.ErrNotExist { // no minifier, write the original
						if _, err := w.Write(t.Data); err != nil {
							return err
						}
					} else {
						return err
					}
				}
			} else if _, err := w.Write(t.Data); err != nil {
				return err
			}
		case xml.CDATAToken:
			if _, err := w.Write(cdataStartBytes); err != nil {
				return err
			}
			t.Data = parse.ReplaceMultiple(parse.Trim(t.Data, parse.IsWhitespace), parse.IsWhitespace, ' ')
			if tag == svg.Style && len(t.Data) > 0 {
				if err := m.Minify("text/css", w, buffer.NewReader(t.Data)); err != nil {
					if err == minify.ErrNotExist { // no minifier, write the original
						if _, err := w.Write(t.Data); err != nil {
							return err
						}
					} else {
						return err
					}
				}
			} else if _, err := w.Write(t.Data); err != nil {
				return err
			}
			if _, err := w.Write(cdataEndBytes); err != nil {
				return err
			}
		case xml.StartTagToken:
			if _, err := w.Write(ltBytes); err != nil {
				return err
			}
			if _, err := w.Write(t.Data); err != nil {
				return err
			}
			tag = svg.ToHash(t.Data)
		case xml.StartTagPIToken:
			for {
				if t := *tb.Shift(); t.TokenType == xml.StartTagClosePIToken {
					break
				}
			}
		case xml.AttributeToken:
			if len(t.AttrVal) < 2 {
				continue
			}
			attr := svg.ToHash(t.Data)
			val := parse.ReplaceMultiple(parse.Trim(t.AttrVal[1:len(t.AttrVal)-1], parse.IsWhitespace), parse.IsWhitespace, ' ')
			if tag == svg.Svg && attr == svg.Version {
				continue
			}

			if _, err := w.Write(spaceBytes); err != nil {
				return err
			}
			if _, err := w.Write(t.Data); err != nil {
				return err
			}
			if _, err := w.Write(isBytes); err != nil {
				return err
			}

			if attr == svg.Style {
				attrMinifyBuffer.Reset()
				if m.Minify("text/css;inline=1", attrMinifyBuffer, buffer.NewReader(val)) == nil {
					val = attrMinifyBuffer.Bytes()
				}
			} else if tag == svg.Path && attr == svg.D {
				val = shortenPathData(val)
			} else if n := parse.Number(val); n > 0 {
				dim := val[n:]
				val = minify.Number(val[:n])
				if len(val) != 1 || val[0] != '0' {
					if len(dim) > 1 { // only percentage is length 1
						parse.ToLower(dim)
					}
					val = append(val, dim...)
				}
			}

			// prefer single or double quotes depending on what occurs more often in value
			val = xml.EscapeAttrVal(&attrByteBuffer, val)
			if _, err := w.Write(val); err != nil {
				return err
			}
		case xml.StartTagCloseToken:
			next := tb.Peek(0)
			skipExtra := false
			if next.TokenType == xml.TextToken && parse.IsAllWhitespace(next.Data) {
				next = tb.Peek(1)
				skipExtra = true
			}
			if next.TokenType == xml.EndTagToken {
				// collapse empty tags to single void tag
				tb.Shift()
				if skipExtra {
					tb.Shift()
				}
				if _, err := w.Write(voidBytes); err != nil {
					return err
				}
			} else {
				if _, err := w.Write(gtBytes); err != nil {
					return err
				}
			}
		case xml.StartTagCloseVoidToken:
			if _, err := w.Write(voidBytes); err != nil {
				return err
			}
		case xml.EndTagToken:
			if _, err := w.Write(endBytes); err != nil {
				return err
			}
			if _, err := w.Write(t.Data); err != nil {
				return err
			}
			if _, err := w.Write(gtBytes); err != nil {
				return err
			}
		}
	}
}

func shortenPathData(b []byte) []byte {
	cmd := byte(0)
	prevDigit := false
	prevDigitRequiresSpace := true
	j := 0
	start := 0
	for i := 0; i < len(b); i++ {
		c := b[i]
		if c == ' ' || c == ',' || c == '\t' || c == '\n' || c == '\r' {
			if start != 0 {
				j += copy(b[j:], b[start:i])
			} else {
				j += i
			}
			start = i + 1
		} else if n := parse.Number(b[i:]); n > 0 {
			if start != 0 {
				j += copy(b[j:], b[start:i])
			} else {
				j += i
			}
			num := minify.Number(b[i : i+n])
			if prevDigit && (num[0] >= '0' && num[0] <= '9' || num[0] == '.' && prevDigitRequiresSpace) {
				b[j] = ' '
				j++
			}
			prevDigit = true
			prevDigitRequiresSpace = true
			for _, c := range num {
				if c == '.' || c == 'e' || c == 'E' {
					prevDigitRequiresSpace = false
					break
				}
			}
			j += copy(b[j:], num)
			start = i + n
			i += n - 1
		} else {
			if cmd == c {
				if start != 0 {
					j += copy(b[j:], b[start:i])
				} else {
					j += i
				}
				start = i + 1
			} else {
				cmd = c
				prevDigit = false
			}
		}
	}
	if start != 0 {
		j += copy(b[j:], b[start:])
		return b[:j]
	}
	return b
}
