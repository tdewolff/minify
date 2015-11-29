// Package svg minifies SVG1.1 following the specifications at http://www.w3.org/TR/SVG11/.
package svg // import "github.com/tdewolff/minify/svg"

import (
	"io"

	"github.com/tdewolff/buffer"
	"github.com/tdewolff/minify"
	minifyCSS "github.com/tdewolff/minify/css"
	"github.com/tdewolff/parse"
	"github.com/tdewolff/parse/css"
	"github.com/tdewolff/parse/svg"
	"github.com/tdewolff/parse/xml"
)

var (
	voidBytes       = []byte("/>")
	isBytes         = []byte("=")
	spaceBytes      = []byte(" ")
	cdataStartBytes = []byte("<![CDATA[")
	cdataEndBytes   = []byte("]]>")
)

////////////////////////////////////////////////////////////////

// Minifier is an SVG minifier.
type Minifier struct{}

// Minify minifies SVG data, it reads from r and writes to w.
func Minify(m *minify.M, w io.Writer, r io.Reader, params map[string]string) error {
	return (&Minifier{}).Minify(m, w, r, params)
}

// Minify minifies SVG data, it reads from r and writes to w.
func (o *Minifier) Minify(m *minify.M, w io.Writer, r io.Reader, _ map[string]string) error {
	var tag svg.Hash
	defaultStyleType := "text/css"
	defaultInlineStyleType := "text/css;inline=1"

	attrMinifyBuffer := buffer.NewWriter(make([]byte, 0, 64))
	attrByteBuffer := make([]byte, 0, 64)

	l := xml.NewLexer(r)
	tb := NewTokenBuffer(l)
	for {
		t := *tb.Shift()
		if t.TokenType == xml.CDATAToken {
			var useText bool
			if t.Data, useText = xml.EscapeCDATAVal(&attrByteBuffer, t.Data); useText {
				t.TokenType = xml.TextToken
			}
		}
	SWITCH:
		switch t.TokenType {
		case xml.ErrorToken:
			if l.Err() == io.EOF {
				return nil
			}
			return l.Err()
		case xml.TextToken:
			t.Data = parse.ReplaceMultipleWhitespace(parse.TrimWhitespace(t.Data))
			if tag == svg.Style && len(t.Data) > 0 {
				if err := m.Minify(defaultStyleType, w, buffer.NewReader(t.Data)); err != nil {
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
			if tag == svg.Style {
				if _, err := w.Write(cdataStartBytes); err != nil {
					return err
				}
				if err := m.Minify(defaultStyleType, w, buffer.NewReader(t.Text)); err != nil {
					if err == minify.ErrNotExist { // no minifier, write the original
						if _, err := w.Write(t.Text); err != nil {
							return err
						}
					} else {
						return err
					}
				}
				if _, err := w.Write(cdataEndBytes); err != nil {
					return err
				}
			} else if _, err := w.Write(t.Data); err != nil {
				return err
			}
		case xml.StartTagPIToken:
			for {
				if t := *tb.Shift(); t.TokenType == xml.StartTagClosePIToken || t.TokenType == xml.ErrorToken {
					break
				}
			}
		case xml.StartTagToken:
			tag = t.Hash
			if containerTagMap[t.Hash] { // skip empty containers
				i := 0
				for {
					next := tb.Peek(i)
					i++
					if next.TokenType == xml.EndTagToken && next.Hash == t.Hash || next.TokenType == xml.StartTagCloseVoidToken || next.TokenType == xml.ErrorToken {
						for j := 0; j < i; j++ {
							tb.Shift()
						}
						break SWITCH
					} else if next.TokenType != xml.AttributeToken && next.TokenType != xml.StartTagCloseToken {
						break
					}
				}
			} else if t.Hash == svg.Metadata {
				for {
					if next := *tb.Shift(); (next.TokenType == xml.EndTagToken || next.TokenType == xml.StartTagCloseVoidToken) && next.Hash == t.Hash || next.TokenType == xml.ErrorToken {
						break
					}
				}
				break
			} else if t.Hash == svg.Line || t.Hash == svg.Rect {
				// TODO: shape2path also for polygon and polyline
				// x1, y1, x2, y2 float64 := 0, 0, 0, 0
				// valid := true
				// i := 0
				// for {
				// 	next := tb.Peek(i)
				// 	i++
				// 	if next.TokenType != xml.AttributeToken {
				// 		break
				// 	}
				// 	v *int
				// 	attr := svg.ToHash(next.Data)
				// 	if tag == svg.Line {
				// 		if attr == svg.X1 {
				// 			v = &x1
				// 		} else if attr == svg.Y1 {
				// 			v = &y1
				// 		} else if attr == svg.X2 {
				// 			v = &x2
				// 		} else if attr == svg.Y2 {
				// 			v = &Y2
				// 		} else {
				// 			continue
				// 		}
				// 	} else if attr == svg.X { // rect
				// 		v = &x1
				// 	} else if attr == svg.Y {
				// 		v = &y1
				// 	} else if attr == svg.Width {
				// 		v = &x2
				// 	} else if attr == svg.Height {
				// 		v = &Y2
				// 	} else if attr == svg.Rx || attr == svg.Ry {
				// 		valid = false
				// 		break
				// 	} else {
				// 		continue
				// 	}

				// }
				// if valid {
				// 	t.Data = pathBytes
				// }
			}
			if _, err := w.Write(t.Data); err != nil {
				return err
			}
		case xml.AttributeToken:
			if len(t.AttrVal) < 2 {
				continue
			}
			val := parse.ReplaceMultipleWhitespace(parse.TrimWhitespace(t.AttrVal[1 : len(t.AttrVal)-1]))
			if tag == svg.Svg && t.Hash == svg.Version {
				continue
			}

			if _, err := w.Write(spaceBytes); err != nil {
				return err
			}
			if _, err := w.Write(t.Text); err != nil {
				return err
			}
			if _, err := w.Write(isBytes); err != nil {
				return err
			}

			if tag == svg.Svg && t.Hash == svg.ContentStyleType {
				val = minify.ContentType(val)
				defaultStyleType = string(val)
				defaultInlineStyleType = defaultStyleType + ";inline=1"
			} else if t.Hash == svg.Style {
				attrMinifyBuffer.Reset()
				if m.Minify(defaultInlineStyleType, attrMinifyBuffer, buffer.NewReader(val)) == nil {
					val = attrMinifyBuffer.Bytes()
				}
			} else if t.Hash == svg.D {
				val = shortenPathData(val)
			} else if t.Hash == svg.ViewBox {
				j := 0
				newVal := val[:0]
				for i := 0; i < 4; i++ {
					if i != 0 {
						if j >= len(val) || val[j] != ' ' && val[j] != ',' {
							newVal = append(newVal, val[j:]...)
							break
						}
						newVal = append(newVal, ' ')
						j++
					}
					if dim, n := shortenDimension(val[j:]); n > 0 {
						newVal = append(newVal, dim...)
						j += n
					} else {
						newVal = append(newVal, val[j:]...)
						break
					}
				}
				val = newVal
			} else if colorAttrMap[t.Hash] && len(val) > 0 {
				parse.ToLower(val)
				if val[0] == '#' {
					if name, ok := minifyCSS.ShortenColorHex[string(val)]; ok {
						val = name
					} else if len(val) == 7 && val[1] == val[2] && val[3] == val[4] && val[5] == val[6] {
						val[2] = val[3]
						val[3] = val[5]
						val = val[:4]
					}
				} else if hex, ok := minifyCSS.ShortenColorName[css.ToHash(val)]; ok {
					val = hex
				} else if len(val) > 5 && parse.Equal(val[:4], []byte("rgb(")) && val[len(val)-1] == ')' {
					// TODO: handle rgb(x, y, z) and hsl(x, y, z)
				}
			} else if n, m := parse.Dimension(val); n+m == len(val) { // TODO: inefficient, temporary measure
				val, _ = shortenDimension(val)
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
				if _, err := w.Write(t.Data); err != nil {
					return err
				}
			}
		case xml.StartTagCloseVoidToken:
			if _, err := w.Write(t.Data); err != nil {
				return err
			}
		case xml.EndTagToken:
			t.Data[2+len(t.Text)] = '>'
			if _, err := w.Write(t.Data[:2+len(t.Text)+1]); err != nil {
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
			for i := 0; i < len(num); i++ {
				c := num[i]
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

func shortenDimension(b []byte) ([]byte, int) {
	if n, m := parse.Dimension(b); n > 0 {
		unit := b[n : n+m]
		b = minify.Number(b[:n])
		if len(b) != 1 || b[0] != '0' {
			if m == 2 && unit[0] == 'p' && unit[1] == 'x' {
				unit = nil
			} else if m > 1 { // only percentage is length 1
				parse.ToLower(unit)
			}
			b = append(b, unit...)
		}
		return b, n + m
	}
	return b, 0
}
