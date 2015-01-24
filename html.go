package minify

import (
	"bytes"
	"io"

	"code.google.com/p/go.net/html"
	"code.google.com/p/go.net/html/atom"
)

var specialTagMap = map[atom.Atom]bool{
	atom.Style:    true,
	atom.Script:   true,
	atom.Pre:      true,
	atom.Code:     true,
	atom.Textarea: true,
	atom.Noscript: true,
}

var inlineTagMap = map[atom.Atom]bool{
	atom.B:       true,
	atom.Big:     true,
	atom.I:       true,
	atom.Small:   true,
	atom.Tt:      true,
	atom.Abbr:    true,
	//atom.Acronym: true,
	atom.Cite:    true,
	atom.Dfn:     true,
	atom.Em:      true,
	atom.Kbd:     true,
	atom.Strong:  true,
	atom.Samp:    true,
	atom.Var:     true,
	atom.A:       true,
	atom.Bdo:     true,
	atom.Img:     true,
	atom.Map:     true,
	atom.Object:  true,
	atom.Q:       true,
	atom.Span:    true,
	atom.Sub:     true,
	atom.Sup:     true,
	atom.Button:  true,
	atom.Input:   true,
	atom.Label:   true,
	atom.Select:  true,
}

var booleanAttrMap = map[atom.Atom]bool{
	//atom.Allowfullscreen: true,
	atom.Async:           true,
	atom.Autofocus:       true,
	atom.Autoplay:        true,
	atom.Checked:         true,
	//atom.Compact:         true,
	atom.Controls:        true,
	//atom.Declare:         true,
	atom.Default:         true,
	//atom.DefaultChecked:  true,
	//atom.DefaultMuted:    true,
	//atom.DefaultSelected: true,
	atom.Defer:           true,
	atom.Disabled:        true,
	atom.Draggable:       true,
	//atom.Enabled:         true,
	atom.Formnovalidate:  true,
	atom.Hidden:          true,
	//atom.Undeterminate:   true,
	atom.Inert:           true,
	atom.Ismap:           true,
	atom.Itemscope:       true,
	atom.Multiple:        true,
	atom.Muted:           true,
	//atom.Nohref:          true,
	//atom.Noresize:        true,
	//atom.Noshade:         true,
	atom.Novalidate:      true,
	//atom.Nowrap:          true,
	atom.Open:            true,
	//atom.Pauseonexit:     true,
	atom.Readonly:        true,
	atom.Required:        true,
	atom.Reversed:        true,
	atom.Scoped:          true,
	atom.Seamless:        true,
	atom.Selected:        true,
	//atom.Sortable:        true,
	atom.Spellcheck:      true,
	atom.Translate:       true,
	//atom.Truespeed:       true,
	atom.Typemustmatch:   true,
	//atom.Visible:         true,
}

var caseInsensitiveAttrMap = map[atom.Atom]bool{
	atom.AcceptCharset:  true,
	atom.Accept:         true,
	atom.Align:          true,
	//atom.Alink:          true,
	//atom.Axis:           true,
	//atom.Bgcolor:        true,
	atom.Charset:        true,
	//atom.Clear:          true,
	//atom.Codetype:       true,
	atom.Color:          true,
	atom.Dir:            true,
	atom.Enctype:        true,
	atom.Face:           true,
	atom.Frame:          true,
	atom.Hreflang:       true,
	atom.HttpEquiv:      true,
	atom.Lang:           true,
	//atom.Language:       true,
	atom.Link:           true,
	atom.Media:          true,
	atom.Method:         true,
	atom.Rel:            true,
	//atom.Rev:            true,
	//atom.Rules:          true,
	atom.Scope:          true,
	//atom.Scrolling:      true,
	atom.Shape:          true,
	atom.Target:         true,
	//atom.Text:           true,
	atom.Type:           true,
	//atom.Valign:         true,
	//atom.Valuetype:      true,
	//atom.Vlink:          true,
}

var urlAttrMap = map[atom.Atom]bool{
	atom.Href:       true,
	atom.Src:        true,
	atom.Cite:       true,
	atom.Action:     true,
	//atom.Profile:    true,
	//atom.Xmlns:      true,
	atom.Formaction: true,
	atom.Poster:     true,
	atom.Manifest:   true,
	atom.Icon:       true,
	//atom.Codebase:   true,
	//atom.Longdesc:   true,
	//atom.Background: true,
	//atom.Classid:    true,
	atom.Usemap:     true,
	atom.Data:       true,
}

////////////////////////////////////////////////////////////////

