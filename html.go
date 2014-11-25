package minify

import (
	"bytes"
	"io"
	"strings"

	"code.google.com/p/go.net/html"
)

func replaceMultipleWhitespace(s []byte) []byte {
	j := 0
	t := make([]byte, len(s))
	previousSpace := false
	for _, x := range s {
		if strings.IndexByte(" \t\n\f\r", x) == -1 {
			previousSpace = false
			t[j] = x
			j++
		} else if !previousSpace {
			previousSpace = true
			t[j] = ' '
			j++
		}
	}
	return t[:j]
}

// HTML minifies HTML5 files, it reads from r and writes to w.
// Removes unnecessary whitespace, tags, attributes, quotes and comments and typically saves 10% in size.
func (m Minifier) HTML(w io.Writer, r io.Reader) error {
	invalidAttrChars := " \t\n\f\r\"'`=<>/"

	booleanAttrMap := make(map[string]bool)
	for _, v := range strings.Split("allowfullscreen|async|autofocus|autoplay|checked|compact|controls|declare|" +
		"default|defaultChecked|defaultMuted|defaultSelected|defer|disabled|draggable|enabled|formnovalidate|hidden|" +
		"undeterminate|inert|ismap|itemscope|multiple|muted|nohref|noresize|noshade|novalidate|nowrap|open|pauseonexit|" +
		"readonly|required|reversed|scoped|seamless|selected|sortable|spellcheck|translate|truespeed|typemustmatch|" +
		"visible", "|") {
		booleanAttrMap[v] = true
	}

	specialTagMap := make(map[string]bool)
	for _, v := range strings.Split("style|script|pre|code|textarea", "|") {
		specialTagMap[v] = true
	}

	inlineTagMap := make(map[string]bool)
	for _, v := range strings.Split("b|big|i|small|tt|abbr|acronym|cite|dfn|em|kbd|strong|samp|var|a|bdo|br|img|map|object|q|span|sub|sup|button|input|label|select", "|") {
		inlineTagMap[v] = true
	}

	// state
	var text []byte             // write text token until next token is received, allows to look forward one token before writing away
	var specialTag []html.Token // stack array of special tags it is in
	var prevElementToken html.Token
	precededBySpace := true 	// on true the next text token must no start with a space
	defaultScriptType := "text/javascript"
	defaultStyleType := "text/css"

	getAttr := func(token html.Token, k string) string {
		for _, attr := range token.Attr {
			if attr.Key == k {
				return strings.ToLower(attr.Val)
			}
		}
		return ""
	}

	z := html.NewTokenizer(r)
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			if z.Err() == io.EOF {
				w.Write(text)
				return nil
			}
			return z.Err()
		case html.DoctypeToken:
			w.Write(bytes.TrimSpace(text))
			text = nil

			w.Write([]byte("<!doctype html>"))
		case html.CommentToken:
			w.Write(text)
			text = nil

			comment := string(z.Token().Data)
			if len(comment) > 0 {
				// TODO: ensure that nested comments are handled properly (tokenizer doesn't handle this!)
				if strings.HasPrefix(comment, "[if") {
					text = []byte("<!--" + comment + "-->")
				} else if strings.HasSuffix(comment, "--") {
					// only occurs when mixed up with conditional comments
					text = []byte("<!" + comment + ">")
				}
			}
		case html.TextToken:
			w.Write(text)
			text = z.Text()

			// CSS and JS minifiers for inline code
			if len(specialTag) > 0 {
				if tag := specialTag[len(specialTag)-1].Data; tag == "style" || tag == "script" {
					mime := getAttr(specialTag[len(specialTag)-1], "type")
					if mime == "" {
						// default mime types
						if tag == "script" {
							mime = defaultScriptType
						} else {
							mime = defaultStyleType
						}
					}

					if err := m.Minify(mime, w, bytes.NewBuffer(text)); err != nil {
						if err == ErrNotExist {
							// no minifier, write the original
							w.Write(text)
						} else {
							return err
						}
					}
				} else {
					w.Write(text)
				}
				text = nil
				break
			}

			// whitespace removal; if after an inline element, trim left if precededBySpace
			text = replaceMultipleWhitespace(text)
			if inlineTagMap[prevElementToken.Data] {
				if precededBySpace && len(text) > 0 && text[0] == ' ' {
					text = text[1:]
				}
				precededBySpace = len(text) > 0 && text[len(text)-1] == ' '
			} else if len(text) > 0 && text[0] == ' ' {
				text = text[1:]
			}
		case html.StartTagToken, html.EndTagToken, html.SelfClosingTagToken:
			token := z.Token()
			prevElementToken = token

			if specialTagMap[token.Data] {
				if tt == html.StartTagToken {
					specialTag = append(specialTag, token)
				} else if tt == html.EndTagToken && len(specialTag) > 0 && specialTag[len(specialTag)-1].Data == token.Data {
					// TODO: test whether the if statement is error proof
					specialTag = specialTag[:len(specialTag)-1]
				}
			}

			// whitespace removal; if we encounter a block or a (closing) inline element, trim the right
			if !inlineTagMap[token.Data] || (tt == html.EndTagToken && len(text) > 0 && text[len(text)-1] == ' ') {
				text = bytes.TrimRight(text, " ")
				precededBySpace = true
			}
			w.Write(text)
			text = nil

			if token.Data == "body" || token.Data == "head" || token.Data == "html" || token.Data == "tbody" ||
				tt == html.EndTagToken && (token.Data == "colgroup" || token.Data == "dd" || token.Data == "dt" || token.Data == "li" ||
					token.Data == "option" || token.Data == "p" || token.Data == "td" || token.Data == "tfoot" ||
					token.Data == "th" || token.Data == "thead" || token.Data == "tr") {
				break
			}

			w.Write([]byte("<"))
			if tt == html.EndTagToken {
				w.Write([]byte("/"))
			}
			w.Write([]byte(token.Data))

			if token.Data == "meta" && getAttr(token, "http-equiv") == "content-type" &&
				getAttr(token, "content") == "text/html; charset=utf-8" {
				w.Write([]byte(" charset=utf-8>"))
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
					attr.Key == "type" && (token.Data == "script" && val == "text/javascript" ||
						token.Data == "style" && val == "text/css" ||
						token.Data == "link" && val == "text/css" ||
						token.Data == "input" && val == "text" ||
						token.Data == "button" && val == "submit") {
					continue
				}
				w.Write([]byte(" " + attr.Key))

				isBoolean := booleanAttrMap[attr.Key]
				if len(val) == 0 && !isBoolean {
					continue
				}

				// booleans have no value
				if !isBoolean {
					var err error
					w.Write([]byte("="))

					// CSS and JS minifiers for attribute inline code
					if attr.Key == "style" {
						val, err = m.MinifyString(defaultStyleType, val)
						if err != nil && err != ErrNotExist {
							return err
						}
					} else if strings.HasPrefix(attr.Key, "on") {
						if strings.HasPrefix(val, "javascript:") {
							val = val[11:]
						}
						val, err = m.MinifyString(defaultScriptType, val)
						if err != nil && err != ErrNotExist {
							return err
						}
					} else if ((attr.Key == "href" || attr.Key == "src" || attr.Key == "cite" || attr.Key == "action") && getAttr(token, "rel") != "external") ||
							attr.Key == "profile" || attr.Key == "xmlns" || attr.Key == "formaction" || attr.Key == "poster" || attr.Key == "manifest" ||
							attr.Key == "icon" || attr.Key == "codebase" || attr.Key == "longdesc" || attr.Key == "background" || attr.Key == "icon" ||
							attr.Key == "classid" || attr.Key == "usemap" || attr.Key == "data" {
						if strings.HasPrefix(val, "http:") {
							val = val[5:]
						}
					} else if token.Data == "meta" && attr.Key == "content" {
						http_equiv := getAttr(token, "http-equiv")
						if http_equiv == "content-type" {
							val = strings.Replace(val, ", ", ",", -1)
						} else if http_equiv == "content-style-type" {
							defaultStyleType = val
						} else if http_equiv == "content-script-type" {
							defaultScriptType = val
						}

						name := getAttr(token, "name")
						if name == "keywords" {
							val = strings.Replace(val, ", ", ",", -1)
						} else if name == "viewport" {
							val = strings.Replace(val, " ", "", -1)
						}
					}

					// no quote if possible, else prefer single or double depending on which occurs more often in value
					if strings.IndexAny(val, invalidAttrChars) == -1 {
						w.Write([]byte(val))
					} else if strings.Count(val, "\"") > strings.Count(val, "'") {
						w.Write([]byte("'" + strings.Replace(val, "'", "&#39;", -1) + "'"))
					} else {
						w.Write([]byte("\"" + strings.Replace(val, "\"", "&quot;", -1) + "\""))
					}
				}
			}
			w.Write([]byte(">"))
		}
	}
}
