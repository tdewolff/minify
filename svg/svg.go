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
	pathBytes       = []byte("path")
	dBytes          = []byte("d")
	zeroBytes       = []byte("0")
)

const maxAttrLookup = 6

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
	attrTokenBuffer := make([]*Token, 0, maxAttrLookup)
	pathDataBuffer := &PathData{}

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
			if containerTagMap[tag] { // skip empty containers
				i := 0
				for {
					next := tb.Peek(i)
					i++
					if next.TokenType == xml.EndTagToken && next.Hash == tag || next.TokenType == xml.StartTagCloseVoidToken || next.TokenType == xml.ErrorToken {
						for j := 0; j < i; j++ {
							tb.Shift()
						}
						break SWITCH
					} else if next.TokenType != xml.AttributeToken && next.TokenType != xml.StartTagCloseToken {
						break
					}
				}
			} else if tag == svg.Metadata {
				skipTag(tb, tag)
				break
			} else if tag == svg.Line {
				getAttributes(&attrTokenBuffer, tb, svg.X1, svg.Y1, svg.X2, svg.Y2)
				i := 0
				x1, y1, x2, y2 := zeroBytes, zeroBytes, zeroBytes, zeroBytes
				if attrTokenBuffer[0] != nil {
					x1 = minify.Number(attrTokenBuffer[0].AttrVal)
					attrTokenBuffer[0].Data = nil
				}
				if attrTokenBuffer[1] != nil {
					y1 = minify.Number(attrTokenBuffer[1].AttrVal)
					attrTokenBuffer[1].Data = nil
					i = 1
				}
				if attrTokenBuffer[2] != nil {
					x2 = minify.Number(attrTokenBuffer[2].AttrVal)
					attrTokenBuffer[2].Data = nil
					i = 2
				}
				if attrTokenBuffer[3] != nil {
					y2 = minify.Number(attrTokenBuffer[3].AttrVal)
					attrTokenBuffer[3].Data = nil
					i = 3
				}

				d := make([]byte, 0, 7+len(x1)+len(y1)+len(x2)+len(y2))
				d = append(d, '"', 'M')
				d = append(d, x1...)
				d = append(d, ' ')
				d = append(d, y1...)
				d = append(d, 'L')
				d = append(d, x2...)
				d = append(d, ' ')
				d = append(d, y2...)
				d = append(d, 'z', '"')
				ShortenPathData(d[1:len(d)-1], pathDataBuffer)

				t.Data = pathBytes
				attrTokenBuffer[i].Data = dBytes
				attrTokenBuffer[i].AttrVal = d
			} else if tag == svg.Rect {
				getAttributes(&attrTokenBuffer, tb, svg.X, svg.Y, svg.Width, svg.Height, svg.Rx, svg.Ry)
				if attrTokenBuffer[4] == nil && attrTokenBuffer[5] == nil {
					i := 0
					x, y, w, h := zeroBytes, zeroBytes, zeroBytes, zeroBytes
					if attrTokenBuffer[0] != nil {
						x = minify.Number(attrTokenBuffer[0].AttrVal)
						attrTokenBuffer[0].Data = nil
					}
					if attrTokenBuffer[1] != nil {
						y = minify.Number(attrTokenBuffer[1].AttrVal)
						attrTokenBuffer[1].Data = nil
						i = 1
					}
					if attrTokenBuffer[2] != nil {
						w = minify.Number(attrTokenBuffer[2].AttrVal)
						attrTokenBuffer[2].Data = nil
						i = 2
					}
					if attrTokenBuffer[3] != nil {
						h = minify.Number(attrTokenBuffer[3].AttrVal)
						attrTokenBuffer[3].Data = nil
						i = 3
					}
					if len(w) == 0 || len(w) == 1 && w[0] == '0' || len(h) == 0 || len(h) == 1 && h[0] == '0' {
						skipTag(tb, tag)
						break
					}

					d := make([]byte, 0, 9+2*len(x)+2*len(y)+len(w)+len(h))
					d = append(d, '"', 'M')
					d = append(d, x...)
					d = append(d, ' ')
					d = append(d, y...)
					d = append(d, 'h')
					d = append(d, w...)
					d = append(d, 'v')
					d = append(d, h...)
					d = append(d, 'H')
					d = append(d, x...)
					d = append(d, 'z', '"')
					ShortenPathData(d[1:len(d)-1], pathDataBuffer)

					t.Data = pathBytes
					attrTokenBuffer[i].Data = dBytes
					attrTokenBuffer[i].AttrVal = d
				}
			} else if tag == svg.Polygon || tag == svg.Polyline {
				getAttributes(&attrTokenBuffer, tb, svg.Points)
				if attrTokenBuffer[0] != nil {
					points := attrTokenBuffer[0].AttrVal

					i := 0
					for i < len(points) && (points[i] == ' ' || points[i] == ',' || points[i] == '\n' || points[i] == '\r' || points[i] == '\t') {
						i++
					}
					if i == len(points) {
						break
					}
					for i < len(points) && !(points[i] == ' ' || points[i] == ',' || points[i] == '\n' || points[i] == '\r' || points[i] == '\t') {
						i++
					}
					for i < len(points) && (points[i] == ' ' || points[i] == ',' || points[i] == '\n' || points[i] == '\r' || points[i] == '\t') {
						i++
					}
					if i == len(points) {
						break
					}
					for i < len(points) && !(points[i] == ' ' || points[i] == ',' || points[i] == '\n' || points[i] == '\r' || points[i] == '\t') {
						i++
					}
					endMoveTo := i
					for i < len(points) && (points[i] == ' ' || points[i] == ',' || points[i] == '\n' || points[i] == '\r' || points[i] == '\t') {
						i++
					}
					startLineTo := i

					d := make([]byte, 0, 2+len(points))
					d = append(d, '"', 'M')
					d = append(d, points[:endMoveTo]...)
					d = append(d, 'L')
					d = append(d, points[startLineTo:]...)
					if tag == svg.Polygon {
						d = append(d, 'z')
					}
					d = append(d, '"')
					ShortenPathData(d[1:len(d)-1], pathDataBuffer)

					t.Data = pathBytes
					attrTokenBuffer[0].Data = dBytes
					attrTokenBuffer[0].AttrVal = d
				}
			}
			if _, err := w.Write(t.Data); err != nil {
				return err
			}
		case xml.AttributeToken:
			if len(t.AttrVal) < 2 || t.Data == nil { // data is nil when attribute has been removed
				continue
			}
			attr := t.Hash
			val := parse.ReplaceMultipleWhitespace(t.AttrVal[1 : len(t.AttrVal)-1])
			if tag == svg.Svg && attr == svg.Version {
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

			if tag == svg.Svg && attr == svg.ContentStyleType {
				val = minify.ContentType(val)
				defaultStyleType = string(val)
				defaultInlineStyleType = defaultStyleType + ";inline=1"
			} else if attr == svg.Style {
				attrMinifyBuffer.Reset()
				if m.Minify(defaultInlineStyleType, attrMinifyBuffer, buffer.NewReader(val)) == nil {
					val = attrMinifyBuffer.Bytes()
				}
			} else if attr == svg.D {
				val = ShortenPathData(val, pathDataBuffer)
			} else if attr == svg.ViewBox {
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
			} else if colorAttrMap[attr] && len(val) > 0 {
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
			if len(t.Data) > 2+len(t.Text) {
				t.Data[2+len(t.Text)] = '>'
				if _, err := w.Write(t.Data[:2+len(t.Text)+1]); err != nil {
					return err
				}
			} else if _, err := w.Write(t.Data); err != nil {
				return err
			}
		}
	}
}

