package html // import "github.com/tdewolff/minify/html"

import (
	"bytes"
	"io"

	"github.com/tdewolff/buffer"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/parse"
	"github.com/tdewolff/parse/html"
)

var ltByte = []byte{'<'}
var gtByte = []byte{'>'}
var isByte = []byte{'='}
var spaceByte = []byte{' '}
var endBytes = []byte{'<', '/'}
var escapedSingleQuoteBytes = []byte("&#39;")
var escapedDoubleQuoteBytes = []byte("&#34;")

var rawTagMap = map[html.Hash]bool{
	html.Code:     true,
	html.Iframe:   true,
	html.Pre:      true,
	html.Script:   true,
	html.Style:    true,
	html.Textarea: true,
}

var inlineTagMap = map[html.Hash]bool{
	html.A:       true,
	html.Abbr:    true,
	html.Acronym: true,
	html.B:       true,
	html.Bdo:     true,
	html.Big:     true,
	html.Cite:    true,
	html.Button:  true,
	html.Dfn:     true,
	html.Em:      true,
	html.I:       true,
	html.Img:     true,
	html.Input:   true,
	html.Kbd:     true,
	html.Label:   true,
	html.Map:     true,
	html.Object:  true,
	html.Q:       true,
	html.Samp:    true,
	html.Select:  true,
	html.Small:   true,
	html.Span:    true,
	html.Strong:  true,
	html.Sub:     true,
	html.Sup:     true,
	html.Tt:      true,
	html.Var:     true,
}

var blockTagMap = map[html.Hash]bool{
	html.Address:    true,
	html.Article:    true,
	html.Aside:      true,
	html.Blockquote: true,
	html.Div:        true,
	html.Dl:         true,
	html.Fieldset:   true,
	html.Footer:     true,
	html.Form:       true,
	html.H1:         true,
	html.H2:         true,
	html.H3:         true,
	html.H4:         true,
	html.H5:         true,
	html.H6:         true,
	html.Header:     true,
	html.Hgroup:     true,
	html.Hr:         true,
	html.Main:       true,
	html.Nav:        true,
	html.Ol:         true,
	html.P:          true,
	html.Pre:        true,
	html.Section:    true,
	html.Table:      true,
	html.Ul:         true,
}

var booleanAttrMap = map[html.Hash]bool{
	html.Allowfullscreen: true,
	html.Async:           true,
	html.Autofocus:       true,
	html.Autoplay:        true,
	html.Checked:         true,
	html.Compact:         true,
	html.Controls:        true,
	html.Declare:         true,
	html.Default:         true,
	html.DefaultChecked:  true,
	html.DefaultMuted:    true,
	html.DefaultSelected: true,
	html.Defer:           true,
	html.Disabled:        true,
	html.Draggable:       true,
	html.Enabled:         true,
	html.Formnovalidate:  true,
	html.Hidden:          true,
	html.Inert:           true,
	html.Ismap:           true,
	html.Itemscope:       true,
	html.Multiple:        true,
	html.Muted:           true,
	html.Nohref:          true,
	html.Noresize:        true,
	html.Noshade:         true,
	html.Novalidate:      true,
	html.Nowrap:          true,
	html.Open:            true,
	html.Pauseonexit:     true,
	html.Readonly:        true,
	html.Required:        true,
	html.Reversed:        true,
	html.Scoped:          true,
	html.Seamless:        true,
	html.Selected:        true,
	html.Sortable:        true,
	html.Spellcheck:      true,
	html.Translate:       true,
	html.Truespeed:       true,
	html.Typemustmatch:   true,
	html.Undeterminate:   true,
	html.Visible:         true,
}

var caseInsensitiveAttrMap = map[html.Hash]bool{
	html.Accept_Charset: true,
	html.Accept:         true,
	html.Align:          true,
	html.Alink:          true,
	html.Axis:           true,
	html.Bgcolor:        true,
	html.Charset:        true,
	html.Clear:          true,
	html.Codetype:       true,
	html.Color:          true,
	html.Dir:            true,
	html.Enctype:        true,
	html.Face:           true,
	html.Frame:          true,
	html.Hreflang:       true,
	html.Http_Equiv:     true,
	html.Lang:           true,
	html.Language:       true,
	html.Link:           true,
	html.Media:          true,
	html.Method:         true,
	html.Rel:            true,
	html.Rev:            true,
	html.Rules:          true,
	html.Scope:          true,
	html.Scrolling:      true,
	html.Shape:          true,
	html.Target:         true,
	html.Text:           true,
	html.Type:           true,
	html.Valign:         true,
	html.Valuetype:      true,
	html.Vlink:          true,
}

