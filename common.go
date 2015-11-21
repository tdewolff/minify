package minify // import "github.com/tdewolff/minify"

import (
	"bytes"
	"encoding/base64"
	"net/url"

	"github.com/tdewolff/parse"
)

// Epsilon is the closest number to zero that is not considered to be zero.
var Epsilon = 0.00001

// ContentType minifies a given mediatype by removing all whitespace.
func ContentType(b []byte) []byte {
	j := 0
	start := 0
	inString := false
	for i := 0; i < len(b); i++ {
		c := b[i]
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
		for i := 0; i < len(dataURI); i++ {
			c := dataURI[i]
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
func Number(num []byte) []byte {
	// omit first + and register mantissa start and end, whether it's negative and the exponent
	neg := false
	start := 0
	dot := -1
	end := len(num)
	exp := int64(0)
	if 0 < len(num) && (num[0] == '+' || num[0] == '-') {
		if num[0] == '-' {
			neg = true
			start++
		} else {
			num = num[1:]
			end--
		}
	}
	for i := 0; i < len(num); i++ {
		c := num[i]
		if c == '.' {
			dot = i
		} else if c == 'e' || c == 'E' {
			end = i
			i++
			if i < len(num) && num[i] == '+' {
				i++
			}
			var ok bool
			if exp, ok = parse.Int(num[i:]); !ok {
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

	// shorten mantissa by increasing/decreasing the exponent
	if end == dot {
		for i := end - 1; i >= start; i-- {
			if num[i] != '0' {
				exp += int64(end - i - 1)
				end = i + 1
				break
			}
		}
	} else {
		exp -= int64(end - dot - 1)
		if start == dot {
			for i = dot + 1; i < end; i++ {
				if num[i] != '0' {
					copy(num[dot:], num[i:end])
					end -= i - dot
					break
				}
			}
		} else {
			copy(num[dot:], num[dot+1:end])
			end--
		}
	}

	// append the exponent or change the mantissa to incorporate the exponent
	relExp := exp + int64(end-start) // exp when the first non-zero digit is directly after the dot
	n := lenInt64(exp)               // number of exp digits
	if exp == 0 {
		if neg {
			start--
			num[start] = '-'
		}
		return num[start:end]
	} else if int(relExp)+n+1 < 0 || 2 < exp { // add exponent for exp 3 and higher and where a lower exp really makes it shorter
		num[end] = 'e'
		end++
		if exp < 0 {
			num[end] = '-'
			end++
			exp = -exp
		}
		for i := end + n - 1; i >= end; i-- {
			num[i] = byte(exp%10) + '0'
			exp /= 10
		}
		end += n
	} else if exp < 0 { // omit exponent
		if relExp > 0 {
			copy(num[start+int(relExp)+1:], num[start+int(relExp):end])
			num[start+int(relExp)] = '.'
			end++
		} else {
			copy(num[start-int(relExp)+1:], num[start:end])
			num[start] = '.'
			for i := 1; i < -int(relExp)+1; i++ {
				num[start+i] = '0'
			}
			end -= int(relExp) - 1
		}
	} else { // for exponent 1 and 2
		num[end] = '0'
		if exp == 2 {
			num[end+1] = '0'
		}
		end += int(exp)
	}

	if neg {
		start--
		num[start] = '-'
	}
	return num[start:end]
}

func lenInt64(i int64) int {
	if i < 0 {
		i = -i
	}
	switch {
	case i < 10:
		return 1
	case i < 100:
		return 2
	case i < 1000:
		return 3
	case i < 10000:
		return 4
	case i < 100000:
		return 5
	case i < 1000000:
		return 6
	case i < 10000000:
		return 7
	case i < 100000000:
		return 8
	case i < 1000000000:
		return 9
	case i < 10000000000:
		return 10
	case i < 100000000000:
		return 11
	case i < 1000000000000:
		return 12
	case i < 10000000000000:
		return 13
	case i < 100000000000000:
		return 14
	case i < 1000000000000000:
		return 15
	case i < 10000000000000000:
		return 16
	case i < 100000000000000000:
		return 17
	case i < 1000000000000000000:
		return 18
	}
	return 19
}
