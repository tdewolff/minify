package html // import "github.com/tdewolff/minify/html"

import (
	"bytes"
	"io"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/parse"
	"github.com/tdewolff/parse/html"
)

var rawTagMap = map[html.Hash]bool{
	html.Code:     true,
	html.Noscript: true,
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
	html.Address:       true,
	html.Article:    true,
	html.Aside: true,
	html.Blockquote:       true,
	html.Div:     true,
	html.Dl:     true,
	html.Fieldset:    true,
	html.Footer:  true,
	html.Form:     true,
	html.H1:      true,
	html.H2:       true,
	html.H3:     true,
	html.H4:   true,
	html.H5:     true,
	html.H6:   true,
	html.Header:     true,
	html.Hgroup:  true,
	html.Hr:       true,
	html.Main:    true,
	html.Nav:  true,
	html.Ol:   true,
	html.P:    true,
	html.Pre:  true,
	html.Section:     true,
	html.Table:     true,
	html.Ul:      true,
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
func replaceMultipleWhitespace(s []byte) []byte {
	i := 0
	t := make([]byte, len(s))
	previousSpace := false
	for _, x := range s {
		if isWhitespace(x) {
			if !previousSpace {
				previousSpace = true
				t[i] = ' '
				i++
			}
		} else {
			previousSpace = false
			t[i] = x
			i++
		}
	}
	return t[:i]
}

func normalizeContentType(s []byte) []byte {
	s = parse.CopyToLower(bytes.TrimSpace(replaceMultipleWhitespace(s)))
	t := make([]byte, len(s))
	w := 0
	start := 0
	for j, x := range s {
		if x == ' ' && (s[j-1] == ';' || s[j-1] == ',') {
			w += copy(t[w:], s[start:j])
			start = j + 1
		}
	}
	w += copy(t[w:], s[start:])
	return t[:w]
}

// escapeAttrVal returns the escape attribute value bytes without quotes.
func escapeAttrVal(s []byte) []byte {
	if len(s) == 0 {
		return []byte("\"\"")
	}
	s = html.Unescape(s)

	amps := 0
	singles := 0
	doubles := 0
	unquoted := true
	for _, c := range s {
		if c == '&' {
			amps++
		} else if c == '"' {
			doubles++
		} else if c == '\'' {
			singles++
		} else if unquoted && (c == '/' || c == '`' || c == '<' || c == '=' || c == '>' || isWhitespace(c)) {
			// TODO: allow slash
			unquoted = false
		}
	}

	if !unquoted || doubles > 0 || singles > 0 {
		// quoted
		c := amps + doubles
		quote := byte('"')
		escapedQuote := []byte("&#34;")
		if doubles > singles {
			c = amps + singles
			quote = byte('\'')
			escapedQuote = []byte("&#39;")
		}

		t := make([]byte, len(s)+c*4+2)
		t[0] = quote
		w := 1
		start := 0
		for j, x := range s {
			if x == '&' {
				w += copy(t[w:], s[start:j])
				w += copy(t[w:], []byte("&amp;"))
				start = j + 1
			} else if x == quote {
				w += copy(t[w:], s[start:j])
				w += copy(t[w:], escapedQuote)
				start = j + 1
			}
		}
		copy(t[w:], s[start:])
		t[len(t)-1] = quote
		return t
	} else {
		// unquoted
		c := amps

		t := make([]byte, len(s)+c*4)
		w := 0
		start := 0
		for j, x := range s {
			if x == '&' {
				w += copy(t[w:], s[start:j])
				w += copy(t[w:], []byte("&amp;"))
				start = j + 1
			}
		}
		copy(t[w:], s[start:])
		return t
	}
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
		return bytes.Equal(attrVal[1:len(attrVal)-1], match)
	}
	return bytes.Equal(attrVal, match)
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
			for {
				if attr := tb.shift(); attr.tt == html.StartTagCloseToken || attr.tt == html.StartTagVoidToken || attr.tt == html.ErrorToken {
					break
				}
			}
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
				if rawTag == html.Style || rawTag == html.Script {
					var mediatype string
					if rawTagMediatype != nil {
						mediatype = string(rawTagMediatype)
					} else if rawTag == html.Script {
						mediatype = defaultScriptType
					} else {
						mediatype = defaultStyleType
					}
					if err := m.Minify(mediatype, w, bytes.NewBuffer(t.data)); err != nil {
						if err == minify.ErrNotExist {
							// no minifier, write the original
							if _, err := w.Write(t.data); err != nil {
								return err
							}
						} else {
							return err
						}
					}
				} else if rawTag == html.Noscript {
					if err := Minify(m, "text/html", w, bytes.NewBuffer(t.data)); err != nil {
						return err
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
					var httpEquiv *token
					var name *token
					var content *token
					hasCharset := false
					i := 0
					for {
						attr := tb.peek(i)
						i++
						if attr.tt != html.AttributeToken {
							break
						}
						if attr.hash == html.Content {
							content = attr
						} else if attr.hash == html.Http_Equiv {
							httpEquiv = attr
						} else if attr.hash == html.Name {
							name = attr
						} else if attr.hash == html.Charset {
							hasCharset = true
						}
					}
					if content != nil {
						content.attrVal = normalizeContentType(content.attrVal)
						if httpEquiv != nil {
							if !hasCharset && attrValEqual(httpEquiv.attrVal, []byte("content-type")) && attrValEqual(content.attrVal, []byte("text/html;charset=utf-8")) {
								httpEquiv.data = nil
								content.data = []byte("charset")
								content.hash = html.Charset
								content.attrVal = []byte("utf-8")
							} else if attrValEqual(httpEquiv.attrVal, []byte("content-style-type")) {
								defaultStyleType = string(content.attrVal)
							} else if attrValEqual(httpEquiv.attrVal, []byte("content-script-type")) {
								defaultScriptType = string(content.attrVal)
							}
						}
						if name != nil {
							parse.ToLower(name.attrVal)
							if attrValEqual(name.attrVal, []byte("keywords")) {
								content.attrVal = bytes.Replace(content.attrVal, []byte(", "), []byte(","), -1)
							} else if attrValEqual(name.attrVal, []byte("viewport")) {
								content.attrVal = bytes.Replace(content.attrVal, []byte(" "), []byte(""), -1)
							}
						}
					}
				}
			}

			// write tag
			if t.tt == html.EndTagToken {
				if _, err := w.Write([]byte("</")); err != nil {
					return err
				}
			} else {
				if _, err := w.Write([]byte("<")); err != nil {
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
					if len(val) > 1 && val[0] == '"' || val[0] == '\'' {
						val = bytes.TrimSpace(val[1:len(val)-1])
					}
					if caseInsensitiveAttrMap[attr.hash] {
						parse.ToLower(val)
						if attr.hash == html.Enctype || attr.hash == html.Codetype || attr.hash == html.Accept || attr.hash == html.Type && (t.hash == html.A || t.hash == html.Link || t.hash == html.Object || t.hash == html.Param || t.hash == html.Script || t.hash == html.Style || t.hash == html.Source) {
							val = normalizeContentType(val)
						}
					}
					if rawTag != 0 && attr.hash == html.Type {
						rawTagMediatype = val
					}

					// default attribute values can be ommited
					if attr.hash == html.Type && (t.hash == html.Script && bytes.Equal(val, []byte("text/javascript")) ||
						t.hash == html.Style && bytes.Equal(val, []byte("text/css")) ||
						t.hash == html.Link && bytes.Equal(val, []byte("text/css")) ||
						t.hash == html.Input && bytes.Equal(val, []byte("text")) ||
						t.hash == html.Button && bytes.Equal(val, []byte("submit"))) ||
						attr.hash == html.Method && bytes.Equal(val, []byte("get")) ||
						attr.hash == html.Enctype && bytes.Equal(val, []byte("application/x-www-form-urlencoded")) ||
						attr.hash == html.Colspan && bytes.Equal(val, []byte("1")) ||
						attr.hash == html.Rowspan && bytes.Equal(val, []byte("1")) ||
						attr.hash == html.Shape && bytes.Equal(val, []byte("rect")) ||
						attr.hash == html.Span && bytes.Equal(val, []byte("1")) ||
						attr.hash == html.Clear && bytes.Equal(val, []byte("none")) ||
						attr.hash == html.Frameborder && bytes.Equal(val, []byte("1")) ||
						attr.hash == html.Scrolling && bytes.Equal(val, []byte("auto")) ||
						attr.hash == html.Valuetype && bytes.Equal(val, []byte("data")) ||
						attr.hash == html.Language && t.hash == html.Script && bytes.Equal(val, []byte("javascript")) {
						continue
					}
					if _, err := w.Write([]byte(" ")); err != nil {
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

						var err error
						if _, err := w.Write([]byte("=")); err != nil {
							return err
						}

						// CSS and JS minifiers for attribute inline code
						if attr.hash == html.Style {
							b := &bytes.Buffer{}
							b.Grow(len(val))
							if err = m.Minify(defaultStyleType, b, bytes.NewReader(val)); err != nil {
								if err != minify.ErrNotExist {
									return err
								}
							} else {
								val = b.Bytes()
							}
						} else if len(attr.data) > 2 && attr.data[0] == 'o' && attr.data[1] == 'n' {
							if len(val) >= 11 && bytes.Equal(parse.CopyToLower(val[:11]), []byte("javascript:")) {
								val = val[11:]
							}
							b := &bytes.Buffer{}
							b.Grow(len(val))
							if err = m.Minify(defaultScriptType, b, bytes.NewReader(val)); err != nil {
								if err != minify.ErrNotExist {
									return err
								}
							} else {
								val = b.Bytes()
							}
						} else if urlAttrMap[attr.hash] {
							if len(val) >= 5 && bytes.Equal(parse.CopyToLower(val[:5]), []byte("http:")) {
								val = val[5:]
							}
						}

						// no quotes if possible, else prefer single or double depending on which occurs more often in value
						val = escapeAttrVal(val)
						if _, err := w.Write(val); err != nil {
							return err
						}
					}
				}
			}
			if _, err := w.Write([]byte(">")); err != nil {
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

		var attrVal []byte = nil
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
