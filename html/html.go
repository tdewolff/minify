package html // import "github.com/tdewolff/minify/html"

import (
	"bytes"
	"io"

	"github.com/tdewolff/minify"
	hash "github.com/tdewolff/parse/html"
	"golang.org/x/net/html"
)

var specialTagMap = map[hash.Hash]bool{
	hash.Code:     true,
	hash.Noscript: true,
	hash.Pre:      true,
	hash.Script:   true,
	hash.Style:    true,
	hash.Textarea: true,
}

var inlineTagMap = map[hash.Hash]bool{
	hash.A:       true,
	hash.Abbr:    true,
	hash.Acronym: true,
	hash.B:       true,
	hash.Bdo:     true,
	hash.Big:     true,
	hash.Cite:    true,
	hash.Button:  true,
	hash.Dfn:     true,
	hash.Em:      true,
	hash.I:       true,
	hash.Img:     true,
	hash.Input:   true,
	hash.Kbd:     true,
	hash.Label:   true,
	hash.Map:     true,
	hash.Object:  true,
	hash.Q:       true,
	hash.Samp:    true,
	hash.Select:  true,
	hash.Small:   true,
	hash.Span:    true,
	hash.Strong:  true,
	hash.Sub:     true,
	hash.Sup:     true,
	hash.Tt:      true,
	hash.Var:     true,
}

var booleanAttrMap = map[hash.Hash]bool{
	hash.Allowfullscreen: true,
	hash.Async:           true,
	hash.Autofocus:       true,
	hash.Autoplay:        true,
	hash.Checked:         true,
	hash.Compact:         true,
	hash.Controls:        true,
	hash.Declare:         true,
	hash.Default:         true,
	hash.DefaultChecked:  true,
	hash.DefaultMuted:    true,
	hash.DefaultSelected: true,
	hash.Defer:           true,
	hash.Disabled:        true,
	hash.Draggable:       true,
	hash.Enabled:         true,
	hash.Formnovalidate:  true,
	hash.Hidden:          true,
	hash.Inert:           true,
	hash.Ismap:           true,
	hash.Itemscope:       true,
	hash.Multiple:        true,
	hash.Muted:           true,
	hash.Nohref:          true,
	hash.Noresize:        true,
	hash.Noshade:         true,
	hash.Novalidate:      true,
	hash.Nowrap:          true,
	hash.Open:            true,
	hash.Pauseonexit:     true,
	hash.Readonly:        true,
	hash.Required:        true,
	hash.Reversed:        true,
	hash.Scoped:          true,
	hash.Seamless:        true,
	hash.Selected:        true,
	hash.Sortable:        true,
	hash.Spellcheck:      true,
	hash.Translate:       true,
	hash.Truespeed:       true,
	hash.Typemustmatch:   true,
	hash.Undeterminate:   true,
	hash.Visible:         true,
}

var caseInsensitiveAttrMap = map[hash.Hash]bool{
	hash.Accept_Charset: true,
	hash.Accept:         true,
	hash.Align:          true,
	hash.Alink:          true,
	hash.Axis:           true,
	hash.Bgcolor:        true,
	hash.Charset:        true,
	hash.Clear:          true,
	hash.Codetype:       true,
	hash.Color:          true,
	hash.Dir:            true,
	hash.Enctype:        true,
	hash.Face:           true,
	hash.Frame:          true,
	hash.Hreflang:       true,
	hash.Http_Equiv:     true,
	hash.Lang:           true,
	hash.Language:       true,
	hash.Link:           true,
	hash.Media:          true,
	hash.Method:         true,
	hash.Rel:            true,
	hash.Rev:            true,
	hash.Rules:          true,
	hash.Scope:          true,
	hash.Scrolling:      true,
	hash.Shape:          true,
	hash.Target:         true,
	hash.Text:           true,
	hash.Type:           true,
	hash.Valign:         true,
	hash.Valuetype:      true,
	hash.Vlink:          true,
}

