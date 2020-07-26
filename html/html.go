// Package html minifies HTML5 following the specifications at http://www.w3.org/TR/html5/syntax.html.
package html

import (
	"bytes"
	"io"

	"github.com/alex-bacart/minify/v2"
	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/buffer"
	"github.com/tdewolff/parse/v2/html"
)

var (
	gtBytes         = []byte(">")
	isBytes         = []byte("=")
	spaceBytes      = []byte(" ")
	doctypeBytes    = []byte("<!doctype html>")
	jsMimeBytes     = []byte("application/javascript")
	cssMimeBytes    = []byte("text/css")
	htmlMimeBytes   = []byte("text/html")
	svgMimeBytes    = []byte("image/svg+xml")
	mathMimeBytes   = []byte("application/mathml+xml")
	dataSchemeBytes = []byte("data:")
	jsSchemeBytes   = []byte("javascript:")
	httpBytes       = []byte("http")
	inlineParams    = map[string]string{"inline": "1"}
)

////////////////////////////////////////////////////////////////

// DefaultMinifier is the default minifier.
var DefaultMinifier = &Minifier{}

// Minifier is an HTML minifier.
type Minifier struct {
	KeepConditionalComments bool
	KeepDefaultAttrVals     bool
	KeepDocumentTags        bool
	KeepEndTags             bool
	KeepQuotes              bool
	KeepWhitespace          bool
	DontLowercaseAttributes bool
}

// Minify minifies HTML data, it reads from r and writes to w.
func Minify(m *minify.M, w io.Writer, r io.Reader, params map[string]string) error {
	return DefaultMinifier.Minify(m, w, r, params)
}