// replaceMultipleWhitespace replaces any series of whitespace characters by a single space
func replaceMultipleWhitespace(s []byte) []byte {
	j := 0
	t := make([]byte, len(s))
	previousSpace := false
	for _, x := range s {
		if x == ' ' || x == '\n' || x == '\r' || x == '\t' || x == '\f' {
			if !previousSpace {
				previousSpace = true
				t[j] = ' '
				j++
			}
		} else {
			previousSpace = false
			t[j] = x
			j++
		}
	}
	return t[:j]
}

// isValidUnquotedAttr returns true when the bytes can be unquoted as an HTML attribute
func isValidUnquotedAttr(s []byte) bool {
	for _, x := range s {
		if x == ' ' || x == '/' || x == '"' || x == '\'' || x == '`' || x >= '<' && x <= '>' || x >= '\n' && x <= '\r' {
			return false
		}
	}
	return true
}

func moreDoubleQuotes(s []byte) bool {
	singles := 0
	doubles := 0
	for _, x := range s {
		if x == '"' {
			doubles++
		} else if x == '\'' {
			singles++
		}
	}
	return doubles > singles
}

////////////////////////////////////////////////////////////////

type attribute struct {
	key atom.Atom
	keyRaw, val []byte
}

type token struct {
	tt    html.TokenType
	token atom.Atom
	tokenRaw []byte
	text  []byte
	attr  []attribute
	attrKey  map[atom.Atom]int
}

