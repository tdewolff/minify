package minify

// TODO: use a better tokenizer
/* TODO: (non-exhaustive)
- remove space before !important
- collapse margin/padding/border/background/list/etc. definitions into one
- remove empty or with duplicate selector blocks
- shorten zero values (none/0px/0pt etc. become 0)
- remove quotes within url()?
*/

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Minifies CSS files, reads from r and writes to w.
// It does a mediocre job of minifying CSS files and should be improved in the future.
func (m Minifier) CSS(w io.Writer, r io.Reader) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	s := string(b)

	inline := false
	if strings.IndexRune(s, '{') == -1 {
		inline = true
	}

	// hex values with a shorter color name
	hexColors := map[string]string{
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

	colorNames := map[string]string{
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

	whitespace := regexp.MustCompile("\\s+")
	selectors := regexp.MustCompile("\\s?([>,+~]|[~|^$*]?=)\\s?")
	afterValue := false
	var prop string

	l := lex("cssminify", s, inline)
	for {
		i := l.NextItem()
		switch i.typ {
		case itemEOF:
			return nil
		case itemError:
			return errors.New(i.val)
		case itemSelector:
			val := whitespace.ReplaceAllString(i.val, " ")
			val = selectors.ReplaceAllString(val, "$1")
			w.Write([]byte(val))
		case itemProperty:
			prop = i.val
			if afterValue {
				w.Write([]byte(";"))
			}
			w.Write([]byte(i.val+":"))
		case itemValue:
			val := strings.Replace(i.val, ", ", ",", -1)

			if prop == "font-weight" {
				if val == "bold" {
					val = "700"
				} else if val == "normal" {
					val = "400"
				}
			} else if prop == "outline" && val == "none" {
				val = "0"
			}

			if len(val) >= 5 && (strings.Index(val, "rgb(") == 0 || strings.Index(val, "rgba(") == 0) {
				if strings.Index(val, "rgb(") == 0 {
					val = val[4 : len(val)-1]
				} else {
					val = val[5 : len(val)-1]
				}

				params := strings.Split(val, ",")
				if len(params) == 4 {
					if n, _ := strconv.ParseFloat(params[3], 32); n == 1.0 {
						params = params[:3]
					}
				}

				if len(params) == 3 {
					var hexVal uint32
					var err error
					for _, param := range params {
						var f float64
						var n uint32
						if param[len(param)-1] == '%' {
							f, err = strconv.ParseFloat(param[:len(param)-1], 32)
							n = uint32(f*255.0 + 0.5)
						} else {
							f, err = strconv.ParseFloat(param, 32)
							n = uint32(f + 0.5)
						}

						hexVal *= 256
						hexVal += n

						if err != nil {
							break
						}
					}

					if err == nil {
						b := make([]byte, 4)
						binary.LittleEndian.PutUint32(b, hexVal)
						val = "#" + hex.EncodeToString(b)
					}
				}
			}

			if len(val) >= 4 && val[0] == '#' {
				if len(val) == 7 && val[1] == val[2] && val[3] == val[4] && val[5] == val[6] {
					val = "#" + string(val[1]) + string(val[3]) + string(val[5])
				} else if name, ok := hexColors[strings.ToUpper(val[1:])]; ok {
					val = name
				}
			} else if hex, ok := colorNames[strings.ToLower(val)]; ok {
				val = "#" + hex
			}

			w.Write([]byte(val))
			afterValue = true
		case itemComment:
		default:
			w.Write([]byte(i.val))
			afterValue = false
		}
	}
}

////////////////////////////////////////////////////////////////

type item struct {
	typ itemType
	val string
}

type itemType int

const (
	itemError itemType = iota
	itemEOF
	itemSelector
	itemProperty
	itemValue
	itemStartBlock
	itemEndBlock
	itemComment
)

func (i item) String() string {
	switch i.typ {
	case itemEOF:
		return "EOF"
	case itemError:
		return i.val
	}
	return fmt.Sprintf("%d %q", i.typ, i.val)
}

////////////////////////////////////////////////////////////////

type stateFn func(*lexer) stateFn

const eof = 0

type lexer struct {
	name      string // used only for error reports.
	input     string // the string being scanned.
	start     int    // start position of this item.
	pos       int    // current position in the input.
	width     int    // width of last rune read from input.
	state     stateFn
	prevState stateFn
	items     chan item // channel of scanned items.
}

func lex(name, input string, inline bool) *lexer {
	state := lexText
	if inline {
		state = lexBlock
	}

	l := &lexer{
		name:  name,
		input: input,
		state: state,
		items: make(chan item, 2), // Two items sufficient.
	}
	return l
}

func (l *lexer) NextItem() item {
	for {
		select {
		case item := <-l.items:
			return item
		default:
			nextState := l.state(l)
			l.prevState = l.state
			l.state = nextState
		}
	}
	panic("not reached")
}

func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.input[l.start:l.pos]}
	l.start = l.pos
}

