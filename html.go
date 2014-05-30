package minify

import (
	"io"
	"io/ioutil"
	"strings"
	"regexp"
	"bytes"
	"code.google.com/p/go.net/html"
)

func (minify Minify) Html(r io.ReadCloser) (io.ReadCloser, error) {
	multipleWhitespaceRegexp := regexp.MustCompile("\\s+")
	validAttrRegexp := regexp.MustCompile("^[^\\s\"'`=<>/]*$")
	booleanAttrRegexp := regexp.MustCompile("^(allowfullscreen|async|autofocus|autoplay|checked|compact|controls|declare|"+
		"default|defaultChecked|defaultMuted|defaultSelected|defer|disabled|draggable|enabled|formnovalidate|hidden|"+
		"undeterminate|inert|ismap|itemscope|multiple|muted|nohref|noresize|noshade|novalidate|nowrap|open|pauseonexit|"+
		"readonly|required|reversed|scoped|seamless|selected|sortable|spellcheck|translate|truespeed|typemustmatch|"+
		"visible)$")
	eventAttrRegexp := regexp.MustCompile("^on[a-z]+$")
	// eventAttrRegexp := regexp.MustCompile("^(onabort|onafterprint|onbeforeprint|onbeforeunload|onblur|oncanplay|"+
	// 	"oncanplaythrough|onchange|onclick|oncontextmenu|ondblclick|ondrag|ondragend|ondragenter|ondragleave|ondragover|"+
	// 	"ondragstart|ondrop|ondurationchange|onemptied|onended|onerror|onfocus|onformchange|onforminput|onhaschange|"+
	// 	"oninput|oninvalid|onkeydown|onkeypress|onkeyup|onload|onloadeddata|onloadedmetadata|onloadstart|onmessage|"+
	// 	"onmousedown|onmousemove|onmouseout|onmouseover|onmouseup|onmousewheel|onoffline|ononline|onpagehide|onpageshow|"+
	// 	"onpause|onplay|onplaying|onpopstate|onprogress|onratechange|onreadystatechange|onredo|onresize|onscroll|"+
	// 	"onseeked|onseeking|onselect|onshow|onstalled|onstorage|onsubmit|onsuspend|ontimeupdate|onundo|onunload|"+
	// 	"onvolumechange|onwaiting)$")
	specialTagRegexp := regexp.MustCompile("^(style|script|pre|code|textarea)$")
	inlineTagRegexp := regexp.MustCompile("^(b|big|i|small|tt|abbr|acronym|cite|dfn|em|kbd|strong|samp|var|a|bdo|br|img|"+
		"map|object|q|span|sub|sup|button|input|label|select)$")

	// state
	var text []byte // write text token until next token is received, allows to look forward one token
	var specialTag []html.Token // stack array of special tags it is in
	var prevElementToken html.Token
	precededBySpace := true //  on true the next text token must no start with a space

	defaultScriptType := "text/javascript"
	defaultStyleType := "text/css"

	getAttr := func(token html.Token, k string) string {
		for _, attr := range token.Attr {
			if attr.Key == k {
				return attr.Val
			}
		}
		return ""
	}

	buffer := new(bytes.Buffer)
	z := html.NewTokenizer(r)
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			if z.Err() == io.EOF {
				buffer.Write(text)
				return ioutil.NopCloser(buffer), nil
			}
			return nil, z.Err()
		case html.DoctypeToken:
			buffer.Write(bytes.TrimSpace(text)); text = nil

			// https://developers.google.com/speed/articles/html5-performance
			buffer.WriteString("<!doctype html>")
		case html.TextToken:
			buffer.Write(text)
			text = z.Text()

			// CSS and JS minifiers for inline code
			if len(specialTag) > 0 {
				tag := specialTag[len(specialTag) - 1].Data
				if tag == "style" || tag == "script" {
					val := getAttr(specialTag[len(specialTag) - 1], "type")
					if val == "" {
						// default
						if tag == "script" {
							val = defaultScriptType
						} else {
							val = defaultStyleType
						}
					}

					if val == "text/javascript" {
						text = inline(minify.Js, text)
					} else if val == "text/css" {
						text = inline(minify.Css, text)
					}
				}
				buffer.Write(text); text = nil
				break
			}

			// whitespace removal; if after an inline element, trim left if precededBySpace
			text = multipleWhitespaceRegexp.ReplaceAll(text, []byte(" "))
			if inlineTagRegexp.MatchString(prevElementToken.Data) {
				if precededBySpace {
					text = bytes.TrimLeft(text, " ")
				}
				precededBySpace = len(text) > 0 && text[len(text) - 1] == ' '
			} else {
				text = bytes.TrimLeft(text, " ")
			}
		case html.StartTagToken, html.EndTagToken, html.SelfClosingTagToken:
			token := z.Token()
			prevElementToken = token

			if specialTagRegexp.MatchString(token.Data) {
				if tt == html.StartTagToken {
					specialTag = append(specialTag, token)
				} else if tt == html.EndTagToken {
					specialTag = specialTag[:len(specialTag) - 1]
				}
			}

			// whitespace removal; if we encounter a block or a (closing) inline element, trim the right
			if !inlineTagRegexp.MatchString(token.Data) || (tt == html.EndTagToken && len(text) > 0 && text[len(text) - 1] == ' ') {
				text = bytes.TrimRight(text, " ")
				precededBySpace = true
			}
			buffer.Write(text); text = nil

			if token.Data == "body" || token.Data == "head" || token.Data == "html" || token.Data == "tbody" ||
					tt == html.EndTagToken && (
						token.Data == "colgroup" || token.Data == "dd" || token.Data == "dt" || token.Data == "li" ||
						token.Data == "option" || token.Data == "p" || token.Data == "td" || token.Data == "tfoot" ||
						token.Data == "th" || token.Data == "thead" || token.Data == "tr") {
				break
			}

			buffer.WriteByte('<')
			if tt == html.EndTagToken {
				buffer.WriteByte('/')
			}
			buffer.WriteString(token.Data)

			// rewrite charset https://developers.google.com/speed/articles/html5-performance
			if token.Data == "meta" && getAttr(token, "http-equiv") == "content-type" &&
					getAttr(token, "content") == "text/html; charset=utf-8" {
				buffer.WriteString(" charset=utf-8>")
				break
			}

			// output attributes
			for _, attr := range token.Attr {
				val := strings.TrimSpace(attr.Val)
				val = strings.Replace(val, "&", "&amp;", -1)
				val = strings.Replace(val, "<", "&lt;", -1)

				// default attribute values can be ommited
				if attr.Key == "clear" && val == "none" ||
						attr.Key == "colspan" && val == "1" ||
						attr.Key == "enctype" && val == "application/x-www-form-urlencoded" ||
						attr.Key == "frameborder" && val == "1" ||
						attr.Key == "method" && val == "get" ||
						attr.Key == "rowspan" && val == "1" ||
						attr.Key == "scrolling" && val == "auto" ||
						attr.Key == "shape" && val == "rect" ||
						attr.Key == "span" && val == "1" ||
						attr.Key == "valuetype" && val == "data" ||
						attr.Key == "type" && (
							token.Data == "script" && val == "text/javascript" ||
							token.Data == "style" && val == "text/css" ||
							token.Data == "link" && val == "text/css" ||
							token.Data == "input" && val == "text" ||
							token.Data == "button" && val == "submit") {
					continue
				}

				buffer.WriteByte(' ')
				buffer.WriteString(attr.Key)

				isBoolean := booleanAttrRegexp.MatchString(attr.Key)
				if len(val) == 0 && !isBoolean {
					continue
				}

				// booleans have no value
				if !isBoolean {
					buffer.WriteByte('=')

					// CSS and JS minifiers for attribute inline code
					if attr.Key == "style" {
						val = inlineString(minify.Css, val)
					} else if eventAttrRegexp.MatchString(attr.Key) {
						val = strings.TrimLeft(val, "javascript:")
						val = inlineString(minify.Js, val)
					} else if (attr.Key == "href" || attr.Key == "src" || attr.Key == "cite" || attr.Key == "action") &&
							getAttr(token, "rel") != "external" || attr.Key == "profile" || attr.Key == "xmlns" {
						val = strings.TrimLeft(val, "http:")
						val = strings.TrimLeft(val, "https:")
					} else if token.Data == "meta" && attr.Key == "content" {
						http_equiv := getAttr(token, "http-equiv")
						if http_equiv == "content-type" {
							val = strings.Replace(val, ", ", ",", -1)
						} else if http_equiv == "content-script-type" {
							defaultScriptType = val
						} else if http_equiv == "content-style-type" {
							defaultStyleType = val
						}

						name := getAttr(token, "name")
						if name == "keywords" {
							val = strings.Replace(val, ", ", ",", -1)
						} else if name == "viewport" {
							val = strings.Replace(val, " ", "", -1)
						}
					}

					// no quote if possible, else prefer single or double depending on which occurs more often in value
					if validAttrRegexp.MatchString(val) {
						buffer.WriteString(val)
					} else if strings.Count(val, "\"") > strings.Count(val, "'") {
						buffer.WriteByte('\'')
						buffer.WriteString(strings.Replace(val, "'", "&#39;", -1))
						buffer.WriteByte('\'')
					} else {
						buffer.WriteByte('"')
						buffer.WriteString(strings.Replace(val, "\"", "&quot;", -1))
						buffer.WriteByte('"')
					}
				}
			}

			if tt == html.SelfClosingTagToken {
				buffer.WriteByte('/')
			}
			buffer.WriteByte('>')
		}
	}
}