func (t *token) getAttrVal(a atom.Atom) []byte {
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

func deepCopy(src []byte) []byte {
	dst := make([]byte, len(src))
	copy(dst, src)
	return dst
}

func (tf *tokenFeed) peek(pos int) *token {
	if pos == len(tf.buf) {
		if len(tf.buf) > 0 {
			t := tf.buf[len(tf.buf)-1]
			t.tokenRaw = deepCopy(t.tokenRaw)
			t.text = deepCopy(t.text)
			for _, attr := range t.attr {
				attr.keyRaw = deepCopy(attr.keyRaw)
				attr.val = deepCopy(attr.val)
			}
		}

		t := &token{tf.z.Next(), 0, nil, nil, nil, nil}
		switch t.tt {
		case html.TextToken, html.CommentToken, html.DoctypeToken:
			t.text = tf.z.Text()
			if t.tt == html.TextToken {
				t.text = replaceMultipleWhitespace(t.text)
			}
		case html.StartTagToken, html.SelfClosingTagToken, html.EndTagToken:
			var moreAttr bool
			var keyRaw, val []byte
			t.tokenRaw, moreAttr = tf.z.TagName()
			t.token = atom.Lookup(t.tokenRaw)
			if moreAttr {
				t.attr = []attribute{}
				t.attrKey = make(map[atom.Atom]int)
				for moreAttr {
					keyRaw, val, moreAttr = tf.z.TagAttr()
					key := atom.Lookup(keyRaw)
					t.attr = append(t.attr, attribute{key, keyRaw, bytes.TrimSpace(val)})
					t.attrKey[key] = len(t.attr)-1
				}
			}
		}
		tf.buf = append(tf.buf, t)
		return t
	}
	return tf.buf[pos]
}

////////////////////////////////////////////////////////////////

// HTML minifies HTML5 files, it reads from r and writes to w.
// Removes unnecessary whitespace, tags, attributes, quotes and comments and typically saves 10% in size.
func (m Minifier) HTML(w io.Writer, r io.Reader) error {
	var prevText []byte         // write prevText token until next token is received, allows to look forward one token before writing away
	var specialTag []*token // stack array of special tags it is in
	var prevTagToken *token
	precededBySpace := true // on true the next prevText token must no start with a space
	defaultScriptType := "text/javascript"
	defaultStyleType := "text/css"

	z := html.NewTokenizer(r)
	tf := newTokenFeed(z)
	for {
		t := tf.shift()
		switch t.tt {
		case html.ErrorToken:
			if z.Err() == io.EOF {
				if _, err := w.Write(prevText); err != nil {
					return err
				}
				return nil
			}
			return z.Err()
		case html.DoctypeToken:
			if _, err := w.Write(bytes.TrimSpace(prevText)); err != nil {
				return err
			}
			prevText = nil
			if _, err := w.Write([]byte("<!doctype html>")); err != nil {
				return err
			}
		case html.CommentToken:
			if _, err := w.Write(prevText); err != nil {
				return err
			}
			prevText = nil

			comment := t.text
			// TODO: ensure that nested comments are handled properly (tokenizer doesn't handle this!)
			if bytes.HasPrefix(comment, []byte("[if")) {
				prevText = append(append([]byte("<!--"), comment...), []byte("-->")...)
			} else if bytes.HasSuffix(comment, []byte("--")) {
				// only occurs when mixed up with conditional comments
				prevText = append(append([]byte("<!"), comment...), []byte(">")...)
			}
		case html.TextToken:
			if _, err := w.Write(prevText); err != nil {
				return err
			}
			prevText = nil

			// CSS and JS minifiers for inline code
			if len(specialTag) > 0 {
				token := specialTag[len(specialTag)-1].token
				if token == atom.Style || token == atom.Script {
					var mediatype string
					mediatypeRaw := specialTag[len(specialTag)-1].getAttrVal(atom.Type)
					if len(mediatypeRaw) > 0 {
						mediatype = string(mediatypeRaw)
					} else if token == atom.Script {
						mediatype = defaultScriptType
					} else {
						mediatype = defaultStyleType
					}

					if err := m.Minify(mediatype, w, bytes.NewBuffer(t.text)); err != nil {
						if err == ErrNotExist {
							// no minifier, write the original
							if _, err := w.Write(t.text); err != nil {
								return err
							}
						} else {
							return err
						}
					}
				} else if token == atom.Noscript {
					if err := m.HTML(w, bytes.NewBuffer(t.text)); err != nil {
						return err
					}
				} else if _, err := w.Write(t.text); err != nil {
					return err
				}
				break
			}

			// whitespace removal; if after an inline element, trim left if precededBySpace
			prevText = t.text
			if prevTagToken != nil && inlineTagMap[prevTagToken.token] {
				if precededBySpace && len(prevText) > 0 && prevText[0] == ' ' {
					prevText = prevText[1:]
				}
				precededBySpace = len(prevText) > 0 && prevText[len(prevText)-1] == ' '
			} else if len(prevText) > 0 && prevText[0] == ' ' {
				prevText = prevText[1:]
			}
		case html.StartTagToken, html.EndTagToken, html.SelfClosingTagToken:
			prevTagToken = t

			if specialTagMap[t.token] {
				if t.tt == html.StartTagToken {
					specialTag = append(specialTag, t)
				} else if t.tt == html.EndTagToken && len(specialTag) > 0 && specialTag[len(specialTag)-1].token == t.token {
					specialTag = specialTag[:len(specialTag)-1]
				}
			}

			// whitespace removal; if we encounter a block or a (closing) inline element, trim the right
			if !inlineTagMap[t.token] || (t.tt == html.EndTagToken && len(prevText) > 0 && prevText[len(prevText)-1] == ' ') {
				precededBySpace = true
				// do not remove when next token is text and doesn't start with a space
				if len(prevText) > 0 {
					trim := false
					i := 0
					for {
						nextT := tf.peek(i)
						// remove if the tag is not an inline tag (but a block tag)
						if nextT.tt == html.ErrorToken || ((nextT.tt == html.StartTagToken || nextT.tt == html.EndTagToken || nextT.tt == html.SelfClosingTagToken) && !inlineTagMap[nextT.token]) {
							trim = true
							break
						} else if nextT.tt == html.TextToken {
							// remove if the text token starts with a whitespace
							trim = len(nextT.text) > 0 && nextT.text[0] == ' '
							break
						}
						i++
					}
					if trim {
						prevText = bytes.TrimRight(prevText, " ")
						precededBySpace = false
					}
				}
			}
			if _, err := w.Write(prevText); err != nil {
				return err
			}
			prevText = nil

			if t.attr == nil && (t.token == atom.Body || t.token == atom.Head || t.token == atom.Html ||
				t.tt == html.EndTagToken && (t.token == atom.Colgroup || t.token == atom.Dd || t.token == atom.Dt ||
					t.token == atom.Option || t.token == atom.Td || t.token == atom.Tfoot ||
					t.token == atom.Th || t.token == atom.Thead || t.token == atom.Tbody || t.token == atom.Tr)) {
				break
			} else if t.tt == html.EndTagToken && (t.token == atom.P || t.token == atom.Li) {
				remove := false
				i := 1
				for {
					nextT := tf.peek(i)
					// continue if text token is empty or whitespace
					if nextT.tt != html.TextToken || (len(nextT.text) > 0 && string(nextT.text) != " ") { // TODO: could write len == 1 and byte 0 == space
						// remove only when encountering EOF, end tag (from parent) or a start tag of the same tag
						remove = (nextT.tt == html.ErrorToken || nextT.tt == html.EndTagToken || (nextT.tt == html.StartTagToken && nextT.token == t.token))
						break
					}
					i++
				}
				if remove {
					break
				}
			}
			if t.token == atom.Script || t.token == atom.Style {
				if nextT := tf.peek(1); nextT.tt == html.EndTagToken {
					tf.shift()
					break
				}
			}

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

			if t.attr != nil && t.token == atom.Meta && bytes.Equal(t.getAttrVal(atom.HttpEquiv), []byte("content-type")) &&
				bytes.Equal(bytes.ToLower(t.getAttrVal(atom.Content)), []byte("text/html; charset=utf-8")) {
				if _, err := w.Write([]byte(" charset=utf-8>")); err != nil {
					return err
				}
				break
			}

			// output attributes
			for _, attr := range t.attr {
				key := attr.key
				val := attr.val
				if caseInsensitiveAttrMap[key] {
					val = bytes.ToLower(val)
				}

				// default attribute values can be ommited
				if key == atom.Colspan && bytes.Equal(val, []byte("1")) ||
					//bytes.Equal(attr.keyRaw, []byte("clear")) && bytes.Equal(val, []byte("none")) ||
					key == atom.Enctype && bytes.Equal(val, []byte("application/x-www-form-urlencoded")) ||
					//bytes.Equal(attr.keyRaw, []byte("frameborder")) && bytes.Equal(val, []byte("1")) ||
					key == atom.Method && bytes.Equal(val, []byte("get")) ||
					key == atom.Rowspan && bytes.Equal(val, []byte("1")) ||
					//bytes.Equal(attr.keyRaw, []byte("scrolling")) && bytes.Equal(val, []byte("auto")) ||
					key == atom.Shape && bytes.Equal(val, []byte("rect")) ||
					key == atom.Span && bytes.Equal(val, []byte("1")) ||
					//bytes.Equal(attr.keyRaw, []byte("valuetype")) && bytes.Equal(val, []byte("data")) ||
					//bytes.Equal(attr.keyRaw, []byte("language")) && t.token == atom.Script && bytes.Equal(val, []byte("javascript")) ||
					key == atom.Type && (t.token == atom.Script && bytes.Equal(val, []byte("text/javascript")) ||
						t.token == atom.Style && bytes.Equal(val, []byte("text/css")) ||
						t.token == atom.Link && bytes.Equal(val, []byte("text/css")) ||
						t.token == atom.Input && bytes.Equal(val, []byte("text")) ||
						t.token == atom.Button && bytes.Equal(val, []byte("submit"))) {
					continue
				}
				if _, err := w.Write([]byte(" ")); err != nil {
					return err
				}
				if _, err := w.Write(attr.keyRaw); err != nil {
					return err
				}

				isBoolean := booleanAttrMap[key]
				if len(val) == 0 && !isBoolean {
					continue
				}

				// booleans have no value
				if !isBoolean {
					var err error
					if _, err := w.Write([]byte("=")); err != nil {
						return err
					}

					// CSS and JS minifiers for attribute inline code
					if key == atom.Style {
						val, err = m.MinifyBytes(defaultStyleType, val)
						if err != nil && err != ErrNotExist {
							return err
						}
					} else if len(attr.keyRaw) > 2 && attr.keyRaw[0] == 'o' && attr.keyRaw[1] == 'n' {
						if len(val) >= 11 && bytes.Equal(bytes.ToLower(val[:11]), []byte("javascript:")) {
							val = val[11:]
						}
						val, err = m.MinifyBytes(defaultScriptType, val)
						if err != nil && err != ErrNotExist {
							return err
						}
					} else if urlAttrMap[key] {
						if len(val) >= 5 && bytes.Equal(bytes.ToLower(val[:5]), []byte("http:")) {
							val = val[5:]
						}
					} else if t.token == atom.Meta && key == atom.Content {
						httpEquiv := t.getAttrVal(atom.HttpEquiv)
						if bytes.Equal(httpEquiv, []byte("content-type")) {
							val = bytes.Replace(val, []byte(", "), []byte(","), -1)
						} else if bytes.Equal(httpEquiv, []byte("content-style-type")) {
							defaultStyleType = string(val)
						} else if bytes.Equal(httpEquiv, []byte("content-script-type")) {
							defaultScriptType = string(val)
						}

						name := bytes.ToLower(t.getAttrVal(atom.Name))
						if bytes.Equal(name, []byte("keywords")) {
							val = bytes.Replace(val, []byte(", "), []byte(","), -1)
						} else if bytes.Equal(name, []byte("viewport")) {
							val = bytes.Replace(val, []byte(" "), []byte(""), -1)
						}
					}

					// no quote if possible, else prefer single or double depending on which occurs more often in value
					if isValidUnquotedAttr(val) {
						if _, err := w.Write(val); err != nil {
							return err
						}
					} else if moreDoubleQuotes(val) {
						if _, err := w.Write([]byte("'")); err != nil {
							return err
						}
						if _, err := w.Write(bytes.Replace(val, []byte("'"), []byte("&#39;"), -1)); err != nil {
							return err
						}
						if _, err := w.Write([]byte("'")); err != nil {
							return err
						}
					} else {
						if _, err := w.Write([]byte("\"")); err != nil {
							return err
						}
						if _, err := w.Write(bytes.Replace(val, []byte("\""), []byte("&quot;"), -1)); err != nil {
							return err
						}
						if _, err := w.Write([]byte("\"")); err != nil {
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