func (l *lexer) next() (r rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return r
}

func (l *lexer) ignore() {
	l.start = l.pos
}

func (l *lexer) backup() {
	l.pos -= l.width
}

func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

func (l *lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}

func (l *lexer) ignoreRun(valid string) {
	l.acceptRun(valid)
	l.ignore()
}

func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{
		itemError,
		fmt.Sprintf(format, args...),
	}
	return nil
}

////////////////////////////////////////////////////////////////

func isWhitespace(r rune) bool {
	return r == ' ' || r == '\n' || r == '\r' || r == '\t' || r == '\f'
}

func isSelector(r rune) bool {
	return r != eof && r != '{' && !isWhitespace(r)
}

func isProperty(r rune) bool {
	return r != eof && r != ':' && !isWhitespace(r)
}

func isValue(r rune) bool {
	return r != eof && r != '}' && r != ';'
}

////////////////////////////////////////////////////////////////

func lexText(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == eof:
			l.emit(itemEOF)
			return nil
		case r == '/' && l.peek() == '*':
			return lexComment
		case isSelector(r):
			l.backup()
			return lexSelector
		default:
			l.ignore()
		}
	}
}

func lexSelector(l *lexer) stateFn {
	lastPos := l.pos
	for {
		switch r := l.next(); {
		case r == eof:
			l.emit(itemEOF)
			return nil
		case r == '{':
			oldPos := l.pos
			l.pos = lastPos
			l.emit(itemSelector)

			l.start = oldPos - l.width
			l.pos = oldPos
			l.emit(itemStartBlock)
			return lexBlock
		case isWhitespace(r):
		default:
			lastPos = l.pos
		}
	}
}

func lexBlock(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == eof:
			l.emit(itemEOF)
			return nil
		case r == '/':
			return lexComment
		case r == ';':
			l.ignore()
		case r == '}':
			l.emit(itemEndBlock)
			return lexText
		case isProperty(r):
			l.backup()
			return lexProperty
		default:
			l.ignore()
		}
	}
}

func lexComment(l *lexer) stateFn {
	l.next()
	var r, prevR rune
	for {
		prevR = r
		r = l.next()

		if r == '/' && prevR == '*' {
			l.emit(itemComment)
			return l.prevState
		} else if r == eof {
			l.emit(itemEOF)
			return nil
		}
	}
}

// func lexString(l *lexer) stateFn {
// 	l.next()
// 	var r, prevR rune
// 	for {
// 		prevR = r
// 		r = l.next()

// 		if r == '"' && prevR != '\\' {
// 			l.emit(itemText)
// 			return l.prevState
// 		} else if r == eof {
// 			l.emit(itemEOF)
// 			return nil
// 		}
// 	}
// }

func lexProperty(l *lexer) stateFn {
	for isProperty(l.next()) {
	}
	l.backup()
	l.emit(itemProperty)

	l.ignoreRun("\r\n\f\t :")
	return lexValue
}

func lexValue(l *lexer) stateFn {
	lastPos := l.pos
	for {
		switch r := l.next(); {
		case r == eof || r == ';' || r == '}':
			oldPos := l.pos
			l.pos = lastPos
			l.emit(itemValue)

			if r == eof {
				l.emit(itemEOF)
				return nil
			}
			l.pos = oldPos
			if r == '}' {
				l.backup()
				l.ignore()
			}
			return lexBlock
		case isWhitespace(r):
		default:
			lastPos = l.pos
		}
	}
}
