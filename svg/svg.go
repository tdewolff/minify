package svg // import "github.com/tdewolff/minify/svg"

import (
	"io"

	"github.com/tdewolff/minify"
	xmlMinify "github.com/tdewolff/minify/xml"
	"github.com/tdewolff/parse"
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
	CDATAStartBytes = []byte("<![CDATA[")
	CDATAEndBytes   = []byte("]]>")
)

////////////////////////////////////////////////////////////////

// Minify minifies XML files, it reads from r and writes to w.
// Removes unnecessary whitespace, tags, attributes, quotes and comments and typically saves 10% in size.
func Minify(m minify.Minifier, _ string, w io.Writer, r io.Reader) error {
	escapeBuffer := make([]byte, 0, 64)

	z := xml.NewTokenizer(r)
	tb := xmlMinify.NewTokenBuffer(z)
	for {
		t := *tb.Shift()
		if t.TokenType == xml.CDATAToken {
			var useCDATA bool
			if t.Data, useCDATA = xmlMinify.EscapeCDATAVal(&escapeBuffer, t.Data); !useCDATA {
				t.TokenType = xml.TextToken
			}
		}
		switch t.TokenType {
		case xml.ErrorToken:
			if z.Err() == io.EOF {
				return nil
			}
			return z.Err()
		case xml.CDATAToken:
			if _, err := w.Write(CDATAStartBytes); err != nil {
				return err
			}
			if _, err := w.Write(t.Data); err != nil {
				return err
			}
			if _, err := w.Write(CDATAEndBytes); err != nil {
				return err
			}
		case xml.StartTagToken:
			if _, err := w.Write(ltBytes); err != nil {
				return err
			}
			if _, err := w.Write(t.Data); err != nil {
				return err
			}
		case xml.StartTagPIToken:
			for {
				if t := *tb.Shift(); t.TokenType != xml.StartTagClosePIToken {
					break
				}
			}
		case xml.AttributeToken:
			val := t.AttrVal
			if len(val) < 2 {
				if _, err := w.Write(emptyBytes); err != nil {
					return err
				}
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

			// prefer single or double quotes depending on what occurs more often in value
			val = val[1 : len(val)-1]
			val = xmlMinify.EscapeAttrVal(&escapeBuffer, val)
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
