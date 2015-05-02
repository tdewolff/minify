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
	ltBytes                 = []byte("<")
	gtBytes                 = []byte(">")
	isBytes                 = []byte("=")
	spaceBytes              = []byte(" ")
	endBytes                = []byte("</")
	escapedSingleQuoteBytes = []byte("&#39;")
	escapedDoubleQuoteBytes = []byte("&#34;")
	externalBytes           = []byte("external")
	httpBytes               = []byte("http")
)

type token struct {
	tt      html.TokenType
	data    []byte
	attrVal []byte
	hash    html.Hash
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
	defaultInlineStyleType := "text/css;inline=1"

	attrMinifyBuffer := buffer.NewWriter(make([]byte, 0, 64))
	attrEscapeBuffer := make([]byte, 0, 64)
	attrIndexBuffer := make([]int, 0, 4)
	attrTokenBuffer := make([]*token, 0, 4)

	z := html.NewTokenizer(r)
	tb := newTokenBuffer(z)
	for {
		t := *tb.Shift()
		switch t.tt {
		case html.ErrorToken:
			if z.Err() == io.EOF {
				return nil
			}
			return z.Err()
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
					// ignore CDATA in embedded HTML
					if mediatype == "text/html" {
						trimmedData := parse.Trim(t.data, parse.IsWhitespace)
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
			} else if t.data = parse.ReplaceMultiple(t.data, parse.IsWhitespace, ' '); len(t.data) > 0 {
				// whitespace removal; trim left
				if precededBySpace && t.data[0] == ' ' {
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
						next := tb.Peek(i)
						// trim if EOF, text token with whitespace begin or block token
						if next.tt == html.ErrorToken {
							trim = true
							break
						} else if next.tt == html.TextToken {
							// remove if the text token starts with a whitespace
							trim = (len(next.data) > 0 && parse.IsWhitespace(next.data[0]))
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
				if next := tb.Peek(0); next.tt == html.AttributeToken {
					hasAttributes = true
				}
			}

			if !inlineTagMap[t.hash] {
				precededBySpace = true
				if t.tt == html.StartTagToken && rawTagMap[t.hash] {
					// ignore empty script and style tags
					if !hasAttributes && (t.hash == html.Script || t.hash == html.Style) {
						if next := tb.Peek(1); next.tt == html.EndTagToken {
							tb.Shift()
							tb.Shift()
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
							next := tb.Peek(i)
							i++
							// continue if text token is empty or whitespace
							if next.tt == html.TextToken && parse.IsAllWhitespace(next.data) {
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
			}

			// rewrite attributes with interdependent conditions
			if hasAttributes {
				if t.hash == html.A {
					if attr := getAttributes(tb, &attrIndexBuffer, &attrTokenBuffer, html.Id, html.Name, html.Rel, html.Href); attr != nil {
						if id := attr[0]; id != nil {
							if name := attr[1]; name != nil && parse.Equal(id.attrVal, name.attrVal) {
								name.data = nil
							}
						}
						if rel := attr[2]; rel == nil || !parse.EqualCaseInsensitive(rel.attrVal, externalBytes) {
							if href := attr[3]; href != nil {
								if len(href.attrVal) > 5 && parse.EqualCaseInsensitive(href.attrVal[:4], []byte{'h', 't', 't', 'p'}) {
									if href.attrVal[4] == ':' {
										href.attrVal = href.attrVal[5:]
									} else if (href.attrVal[4] == 's' || href.attrVal[4] == 'S') && href.attrVal[5] == ':' {
										href.attrVal = href.attrVal[6:]
									}
								}
							}
						}
					}
				} else if t.hash == html.Meta {
					if attr := getAttributes(tb, &attrIndexBuffer, &attrTokenBuffer, html.Content, html.Http_Equiv, html.Charset, html.Name); attr != nil {
						if content := attr[0]; content != nil {
							if httpEquiv := attr[1]; httpEquiv != nil {
								content.attrVal = parse.NormalizeContentType(content.attrVal)
								if charset := attr[2]; charset == nil && parse.EqualCaseInsensitive(httpEquiv.attrVal, []byte("content-type")) && parse.Equal(content.attrVal, []byte("text/html;charset=utf-8")) {
									httpEquiv.data = nil
									content.data = []byte("charset")
									content.hash = html.Charset
									content.attrVal = []byte("utf-8")
								} else if parse.EqualCaseInsensitive(httpEquiv.attrVal, []byte("content-style-type")) {
									defaultStyleType = string(content.attrVal)
									defaultInlineStyleType = defaultStyleType + ";inline=1"
								} else if parse.EqualCaseInsensitive(httpEquiv.attrVal, []byte("content-script-type")) {
									defaultScriptType = string(content.attrVal)
								}
							}
							if name := attr[3]; name != nil {
								if parse.EqualCaseInsensitive(name.attrVal, []byte("keywords")) {
									content.attrVal = bytes.Replace(content.attrVal, []byte(", "), []byte(","), -1)
								} else if parse.EqualCaseInsensitive(name.attrVal, []byte("viewport")) {
									content.attrVal = bytes.Replace(content.attrVal, []byte(" "), []byte(""), -1)
								}
							}
						}
					}
				} else if t.hash == html.Script {
					if attr := getAttributes(tb, &attrIndexBuffer, &attrTokenBuffer, html.Src, html.Charset); attr != nil {
						if src := attr[0]; src != nil {
							if charset := attr[1]; charset != nil {
								charset.data = nil
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
				if _, err := w.Write(ltBytes); err != nil {
					return err
				}
			}
			if _, err := w.Write(t.data); err != nil {
				return err
			}

			// write attributes
			if hasAttributes {
				for {
					attr := *tb.Shift()
					if attr.tt != html.AttributeToken {
						break
					} else if attr.data == nil {
						continue // removed attribute
					}

					val := attr.attrVal
					if len(val) > 1 && (val[0] == '"' || val[0] == '\'') {
						val = parse.Trim(val[1:len(val)-1], parse.IsWhitespace)
					}
					if len(val) == 0 && (attr.hash == html.Class ||
						attr.hash == html.Dir ||
						attr.hash == html.Id ||
						attr.hash == html.Lang ||
						attr.hash == html.Name ||
						attr.hash == html.Style ||
						attr.hash == html.Title ||
						attr.hash == html.Action && t.hash == html.Form ||
						attr.hash == html.Value && t.hash == html.Input ||
						len(attr.data) > 2 && attr.data[0] == 'o' && attr.data[1] == 'n') {
						continue // omit empty attribute values
					}
					if caseInsensitiveAttrMap[attr.hash] {
						val = parse.ToLower(val)
						if attr.hash == html.Enctype || attr.hash == html.Codetype || attr.hash == html.Accept || attr.hash == html.Type && (t.hash == html.A || t.hash == html.Link || t.hash == html.Object || t.hash == html.Param || t.hash == html.Script || t.hash == html.Style || t.hash == html.Source) {
							val = parse.NormalizeContentType(val)
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
						attr.hash == html.Language && t.hash == html.Script ||
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
						attr.hash == html.Media && t.hash == html.Style && parse.Equal(val, []byte("all")) {
						continue
					}
					if _, err := w.Write(spaceBytes); err != nil {
						return err
					}
					if _, err := w.Write(attr.data); err != nil {
						return err
					}

					if len(val) > 0 && !booleanAttrMap[attr.hash] {
						if _, err := w.Write(isBytes); err != nil {
							return err
						}
						// CSS and JS minifiers for attribute inline code
						if attr.hash == html.Style {
							attrMinifyBuffer.Reset()
							if m.Minify(defaultInlineStyleType, attrMinifyBuffer, buffer.NewReader(val)) == nil {
								val = attrMinifyBuffer.Bytes()
							}
						} else if len(attr.data) > 2 && attr.data[0] == 'o' && attr.data[1] == 'n' {
							if len(val) >= 11 && parse.EqualCaseInsensitive(val[:11], []byte("javascript:")) {
								val = val[11:]
							}
							attrMinifyBuffer.Reset()
							if m.Minify(defaultScriptType, attrMinifyBuffer, buffer.NewReader(val)) == nil {
								val = attrMinifyBuffer.Bytes()
							}
						} else if t.hash != html.A && urlAttrMap[attr.hash] { // anchors are already handled
							if len(val) > 5 && parse.EqualCaseInsensitive(val[:4], []byte{'h', 't', 't', 'p'}) {
								if val[4] == ':' {
									val = val[5:]
								} else if (val[4] == 's' || val[4] == 'S') && val[5] == ':' {
									val = val[6:]
								}
							}
						}
						// no quotes if possible, else prefer single or double depending on which occurs more often in value
						val = escapeAttrVal(&attrEscapeBuffer, attr.attrVal, val)
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

func getAttributes(tb *tokenBuffer, attrIndexBuffer *[]int, attrTokenBuffer *[]*token, hashes ...html.Hash) []*token {
	if cap(*attrIndexBuffer) < len(hashes) || cap(*attrTokenBuffer) < len(hashes) {
		*attrIndexBuffer = make([]int, 0, len(hashes))
		*attrTokenBuffer = make([]*token, 0, len(hashes))
	}
	*attrIndexBuffer = (*attrIndexBuffer)[:len(hashes)]
	*attrTokenBuffer = (*attrTokenBuffer)[:len(hashes)]
	i := 0
	for {
		t := tb.Peek(i)
		if t.tt != html.AttributeToken {
			break
		}
		for j, hash := range hashes {
			if t.hash == hash {
				(*attrIndexBuffer)[j] = i + 1
			}
		}
		i++
	}
	for j, i := range *attrIndexBuffer {
		if i > 0 {
			t := tb.Peek(i - 1)
			if len(t.attrVal) > 1 && (t.attrVal[0] == '"' || t.attrVal[0] == '\'') {
				t.attrVal = parse.Trim(t.attrVal[1:len(t.attrVal)-1], parse.IsWhitespace) // quotes will be readded in attribute loop if necessary
			}
			(*attrTokenBuffer)[j] = t
		} else {
			(*attrTokenBuffer)[j] = nil
		}
	}
	return *attrTokenBuffer
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

// escapeAttrVal returns the escaped attribute value bytes without quotes.
func escapeAttrVal(buf *[]byte, orig []byte, b []byte) []byte {
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
		} else if unquoted && (c == '`' || c == '<' || c == '=' || c == '>' || parse.IsWhitespace(c)) {
			unquoted = false
		}
	}
	if unquoted {
		return b
	} else if singles == 0 && doubles == 0 && len(orig) == len(b)+2 {
		return orig
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

	if len(b)+2 > cap(*buf) {
		*buf = make([]byte, 0, len(b)+2) // maximum size, not actual size
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
