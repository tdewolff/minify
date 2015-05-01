package css // import "github.com/tdewolff/minify/css"

/*
Uses http://www.w3.org/TR/2010/PR-css3-color-20101028/ for colors
*/

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"io"
	"math"
	"mime"
	"net/url"
	"strconv"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/parse"
	"github.com/tdewolff/parse/css"
)

var epsilon = 0.00001

var (
	spaceBytes        = []byte(" ")
	colonBytes        = []byte(":")
	semicolonBytes    = []byte(";")
	leftBracketBytes  = []byte("{")
	rightBracketBytes = []byte("}")
	zeroBytes         = []byte("0")
	msfilterBytes     = []byte("-ms-filter")
)

type cssMinifier struct {
	m minify.Minifier
	w io.Writer
	p *css.Parser
}

////////////////////////////////////////////////////////////////

// Minify minifies CSS files, it reads from r and writes to w.
func Minify(m minify.Minifier, mediatype string, w io.Writer, r io.Reader) error {
	isStylesheet := true
	if _, params, err := mime.ParseMediaType(mediatype); err == nil && params["inline"] == "1" {
		isStylesheet = false
	}
	c := &cssMinifier{
		m: m,
		w: w,
		p: css.NewParser(r, isStylesheet),
	}

	if err := c.minifyGrammar(); err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}
	return nil
}

func (c *cssMinifier) minifyGrammar() error {
	semicolonQueued := false
	for {
		gt, _, data := c.p.Next()
		if gt == css.ErrorGrammar {
			return c.p.Err()
		} else if gt == css.EndAtRuleGrammar || gt == css.EndRulesetGrammar {
			if _, err := c.w.Write(rightBracketBytes); err != nil {
				return err
			}
			semicolonQueued = false
			continue
		}

		if semicolonQueued {
			if _, err := c.w.Write(semicolonBytes); err != nil {
				return err
			}
			semicolonQueued = false
		}

		if gt == css.AtRuleGrammar {
			if _, err := c.w.Write(data); err != nil {
				return err
			}
			for _, val := range c.p.Values() {
				if _, err := c.w.Write(val.Data); err != nil {
					return err
				}
			}
			semicolonQueued = true
		} else if gt == css.BeginAtRuleGrammar {
			if _, err := c.w.Write(data); err != nil {
				return err
			}
			for _, val := range c.p.Values() {
				if _, err := c.w.Write(val.Data); err != nil {
					return err
				}
			}
			if _, err := c.w.Write(leftBracketBytes); err != nil {
				return err
			}
		} else if gt == css.BeginRulesetGrammar {
			if err := c.minifySelectors(data, c.p.Values()); err != nil {
				return err
			}
			if _, err := c.w.Write(leftBracketBytes); err != nil {
				return err
			}
		} else if gt == css.DeclarationGrammar {
			if _, err := c.w.Write(data); err != nil {
				return err
			}
			if _, err := c.w.Write(colonBytes); err != nil {
				return err
			}
			if err := c.minifyDeclaration(data, c.p.Values()); err != nil {
				return err
			}
			semicolonQueued = true
		} else if _, err := c.w.Write(data); err != nil {
			return err
		}
	}
	return nil
}

func (c *cssMinifier) minifySelectors(property []byte, values []css.Token) error {
	inAttr := false
	isClass := false
	for _, val := range c.p.Values() {
		if !inAttr && val.TokenType == css.LeftBracketToken {
			inAttr = true
		} else if inAttr && val.TokenType == css.RightBracketToken {
			inAttr = false
		} else if inAttr && val.TokenType == css.StringToken {
			s := val.Data[1 : len(val.Data)-1]
			if css.IsIdent([]byte(s)) {
				if _, err := c.w.Write(s); err != nil {
					return err
				}
				continue
			}
		} else if !inAttr && val.TokenType == css.DelimToken && val.Data[0] == '.' {
			isClass = true
		} else if !inAttr && val.TokenType == css.IdentToken {
			if !isClass {
				parse.ToLower(val.Data)
			}
			isClass = false
		}
		if _, err := c.w.Write(val.Data); err != nil {
			return err
		}
	}
	return nil
}

