package minify // import "github.com/tdewolff/minify"

import (
	"bytes"
	"encoding/base64"
	"net/url"

	"github.com/tdewolff/parse"
	"github.com/tdewolff/strconv"
)

// Epsilon is the closest number to zero that is not considered to be zero.
var Epsilon = 0.00001

// ContentType minifies a given mediatype by removing all whitespace.
func ContentType(b []byte) []byte {
	j := 0
	start := 0
	inString := false
	for i, c := range b {
		if !inString && parse.IsWhitespace(c) {
			if start != 0 {
				j += copy(b[j:], b[start:i])
			} else {
				j += i
			}
			start = i + 1
		} else if c == '"' {
			inString = !inString
		}
	}
	if start != 0 {
		j += copy(b[j:], b[start:])
		return parse.ToLower(b[:j])
	}
	return parse.ToLower(b)
}

// DataURI minifies a data URI and calls a minifier by the specified mediatype. Specifications: https://www.ietf.org/rfc/rfc2397.txt.
func DataURI(m *M, dataURI []byte) []byte {
	if mediatype, data, err := parse.DataURI(dataURI); err == nil {
		dataURI, _ = m.Bytes(string(mediatype), data)
		base64Len := len(";base64") + base64.StdEncoding.EncodedLen(len(dataURI))
		asciiLen := len(dataURI)
		for _, c := range dataURI {
			if 'A' <= c && c <= 'Z' || 'a' <= c && c <= 'z' || '0' <= c && c <= '9' || c == '-' || c == '_' || c == '.' || c == '~' || c == ' ' {
				asciiLen++
			} else {
				asciiLen += 2
			}
			if asciiLen > base64Len {
				break
			}
		}
		if asciiLen > base64Len {
			encoded := make([]byte, base64Len-len(";base64"))
			base64.StdEncoding.Encode(encoded, dataURI)
			dataURI = encoded
			mediatype = append(mediatype, []byte(";base64")...)
		} else {
			dataURI = []byte(url.QueryEscape(string(dataURI)))
			dataURI = bytes.Replace(dataURI, []byte("\""), []byte("\\\""), -1)
		}
		if len("text/plain") <= len(mediatype) && parse.EqualFold(mediatype[:len("text/plain")], []byte("text/plain")) {
			mediatype = mediatype[len("text/plain"):]
		}
		for i := 0; i+len(";charset=us-ascii") <= len(mediatype); i++ {
			// must start with semicolon and be followed by end of mediatype or semicolon
			if mediatype[i] == ';' && parse.EqualFold(mediatype[i+1:i+len(";charset=us-ascii")], []byte("charset=us-ascii")) && (i+len(";charset=us-ascii") >= len(mediatype) || mediatype[i+len(";charset=us-ascii")] == ';') {
				mediatype = append(mediatype[:i], mediatype[i+len(";charset=us-ascii"):]...)
				break
			}
		}
		dataURI = append(append(append([]byte("data:"), mediatype...), ','), dataURI...)
	}
	return dataURI
}

// Number minifies a given byte slice containing a number (see parse.Number) and removes superfluous characters.
func Number(num []byte, prec int) []byte {
	// omit first + and register mantissa start and end, whether it's negative and the exponent
	neg := false
	start := 0
	dot := -1
	end := len(num)
	origExp := int64(0)
	if 0 < len(num) && (num[0] == '+' || num[0] == '-') {
		if num[0] == '-' {
			neg = true
		}
		start++
	}
	for i := start; i < len(num); i++ {
		c := num[i]
		if c == '.' {
			dot = i
		} else if c == 'e' || c == 'E' {
			end = i
			i++
			if i < len(num) && num[i] == '+' {
				i++
			}
			var n int
			if origExp, n = strconv.ParseInt(num[i:]); n == 0 {
				return num
			}
			break
		}
	}
	if dot == -1 {
		dot = end
	}

	// trim leading zeros but leave at least one digit
	for start < end-1 && num[start] == '0' {
		start++
	}
	// trim trailing zeros
	i := end - 1
	for ; i > dot; i-- {
		if num[i] != '0' {
			end = i + 1
			break
		}
	}
	if i == dot {
		end = dot
		if start == end {
			num[start] = '0'
			return num[start : start+1]
		}
	} else if start == end-1 && num[start] == '0' {
		return num[start:end]
	}

	normExp := 0
	n := 0
	if dot == start {
		for i = dot + 1; i < end; i++ {
			if num[i] != '0' {
				normExp = dot - i + 1
				n = end - i
				break
			}
		}
	} else if dot == end {
		normExp = end - start
		for i := end - 1; i >= start; i-- {
			if num[i] != '0' {
				n = i + 1 - start
				end = i + 1
				break
			}
		}
	} else {
		normExp = dot - start
		n = end - start - 1
	}
	normExp += int(origExp)

	exp := int64(normExp - n)
	lenExp := strconv.LenInt(exp)

	//fmt.Println("normExp", normExp, "n", n, "exp", exp)

	if normExp >= n {
		if dot < end {
			if dot == start {
				start = end - n
			} else {
				// TODO: copy the other part if shorter?
				//fmt.Println("COPY", end-dot-1)
				copy(num[dot:], num[dot+1:end])
				end--
			}
		}
		if normExp >= n+3 {
			num[end] = 'e'
			end++
			for i := end + lenExp - 1; i >= end; i-- {
				num[i] = byte(exp%10) + '0'
				exp /= 10
			}
			end += lenExp
		} else if normExp == n+2 {
			num[end] = '0'
			num[end+1] = '0'
			end += 2
		} else if normExp == n+1 {
			num[end] = '0'
			end++
		}
		//fmt.Println("A", string(num[start:end]))
	} else if normExp >= -lenExp-1 {
		if normExp >= 0 {
			newDot := start + normExp
			if dot != newDot {
				// TODO: copy the other part if shorter
				if dot < newDot {
					//fmt.Println("COPY", newDot-dot)
					copy(num[dot:], num[dot+1:newDot+1])
				} else {
					//fmt.Println("COPY", dot-newDot)
					copy(num[newDot+1:], num[newDot:dot])
				}
				num[newDot] = '.'
				if dot == end {
					end++
				}
			}
		} else if origExp != 0 {
			zeroes := -normExp
			//fmt.Println("COPY", n)
			copy(num[dot+1+zeroes:], num[end-n:end])
			//fmt.Println("COPY", dot-start)
			copy(num[start+1+zeroes:], num[start:dot])
			num[start] = '.'
			for i := 0; i < zeroes; i++ {
				num[start+1+i] = '0'
			}
			end += zeroes
		}
		//fmt.Println("B", string(num[start:end]))
	} else {
		if dot < end {
			if dot == start {
				//fmt.Println("COPY", n)
				copy(num[start:], num[end-n:end])
				end = start + n
			} else {
				//fmt.Println("COPY", end-dot-1)
				copy(num[dot:], num[dot+1:end])
				end--
			}
		}
		num[end] = 'e'
		num[end+1] = '-'
		end += 2
		exp = -exp
		for i := end + lenExp - 1; i >= end; i-- {
			num[i] = byte(exp%10) + '0'
			exp /= 10
		}
		end += lenExp
		//fmt.Println("C", string(num[start:end]))
	}

	if neg {
		start--
		num[start] = '-'
	}
	return num[start:end]
}