// type pathInstruction struct {
// 	cmd   byte
// 	param [7][]byte // elliptical arc has seven parameters
// }

// func readPathData(pathBuffer *[]pathInstruction, b []byte) {
// 	for i := 0; i < len(b); {
// 		c := b[i]
// 	COMMAND:
// 		if c == 'H' || c == 'h' || c == 'V' || c == 'v' || c == 'M' || c == 'm' || c == 'L' || c == 'l' || c == 'T' || c == 't' || c == 'S' || c == 's' || c == 'Q' || c == 'q' || c == 'C' || c == 'c' || c == 'A' || c == 'a' {
// 			instruction := pathInstruction{cmd: c}
// 			i++
// 			n := 2
// 			if c == 'H' || c == 'h' || c == 'V' || c == 'v' {
// 				n = 1
// 			} else if c == 'S' || c == 's' || c == 'Q' || c == 'q' {
// 				n = 4
// 			} else if c == 'C' || c == 'c' {
// 				n = 6
// 			} else if c == 'A' || c == 'a' {
// 				n = 7
// 			}
// 			for j := 0; j < n; j++ {
// 				for len(b) > i && (b[i] < '0' || b[i] > '9') && b[i] != '-' && b[i] != '.' && b[i] != '+' {
// 					i++
// 				}
// 				if n := parse.Number(b[i:]); n > 0 {
// 					instruction.param[j] = b[i : i+n]
// 					i += n
// 				}
// 			}
// 			*pathBuffer = append(*pathBuffer, instruction)
// 		} else {
// 			i++
// 			continue
// 		}
// 		for len(b) > i && (b[i] == ' ' || b[i] == ',' || b[i] == '\n' || b[i] == '\r' || b[i] == '\t') {
// 			i++
// 		}
// 		if len(b) > i && (b[i] >= '0' && b[i] <= '9' || b[i] == '-' || b[i] == '.' || b[i] == '+') {
// 			goto COMMAND
// 		}
// 	}
// }

