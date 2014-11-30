package minify

// TODO: use a better tokenizer
/* TODO: (non-exhaustive)
- remove space before !important
- collapse margin/padding/border/background/list/etc. definitions into one
- remove empty or with duplicate selector blocks
- shorten zero values (none/0px/0pt etc. become 0)
- remove quotes within url()?
*/

/*
Uses http://www.w3.org/TR/2010/PR-css3-color-20101028/ for colors
*/

import (
	"encoding/hex"
	"errors"
	"io"
	"strconv"
	"strings"

	"github.com/tdewolff/css"
)

var epsilon = 0.00001

// hex values with a shorter color name
var hexColors = map[string]string{
	"000080": "navy",
	"008000": "green",
	"008080": "teal",
	"4B0082": "indigo",
	"800000": "maroon",
	"800080": "purple",
	"808000": "olive",
	"808080": "gray",
	"A0522D": "sienna",
	"A52A2A": "brown",
	"C0C0C0": "silver",
	"CD853F": "peru",
	"D2B48C": "tan",
	"DA70D6": "orchid",
	"DDA0DD": "plum",
	"EE82EE": "violet",
	"F0E68C": "khaki",
	"F0FFFF": "azure",
	"F5DEB3": "wheat",
	"F5F5DC": "beige",
	"FA8072": "salmon",
	"FAF0E6": "linen",
	"FF6347": "tomato",
	"FF7F50": "coral",
	"FFA500": "orange",
	"FFC0CB": "pink",
	"FFD700": "gold",
	"FFE4C4": "bisque",
	"FFFAFA": "snow",
	"FFFFF0": "ivory",
	"FF0000": "red",
	"F00":    "red",
}

var colorNames = map[string]string{
	"black":                "000",
	"darkblue":             "00008B",
	"mediumblue":           "0000CD",
	"darkgreen":            "006400",
	"darkcyan":             "008B8B",
	"deepskyblue":          "00BFFF",
	"darkturquoise":        "00CED1",
	"mediumspringgreen":    "00FA9A",
	"springgreen":          "00FF7F",
	"midnightblue":         "191970",
	"dodgerblue":           "1E90FF",
	"lightseagreen":        "20B2AA",
	"forestgreen":          "228B22",
	"seagreen":             "2E8B57",
	"darkslategray":        "2F4F4F",
	"limegreen":            "32CD32",
	"mediumseagreen":       "3CB371",
	"turquoise":            "40E0D0",
	"royalblue":            "4169E1",
	"steelblue":            "4682B4",
	"darkslateblue":        "483D8B",
	"mediumturquoise":      "48D1CC",
	"darkolivegreen":       "556B2F",
	"cadetblue":            "5F9EA0",
	"cornflowerblue":       "6495ED",
	"mediumaquamarine":     "66CDAA",
	"slateblue":            "6A5ACD",
	"olivedrab":            "6B8E23",
	"slategray":            "708090",
	"lightslateblue":       "789",
	"mediumslateblue":      "7B68EE",
	"lawngreen":            "7CFC00",
	"chartreuse":           "7FFF00",
	"aquamarine":           "7FFFD4",
	"lightskyblue":         "87CEFA",
	"blueviolet":           "8A2BE2",
	"darkmagenta":          "8B008B",
	"saddlebrown":          "8B4513",
	"darkseagreen":         "8FBC8F",
	"lightgreen":           "90EE90",
	"mediumpurple":         "9370DB",
	"darkviolet":           "9400D3",
	"palegreen":            "98FB98",
	"darkorchid":           "9932CC",
	"yellowgreen":          "9ACD32",
	"darkgray":             "A9A9A9",
	"lightblue":            "ADD8E6",
	"greenyellow":          "ADFF2F",
	"paleturquoise":        "AFEEEE",
	"lightsteelblue":       "B0C4DE",
	"powderblue":           "B0E0E6",
	"firebrick":            "B22222",
	"darkgoldenrod":        "B8860B",
	"mediumorchid":         "BA55D3",
	"rosybrown":            "BC8F8F",
	"darkkhaki":            "BDB76B",
	"mediumvioletred":      "C71585",
	"indianred":            "CD5C5C",
	"chocolate":            "D2691E",
	"lightgray":            "D3D3D3",
	"goldenrod":            "DAA520",
	"palevioletred":        "DB7093",
	"gainsboro":            "DCDCDC",
	"burlywood":            "DEB887",
	"lightcyan":            "E0FFFF",
	"lavender":             "E6E6FA",
	"darksalmon":           "E9967A",
	"palegoldenrod":        "EEE8AA",
	"lightcoral":           "F08080",
	"aliceblue":            "F0F8FF",
	"honeydew":             "F0FFF0",
	"sandybrown":           "F4A460",
	"whitesmoke":           "F5F5F5",
	"mintcream":            "F5FFFA",
	"ghostwhite":           "F8F8FF",
	"antiquewhite":         "FAEBD7",
	"lightgoldenrodyellow": "FAFAD2",
	"fuchsia":              "F0F",
	"magenta":              "F0F",
	"deeppink":             "FF1493",
	"orangered":            "FF4500",
	"darkorange":           "FF8C00",
	"lightsalmon":          "FFA07A",
	"lightpink":            "FFB6C1",
	"peachpuff":            "FFDAB9",
	"navajowhite":          "FFDEAD",
	"moccasin":             "FFE4B5",
	"mistyrose":            "FFE4E1",
	"blanchedalmond":       "FFEBCD",
	"papayawhip":           "FFEFD5",
	"lavenderblush":        "FFF0F5",
	"seashell":             "FFF5EE",
	"cornsilk":             "FFF8DC",
	"lemonchiffon":         "FFFACD",
	"floralwhite":          "FFFAF0",
	"yellow":               "FF0",
	"lightyellow":          "FFFFE0",
	"white":                "FFF",
}

