// Package html minifies HTML5 following the specifications at http://www.w3.org/TR/html5/syntax.html.
package html // import "github.com/tdewolff/minify/html"

import (
	"bytes"
	"io"

	"github.com/tdewolff/buffer"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/parse"
	"github.com/tdewolff/parse/html"
)

var (
	gtBytes         = []byte(">")
	isBytes         = []byte("=")
	spaceBytes      = []byte(" ")
	cdataBytes      = []byte("<![CDATA[")
	cdataEndBytes   = []byte("]]>")
	doctypeBytes    = []byte("<!doctype html>")
	jsMimeBytes     = []byte("text/javascript")
	cssMimeBytes    = []byte("text/css")
	htmlMimeBytes   = []byte("text/html")
	svgMimeBytes    = []byte("image/svg+xml")
	mathMimeBytes   = []byte("application/mathml+xml")
	dataSchemeBytes = []byte("data:")
	jsSchemeBytes   = []byte("javascript:")
	httpBytes       = []byte("http")
)

const maxAttrLookup = 4

////////////////////////////////////////////////////////////////

// Minifier is an HTML minifier.
type Minifier struct {
	KeepDefaultAttrVals bool
	KeepWhitespace      bool
}

// Minify minifies HTML data, it reads from r and writes to w.
func Minify(m *minify.M, w io.Writer, r io.Reader, params map[string]string) error {
	return (&Minifier{}).Minify(m, w, r, params)
}