// Minify minifies HTML data, it reads from r and writes to w.
func (o *Minifier) Minify(m *minify.M, w io.Writer, r io.Reader, _ map[string]string) error {
	var rawTagHash Hash
	var rawTagMediatype []byte

	omitSpace := true // if true the next leading space is omitted
	inPre := false

	attrMinifyBuffer := buffer.NewWriter(make([]byte, 0, 64))
	attrByteBuffer := make([]byte, 0, 64)

	l := html.NewLexer(r)
	defer l.Restore()

	tb := NewTokenBuffer(l)
	for {
		t := *tb.Shift()
		switch t.TokenType {
		case html.ErrorToken:
			if l.Err() == io.EOF {
				return nil
			}
			return l.Err()
		case html.DoctypeToken:
			if _, err := w.Write(doctypeBytes); err != nil {
				return err
			}
		case html.CommentToken:
			if o.KeepConditionalComments && len(t.Text) > 6 && (bytes.HasPrefix(t.Text, []byte("[if ")) || bytes.HasSuffix(t.Text, []byte("[endif]")) || bytes.HasSuffix(t.Text, []byte("[endif]--"))) {
				// [if ...] is always 7 or more characters, [endif] is only encountered for downlevel-revealed
				// see https://msdn.microsoft.com/en-us/library/ms537512(v=vs.85).aspx#syntax
				if bytes.HasPrefix(t.Data, []byte("<!--[if ")) && bytes.HasSuffix(t.Data, []byte("<![endif]-->")) { // downlevel-hidden
					begin := bytes.IndexByte(t.Data, '>') + 1
					end := len(t.Data) - len("<![endif]-->")
					if _, err := w.Write(t.Data[:begin]); err != nil {
						return err
					}
					if err := o.Minify(m, w, buffer.NewReader(t.Data[begin:end]), nil); err != nil {
						return err
					}
					if _, err := w.Write(t.Data[end:]); err != nil {
						return err
					}
				} else if _, err := w.Write(t.Data); err != nil { // downlevel-revealed or short downlevel-hidden
					return err
				}
			}
		case html.SvgToken:
			if err := m.MinifyMimetype(svgMimeBytes, w, buffer.NewReader(t.Data), nil); err != nil {
				if err != minify.ErrNotExist {
					return err
				} else if _, err := w.Write(t.Data); err != nil {
					return err
				}
			}
		case html.MathToken:
			if err := m.MinifyMimetype(mathMimeBytes, w, buffer.NewReader(t.Data), nil); err != nil {
				if err != minify.ErrNotExist {
					return err
				} else if _, err := w.Write(t.Data); err != nil {
					return err
				}
			}
		case html.TextToken:
			// CSS and JS minifiers for inline code
			if rawTagHash != 0 {
				if rawTagHash == Style || rawTagHash == Script || rawTagHash == Iframe {
					var mimetype []byte
					var params map[string]string
					if rawTagHash == Iframe {
						mimetype = htmlMimeBytes
					} else if len(rawTagMediatype) > 0 {
						mimetype, params = parse.Mediatype(rawTagMediatype)
					} else if rawTagHash == Script {
						mimetype = jsMimeBytes
					} else if rawTagHash == Style {
						mimetype = cssMimeBytes
					}
					if err := m.MinifyMimetype(mimetype, w, buffer.NewReader(t.Data), params); err != nil {
						if err != minify.ErrNotExist {
							return err
						} else if _, err := w.Write(t.Data); err != nil {
							return err
						}
					}
				} else if _, err := w.Write(t.Data); err != nil {
					return err
				}
			} else if inPre {
				if _, err := w.Write(t.Data); err != nil {
					return err
				}
			} else {
				t.Data = parse.ReplaceMultipleWhitespaceAndEntities(t.Data, EntitiesMap, TextRevEntitiesMap)

				// whitespace removal; trim left
				if omitSpace && parse.IsWhitespace(t.Data[0]) {
					t.Data = t.Data[1:]
				}

				// whitespace removal; trim right
				omitSpace = false
				if len(t.Data) == 0 {
					omitSpace = true
				} else if parse.IsWhitespace(t.Data[len(t.Data)-1]) {
					omitSpace = true
					i := 0
					for {
						next := tb.Peek(i)
						// trim if EOF, text token with leading whitespace or block token
						if next.TokenType == html.ErrorToken {
							t.Data = t.Data[:len(t.Data)-1]
							omitSpace = false
							break
						} else if next.TokenType == html.TextToken {
							// this only happens when a comment, doctype or phrasing end tag (only for !o.KeepWhitespace) was in between
							// remove if the text token starts with a whitespace
							if len(next.Data) > 0 && parse.IsWhitespace(next.Data[0]) {
								t.Data = t.Data[:len(t.Data)-1]
								omitSpace = false
							}
							break
						} else if next.TokenType == html.StartTagToken || next.TokenType == html.EndTagToken {
							if o.KeepWhitespace {
								break
							}
							// remove when followed up by a block tag
							if next.Traits&nonPhrasingTag != 0 {
								t.Data = t.Data[:len(t.Data)-1]
								omitSpace = false
								break
							} else if next.TokenType == html.StartTagToken {
								break
							}
						}
						i++
					}
				}

				if _, err := w.Write(t.Data); err != nil {
					return err
				}
			}
		case html.StartTagToken, html.EndTagToken:
			rawTagHash = 0
			hasAttributes := false
			if t.TokenType == html.StartTagToken {
				if next := tb.Peek(0); next.TokenType == html.AttributeToken {
					hasAttributes = true
				}
				if t.Traits&rawTag != 0 {
					// ignore empty script and style tags
					if !hasAttributes && (t.Hash == Script || t.Hash == Style) {
						if next := tb.Peek(1); next.TokenType == html.EndTagToken {
							tb.Shift()
							tb.Shift()
							break
						}
					}
					rawTagHash = t.Hash
					rawTagMediatype = nil

					// do not minify content of <style amp-boilerplate>
					if hasAttributes && t.Hash == Style {
						if attrs := tb.Attributes(Amp_Boilerplate); attrs[0] != nil {
							rawTagHash = 0
						}
					}
				}
			} else if t.Hash == Template {
				omitSpace = true // EndTagToken
			}

			if t.Hash == Pre {
				inPre = t.TokenType == html.StartTagToken
			}

			// remove superfluous tags, except for html, head and body tags when KeepDocumentTags is set
			if !hasAttributes && (!o.KeepDocumentTags && (t.Hash == Html || t.Hash == Head || t.Hash == Body) || t.Hash == Colgroup) {
				break
			} else if t.TokenType == html.EndTagToken {
				omitEndTag := false
				if !o.KeepEndTags {
					if t.Hash == Thead || t.Hash == Tbody || t.Hash == Tfoot || t.Hash == Tr || t.Hash == Th ||
						t.Hash == Td || t.Hash == Option || t.Hash == Dd || t.Hash == Dt || t.Hash == Li ||
						t.Hash == Rb || t.Hash == Rt || t.Hash == Rtc || t.Hash == Rp {
						omitEndTag = true // omit end tags
					} else if t.Hash == P {
						i := 0
						for {
							next := tb.Peek(i)
							i++
							// continue if text token is empty or whitespace
							if next.TokenType == html.TextToken && parse.IsAllWhitespace(next.Data) {
								continue
							}
							if next.TokenType == html.ErrorToken || next.TokenType == html.EndTagToken && next.Traits&keepPTag == 0 || next.TokenType == html.StartTagToken && next.Traits&omitPTag != 0 {
								omitEndTag = true // omit p end tag
							}
							break
						}
					} else if t.Hash == Optgroup {
						i := 0
						for {
							next := tb.Peek(i)
							i++
							// continue if text token
							if next.TokenType == html.TextToken {
								continue
							}
							if next.TokenType == html.ErrorToken || next.Hash != Option {
								omitEndTag = true // omit optgroup end tag
							}
							break
						}
					}
				}

				if o.KeepWhitespace || t.Traits&objectTag != 0 {
					omitSpace = false
				} else if t.Traits&nonPhrasingTag != 0 {
					omitSpace = true // omit spaces after block elements
				}

				if !omitEndTag {
					if len(t.Data) > 3+len(t.Text) {
						t.Data[2+len(t.Text)] = '>'
						t.Data = t.Data[:3+len(t.Text)]
					}
					if _, err := w.Write(t.Data); err != nil {
						return err
					}
				}

				// skip text in select and optgroup tags
				if t.Hash == Option || t.Hash == Optgroup {
					if next := tb.Peek(0); next.TokenType == html.TextToken {
						tb.Shift()
					}
				}
				break
			}

			if o.KeepWhitespace || t.Traits&objectTag != 0 {
				omitSpace = false
			} else if t.Traits&nonPhrasingTag != 0 {
				omitSpace = true // omit spaces after block elements
			}

			if _, err := w.Write(t.Data); err != nil {
				return err
			}

			if hasAttributes {
				if t.Hash == Meta {
					attrs := tb.Attributes(Content, Http_Equiv, Charset, Name)
					if content := attrs[0]; content != nil {
						if httpEquiv := attrs[1]; httpEquiv != nil {
							if charset := attrs[2]; charset == nil && parse.EqualFold(httpEquiv.AttrVal, []byte("content-type")) {
								content.AttrVal = minify.Mediatype(content.AttrVal)
								if bytes.Equal(content.AttrVal, []byte("text/html;charset=utf-8")) {
									httpEquiv.Text = nil
									content.Text = []byte("charset")
									content.Hash = Charset
									content.AttrVal = []byte("utf-8")
								}
							}
						}
						if name := attrs[3]; name != nil {
							if parse.EqualFold(name.AttrVal, []byte("keywords")) {
								content.AttrVal = bytes.Replace(content.AttrVal, []byte(", "), []byte(","), -1)
							} else if parse.EqualFold(name.AttrVal, []byte("viewport")) {
								content.AttrVal = bytes.Replace(content.AttrVal, []byte(" "), []byte(""), -1)
								for i := 0; i < len(content.AttrVal); i++ {
									if content.AttrVal[i] == '=' && i+2 < len(content.AttrVal) {
										i++
										if n := parse.Number(content.AttrVal[i:]); n > 0 {
											minNum := minify.Number(content.AttrVal[i:i+n], -1)
											if len(minNum) < n {
												copy(content.AttrVal[i:i+len(minNum)], minNum)
												copy(content.AttrVal[i+len(minNum):], content.AttrVal[i+n:])
												content.AttrVal = content.AttrVal[:len(content.AttrVal)+len(minNum)-n]
											}
											i += len(minNum)
										}
										i-- // mitigate for-loop increase
									}
								}
							}
						}
					}
				} else if t.Hash == Script {
					attrs := tb.Attributes(Src, Charset)
					if attrs[0] != nil && attrs[1] != nil {
						attrs[1].Text = nil
					}
				} else if t.Hash == Input {
					attrs := tb.Attributes(Type, Value)
					if t, value := attrs[0], attrs[1]; t != nil && value != nil {
						isRadio := parse.EqualFold(t.AttrVal, []byte("radio"))
						if !isRadio && len(value.AttrVal) == 0 {
							value.Text = nil
						} else if isRadio && parse.EqualFold(value.AttrVal, []byte("on")) {
							value.Text = nil
						}
					}
				}

				// write attributes
				htmlEqualIdName := false
				for {
					attr := *tb.Shift()
					if attr.TokenType != html.AttributeToken {
						break
					} else if attr.Text == nil {
						continue // removed attribute
					}

					if t.Hash == A && (attr.Hash == Id || attr.Hash == Name) {
						if attr.Hash == Id {
							if name := tb.Attributes(Name)[0]; name != nil && bytes.Equal(attr.AttrVal, name.AttrVal) {
								htmlEqualIdName = true
							}
						} else if htmlEqualIdName {
							continue
						} else if id := tb.Attributes(Id)[0]; id != nil && bytes.Equal(id.AttrVal, attr.AttrVal) {
							continue
						}
					}

					val := attr.AttrVal
					if attr.Traits&trimAttr != 0 {
						val = parse.ReplaceMultipleWhitespaceAndEntities(val, EntitiesMap, nil)
					} else {
						val = parse.ReplaceEntities(val, EntitiesMap, nil)
					}
					if t.Traits != 0 {
						if len(val) == 0 && (attr.Hash == Class ||
							attr.Hash == Dir ||
							attr.Hash == Id ||
							attr.Hash == Lang ||
							attr.Hash == Name ||
							attr.Hash == Title ||
							attr.Hash == Action && t.Hash == Form) {
							continue // omit empty attribute values
						}
						if attr.Traits&caselessAttr != 0 {
							if !o.DontLowercaseAttributes {
								val = parse.ToLower(val)
							}
							if attr.Hash == Enctype || attr.Hash == Codetype || attr.Hash == Accept || attr.Hash == Type && (t.Hash == A || t.Hash == Link || t.Hash == Embed || t.Hash == Object || t.Hash == Source || t.Hash == Script || t.Hash == Style) {
								val = minify.Mediatype(val)
							}
						}
						if rawTagHash != 0 && attr.Hash == Type {
							rawTagMediatype = parse.Copy(val)
						}

						// default attribute values can be omitted
						if !o.KeepDefaultAttrVals && (attr.Hash == Type && (t.Hash == Script && jsMimetypes[string(val)] ||
							t.Hash == Style && bytes.Equal(val, []byte("text/css")) ||
							t.Hash == Link && bytes.Equal(val, []byte("text/css")) ||
							t.Hash == Input && bytes.Equal(val, []byte("text")) ||
							t.Hash == Button && bytes.Equal(val, []byte("submit"))) ||
							attr.Hash == Language && t.Hash == Script ||
							attr.Hash == Method && bytes.Equal(val, []byte("get")) ||
							attr.Hash == Enctype && bytes.Equal(val, []byte("application/x-www-form-urlencoded")) ||
							attr.Hash == Colspan && bytes.Equal(val, []byte("1")) ||
							attr.Hash == Rowspan && bytes.Equal(val, []byte("1")) ||
							attr.Hash == Shape && bytes.Equal(val, []byte("rect")) ||
							attr.Hash == Span && bytes.Equal(val, []byte("1")) ||
							attr.Hash == Clear && bytes.Equal(val, []byte("none")) ||
							attr.Hash == Frameborder && bytes.Equal(val, []byte("1")) ||
							attr.Hash == Scrolling && bytes.Equal(val, []byte("auto")) ||
							attr.Hash == Valuetype && bytes.Equal(val, []byte("data")) ||
							attr.Hash == Media && t.Hash == Style && bytes.Equal(val, []byte("all"))) {
							continue
						}

						if attr.Hash == Style {
							// CSS minifier for attribute inline code
							val = parse.TrimWhitespace(val)
							attrMinifyBuffer.Reset()
							if err := m.MinifyMimetype(cssMimeBytes, attrMinifyBuffer, buffer.NewReader(val), inlineParams); err == nil {
								val = attrMinifyBuffer.Bytes()
							} else if err != minify.ErrNotExist {
								return err
							}
							if len(val) == 0 {
								continue
							}
						} else if len(attr.Text) > 2 && attr.Text[0] == 'o' && attr.Text[1] == 'n' {
							// JS minifier for attribute inline code
							val = parse.TrimWhitespace(val)
							if len(val) >= 11 && parse.EqualFold(val[:11], jsSchemeBytes) {
								val = val[11:]
							}
							attrMinifyBuffer.Reset()
							if err := m.MinifyMimetype(jsMimeBytes, attrMinifyBuffer, buffer.NewReader(val), nil); err == nil {
								val = attrMinifyBuffer.Bytes()
							} else if err != minify.ErrNotExist {
								return err
							}
							if len(val) == 0 {
								continue
							}
						} else if attr.Traits&urlAttr != 0 { // anchors are already handled
							val = parse.TrimWhitespace(val)
							if 5 < len(val) {
								if parse.EqualFold(val[:4], httpBytes) {
									if val[4] == ':' {
										if m.URL != nil && m.URL.Scheme == "http" {
											val = val[5:]
										} else {
											parse.ToLower(val[:4])
										}
									} else if (val[4] == 's' || val[4] == 'S') && val[5] == ':' {
										if m.URL != nil && m.URL.Scheme == "https" {
											val = val[6:]
										} else {
											parse.ToLower(val[:5])
										}
									}
								} else if parse.EqualFold(val[:5], dataSchemeBytes) {
									val = minify.DataURI(m, val)
								}
							}
						}
					}

					if _, err := w.Write(spaceBytes); err != nil {
						return err
					}
					if _, err := w.Write(attr.Text); err != nil {
						return err
					}
					if len(val) > 0 && attr.Traits&booleanAttr == 0 {
						if _, err := w.Write(isBytes); err != nil {
							return err
						}

						// use double quotes for RDFa attributes
						isXML := attr.Hash == Vocab || attr.Hash == Typeof || attr.Hash == Property || attr.Hash == Resource || attr.Hash == Prefix || attr.Hash == Content || attr.Hash == About || attr.Hash == Rev || attr.Hash == Datatype || attr.Hash == Inlist

						// no quotes if possible, else prefer single or double depending on which occurs more often in value
						val = html.EscapeAttrVal(&attrByteBuffer, attr.AttrVal, val, o.KeepQuotes || isXML)
						if _, err := w.Write(val); err != nil {
							return err
						}
					}
				}
			} else {
				_ = tb.Shift() // StartTagClose
			}
			if _, err := w.Write(gtBytes); err != nil {
				return err
			}

			// skip text in select and optgroup tags
			if t.Hash == Select || t.Hash == Optgroup {
				if next := tb.Peek(0); next.TokenType == html.TextToken {
					tb.Shift()
				}
			}
		}
	}
}