// func shortenPathData(b []byte) []byte {
// 	cmd := byte(0)
// 	coords := [][]byte{}

// 	var x, y, x0, y0 float64

// 	j := 0
// 	for i := 0; i < len(b); i++ {
// 		c := b[i]
// 		if c == ' ' || c == ',' || c == '\n' || c == '\r' || c == '\t' {
// 			continue
// 		} else if c >= 'A' { // any command
// 			if cmd == 0 {
// 				cmd = c
// 			} else if c != cmd {
// 				x1, y1 := x0, y0
// 				if cmd == 'M' {
// 					x1 = toFloat(coords[len(coords)-2])
// 					y1 = toFloat(coords[len(coords)-1])
// 				} else if cmd == 'm' {
// 					x1 += toFloat(coords[len(coords)-2])
// 					y1 += toFloat(coords[len(coords)-1])
// 				}
// 				j += shortenPathDataInstruction(b[j:], cmd, coords, &x, &y)
// 				if cmd == 'M' || cmd == 'm' || cmd == 'Z' || cmd == 'z' {
// 					x0 = x1
// 					y0 = y1
// 					x = x0
// 					y = y0
// 				}
// 				cmd = c
// 				coords = coords[:0]
// 			}
// 		} else if n := parse.Number(b[i:]); n > 0 {
// 			coords = append(coords, minify.Number(b[i:i+n]))
// 			i += n - 1
// 		}
// 	}
// 	j += shortenPathDataInstruction(b[j:], cmd, coords, &x, &y)
// 	return b[:j]
// }

// func shortenPathDataInstruction(b []byte, cmd byte, coords [][]byte, x *float64, y *float64) int {
// 	n := len(coords)
// 	cmdIsRelative := cmd >= 'a'

// 	// get new cursor coordinates
// 	ax, ay := *x, *y
// 	if n >= 2 && (cmd == 'L' || cmd == 'l' || cmd == 'C' || cmd == 'c' || cmd == 'S' || cmd == 's' || cmd == 'Q' || cmd == 'q' || cmd == 'T' || cmd == 't' || cmd == 'A' || cmd == 'a') {
// 		ax = toFloat(coords[n-2])
// 		ay = toFloat(coords[n-1])
// 	} else if n >= 1 && (cmd == 'H' || cmd == 'h' || cmd == 'V' || cmd == 'v') {
// 		if cmd == 'H' || cmd == 'h' {
// 			ax = toFloat(coords[n-1])
// 		} else {
// 			ay = toFloat(coords[n-1])
// 		}
// 	}

// 	// make an alternative path with absolute/relative altered
// 	bAlter := make([]byte, 0, len(b))
// 	cmdAlter := cmd - 'A' + 'a'
// 	dx, dy := -*x, -*y
// 	if cmdIsRelative {
// 		cmdAlter = cmd - 'a' + 'A'
// 		dx, dy = *x, *y
// 	}
// 	bAlter = shortenPathDataInstructionCoordsAlter(bAlter, cmdAlter, coords, dx, dy)

// 	// choose shortest, relative or absolute path?
// 	j := shortenPathDataInstructionCoords(b, cmd, coords)
// 	jAlter := len(bAlter)
// 	if jAlter < j {
// 		j = jAlter
// 		copy(b, bAlter)
// 	}

// 	// set new cursor coordinates
// 	if cmdIsRelative {
// 		*x += ax
// 		*y += ay
// 	} else {
// 		*x = ax
// 		*y = ay
// 	}
// 	return j
// }

// func shortenPathDataInstructionCoords(b []byte, cmd byte, coords [][]byte) int {
// 	prevDigit := false
// 	prevDigitRequiresSpace := true

// 	b[0] = cmd
// 	j := 1
// 	for _, coord := range coords {
// 		if prevDigit && (coord[0] >= '0' && coord[0] <= '9' || coord[0] == '.' && prevDigitRequiresSpace) {
// 			b[j] = ' '
// 			j++
// 		}
// 		prevDigit = true
// 		prevDigitRequiresSpace = true
// 		for _, c := range coord {
// 			if c == '.' || c == 'e' || c == 'E' {
// 				prevDigitRequiresSpace = false
// 				break
// 			}
// 		}
// 		j += copy(b[j:], coord)
// 	}
// 	return j
// }

// func shortenPathDataInstructionCoordsAlter(b []byte, cmd byte, coords [][]byte, dx, dy float64) []byte {
// 	prevDigit := false
// 	prevDigitRequiresSpace := true

// 	coordBuf := []byte{}

