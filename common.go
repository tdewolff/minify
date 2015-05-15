package minify // import "github.com/tdewolff/minify"

import (
	"bytes"
	"encoding/base64"
	"math"
	"net/url"
	"strconv"

	"github.com/tdewolff/parse"
)

// Epsilon is the closest number to zero that is not considered to be zero.
var Epsilon = 0.00001

var (
	zeroBytes = []byte("0")
)

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
func DataURI(m Minifier, dataURI []byte) []byte {
	if mediatype, data, err := parse.DataURI(dataURI); err == nil {
		dataURI, _ = Bytes(m, string(mediatype), data)
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
		if len(mediatype) >= len("text/plain") && bytes.HasPrefix(mediatype, []byte("text/plain")) {
			mediatype = mediatype[len("text/plain"):]
		}
		dataURI = append(append(append([]byte("data:"), mediatype...), ','), dataURI...)
	}
	return dataURI
}

// TODO: omit ParseFloat in favor of counting digit-zero bytes on the left, also useful for using exponents as shorter notations
// Number minifies a given byte slice containing a number (see parse.Number) and remove superfluous characters.
func Number(num []byte) []byte {
	f, err := strconv.ParseFloat(string(num), 64)
	if err != nil {
		return num
	}
	if math.Abs(f) < Epsilon {
		return zeroBytes
	}
	if num[0] == '-' {
		n := 1
		for n < len(num) && num[n] == '0' {
			n++
		}
		num = num[n-1:]
		num[0] = '-'
	} else {
		if num[0] == '+' {
			num = num[1:]
		}
		// trim 0 left
		for len(num) > 0 && num[0] == '0' {
			num = num[1:]
		}
	}
	// trim 0 right
	for i, digit := range num {
		if digit == '.' {
			j := len(num) - 1
			for ; j > i; j-- {
				if num[j] == '0' {
					num = num[:len(num)-1]
				} else {
					break
				}
			}
			if j == i {
				num = num[:len(num)-1] // remove .
			}
			break
		}
	}
	return num
}
