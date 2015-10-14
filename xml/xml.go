// Package xml minifies XML1.0 following the specifications at http://www.w3.org/TR/xml/.
package xml // import "github.com/tdewolff/minify/xml"

import (
	"io"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/parse"
	"github.com/tdewolff/parse/xml"
)

var (
	ltBytes         = []byte("<")
	gtBytes         = []byte(">")
	voidBytes       = []byte("/>")
	ltPIBytes       = []byte("<?")
	gtPIBytes       = []byte("?>")
	endBytes        = []byte("</")
	doctypeBytes    = []byte("<!DOCTYPE ")
	cdataStartBytes = []byte("<![CDATA[")
	cdataEndBytes   = []byte("]]>")
	isBytes         = []byte("=")
	spaceBytes      = []byte(" ")
)

////////////////////////////////////////////////////////////////

// Minify minifies XML data, it reads from r and writes to w.
func Minify(m minify.Minifier, _ string, w io.Writer, r io.Reader) error {
	precededBySpace := true // on true the next text token must not start with a space

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
		switch t.TokenType {
		case xml.ErrorToken:
			if l.Err() == io.EOF {
				return nil
			}
			return l.Err()
		case xml.DOCTYPEToken:
			if _, err := w.Write(doctypeBytes); err != nil {
				return err
			}
			if _, err := w.Write(t.Data); err != nil {
				return err
			}
			if _, err := w.Write(gtBytes); err != nil {
				return err
			}
		case xml.CDATAToken:
			if _, err := w.Write(cdataStartBytes); err != nil {
				return err
			}
			if _, err := w.Write(t.Data); err != nil {
				return err
			}
			if _, err := w.Write(cdataEndBytes); err != nil {
				return err
			}
		case xml.TextToken:
			if t.Data = parse.ReplaceMultipleWhitespace(t.Data); len(t.Data) > 0 {
				// whitespace removal; trim left
				if t.Data[0] == ' ' && precededBySpace {
					t.Data = t.Data[1:]
				}

				// whitespace removal; trim right
				precededBySpace = false
				if len(t.Data) == 0 {
					precededBySpace = true
				} else if t.Data[len(t.Data)-1] == ' ' {
					precededBySpace = true
					i := 0
					for {
						next := tb.Peek(i)
						// trim if EOF, text token with whitespace begin or block token
						if next.TokenType == xml.StartTagToken || next.TokenType == xml.EndTagToken || next.TokenType == xml.ErrorToken {
							t.Data = t.Data[:len(t.Data)-1]
							precededBySpace = false
							break
						} else if next.TokenType == xml.TextToken {
							// remove if the text token starts with a whitespace
							if len(next.Data) > 0 && parse.IsWhitespace(next.Data[0]) {
								t.Data = t.Data[:len(t.Data)-1]
								precededBySpace = false

							}
							break
						}
						i++
					}
				}
				if _, err := w.Write(t.Data); err != nil {
					return err
				}
			}
		case xml.StartTagToken:
			if _, err := w.Write(ltBytes); err != nil {
				return err
			}
			if _, err := w.Write(t.Data); err != nil {
				return err
			}
		case xml.StartTagPIToken:
			if _, err := w.Write(ltPIBytes); err != nil {
				return err
			}
			if _, err := w.Write(t.Data); err != nil {
				return err
			}
		case xml.AttributeToken:
			if _, err := w.Write(spaceBytes); err != nil {
				return err
			}
			if _, err := w.Write(t.Data); err != nil {
				return err
			}
			if _, err := w.Write(isBytes); err != nil {
				return err
			}

			if len(t.AttrVal) < 2 {
				if _, err := w.Write(t.AttrVal); err != nil {
					return err
				}
			} else {
				// prefer single or double quotes depending on what occurs more often in value
				val := xml.EscapeAttrVal(&attrByteBuffer, t.AttrVal[1:len(t.AttrVal)-1])
				if _, err := w.Write(val); err != nil {
					return err
				}
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
		case xml.StartTagClosePIToken:
			if _, err := w.Write(gtPIBytes); err != nil {
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