// 	b = append(b, cmd)
// 	for i, coord := range coords {
// 		f := toFloat(coord)
// 		if cmd == 'L' || cmd == 'l' || cmd == 'C' || cmd == 'c' || cmd == 'S' || cmd == 's' || cmd == 'Q' || cmd == 'q' || cmd == 'T' || cmd == 't' || cmd == 'M' || cmd == 'm' {
// 			if i%2 == 0 {
// 				f += dx
// 			} else {
// 				f += dy
// 			}
// 		} else if cmd == 'H' || cmd == 'h' {
// 			f += dx
// 		} else if cmd == 'V' || cmd == 'v' {
// 			f += dy
// 		} else if cmd == 'A' || cmd == 'a' {
// 			if i%7 == 5 {
// 				f += dx
// 			} else if i%7 == 6 {
// 				f += dy
// 			}
// 		} else {
// 			continue
// 		}
// 		coordBuf = strconv.AppendFloat(coordBuf[:0], f, 'f', -1, 32)
// 		coordBuf = minify.Number(coordBuf)

// 		if prevDigit && (coordBuf[0] >= '0' && coordBuf[0] <= '9' || coordBuf[0] == '.' && prevDigitRequiresSpace) {
// 			b = append(b, ' ')
// 		}
// 		prevDigit = true
// 		prevDigitRequiresSpace = true
// 		for _, c := range coordBuf {
// 			if c == '.' || c == 'e' || c == 'E' {
// 				prevDigitRequiresSpace = false
// 				break
// 			}
// 		}
// 		b = append(b, coordBuf...)
// 	}
// 	return b
// }

// func toFloat(b []byte) float64 {
// 	f, err := strconv.ParseFloat(string(b), 64)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return f
// }

// func shortenPathDataOld(b []byte) []byte {
// 	cmd := byte(0)
// 	coords := [][]byte{}
// 	nCoord := 0

// 	//var x, y, x0, y0 float64

// 	j := 0
// 	start := 0
// 	for i := 0; i < len(b); i++ {
// 		c := b[i]
// 		if c == ' ' || c == ',' || c == '\n' || c == '\r' || c == '\t' {
// 			if start != 0 {
// 				j += copy(b[j:], b[start:i])
// 			} else {
// 				j += i
// 			}
// 			start = i + 1
// 		} else if n := parse.Number(b[i:]); n > 0 {
// 			if start != 0 {
// 				j += copy(b[j:], b[start:i])
// 			} else {
// 				j += i
// 			}
// 			num := minify.Number(b[i : i+n])
// 			if prevDigit && (num[0] >= '0' && num[0] <= '9' || num[0] == '.' && prevDigitRequiresSpace) {
// 				b[j] = ' '
// 				j++
// 			}
// 			prevDigit = true
// 			prevDigitRequiresSpace = true
// 			for _, c := range num {
// 				if c == '.' || c == 'e' || c == 'E' {
// 					prevDigitRequiresSpace = false
// 					break
// 				}
// 			}
// 			j += copy(b[j:], num)
// 			start = i + n
// 			i += n - 1
// 		} else {
// 			if cmd == c {
// 				if start != 0 {
// 					j += copy(b[j:], b[start:i])
// 				} else {
// 					j += i
// 				}
// 				start = i + 1
// 			} else {
// 				cmd = c
// 				prevDigit = false
// 			}
// 		}
// 	}
// 	if start != 0 {
// 		j += copy(b[j:], b[start:])
// 		return b[:j]
// 	}
// 	return b
// }

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

////////////////////////////////////////////////////////////////

func getAttributes(attrTokenBuffer *[]*Token, tb *TokenBuffer, hashes ...svg.Hash) {
	*attrTokenBuffer = (*attrTokenBuffer)[:len(hashes)]
	for j, _ := range *attrTokenBuffer {
		(*attrTokenBuffer)[j] = nil
	}
	for i := 0; ; i++ {
		t := tb.Peek(i)
		if t.TokenType != xml.AttributeToken {
			break
		}
		for j, hash := range hashes {
			if t.Hash == hash {
				if len(t.AttrVal) > 1 && t.AttrVal[0] == '"' {
					t.AttrVal = parse.TrimWhitespace(t.AttrVal[1 : len(t.AttrVal)-1]) // quotes will be readded in attribute loop if necessary
				}
				(*attrTokenBuffer)[j] = t
				break
			}
		}
	}
}

func skipTag(tb *TokenBuffer, tag svg.Hash) {
	for {
		if t := *tb.Shift(); (t.TokenType == xml.EndTagToken || t.TokenType == xml.StartTagCloseVoidToken) && t.Hash == tag || t.TokenType == xml.ErrorToken {
			break
		}
	}
}