func (c *cssMinifier) minifyDeclaration(property []byte, values []css.Token) error {
	if len(values) == 0 {
		return nil
	}
	prop := css.ToHash(property)
	inProgid := false
	for i, value := range values {
		if inProgid {
			if value.TokenType == css.FunctionToken {
				inProgid = false
			}
			continue
		} else if value.TokenType == css.IdentToken && bytes.Equal(value.Data, []byte("progid")) {
			inProgid = true
			continue
		}
		values[i].TokenType, values[i].Data = c.shortenToken(value.TokenType, value.Data)
		if prop == css.Font || prop == css.Font_Family || prop == css.Font_Weight {
			if value.TokenType == css.IdentToken && (prop == css.Font || prop == css.Font_Weight) {
				val := css.ToHash(value.Data)
				if val == css.Normal && prop == css.Font_Weight {
					// normal could also be specified for font-variant, not just font-weight
					values[i].TokenType = css.NumberToken
					values[i].Data = []byte("400")
				} else if val == css.Bold {
					values[i].TokenType = css.NumberToken
					values[i].Data = []byte("700")
				}
			} else if value.TokenType == css.StringToken && (prop == css.Font || prop == css.Font_Family) {
				parse.ToLower(value.Data)
				s := value.Data[1 : len(value.Data)-1]
				unquote := true
				for _, split := range bytes.Split(s, spaceBytes) {
					val := css.ToHash(split)
					// if len is zero, it contains two consecutive spaces
					if val == css.Inherit || val == css.Serif || val == css.Sans_Serif || val == css.Monospace || val == css.Fantasy || val == css.Cursive || val == css.Initial || val == css.Default ||
						len(split) == 0 || !css.IsIdent(split) {
						unquote = false
						break
					}
				}
				if unquote {
					values[i].Data = s
				}
			}
		} else if prop == css.Outline || prop == css.Background || prop == css.Border || prop == css.Border_Bottom || prop == css.Border_Left || prop == css.Border_Right || prop == css.Border_Top {
			if css.ToHash(values[i].Data) == css.None {
				values[i].TokenType = css.NumberToken
				values[i].Data = zeroBytes
			}
		}
	}

	important := false
	if len(values) > 2 && values[len(values)-2].Data[0] == '!' && bytes.Equal(values[len(values)-1].Data, []byte("important")) {
		values = values[:len(values)-2]
		important = true
	}

	if len(values) == 1 {
		if bytes.Equal(property, msfilterBytes) {
			alpha := []byte("progid:DXImageTransform.Microsoft.Alpha(Opacity=")
			if values[0].TokenType == css.StringToken && bytes.HasPrefix(values[0].Data[1:len(values[0].Data)-1], alpha) {
				values[0].Data = append(append([]byte{values[0].Data[0]}, []byte("alpha(opacity=")...), values[0].Data[1+len(alpha):]...)
			}
		}
	} else {
		if prop == css.Margin || prop == css.Padding || prop == css.Border_Width {
			if values[0].TokenType == css.NumberToken && (len(values)+1)%2 == 0 {
				valid := true
				for i := 1; i < len(values); i += 2 {
					if values[i].TokenType != css.WhitespaceToken || values[i+1].TokenType != css.NumberToken && values[i+1].TokenType != css.DimensionToken && values[i+1].TokenType != css.PercentageToken {
						valid = false
						break
					}
				}
				if valid {
					n := (len(values) + 1) / 2
					if n == 2 {
						if bytes.Equal(values[0].Data, values[2].Data) {
							values = values[:1]
						}
					} else if n == 3 {
						if bytes.Equal(values[0].Data, values[2].Data) && bytes.Equal(values[0].Data, values[4].Data) {
							values = values[:1]
						} else if bytes.Equal(values[0].Data, values[4].Data) {
							values = values[:3]
						}
					} else if n == 4 {
						if bytes.Equal(values[0].Data, values[2].Data) && bytes.Equal(values[0].Data, values[4].Data) && bytes.Equal(values[0].Data, values[6].Data) {
							values = values[:1]
						} else if bytes.Equal(values[0].Data, values[4].Data) && bytes.Equal(values[2].Data, values[6].Data) {
							values = values[:3]
						} else if bytes.Equal(values[2].Data, values[6].Data) {
							values = values[:5]
						}
					}
				}
			}
		} else if prop == css.Filter && len(values) == 11 {
			if bytes.Equal(values[0].Data, []byte("progid")) &&
				values[1].TokenType == css.ColonToken &&
				bytes.Equal(values[2].Data, []byte("DXImageTransform")) &&
				values[3].Data[0] == '.' &&
				bytes.Equal(values[4].Data, []byte("Microsoft")) &&
				values[5].Data[0] == '.' &&
				bytes.Equal(values[6].Data, []byte("Alpha(")) &&
				bytes.Equal(parse.ToLower(values[7].Data), []byte("opacity")) &&
				values[8].Data[0] == '=' &&
				values[10].Data[0] == ')' {
				values = values[6:]
				values[0].Data = []byte("alpha(")
			}
		}
	}

	for i := 0; i < len(values); i++ {
		if values[i].TokenType == css.FunctionToken {
			n, err := c.minifyFunction(values[i:])
			if err != nil {
				return err
			}
			i += n - 1
		} else if _, err := c.w.Write(values[i].Data); err != nil {
			return err
		}
	}
	if important {
		if _, err := c.w.Write([]byte("!important")); err != nil {
			return err
		}
	}
	return nil
}