var urlAttrMap = map[hash.Hash]bool{
	hash.Action:     true,
	hash.Background: true,
	hash.Cite:       true,
	hash.Classid:    true,
	hash.Codebase:   true,
	hash.Data:       true,
	hash.Formaction: true,
	hash.Href:       true,
	hash.Icon:       true,
	hash.Longdesc:   true,
	hash.Manifest:   true,
	hash.Poster:     true,
	hash.Profile:    true,
	hash.Src:        true,
	hash.Usemap:     true,
	hash.Xmlns:      true,
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
	s = bytes.ToLower(bytes.TrimSpace(replaceMultipleWhitespace(s)))
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

// escapeText escapes ampersands.
func escapeText(s []byte) []byte {
	amps := 0
	for _, x := range s {
		if x == '&' {
			amps++
		}
	}

	t := make([]byte, len(s)+amps*4)
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

// escapeAttrVal returns the escape attribute value bytes without quotes.
func escapeAttrVal(s []byte) []byte {
	if len(s) == 0 {
		return []byte("\"\"")
	}

	amps := 0
	singles := 0
	doubles := 0
	unquoted := true
	for _, x := range s {
		if x == '&' {
			amps++
		} else if x == '"' {
			doubles++
		} else if x == '\'' {
			singles++
		} else if unquoted && (x == '/' || x == '`' || x == '<' || x == '=' || x == '>' || isWhitespace(x)) {
			// no slash either because it causes difficulties!
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
func isWhitespace(x byte) bool {
	return x == ' ' || x == '\t' || x == '\n' || x == '\r' || x == '\f'
}

// copyBytes copies bytes to the same position.
// This is required because the referenced slices from the tokenizer might be overwritten on subsequent Next calls.
func copyBytes(src []byte) []byte {
	dst := make([]byte, len(src))
	copy(dst, src)
	return dst
}

////////////////////////////////////////////////////////////////

type attribute struct {
	key         hash.Hash
	keyRaw, val []byte
}

type token struct {
	tt       html.TokenType
	token    hash.Hash
	tokenRaw []byte
	text     []byte
	attr     []attribute
	attrKey  map[hash.Hash]int
}

func (t *token) getAttrVal(a hash.Hash) []byte {
	if i, ok := t.attrKey[a]; ok {
		return t.attr[i].val
	}
	return []byte{}
}

type tokenFeed struct {
	z   *html.Tokenizer
	buf []*token
}

func newTokenFeed(z *html.Tokenizer) *tokenFeed {
	return &tokenFeed{z: z}
}

func (tf *tokenFeed) shift() *token {
	if len(tf.buf) > 0 {
		tf.buf = tf.buf[1:]
	}
	return tf.peek(0)
}

func (tf *tokenFeed) peek(pos int) *token {
	if pos == len(tf.buf) {
		t := &token{tf.z.Next(), 0, nil, nil, nil, nil}
		switch t.tt {
		case html.TextToken, html.CommentToken, html.DoctypeToken:
			t.text = copyBytes(tf.z.Text())
		case html.StartTagToken, html.SelfClosingTagToken, html.EndTagToken:
			var moreAttr bool
			var keyRaw, val []byte
			t.tokenRaw, moreAttr = tf.z.TagName()
			t.tokenRaw = copyBytes(t.tokenRaw)
			t.token = hash.ToHash(t.tokenRaw)
			if moreAttr {
				t.attr = make([]attribute, 0, 3)
				t.attrKey = make(map[hash.Hash]int)
				for moreAttr {
					keyRaw, val, moreAttr = tf.z.TagAttr()
					key := hash.ToHash(keyRaw)
					t.attr = append(t.attr, attribute{key, copyBytes(keyRaw), copyBytes(bytes.TrimSpace(val))})
					t.attrKey[key] = len(t.attr) - 1
				}
			}
		}
		tf.buf = append(tf.buf, t)
		return t
	}
	return tf.buf[pos]
}

func (tf tokenFeed) err() error {
	return tf.z.Err()
}

////////////////////////////////////////////////////////////////

// Minify minifies HTML5 files, it reads from r and writes to w.
// Removes unnecessary whitespace, tags, attributes, quotes and comments and typically saves 10% in size.
func Minify(m minify.Minifier, _ string, w io.Writer, r io.Reader) error {
	var specialTag []*token // stack array of special tags it is in
	precededBySpace := true // on true the next text token must not start with a space
	defaultScriptType := "text/javascript"
	defaultStyleType := "text/css"

	tf := newTokenFeed(html.NewTokenizer(r))
	for {
		t := tf.shift()
		switch t.tt {
		case html.ErrorToken:
			if tf.err() == io.EOF {
				return nil
			}
			return tf.err()
		case html.DoctypeToken:
			if _, err := w.Write([]byte("<!doctype html>")); err != nil {
				return err
			}
		case html.CommentToken:
			// TODO: ensure that nested comments are handled properly (tokenizer doesn't handle this!)
			var text []byte
			if bytes.HasPrefix(t.text, []byte("[if")) {
				text = append(append([]byte("<!--"), t.text...), []byte("-->")...)
			} else if bytes.HasSuffix(t.text, []byte("--")) {
				// only occurs when mixed up with conditional comments
				text = append(append([]byte("<!"), t.text...), '>')
			}
			if _, err := w.Write(text); err != nil {
				return err
			}
		case html.TextToken:
			// CSS and JS minifiers for inline code
			if len(specialTag) > 0 {
				token := specialTag[len(specialTag)-1].token
				if token == hash.Style || token == hash.Script {
					var mediatype string
					mediatypeRaw := specialTag[len(specialTag)-1].getAttrVal(hash.Type)
					if len(mediatypeRaw) > 0 {
						mediatype = string(mediatypeRaw)
					} else if token == hash.Script {
						mediatype = defaultScriptType
					} else {
						mediatype = defaultStyleType
					}
					if err := m.Minify(mediatype, w, bytes.NewBuffer(t.text)); err != nil {
						if err == minify.ErrNotExist {
							// no minifier, write the original
							if _, err := w.Write(t.text); err != nil {
								return err
							}
						} else {
							return err
						}
					}
				} else if token == hash.Noscript {
					if err := Minify(m, "text/html", w, bytes.NewBuffer(t.text)); err != nil {
						return err
					}
				} else if _, err := w.Write(t.text); err != nil {
					return err
				}
			} else if t.text = escapeText(replaceMultipleWhitespace(t.text)); len(t.text) > 0 {
				// whitespace removal; trim left
				if t.text[0] == ' ' && precededBySpace {
					t.text = t.text[1:]
				}

				// whitespace removal; trim right
				precededBySpace = false
				if len(t.text) == 0 {
					precededBySpace = true
				} else if t.text[len(t.text)-1] == ' ' {
					precededBySpace = true
					trim := false
					i := 1
					for {
						next := tf.peek(i)
						// trim if EOF, text token with whitespace begin or block token
						if next.tt == html.ErrorToken {
							trim = true
							break
						} else if next.tt == html.TextToken {
							// remove if the text token starts with a whitespace
							trim = (len(next.text) > 0 && isWhitespace(next.text[0]))
							break
						} else if next.tt == html.StartTagToken || next.tt == html.EndTagToken || next.tt == html.SelfClosingTagToken {
							if !inlineTagMap[next.token] {
								trim = true
								break
							} else if next.tt == html.StartTagToken {
								break
							}
						}
						i++
					}
					if trim {
						t.text = t.text[:len(t.text)-1]
						precededBySpace = false
					}
				}
				if _, err := w.Write(t.text); err != nil {
					return err
				}
			}
		case html.StartTagToken, html.EndTagToken, html.SelfClosingTagToken:
			if !inlineTagMap[t.token] {
				precededBySpace = true
				if specialTagMap[t.token] {
					if t.tt == html.StartTagToken {
						specialTag = append(specialTag, t)
					} else if t.tt == html.EndTagToken && len(specialTag) > 0 && specialTag[len(specialTag)-1].token == t.token {
						specialTag = specialTag[:len(specialTag)-1]
					}

					// ignore empty script and style tags
					if t.attr == nil && (t.token == hash.Script || t.token == hash.Style) {
						if next := tf.peek(1); next.tt == html.EndTagToken {
							tf.shift()
							break
						}
					}
				}

				// remove superfluous ending tags
				if t.attr == nil && (t.token == hash.Html || t.token == hash.Head || t.token == hash.Body ||
					t.tt == html.EndTagToken && (t.token == hash.Td || t.token == hash.Tr || t.token == hash.Th || t.token == hash.Thead || t.token == hash.Tbody || t.token == hash.Tfoot ||
						t.token == hash.Option || t.token == hash.Colgroup || t.token == hash.Dd || t.token == hash.Dt)) {
					break
				} else if t.tt == html.EndTagToken && (t.token == hash.P || t.token == hash.Li) {
					remove := false
					i := 1
					for {
						next := tf.peek(i)
						i++
						// continue if text token is empty or whitespace
						if next.tt == html.TextToken && (len(next.text) == 0 || isWhitespace(next.text[0]) && len(replaceMultipleWhitespace(next.text)) == 1) {
							continue
						}
						remove = (next.tt == html.ErrorToken || next.tt == html.EndTagToken || next.tt == html.StartTagToken && next.token == t.token)
						break
					}
					if remove {
						break
					}
				}

				// rewrite meta tag with charset
				if t.attr != nil && t.token == hash.Meta {
					if _, ok := t.attrKey[hash.Charset]; !ok {
						if iHttpEquiv, ok := t.attrKey[hash.Http_Equiv]; ok && bytes.Equal(bytes.ToLower(t.attr[iHttpEquiv].val), []byte("content-type")) {
							if iContent, ok := t.attrKey[hash.Content]; ok && bytes.Equal(normalizeContentType(t.attr[iContent].val), []byte("text/html;charset=utf-8")) {
								delete(t.attrKey, hash.Http_Equiv)
								delete(t.attrKey, hash.Content)
								t.attr = append(t.attr[:iContent], t.attr[iContent+1:]...)
								t.attr[iHttpEquiv] = attribute{hash.Charset, []byte("charset"), []byte("utf-8")}
								t.attrKey[hash.Charset] = iHttpEquiv
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
			if _, err := w.Write(t.tokenRaw); err != nil {
				return err
			}

			// write attributes
			for _, attr := range t.attr {
				val := attr.val
				if caseInsensitiveAttrMap[attr.key] {
					val = bytes.ToLower(val)
					if attr.key == hash.Enctype || attr.key == hash.Codetype || attr.key == hash.Accept || attr.key == hash.Type && (t.token == hash.A || t.token == hash.Link || t.token == hash.Object || t.token == hash.Param || t.token == hash.Script || t.token == hash.Style || t.token == hash.Source) {
						val = normalizeContentType(val)
					}
				}

				// default attribute values can be ommited
				if attr.key == hash.Type && (t.token == hash.Script && bytes.Equal(val, []byte("text/javascript")) ||
					t.token == hash.Style && bytes.Equal(val, []byte("text/css")) ||
					t.token == hash.Link && bytes.Equal(val, []byte("text/css")) ||
					t.token == hash.Input && bytes.Equal(val, []byte("text")) ||
					t.token == hash.Button && bytes.Equal(val, []byte("submit"))) ||
					attr.key == hash.Method && bytes.Equal(val, []byte("get")) ||
					attr.key == hash.Enctype && bytes.Equal(val, []byte("application/x-www-form-urlencoded")) ||
					attr.key == hash.Colspan && bytes.Equal(val, []byte("1")) ||
					attr.key == hash.Rowspan && bytes.Equal(val, []byte("1")) ||
					attr.key == hash.Shape && bytes.Equal(val, []byte("rect")) ||
					attr.key == hash.Span && bytes.Equal(val, []byte("1")) ||
					attr.key == hash.Clear && bytes.Equal(val, []byte("none")) ||
					attr.key == hash.Frameborder && bytes.Equal(val, []byte("1")) ||
					attr.key == hash.Scrolling && bytes.Equal(val, []byte("auto")) ||
					attr.key == hash.Valuetype && bytes.Equal(val, []byte("data")) ||
					attr.key == hash.Language && t.token == hash.Script && bytes.Equal(val, []byte("javascript")) {
					continue
				}
				if _, err := w.Write([]byte(" ")); err != nil {
					return err
				}
				if _, err := w.Write(attr.keyRaw); err != nil {
					return err
				}

				// booleans have no value
				if !booleanAttrMap[attr.key] {
					if len(val) == 0 {
						continue
					}

					var err error
					if _, err := w.Write([]byte("=")); err != nil {
						return err
					}

					// CSS and JS minifiers for attribute inline code
					if attr.key == hash.Style {
						b := &bytes.Buffer{}
						b.Grow(len(val))
						if err = m.Minify(defaultStyleType, b, bytes.NewReader(val)); err != nil {
							if err != minify.ErrNotExist {
								return err
							}
						} else {
							val = b.Bytes()
						}
					} else if len(attr.keyRaw) > 2 && attr.keyRaw[0] == 'o' && attr.keyRaw[1] == 'n' {
						if len(val) >= 11 && bytes.Equal(bytes.ToLower(val[:11]), []byte("javascript:")) {
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
					} else if urlAttrMap[attr.key] {
						if len(val) >= 5 && bytes.Equal(bytes.ToLower(val[:5]), []byte("http:")) {
							val = val[5:]
						}
					} else if t.token == hash.Meta && attr.key == hash.Content {
						httpEquiv := t.getAttrVal(hash.Http_Equiv)
						if bytes.Equal(httpEquiv, []byte("content-type")) {
							val = normalizeContentType(val)
						} else if bytes.Equal(httpEquiv, []byte("content-style-type")) {
							defaultStyleType = string(normalizeContentType(val))
						} else if bytes.Equal(httpEquiv, []byte("content-script-type")) {
							defaultScriptType = string(normalizeContentType(val))
						}

						name := bytes.ToLower(t.getAttrVal(hash.Name))
						if bytes.Equal(name, []byte("keywords")) {
							val = bytes.Replace(val, []byte(", "), []byte(","), -1)
						} else if bytes.Equal(name, []byte("viewport")) {
							val = bytes.Replace(val, []byte(" "), []byte(""), -1)
						}
					}

					// no quotes if possible, else prefer single or double depending on which occurs more often in value
					val = escapeAttrVal(val)
					if _, err := w.Write(val); err != nil {
						return err
					}
				}
			}
			if _, err := w.Write([]byte(">")); err != nil {
				return err
			}
		}
	}
}
