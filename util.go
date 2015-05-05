package minify // import "github.com/tdewolff/minify"

import (
	"bytes"
	"encoding/base64"
	"math"
	"net/url"
	"strconv"

	"github.com/tdewolff/parse"
)

var Epsilon = 0.00001

var (
	zeroBytes = []byte("0")
)

func MinifyDataURI(m Minifier, dataURI []byte) []byte {
	if mediatype, data, ok := parse.SplitDataURI(dataURI); ok {
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

func MinifyNumber(num []byte) []byte {
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