var errParse = errors.New("parse error")

func shortenHex(s string) string {
	s = strings.ToUpper(s)
	if ident, ok := hexColors[s[1:]]; ok {
		return ident
	}

	if len(s) == 7 && s[1] == s[2] && s[3] == s[4] && s[5] == s[6] {
		return "#"+string(s[1])+string(s[3])+string(s[5])
	}
	return s
}

type Token struct {
	tt css.TokenType
	val string
}

func propVals(z *css.Tokenizer) ([]string, string) {
	raw := ""
	vals := []string{}
loop:
	for {
		tt := z.Next()
		switch tt {
		case css.SemicolonToken, css.RightBraceToken, css.ErrorToken:
			break loop
		case css.WhitespaceToken:
			raw += " "
		default:
			vals = append(vals, z.String())
			raw += z.String()
		}
	}
	return vals, strings.TrimSpace(raw)
}

func funcParams(z *css.Tokenizer) ([]Token, string) {
	raw := ""
	params := []Token{}
loop:
	for {
		tt := z.Next()
		switch tt {
		case css.ErrorToken:
			break loop
		case css.RightParenthesisToken:
			raw += z.String()
			break loop
		case css.WhitespaceToken:
		case css.CommaToken:
			raw += z.String()
		default:
			params = append(params, Token{tt, z.String()})
			raw += z.String()
		}
	}
	return params, raw
}

// CSS minifies CSS files, it reads from r and writes to w.
// It does a mediocre job of minifying CSS files and should be improved in the future.
func (m Minifier) CSS(w io.Writer, r io.Reader) error {
	semicolonQueued := false
	lastToken := css.ErrorToken
	lastIdent := ""

	z := css.NewTokenizer(r)
	var tt css.TokenType
	for {
		if tt != css.WhitespaceToken {
			lastToken = tt
		}
		tt = z.Next()
		if tt == css.WhitespaceToken {
			continue
		}

		// whitespace removal correction
		if (lastToken == css.NumberToken || lastToken == css.IdentToken) && (tt == css.NumberToken || tt == css.IdentToken) {
			w.Write([]byte(" "))
		}

		// semicolon removal correction
		if semicolonQueued {
			if tt != css.RightBraceToken && tt != css.ErrorToken {
				w.Write([]byte(";"))
			}
			semicolonQueued = false
		}

		switch tt {
		case css.ErrorToken:
			if z.Err() != io.EOF {
				return z.Err()
			}
			return nil
		case css.SemicolonToken:
			semicolonQueued = true
		case css.HashToken:
			h := shortenHex(z.String())
			w.Write([]byte(h))
		case css.IdentToken:
			ident := z.String()
			if h, ok := colorNames[ident]; ok {
				w.Write([]byte("#"+h))
				break
			}

			if lastIdent == "font-weight" && ident == "bold" {
				ident = "700"
			} else if lastIdent == "font-weight" && ident == "normal" {
				ident = "400"
			} else if lastIdent == "outline" && ident == "none" {
				ident = "0"
			}
			w.Write([]byte(ident))
			lastIdent = ident
		case css.NumberToken, css.DimensionToken, css.PercentageToken:
			if lastIdent == "margin" || lastIdent == "padding" {
				curr := z.String()
				vals, raw := propVals(z)
				vals = append([]string{curr}, vals...)
				raw = curr+" "+raw
				if len(vals) == 2 {
					if vals[0] == vals[1] {
						w.Write([]byte(vals[0]))
						break
					}
				} else if len(vals) == 3 {
					if vals[0] == vals[1] && vals[0] == vals[2] {
						w.Write([]byte(vals[0]))
						break
					} else if vals[0] == vals[2] {
						w.Write([]byte(vals[0]+" "+vals[1]))
						break
					}
				} else if len(vals) == 4 {
					if vals[0] == vals[1] && vals[0] == vals[2] && vals[0] == vals[3] {
						w.Write([]byte(vals[0]))
						break
					} else if vals[0] == vals[2] && vals[1] == vals[3] {
						w.Write([]byte(vals[0]+" "+vals[1]))
						break
					} else if vals[1] == vals[3] {
						w.Write([]byte(vals[0]+" "+vals[1]+" "+vals[2]))
						break
					}
				}
				w.Write([]byte(raw))
				break
			}
		case css.FunctionToken:
			var err error
			f := z.String()
			params, raw := funcParams(z)
			raw = f+raw
			if f == "rgba(" && len(params) == 4 {
				d, _ := strconv.ParseFloat(params[3].val[:len(params[3].val)-1], 32)
				if d - 1.0 < epsilon {
					f = "rgb("
					params = params[:len(params)-1]
				}
			}
			if f == "rgb(" && len(params) == 3 {
				rgb := make([]byte, 3)
				for i := 0; i < 3; i++ {
					if params[i].tt == css.NumberToken {
						var d int64
						d, err = strconv.ParseInt(params[i].val, 10, 32)
						if d < 0 {
							d = 0
						} else if d > 255 {
							d = 255
						}
						rgb[i] = byte(d)
					} else if params[i].tt == css.PercentageToken {
						var d float64
						d, err = strconv.ParseFloat(params[i].val[:len(params[i].val)-1], 32)
						if d < 0.0 {
							d = 0.0
						} else if d > 100.0 {
							d = 100.0
						}
						rgb[i] = byte((d/100.0*255.0)+0.5)
					} else {
						err = errParse
					}
				}
				if err == nil {
					h := shortenHex("#"+hex.EncodeToString(rgb))
					w.Write([]byte(h))
					break
				}
			}
			w.Write([]byte(raw))
		default:
			w.Write(z.Bytes())
		}
	}
}
