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
	ltBytes       = []byte("<")
	gtBytes       = []byte(">")
	isBytes       = []byte("=")
	spaceBytes    = []byte(" ")
	endBytes      = []byte("</")
	externalBytes = []byte("external")
	httpBytes     = []byte("http")
)

const maxAttrLookup = 4

////////////////////////////////////////////////////////////////

// Minify minifies HTML data, it reads from r and writes to w.
func Minify(m minify.Minifier, _ string, w io.Writer, r io.Reader) error {
	var rawTag html.Hash
	var rawTagMediatype []byte
	precededBySpace := true // on true the next text token must not start with a space
	defaultScriptType := "text/javascript"
	defaultStyleType := "text/css"
	defaultInlineStyleType := "text/css;inline=1"

	attrMinifyBuffer := buffer.NewWriter(make([]byte, 0, 64))
	attrByteBuffer := make([]byte, 0, 64)
	attrIntBuffer := make([]int, 0, maxAttrLookup)
	attrTokenBuffer := make([]*html.Token, 0, maxAttrLookup)

	l := html.NewLexer(r)
	tb := html.NewTokenBuffer(l)
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
			if _, err := w.Write([]byte("<!doctype html>")); err != nil {
				return err
			}
		case html.CommentToken:
			// TODO: ensure that nested comments are handled properly (lexer doesn't handle this!)
			var comment []byte
			if bytes.HasPrefix(t.Data, []byte("[if")) {
				comment = append(append([]byte("<!--"), t.Data...), []byte("-->")...)
			} else if bytes.HasSuffix(t.Data, []byte("--")) {
				// only occurs when mixed up with conditional comments
				comment = append(append([]byte("<!"), t.Data...), '>')
			}
			if _, err := w.Write(comment); err != nil {
				return err
			}
		case html.TextToken:
			// CSS and JS minifiers for inline code
			if rawTag != 0 {
				if rawTag == html.Style || rawTag == html.Script || rawTag == html.Iframe || rawTag == html.Svg || rawTag == html.Math {
					var mediatype string
					if rawTag == html.Iframe {
						mediatype = "text/html"
					} else if len(rawTagMediatype) > 0 {
						mediatype = string(rawTagMediatype)
					} else if rawTag == html.Script {
						mediatype = defaultScriptType
					} else if rawTag == html.Style {
						mediatype = defaultStyleType
					} else if rawTag == html.Svg {
						mediatype = "image/svg+xml"
					} else if rawTag == html.Math {
						mediatype = "application/mathml+xml"
					}
					// ignore CDATA
					if trimmedData := parse.Trim(t.Data, parse.IsWhitespace); len(trimmedData) > 12 && bytes.Equal(trimmedData[:9], []byte("<![CDATA[")) && bytes.Equal(trimmedData[len(trimmedData)-3:], []byte("]]>")) {
						t.Data = trimmedData[9 : len(trimmedData)-3]
					}
					if err := m.Minify(mediatype, w, buffer.NewReader(t.Data)); err != nil {
						if _, err := w.Write(t.Data); err != nil {
							return err
						}
					}
				} else if _, err := w.Write(t.Data); err != nil {
					return err
				}
			} else if t.Data = parse.ReplaceMultiple(t.Data, parse.IsWhitespace, ' '); len(t.Data) > 0 {
				// whitespace removal; trim left
				if precededBySpace && t.Data[0] == ' ' {
					t.Data = t.Data[1:]
				}

				// whitespace removal; trim right
				precededBySpace = false
				if len(t.Data) == 0 {
					precededBySpace = true
				} else if t.Data[len(t.Data)-1] == ' ' {
					precededBySpace = true
					trim := false
					i := 0
					for {
						next := tb.Peek(i)
						// trim if EOF, text token with whitespace begin or block token
						if next.TokenType == html.ErrorToken {
							trim = true
							break
						} else if next.TokenType == html.TextToken {
							// remove if the text token starts with a whitespace
							trim = (len(next.Data) > 0 && parse.IsWhitespace(next.Data[0]))
							break
						} else if next.TokenType == html.StartTagToken || next.TokenType == html.EndTagToken {
							if !inlineTagMap[next.Hash] {
								trim = true
								break
							} else if next.TokenType == html.StartTagToken {
								break
							}
						}
						i++
					}
					if trim {
						t.Data = t.Data[:len(t.Data)-1]
						precededBySpace = false
					}
				}
				if _, err := w.Write(t.Data); err != nil {
					return err
				}
			}
		case html.StartTagToken, html.EndTagToken:
			rawTag = 0
			hasAttributes := false
			if t.TokenType == html.StartTagToken {
				if next := tb.Peek(0); next.TokenType == html.AttributeToken {
					hasAttributes = true
				}
			}

			if !inlineTagMap[t.Hash] {
				precededBySpace = true
				if t.TokenType == html.StartTagToken && rawTagMap[t.Hash] {
					// ignore empty script and style tags
					if !hasAttributes && (t.Hash == html.Script || t.Hash == html.Style) {
						if next := tb.Peek(1); next.TokenType == html.EndTagToken {
							tb.Shift()
							tb.Shift()
							break
						}
					}
					rawTag = t.Hash
					rawTagMediatype = []byte{}
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
							if next.TokenType == html.ErrorToken || next.TokenType == html.EndTagToken && next.Hash != html.A || next.TokenType == html.StartTagToken && blockTagMap[next.Hash] {
								break SWITCH
							}
							break
						}
					}
				}
			}

			// write tag
			if t.TokenType == html.EndTagToken {
				if _, err := w.Write(endBytes); err != nil {
					return err
				}
			} else {
				if _, err := w.Write(ltBytes); err != nil {
					return err
				}
			}
			if _, err := w.Write(t.Data); err != nil {
				return err
			}

			if hasAttributes {
				// rewrite attributes with interdependent conditions
				if t.Hash == html.A {
					if attr := getAttributes(tb, &attrIntBuffer, &attrTokenBuffer, html.Id, html.Name, html.Rel, html.Href); attr != nil {
						if id := attr[0]; id != nil {
							if name := attr[1]; name != nil && parse.Equal(id.AttrVal, name.AttrVal) {
								name.Data = nil
							}
						}
						if rel := attr[2]; rel == nil || !parse.EqualFold(rel.AttrVal, externalBytes) {
							if href := attr[3]; href != nil {
								if len(href.AttrVal) > 5 && parse.EqualFold(href.AttrVal[:4], []byte{'h', 't', 't', 'p'}) {
									if href.AttrVal[4] == ':' {
										href.AttrVal = href.AttrVal[5:]
									} else if (href.AttrVal[4] == 's' || href.AttrVal[4] == 'S') && href.AttrVal[5] == ':' {
										href.AttrVal = href.AttrVal[6:]
									}
								}
							}
						}
					}
				} else if t.Hash == html.Meta {
					if attr := getAttributes(tb, &attrIntBuffer, &attrTokenBuffer, html.Content, html.Http_Equiv, html.Charset, html.Name); attr != nil {
						if content := attr[0]; content != nil {
							if httpEquiv := attr[1]; httpEquiv != nil {
								content.AttrVal = minify.ContentType(content.AttrVal)
								if charset := attr[2]; charset == nil && parse.EqualFold(httpEquiv.AttrVal, []byte("content-type")) && parse.Equal(content.AttrVal, []byte("text/html;charset=utf-8")) {
									httpEquiv.Data = nil
									content.Data = []byte("charset")
									content.Hash = html.Charset
									content.AttrVal = []byte("utf-8")
								} else if parse.EqualFold(httpEquiv.AttrVal, []byte("content-style-type")) {
									defaultStyleType = string(content.AttrVal)
									defaultInlineStyleType = defaultStyleType + ";inline=1"
								} else if parse.EqualFold(httpEquiv.AttrVal, []byte("content-script-type")) {
									defaultScriptType = string(content.AttrVal)
								}
							}
							if name := attr[3]; name != nil {
								if parse.EqualFold(name.AttrVal, []byte("keywords")) {
									content.AttrVal = bytes.Replace(content.AttrVal, []byte(", "), []byte(","), -1)
								} else if parse.EqualFold(name.AttrVal, []byte("viewport")) {
									content.AttrVal = bytes.Replace(content.AttrVal, []byte(" "), []byte(""), -1)
								}
							}
						}
					}
				} else if t.Hash == html.Script {
					if attr := getAttributes(tb, &attrIntBuffer, &attrTokenBuffer, html.Src, html.Charset); attr != nil {
						if src := attr[0]; src != nil {
							if charset := attr[1]; charset != nil {
								charset.Data = nil
							}
						}
					}
				}

				// write attributes
				for {
					attr := *tb.Shift()
					if attr.TokenType != html.AttributeToken {
						break
					} else if attr.Data == nil {
						continue // removed attribute
					}

					val := attr.AttrVal
					if len(val) > 1 && (val[0] == '"' || val[0] == '\'') {
						val = parse.Trim(val[1:len(val)-1], parse.IsWhitespace)
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
					if caseInsensitiveAttrMap[attr.Hash] {
						val = parse.ToLower(val)
						if attr.Hash == html.Enctype || attr.Hash == html.Codetype || attr.Hash == html.Accept || attr.Hash == html.Type && (t.Hash == html.A || t.Hash == html.Link || t.Hash == html.Object || t.Hash == html.Param || t.Hash == html.Script || t.Hash == html.Style || t.Hash == html.Source) {
							val = minify.ContentType(val)
						}
					}
					if rawTag != 0 && attr.Hash == html.Type {
						rawTagMediatype = val
					}

					// default attribute values can be ommited
					if attr.Hash == html.Type && (t.Hash == html.Script && parse.Equal(val, []byte("text/javascript")) ||
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
						attr.Hash == html.Media && t.Hash == html.Style && parse.Equal(val, []byte("all")) {
						continue
					}
					// CSS and JS minifiers for attribute inline code
					if attr.Hash == html.Style {
						attrMinifyBuffer.Reset()
						if m.Minify(defaultInlineStyleType, attrMinifyBuffer, buffer.NewReader(val)) == nil {
							val = attrMinifyBuffer.Bytes()
						}
						if len(val) == 0 {
							continue
						}
					} else if len(attr.Data) > 2 && attr.Data[0] == 'o' && attr.Data[1] == 'n' {
						if len(val) >= 11 && parse.EqualFold(val[:11], []byte("javascript:")) {
							val = val[11:]
						}
						attrMinifyBuffer.Reset()
						if m.Minify(defaultScriptType, attrMinifyBuffer, buffer.NewReader(val)) == nil {
							val = attrMinifyBuffer.Bytes()
						}
						if len(val) == 0 {
							continue
						}
					} else if t.Hash != html.A && urlAttrMap[attr.Hash] && len(val) > 5 { // anchors are already handled
						if parse.EqualFold(val[:4], []byte{'h', 't', 't', 'p'}) {
							if val[4] == ':' {
								val = val[5:]
							} else if (val[4] == 's' || val[4] == 'S') && val[5] == ':' {
								val = val[6:]
							}
						} else if parse.EqualFold(val[:5], []byte{'d', 'a', 't', 'a', ':'}) {
							val = minify.DataURI(m, val)
						}
						if len(val) == 0 {
							continue
						}
					}

					if _, err := w.Write(spaceBytes); err != nil {
						return err
					}
					if _, err := w.Write(attr.Data); err != nil {
						return err
					}
					if len(val) > 0 && !booleanAttrMap[attr.Hash] {
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

func getAttributes(tb *html.TokenBuffer, attrIndexBuffer *[]int, attrTokenBuffer *[]*html.Token, hashes ...html.Hash) []*html.Token {
	*attrIndexBuffer = (*attrIndexBuffer)[:len(hashes)]
	*attrTokenBuffer = (*attrTokenBuffer)[:len(hashes)]
	for j, _ := range *attrIndexBuffer {
		(*attrIndexBuffer)[j] = 0
	}
	i := 0
	for {
		t := tb.Peek(i)
		if t.TokenType != html.AttributeToken {
			break
		}
		for j, hash := range hashes {
			if t.Hash == hash {
				(*attrIndexBuffer)[j] = i + 1
			}
		}
		i++
	}
	for j, i := range *attrIndexBuffer {
		if i > 0 {
			t := tb.Peek(i - 1)
			if len(t.AttrVal) > 1 && (t.AttrVal[0] == '"' || t.AttrVal[0] == '\'') {
				t.AttrVal = parse.Trim(t.AttrVal[1:len(t.AttrVal)-1], parse.IsWhitespace) // quotes will be readded in attribute loop if necessary
			}
			(*attrTokenBuffer)[j] = t
		} else {
			(*attrTokenBuffer)[j] = nil
		}
	}
	return *attrTokenBuffer
}