func (c *cssMinifier) minifyFunction(values []css.Token) (int, error) {
	n := 1
	simple := true
	for i, value := range values[1:] {
		if value.TokenType == css.RightParenthesisToken {
			n++
			break
		}
		if i%2 == 0 && (value.TokenType != css.NumberToken && value.TokenType != css.PercentageToken) || (i%2 == 1 && value.TokenType != css.CommaToken) {
			simple = false
		}
		n++
	}
	values = values[:n]
	if simple && (n-1)%2 == 0 {
		fun := css.ToHash(values[0].Data[:len(values[0].Data)-1])
		nArgs := (n - 1) / 2
		if fun == css.Rgba && nArgs == 4 {
			d, _ := strconv.ParseFloat(string(values[7].Data), 32)
			if math.Abs(d-1.0) < epsilon {
				values[0].Data = []byte("rgb")
				values = values[:len(values)-2]
				fun = css.Rgb
				nArgs = 3
			}
		}
		if fun == css.Rgb && nArgs == 3 {
			var err error
			rgb := [3]byte{}
			for j := 0; j < 3; j++ {
				val := values[j*2+1]
				if val.TokenType == css.NumberToken {
					var d int64
					d, err = strconv.ParseInt(string(val.Data), 10, 32)
					if d < 0 {
						d = 0
					} else if d > 255 {
						d = 255
					}
					rgb[j] = byte(d)
				} else if val.TokenType == css.PercentageToken {
					var d float64
					d, err = strconv.ParseFloat(string(val.Data[:len(val.Data)-1]), 32)
					if d < 0.0 {
						d = 0.0
					} else if d > 100.0 {
						d = 100.0
					}
					rgb[j] = byte((d / 100.0 * 255.0) + 0.5)
				}
			}
			if err == nil {
				val := make([]byte, 7)
				val[0] = '#'
				hex.Encode(val[1:], rgb[:])
				parse.ToLower(val)
				if s, ok := shortenColorHex[string(val)]; ok {
					if _, err := c.w.Write(s); err != nil {
						return 0, err
					}
				} else {
					if len(val) == 7 && val[1] == val[2] && val[3] == val[4] && val[5] == val[6] {
						val[2] = val[3]
						val[3] = val[5]
						val = val[:4]
					}
					if _, err := c.w.Write(val); err != nil {
						return 0, err
					}
				}
				return n, nil
			}
		}
	}
	for _, value := range values {
		if _, err := c.w.Write(value.Data); err != nil {
			return 0, err
		}
	}
	return n, nil
}