var urlAttrMap = map[html.Hash]bool{
	html.Action:     true,
	html.Background: true,
	html.Cite:       true,
	html.Classid:    true,
	html.Codebase:   true,
	html.Data:       true,
	html.Formaction: true,
	html.Href:       true,
	html.Icon:       true,
	html.Longdesc:   true,
	html.Manifest:   true,
	html.Poster:     true,
	html.Profile:    true,
	html.Src:        true,
	html.Usemap:     true,
	html.Xmlns:      true,
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

func normalizeContentType(b []byte) []byte {
	j := 0
	start := 0
	b = parse.ToLower(bytes.TrimSpace(replaceMultipleWhitespace(b)))
	for i, c := range b {
		if c == ' ' && (b[i-1] == ';' || b[i-1] == ',') {
			if start != 0 {
				j += copy(b[j:], b[start:i])
			} else {
				j += i - start
			}
			start = i + 1
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
func escapeAttrVal(b []byte, buf *[]byte) []byte {
	singles := 0
	doubles := 0
	unquoted := true
	for i, c := range b {
		if c == '&' {
			if quote, _, ok := isAtQuoteEntity(b[i:]); ok {
				if quote == '"' {
					doubles++
					unquoted = false
				} else {
					singles++
					unquoted = false
				}
			}
		} else if c == '"' {
			doubles++
			unquoted = false
		} else if c == '\'' {
			singles++
			unquoted = false
		} else if unquoted && (c == '`' || c == '<' || c == '=' || c == '>' || isWhitespace(c)) {
			unquoted = false
		}
	}

	if unquoted {
		return b
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

func attrValEqual(attrVal, match []byte) bool {
	if len(attrVal) > 0 && (attrVal[0] == '"' || attrVal[0] == '\'') {
		attrVal = attrVal[1 : len(attrVal)-1]
	}
	return parse.Equal(attrVal, match)
}

func attrValEqualCaseInsensitive(attrVal, match []byte) bool {
	if len(attrVal) > 0 && (attrVal[0] == '"' || attrVal[0] == '\'') {
		attrVal = attrVal[1 : len(attrVal)-1]
	}
	return parse.EqualCaseInsensitive(attrVal, match)
}

////////////////////////////////////////////////////////////////

// Minify minifies HTML5 files, it reads from r and writes to w.
// Removes unnecessary whitespace, tags, attributes, quotes and comments and typically saves 10% in size.
func Minify(m minify.Minifier, _ string, w io.Writer, r io.Reader) error {
	var rawTag html.Hash
	var rawTagMediatype []byte
	precededBySpace := true // on true the next text token must not start with a space
	defaultScriptType := "text/javascript"
	defaultStyleType := "text/css"

	attrMinifyBuffer := make([]byte, 0, 64)
	attrEscapeBuffer := make([]byte, 0, 64)

	tb := newTokenBuffer(html.NewTokenizer(r))
	for {
		t := tb.shift()
		switch t.tt {
		case html.ErrorToken:
			if tb.err() == io.EOF {
				return nil
			}
			return tb.err()
		case html.DoctypeToken:
			if _, err := w.Write([]byte("<!doctype html>")); err != nil {
				return err
			}
		case html.CommentToken:
			// TODO: ensure that nested comments are handled properly (tokenizer doesn't handle this!)
			var comment []byte
			if bytes.HasPrefix(t.data, []byte("[if")) {
				comment = append(append([]byte("<!--"), t.data...), []byte("-->")...)
			} else if bytes.HasSuffix(t.data, []byte("--")) {
				// only occurs when mixed up with conditional comments
				comment = append(append([]byte("<!"), t.data...), '>')
			}
			if _, err := w.Write(comment); err != nil {
				return err
			}
		case html.TextToken:
			// CSS and JS minifiers for inline code
			if rawTag != 0 {
				if rawTag == html.Style || rawTag == html.Script || rawTag == html.Iframe {
					var mediatype string
					if rawTag == html.Iframe {
						mediatype = "text/html"
					} else if len(rawTagMediatype) > 0 {
						mediatype = string(rawTagMediatype)
					} else if rawTag == html.Script {
						mediatype = defaultScriptType
					} else {
						mediatype = defaultStyleType
					}
					// ignore CDATA in embedded HTML
					if mediatype == "text/html" {
						trimmedData := bytes.TrimSpace(t.data)
						if len(trimmedData) > 12 && bytes.Equal(trimmedData[:9], []byte("<![CDATA[")) && bytes.Equal(trimmedData[len(trimmedData)-3:], []byte("]]>")) {
							if _, err := w.Write([]byte("<![CDATA[")); err != nil {
								return err
							}
							t.data = trimmedData[9:]
						}
					}
					if err := m.Minify(mediatype, w, buffer.NewReader(t.data)); err != nil {
						if err == minify.ErrNotExist {
							// no minifier, write the original
							if _, err := w.Write(t.data); err != nil {
								return err
							}
						} else {
							return err
						}
					}
				} else if _, err := w.Write(t.data); err != nil {
					return err
				}
			} else if t.data = replaceMultipleWhitespace(t.data); len(t.data) > 0 {
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
						next := tb.peek(i)
						// trim if EOF, text token with whitespace begin or block token
						if next.tt == html.ErrorToken {
							trim = true
							break
						} else if next.tt == html.TextToken {
							// remove if the text token starts with a whitespace
							trim = (len(next.data) > 0 && isWhitespace(next.data[0]))
							break
						} else if next.tt == html.StartTagToken || next.tt == html.EndTagToken {
							if !inlineTagMap[next.hash] {
								trim = true
								break
							} else if next.tt == html.StartTagToken {
								break
							}
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
		case html.StartTagToken, html.EndTagToken:
			rawTag = 0
			hasAttributes := false
			if t.tt == html.StartTagToken {
				if next := tb.peek(0); next.tt != html.StartTagCloseToken && next.tt != html.StartTagVoidToken {
					hasAttributes = true
				}
			}

			if !inlineTagMap[t.hash] {
				precededBySpace = true
				if rawTagMap[t.hash] && t.tt == html.StartTagToken {
					// ignore empty script and style tags
					if !hasAttributes && (t.hash == html.Script || t.hash == html.Style) {
						if next := tb.peek(1); next.tt == html.EndTagToken {
							tb.shift()
							tb.shift()
							break
						}
					}
					rawTag = t.hash
					rawTagMediatype = []byte{}
				}

				// remove superfluous ending tags
				if !hasAttributes && (t.hash == html.Html || t.hash == html.Head || t.hash == html.Body || t.hash == html.Colgroup) {
					break
				} else if t.tt == html.EndTagToken {
					if t.hash == html.Thead || t.hash == html.Tbody || t.hash == html.Tfoot || t.hash == html.Tr || t.hash == html.Th || t.hash == html.Td ||
						t.hash == html.Optgroup || t.hash == html.Option || t.hash == html.Dd || t.hash == html.Dt ||
						t.hash == html.Li || t.hash == html.Rb || t.hash == html.Rt || t.hash == html.Rtc || t.hash == html.Rp {
						break
					} else if t.hash == html.P {
						remove := false
						i := 0
						for {
							next := tb.peek(i)
							i++
							// continue if text token is empty or whitespace
							if next.tt == html.TextToken && isAllWhitespace(next.data) {
								continue
							}
							remove = (next.tt == html.ErrorToken || next.tt == html.EndTagToken && next.hash != html.A || next.tt == html.StartTagToken && blockTagMap[next.hash])
							break
						}
						if remove {
							break
						}
					}
				}

				// rewrite meta tag with charset
				if hasAttributes && t.hash == html.Meta {
					iHTTPEquiv := -1
					iName := -1
					iContent := -1
					hasCharset := false
					i := 0
					for {
						attr := tb.peek(i)
						if attr.tt != html.AttributeToken {
							break
						}
						if attr.hash == html.Http_Equiv {
							iHTTPEquiv = i
						} else if attr.hash == html.Name {
							iName = i
						} else if attr.hash == html.Content {
							iContent = i
						} else if attr.hash == html.Charset {
							hasCharset = true
						}
						i++
					}
					if iContent != -1 {
						content := tb.peek(iContent)
						if iHTTPEquiv != -1 {
							httpEquiv := tb.peek(iHTTPEquiv)
							content.attrVal = normalizeContentType(content.attrVal)
							if !hasCharset && attrValEqualCaseInsensitive(httpEquiv.attrVal, []byte("content-type")) && attrValEqual(content.attrVal, []byte("text/html;charset=utf-8")) {
								httpEquiv.data = nil
								content.data = []byte("charset")
								content.hash = html.Charset
								content.attrVal = []byte("utf-8")
							} else if attrValEqualCaseInsensitive(httpEquiv.attrVal, []byte("content-style-type")) {
								defaultStyleType = string(content.attrVal)
							} else if attrValEqualCaseInsensitive(httpEquiv.attrVal, []byte("content-script-type")) {
								defaultScriptType = string(content.attrVal)
							}
						}
						if iName != -1 {
							name := tb.peek(iName)
							if attrValEqualCaseInsensitive(name.attrVal, []byte("keywords")) {
								content.attrVal = bytes.Replace(content.attrVal, []byte(", "), []byte(","), -1)
							} else if attrValEqualCaseInsensitive(name.attrVal, []byte("viewport")) {
								content.attrVal = bytes.Replace(content.attrVal, []byte(" "), []byte(""), -1)
							}
						}
					}
				}
			}

			// write tag
			if t.tt == html.EndTagToken {
				if _, err := w.Write(endBytes); err != nil {
					return err
				}
			} else {
				if _, err := w.Write(ltByte); err != nil {
					return err
				}
			}
			if _, err := w.Write(t.data); err != nil {
				return err
			}

			// write attributes
			if hasAttributes {
				for {
					attr := tb.shift()
					if attr.tt != html.AttributeToken {
						break
					} else if attr.data == nil {
						// removed attribute
						continue
					}

					val := attr.attrVal
					if len(val) > 1 && (val[0] == '"' || val[0] == '\'') {
						val = bytes.TrimSpace(val[1 : len(val)-1])
					}
					if caseInsensitiveAttrMap[attr.hash] {
						val = parse.ToLower(val)
						if attr.hash == html.Enctype || attr.hash == html.Codetype || attr.hash == html.Accept || attr.hash == html.Type && (t.hash == html.A || t.hash == html.Link || t.hash == html.Object || t.hash == html.Param || t.hash == html.Script || t.hash == html.Style || t.hash == html.Source) {
							val = normalizeContentType(val)
						}
					}
					if rawTag != 0 && attr.hash == html.Type {
						rawTagMediatype = val
					}

					// default attribute values can be ommited
					if attr.hash == html.Type && (t.hash == html.Script && parse.Equal(val, []byte("text/javascript")) ||
						t.hash == html.Style && parse.Equal(val, []byte("text/css")) ||
						t.hash == html.Link && parse.Equal(val, []byte("text/css")) ||
						t.hash == html.Input && parse.Equal(val, []byte("text")) ||
						t.hash == html.Button && parse.Equal(val, []byte("submit"))) ||
						attr.hash == html.Method && parse.Equal(val, []byte("get")) ||
						attr.hash == html.Enctype && parse.Equal(val, []byte("application/x-www-form-urlencoded")) ||
						attr.hash == html.Colspan && parse.Equal(val, []byte("1")) ||
						attr.hash == html.Rowspan && parse.Equal(val, []byte("1")) ||
						attr.hash == html.Shape && parse.Equal(val, []byte("rect")) ||
						attr.hash == html.Span && parse.Equal(val, []byte("1")) ||
						attr.hash == html.Clear && parse.Equal(val, []byte("none")) ||
						attr.hash == html.Frameborder && parse.Equal(val, []byte("1")) ||
						attr.hash == html.Scrolling && parse.Equal(val, []byte("auto")) ||
						attr.hash == html.Valuetype && parse.Equal(val, []byte("data")) ||
						attr.hash == html.Language && t.hash == html.Script && parse.Equal(val, []byte("javascript")) {
						continue
					}
					if _, err := w.Write(spaceByte); err != nil {
						return err
					}
					if _, err := w.Write(attr.data); err != nil {
						return err
					}

					// booleans have no value
					if !booleanAttrMap[attr.hash] {
						if len(val) == 0 {
							continue
						}
						if _, err := w.Write(isByte); err != nil {
							return err
						}
						// CSS and JS minifiers for attribute inline code
						if attr.hash == html.Style {
							out := buffer.NewWriter(attrMinifyBuffer[:0])
							if m.Minify(defaultStyleType, out, buffer.NewReader(val)) == nil {
								val = out.Bytes()
							}
							attrMinifyBuffer = out.Bytes() // reuse resized buffer
						} else if len(attr.data) > 2 && attr.data[0] == 'o' && attr.data[1] == 'n' {
							if len(val) >= 11 && parse.EqualCaseInsensitive(val[:11], []byte("javascript:")) {
								val = val[11:]
							}
							out := buffer.NewWriter(attrMinifyBuffer[:0])
							if m.Minify(defaultScriptType, out, buffer.NewReader(val)) == nil {
								val = out.Bytes()
							}
							attrMinifyBuffer = out.Bytes() // reuse resized buffer
						} else if urlAttrMap[attr.hash] {
							if len(val) >= 5 && parse.EqualCaseInsensitive(val[:5], []byte{'h', 't', 't', 'p', ':'}) {
								val = val[5:]
							}
						}

						// no quotes if possible, else prefer single or double depending on which occurs more often in value
						val = escapeAttrVal(val, &attrEscapeBuffer) // reuse resized buffer
						if _, err := w.Write(val); err != nil {
							return err
						}
					}
				}
			}
			if _, err := w.Write(gtByte); err != nil {
				return err
			}
		}
	}
}

////////////////////////////////////////////////////////////////

type token struct {
	tt      html.TokenType
	data    []byte
	attrVal []byte
	hash    html.Hash
}

type tokenBuffer struct {
	z *html.Tokenizer

	buf []token
	pos int
}

func newTokenBuffer(z *html.Tokenizer) *tokenBuffer {
	return &tokenBuffer{
		z:   z,
		buf: make([]token, 0, 8),
	}
}

func (tb *tokenBuffer) read(p []token) int {
	for i := 0; i < len(p); i++ {
		tt, data := tb.z.Next()
		if !tb.z.IsEOF() {
			data = parse.Copy(data)
		}

		var attrVal []byte
		var hash html.Hash
		if tt == html.AttributeToken {
			attrVal = tb.z.AttrVal()
			if !tb.z.IsEOF() {
				attrVal = parse.Copy(attrVal)
			}
			hash = html.ToHash(data)
		} else if tt == html.StartTagToken || tt == html.EndTagToken {
			hash = tb.z.RawTag()
			if hash == 0 {
				hash = html.ToHash(data)
			}
		}
		p[i] = token{tt, data, attrVal, hash}

		if tt == html.ErrorToken {
			return i + 1
		}
	}
	return len(p)
}

func (tb *tokenBuffer) peek(i int) *token {
	end := tb.pos + i
	if end >= len(tb.buf) {
		if len(tb.buf) > 0 && tb.buf[len(tb.buf)-1].tt == html.ErrorToken {
			return &tb.buf[len(tb.buf)-1]
		}

		c := cap(tb.buf)
		d := len(tb.buf) - tb.pos
		var buf1 []token
		if 2*d > c {
			buf1 = make([]token, d, 2*c)
		} else {
			buf1 = tb.buf[:d]
		}
		copy(buf1, tb.buf[tb.pos:])

		n := tb.read(buf1[d:cap(buf1)])
		end -= tb.pos
		tb.pos, tb.buf = 0, buf1[:d+n]
	}
	return &tb.buf[end]
}

func (tb *tokenBuffer) shift() token {
	shifted := *tb.peek(0)
	tb.pos++
	return shifted
}

func (tb *tokenBuffer) err() error {
	return tb.z.Err()
}