// Minify minifies HTML data, it reads from r and writes to w.
func (o *Minifier) Minify(m *minify.M, w io.Writer, r io.Reader, _ map[string]string) error {
	var rawTagHash html.Hash
	var rawTagTraits traits
	var rawTagMediatype []byte
	omitSpace := true // if true the next leading space is omitted
	defaultScriptType := jsMimeBytes
	defaultScriptParams := map[string]string(nil)
	defaultStyleType := cssMimeBytes
	defaultStyleParams := map[string]string(nil)
	defaultInlineStyleParams := map[string]string{"inline": "1"}

	attrMinifyBuffer := buffer.NewWriter(make([]byte, 0, 64))
	attrByteBuffer := make([]byte, 0, 64)
	attrTokenBuffer := make([]*Token, 0, maxAttrLookup)

	l := html.NewLexer(r)
	tb := NewTokenBuffer(l)
	for {
		t := *tb.Shift()
	SWITCH:
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
			// TODO: ensure that nested comments are handled properly (lexer doesn't handle this!)
			var comment []byte
			if bytes.HasPrefix(t.Text, []byte("[if")) {
				comment = t.Data
			} else if bytes.HasSuffix(t.Text, []byte("--")) {
				// only occurs when mixed up with conditional comments
				comment = append(append([]byte("<!"), t.Text...), '>')
			}
			if _, err := w.Write(comment); err != nil {
				return err
			}
		case html.TextToken:
			// CSS and JS minifiers for inline code
			if rawTagHash != 0 {
				if rawTagHash == html.Style || rawTagHash == html.Script || rawTagHash == html.Iframe || rawTagHash == html.Svg || rawTagHash == html.Math {
					var mimetype []byte
					var params map[string]string
					if rawTagHash == html.Iframe {
						mimetype = htmlMimeBytes
					} else if rawTagHash == html.Svg {
						mimetype = svgMimeBytes
					} else if rawTagHash == html.Math {
						mimetype = mathMimeBytes
					} else if len(rawTagMediatype) > 0 {
						mimetype, params = parse.Mediatype(rawTagMediatype)
					} else if rawTagHash == html.Script {
						mimetype = defaultScriptType
						params = defaultScriptParams
					} else if rawTagHash == html.Style {
						mimetype = defaultStyleType
						params = defaultStyleParams
					}
					// TODO: really necessary?
					// ignore CDATA because that only has meaning in XML
					if trimmedData := parse.TrimWhitespace(t.Data); len(trimmedData) > 12 && bytes.Equal(trimmedData[:9], cdataBytes) && bytes.Equal(trimmedData[len(trimmedData)-3:], cdataEndBytes) {
						t.Data = trimmedData[9 : len(trimmedData)-3]
					}
					if err := m.MinifyMimetype(mimetype, w, buffer.NewReader(t.Data), params); err != nil {
						if _, err := w.Write(t.Data); err != nil {
							return err
						}
					}
				} else if _, err := w.Write(t.Data); err != nil {
					return err
				}
				if rawTagTraits&nonPhrasingTag == 0 && rawTagHash != html.Script {
					omitSpace = len(t.Data) > 0 && (t.Data[len(t.Data)-1] == ' ' || t.Data[len(t.Data)-1] == '\n')
				}
			} else {
				t.Data = parse.ReplaceMultipleWhitespace(t.Data)
				if !o.KeepWhitespace {
					// whitespace removal; trim left
					if omitSpace && (t.Data[0] == ' ' || t.Data[0] == '\n') {
						t.Data = t.Data[1:]
					}

					// whitespace removal; trim right
					omitSpace = false
					if len(t.Data) == 0 {
						omitSpace = true
					} else if t.Data[len(t.Data)-1] == ' ' || t.Data[len(t.Data)-1] == '\n' {
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
								// remove if the text token starts with a whitespace
								if len(next.Data) > 0 && parse.IsWhitespace(next.Data[0]) {
									t.Data = t.Data[:len(t.Data)-1]
									omitSpace = false
								}
								break
							} else if next.TokenType == html.StartTagToken || next.TokenType == html.EndTagToken {
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
					if !hasAttributes && (t.Hash == html.Script || t.Hash == html.Style) {
						if next := tb.Peek(1); next.TokenType == html.EndTagToken {
							tb.Shift()
							tb.Shift()
							break
						}
					}
					rawTagHash = t.Hash
					rawTagTraits = t.Traits
					rawTagMediatype = nil
				}
			}
			if t.Traits&nonPhrasingTag != 0 {
				omitSpace = true // omit spaces after block elements
			}

			// remove superfluous ending tags
			if !hasAttributes && (t.Hash == html.Html || t.Hash == html.Head || t.Hash == html.Body || t.Hash == html.Colgroup) {
				break
			} else if t.TokenType == html.EndTagToken {
				if t.Hash == html.Thead || t.Hash == html.Tbody || t.Hash == html.Tfoot || t.Hash == html.Tr || t.Hash == html.Th || t.Hash == html.Td ||
					t.Hash == html.Optgroup || t.Hash == html.Option || t.Hash == html.Dd || t.Hash == html.Dt ||
					t.Hash == html.Li || t.Hash == html.Rb || t.Hash == html.Rt || t.Hash == html.Rtc || t.Hash == html.Rp {
					break
				} else if t.Hash == html.P {
					i := 0
					for {
						next := tb.Peek(i)
						i++
						// continue if text token is empty or whitespace
						if next.TokenType == html.TextToken && parse.IsAllWhitespace(next.Data) {
							continue
						}
						if next.TokenType == html.ErrorToken || next.TokenType == html.EndTagToken && next.Hash != html.A || next.TokenType == html.StartTagToken && next.Traits&nonPhrasingTag != 0 {
							break SWITCH
						}
						break
					}
				}
			}

			if t.TokenType == html.EndTagToken {
				t.Data[2+len(t.Text)] = '>'
				if _, err := w.Write(t.Data[:2+len(t.Text)+1]); err != nil {
					return err
				}
				break
			}

			if _, err := w.Write(t.Data); err != nil {
				return err
			}

			if hasAttributes {
				// rewrite attributes with interdependent conditions
				if t.Hash == html.A {
					getAttributes(&attrTokenBuffer, tb, html.Id, html.Name, html.Rel, html.Href)
					if id := attrTokenBuffer[0]; id != nil {
						if name := attrTokenBuffer[1]; name != nil && parse.Equal(id.AttrVal, name.AttrVal) {
							name.Text = nil
						}
					}
					if href := attrTokenBuffer[3]; href != nil {
						if len(href.AttrVal) > 5 && parse.EqualFold(href.AttrVal[:4], httpBytes) {
							if href.AttrVal[4] == ':' {
								if m.URL != nil && m.URL.Scheme == "http" {
									href.AttrVal = href.AttrVal[5:]
								} else {
									parse.ToLower(href.AttrVal[:4])
								}
							} else if (href.AttrVal[4] == 's' || href.AttrVal[4] == 'S') && href.AttrVal[5] == ':' {
								if m.URL != nil && m.URL.Scheme == "https" {
									href.AttrVal = href.AttrVal[6:]
								} else {
									parse.ToLower(href.AttrVal[:5])
								}
							}
						}
					}
				} else if t.Hash == html.Meta {
					getAttributes(&attrTokenBuffer, tb, html.Content, html.Http_Equiv, html.Charset, html.Name)
					if content := attrTokenBuffer[0]; content != nil {
						if httpEquiv := attrTokenBuffer[1]; httpEquiv != nil {
							content.AttrVal = minify.ContentType(content.AttrVal)
							if charset := attrTokenBuffer[2]; charset == nil && parse.EqualFold(httpEquiv.AttrVal, []byte("content-type")) && parse.Equal(content.AttrVal, []byte("text/html;charset=utf-8")) {
								httpEquiv.Text = nil
								content.Text = []byte("charset")
								content.Hash = html.Charset
								content.AttrVal = []byte("utf-8")
							} else if parse.EqualFold(httpEquiv.AttrVal, []byte("content-style-type")) {
								defaultStyleType, defaultStyleParams = parse.Mediatype(content.AttrVal)
								if defaultStyleParams != nil {
									defaultInlineStyleParams = defaultStyleParams
									defaultInlineStyleParams["inline"] = "1"
								} else {
									defaultInlineStyleParams = map[string]string{"inline": "1"}
								}
							} else if parse.EqualFold(httpEquiv.AttrVal, []byte("content-script-type")) {
								defaultScriptType, defaultScriptParams = parse.Mediatype(content.AttrVal)
							}
						}
						if name := attrTokenBuffer[3]; name != nil {
							if parse.EqualFold(name.AttrVal, []byte("keywords")) {
								content.AttrVal = bytes.Replace(content.AttrVal, []byte(", "), []byte(","), -1)
							} else if parse.EqualFold(name.AttrVal, []byte("viewport")) {
								content.AttrVal = bytes.Replace(content.AttrVal, []byte(" "), []byte(""), -1)
							}
						}
					}
				} else if t.Hash == html.Script {
					getAttributes(&attrTokenBuffer, tb, html.Src, html.Charset)
					if src := attrTokenBuffer[0]; src != nil {
						if charset := attrTokenBuffer[1]; charset != nil {
							charset.Text = nil
						}
					}
				}

				// write attributes
				for {
					attr := *tb.Shift()
					if attr.TokenType != html.AttributeToken {
						break
					} else if attr.Text == nil {
						continue // removed attribute
					}

					val := attr.AttrVal
					if len(val) > 1 && (val[0] == '"' || val[0] == '\'') {
						val = parse.TrimWhitespace(val[1 : len(val)-1])
					}
					if len(val) == 0 && (attr.Hash == html.Class ||
						attr.Hash == html.Dir ||
						attr.Hash == html.Id ||
						attr.Hash == html.Lang ||
						attr.Hash == html.Name ||
						attr.Hash == html.Title ||
						attr.Hash == html.Action && t.Hash == html.Form ||
						attr.Hash == html.Value && t.Hash == html.Input) {
						continue // omit empty attribute values
					}
					if attr.Traits&caselessAttr != 0 {
						val = parse.ToLower(val)
						if attr.Hash == html.Enctype || attr.Hash == html.Codetype || attr.Hash == html.Accept || attr.Hash == html.Type && (t.Hash == html.A || t.Hash == html.Link || t.Hash == html.Object || t.Hash == html.Param || t.Hash == html.Script || t.Hash == html.Style || t.Hash == html.Source) {
							val = minify.ContentType(val)
						}
					}
					if rawTagHash != 0 && attr.Hash == html.Type {
						rawTagMediatype = val
					}

					// default attribute values can be ommited
					if !o.KeepDefaultAttrVals && (attr.Hash == html.Type && (t.Hash == html.Script && parse.Equal(val, []byte("text/javascript")) ||
						t.Hash == html.Style && parse.Equal(val, []byte("text/css")) ||
						t.Hash == html.Link && parse.Equal(val, []byte("text/css")) ||
						t.Hash == html.Input && parse.Equal(val, []byte("text")) ||
						t.Hash == html.Button && parse.Equal(val, []byte("submit"))) ||
						attr.Hash == html.Language && t.Hash == html.Script ||
						attr.Hash == html.Method && parse.Equal(val, []byte("get")) ||
						attr.Hash == html.Enctype && parse.Equal(val, []byte("application/x-www-form-urlencoded")) ||
						attr.Hash == html.Colspan && parse.Equal(val, []byte("1")) ||
						attr.Hash == html.Rowspan && parse.Equal(val, []byte("1")) ||
						attr.Hash == html.Shape && parse.Equal(val, []byte("rect")) ||
						attr.Hash == html.Span && parse.Equal(val, []byte("1")) ||
						attr.Hash == html.Clear && parse.Equal(val, []byte("none")) ||
						attr.Hash == html.Frameborder && parse.Equal(val, []byte("1")) ||
						attr.Hash == html.Scrolling && parse.Equal(val, []byte("auto")) ||
						attr.Hash == html.Valuetype && parse.Equal(val, []byte("data")) ||
						attr.Hash == html.Media && t.Hash == html.Style && parse.Equal(val, []byte("all"))) {
						continue
					}
					// CSS and JS minifiers for attribute inline code
					if attr.Hash == html.Style {
						attrMinifyBuffer.Reset()
						if m.MinifyMimetype(defaultStyleType, attrMinifyBuffer, buffer.NewReader(val), defaultInlineStyleParams) == nil {
							val = attrMinifyBuffer.Bytes()
						}
						if len(val) == 0 {
							continue
						}
					} else if len(attr.Text) > 2 && attr.Text[0] == 'o' && attr.Text[1] == 'n' {
						if len(val) >= 11 && parse.EqualFold(val[:11], jsSchemeBytes) {
							val = val[11:]
						}
						attrMinifyBuffer.Reset()
						if m.MinifyMimetype(defaultScriptType, attrMinifyBuffer, buffer.NewReader(val), defaultScriptParams) == nil {
							val = attrMinifyBuffer.Bytes()
						}
						if len(val) == 0 {
							continue
						}
					} else if len(val) > 5 && attr.Traits&urlAttr != 0 { // anchors are already handled
						if t.Hash != html.A {
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
							}
						}
						if parse.EqualFold(val[:5], dataSchemeBytes) {
							val = minify.DataURI(m, val)
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
						// no quotes if possible, else prefer single or double depending on which occurs more often in value
						val = html.EscapeAttrVal(&attrByteBuffer, attr.AttrVal, val)
						if _, err := w.Write(val); err != nil {
							return err
						}
					}
				}
			}
			if _, err := w.Write(gtBytes); err != nil {
				return err
			}
		}
	}
}

////////////////////////////////////////////////////////////////

func getAttributes(attrTokenBuffer *[]*Token, tb *TokenBuffer, hashes ...html.Hash) {
	*attrTokenBuffer = (*attrTokenBuffer)[:len(hashes)]
	for j := range *attrTokenBuffer {
		(*attrTokenBuffer)[j] = nil
	}
	for i := 0; ; i++ {
		t := tb.Peek(i)
		if t.TokenType != html.AttributeToken {
			break
		}
		for j, hash := range hashes {
			if t.Hash == hash {
				if len(t.AttrVal) > 1 && (t.AttrVal[0] == '"' || t.AttrVal[0] == '\'') {
					t.AttrVal = parse.TrimWhitespace(t.AttrVal[1 : len(t.AttrVal)-1]) // quotes will be readded in attribute loop if necessary
				}
				(*attrTokenBuffer)[j] = t
				break
			}
		}
	}
}