func (c *cssMinifier) shortenToken(tt css.TokenType, data []byte) (css.TokenType, []byte) {
	if tt == css.NumberToken || tt == css.DimensionToken || tt == css.PercentageToken {
		if len(data) > 0 && data[0] == '+' {
			data = data[1:]
		}
		num, dim := css.SplitNumberToken(data)
		f, err := strconv.ParseFloat(string(num), 64)
		if err != nil {
			return tt, data
		}
		if math.Abs(f) < epsilon {
			data = zeroBytes
		} else if len(num) > 0 {
			if num[0] == '-' {
				n := 1
				for n < len(num) && num[n] == '0' {
					n++
				}
				num = num[n-1:]
				num[0] = '-'
			} else {
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
			if len(dim) > 1 { // only percentage is length 1
				parse.ToLower(dim)
			}
			data = append(num, dim...)
		}
	} else if tt == css.IdentToken {
		parse.ToLower(data)
		if hash, ok := shortenColorName[css.ToHash(data)]; ok {
			tt = css.HashToken
			data = hash
		}
	} else if tt == css.HashToken {
		parse.ToLower(data)
		if ident, ok := shortenColorHex[string(data)]; ok {
			tt = css.IdentToken
			data = ident
		} else if len(data) == 7 && data[1] == data[2] && data[3] == data[4] && data[5] == data[6] {
			tt = css.HashToken
			data[2] = data[3]
			data[3] = data[5]
			data = data[:4]
		}
	} else if tt == css.StringToken {
		// remove any \\\r\n \\\r \\\n
		for i := 1; i < len(data)-2; i++ {
			if data[i] == '\\' && (data[i+1] == '\n' || data[i+1] == '\r') {
				// encountered first replacee, now start to move bytes to the front
				j := i + 2
				if data[i+1] == '\r' && len(data) > i+2 && data[i+2] == '\n' {
					j++
				}
				for ; j < len(data); j++ {
					if data[j] == '\\' && len(data) > j+1 && (data[j+1] == '\n' || data[j+1] == '\r') {
						if data[j+1] == '\r' && len(data) > j+2 && data[j+2] == '\n' {
							j++
						}
						j++
					} else {
						data[i] = data[j]
						i++
					}
				}
				data = data[:i]
				break
			}
		}
	} else if tt == css.URLToken {
		parse.ToLower(data[:3])
		if mediatype, originalData, ok := css.SplitDataURI(data); ok {
			if bytes.HasPrefix(mediatype, []byte("text/css")) {
				data, _ = minify.Bytes(c.m, string(mediatype)+";inline=1", originalData)
			} else {
				data, _ = minify.Bytes(c.m, string(mediatype), originalData)
			}
			base64Len := len(";base64") + base64.StdEncoding.EncodedLen(len(data))
			asciiLen := len(data)
			for _, c := range data {
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
				base64.StdEncoding.Encode(encoded, data)
				data = encoded
				mediatype = append(mediatype, []byte(";base64")...)
			} else {
				data = []byte(url.QueryEscape(string(data)))
				data = bytes.Replace(data, []byte("\""), []byte("\\\""), -1)
			}
			if len(mediatype) >= len("text/plain") && bytes.HasPrefix(mediatype, []byte("text/plain")) {
				mediatype = mediatype[len("text/plain"):]
			}
			data = append(append(append(append([]byte("url(\"data:"), mediatype...), ','), data...), []byte("\")")...)
		}
		s := data[4 : len(data)-1]
		if len(s) > 2 && (s[0] == '"' || s[0] == '\'') && css.IsUrlUnquoted([]byte(s[1:len(s)-1])) {
			data = append(append([]byte("url("), s[1:len(s)-1]...), ')')
		}
	}
	return tt, data
}
