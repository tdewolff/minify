package css // import "github.com/tdewolff/minify/css"

/*
Uses http://www.w3.org/TR/2010/PR-css3-color-20101028/ for colors
*/

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"math"
	"net/url"
	"strconv"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/parse/css"
)

var epsilon = 0.00001

var shortenColorHex = map[string][]byte{
	"#000080": []byte("navy"),
	"#008000": []byte("green"),
	"#008080": []byte("teal"),
	"#4b0082": []byte("indigo"),
	"#800000": []byte("maroon"),
	"#800080": []byte("purple"),
	"#808000": []byte("olive"),
	"#808080": []byte("gray"),
	"#a0522d": []byte("sienna"),
	"#a52a2a": []byte("brown"),
	"#c0c0c0": []byte("silver"),
	"#cd853f": []byte("peru"),
	"#d2b48c": []byte("tan"),
	"#da70d6": []byte("orchid"),
	"#dda0dd": []byte("plum"),
	"#ee82ee": []byte("violet"),
	"#f0e68c": []byte("khaki"),
	"#f0ffff": []byte("azure"),
	"#f5deb3": []byte("wheat"),
	"#f5f5dc": []byte("beige"),
	"#fa8072": []byte("salmon"),
	"#faf0e6": []byte("linen"),
	"#ff6347": []byte("tomato"),
	"#ff7f50": []byte("coral"),
	"#ffa500": []byte("orange"),
	"#ffc0cb": []byte("pink"),
	"#ffd700": []byte("gold"),
	"#ffe4c4": []byte("bisque"),
	"#fffafa": []byte("snow"),
	"#fffff0": []byte("ivory"),
	"#ff0000": []byte("red"),
	"#f00":    []byte("red"),
}

var shortenColorName = map[css.Hash][]byte{
	css.Black:                []byte("#000"),
	css.Darkblue:             []byte("#00008b"),
	css.Mediumblue:           []byte("#0000cd"),
	css.Darkgreen:            []byte("#006400"),
	css.Darkcyan:             []byte("#008b8b"),
	css.Deepskyblue:          []byte("#00bfff"),
	css.Darkturquoise:        []byte("#00ced1"),
	css.Mediumspringgreen:    []byte("#00fa9a"),
	css.Springgreen:          []byte("#00ff7f"),
	css.Midnightblue:         []byte("#191970"),
	css.Dodgerblue:           []byte("#1e90ff"),
	css.Lightseagreen:        []byte("#20b2aa"),
	css.Forestgreen:          []byte("#228b22"),
	css.Seagreen:             []byte("#2e8b57"),
	css.Darkslategray:        []byte("#2f4f4f"),
	css.Limegreen:            []byte("#32cd32"),
	css.Mediumseagreen:       []byte("#3cb371"),
	css.Turquoise:            []byte("#40e0d0"),
	css.Royalblue:            []byte("#4169e1"),
	css.Steelblue:            []byte("#4682b4"),
	css.Darkslateblue:        []byte("#483d8b"),
	css.Mediumturquoise:      []byte("#48d1cc"),
	css.Darkolivegreen:       []byte("#556b2f"),
	css.Cadetblue:            []byte("#5f9ea0"),
	css.Cornflowerblue:       []byte("#6495ed"),
	css.Mediumaquamarine:     []byte("#66cdaa"),
	css.Slateblue:            []byte("#6a5acd"),
	css.Olivedrab:            []byte("#6b8e23"),
	css.Slategray:            []byte("#708090"),
	css.Lightslateblue:       []byte("#789"),
	css.Mediumslateblue:      []byte("#7b68ee"),
	css.Lawngreen:            []byte("#7cfc00"),
	css.Chartreuse:           []byte("#7fff00"),
	css.Aquamarine:           []byte("#7fffd4"),
	css.Lightskyblue:         []byte("#87cefa"),
	css.Blueviolet:           []byte("#8a2be2"),
	css.Darkmagenta:          []byte("#8b008b"),
	css.Saddlebrown:          []byte("#8b4513"),
	css.Darkseagreen:         []byte("#8fbc8f"),
	css.Lightgreen:           []byte("#90ee90"),
	css.Mediumpurple:         []byte("#9370db"),
	css.Darkviolet:           []byte("#9400d3"),
	css.Palegreen:            []byte("#98fb98"),
	css.Darkorchid:           []byte("#9932cc"),
	css.Yellowgreen:          []byte("#9acd32"),
	css.Darkgray:             []byte("#a9a9a9"),
	css.Lightblue:            []byte("#add8e6"),
	css.Greenyellow:          []byte("#adff2f"),
	css.Paleturquoise:        []byte("#afeeee"),
	css.Lightsteelblue:       []byte("#b0c4de"),
	css.Powderblue:           []byte("#b0e0e6"),
	css.Firebrick:            []byte("#b22222"),
	css.Darkgoldenrod:        []byte("#b8860b"),
	css.Mediumorchid:         []byte("#ba55d3"),
	css.Rosybrown:            []byte("#bc8f8f"),
	css.Darkkhaki:            []byte("#bdb76b"),
	css.Mediumvioletred:      []byte("#c71585"),
	css.Indianred:            []byte("#cd5c5c"),
	css.Chocolate:            []byte("#d2691e"),
	css.Lightgray:            []byte("#d3d3d3"),
	css.Goldenrod:            []byte("#daa520"),
	css.Palevioletred:        []byte("#db7093"),
	css.Gainsboro:            []byte("#dcdcdc"),
	css.Burlywood:            []byte("#deb887"),
	css.Lightcyan:            []byte("#e0ffff"),
	css.Lavender:             []byte("#e6e6fa"),
	css.Darksalmon:           []byte("#e9967a"),
	css.Palegoldenrod:        []byte("#eee8aa"),
	css.Lightcoral:           []byte("#f08080"),
	css.Aliceblue:            []byte("#f0f8ff"),
	css.Honeydew:             []byte("#f0fff0"),
	css.Sandybrown:           []byte("#f4a460"),
	css.Whitesmoke:           []byte("#f5f5f5"),
	css.Mintcream:            []byte("#f5fffa"),
	css.Ghostwhite:           []byte("#f8f8ff"),
	css.Antiquewhite:         []byte("#faebd7"),
	css.Lightgoldenrodyellow: []byte("#fafad2"),
	css.Fuchsia:              []byte("#f0f"),
	css.Magenta:              []byte("#f0f"),
	css.Deeppink:             []byte("#ff1493"),
	css.Orangered:            []byte("#ff4500"),
	css.Darkorange:           []byte("#ff8c00"),
	css.Lightsalmon:          []byte("#ffa07a"),
	css.Lightpink:            []byte("#ffb6c1"),
	css.Peachpuff:            []byte("#ffdab9"),
	css.Navajowhite:          []byte("#ffdead"),
	css.Moccasin:             []byte("#ffe4b5"),
	css.Mistyrose:            []byte("#ffe4e1"),
	css.Blanchedalmond:       []byte("#ffebcd"),
	css.Papayawhip:           []byte("#ffefd5"),
	css.Lavenderblush:        []byte("#fff0f5"),
	css.Seashell:             []byte("#fff5ee"),
	css.Cornsilk:             []byte("#fff8dc"),
	css.Lemonchiffon:         []byte("#fffacd"),
	css.Floralwhite:          []byte("#fffaf0"),
	css.Yellow:               []byte("#ff0"),
	css.Lightyellow:          []byte("#ffffe0"),
	css.White:                []byte("#fff"),
}

////////////////////////////////////////////////////////////////

type cssMinifier struct {
	m minify.Minifier
	w io.Writer
	p *css.Parser

	semicolonQueued bool
}

// Minify minifies CSS files, it reads from r and writes to w.
func Minify(m minify.Minifier, w io.Writer, r io.Reader) error {
	c := &cssMinifier{
		m,
		w,
		css.NewParser(r),
		false,
	}
	var err error
	for {
		gt, n := c.p.Next()
		if gt == css.ErrorGrammar {
			err = c.p.Err()
			break
		} else if err = c.minifyRecursively(gt, n); err != nil {
			break
		}
	}
	if err != io.EOF {
		return err
	}
	return nil
}

func (c *cssMinifier) minifyRecursively(rootGt css.GrammarType, n css.Node) error {
	if rootGt != css.ErrorGrammar && rootGt != css.TokenGrammar && c.semicolonQueued { // it is only TokenGrammar for CDO and CDC
		if err := c.write([]byte(";")); err != nil {
			return err
		}
		c.semicolonQueued = false
	}

	if rootGt == css.AtRuleGrammar {
		atRule := n.(*css.AtRuleNode)
		if err := c.write(atRule.At.Data); err != nil {
			return err
		}
		if err := c.minifyAtRuleNodes(atRule.Nodes); err != nil {
			return err
		}
		hasRules := false
		for {
			gt, m := c.p.Next()
			if gt == css.ErrorGrammar {
				return c.p.Err()
			} else if gt == css.EndAtRuleGrammar {
				break
			} else if !hasRules {
				if err := c.write([]byte("{")); err != nil {
					return err
				}
				hasRules = true
			}
			if err := c.minifyRecursively(gt, m); err != nil {
				return err
			}
		}
		if hasRules {
			if err := c.write([]byte("}")); err != nil {
				return err
			}
			c.semicolonQueued = false
		} else {
			c.semicolonQueued = true
		}
	} else if rootGt == css.RulesetGrammar {
		ruleset := n.(*css.RulesetNode)
		hasRules := false
		for {
			gt, m := c.p.Next()
			if gt == css.ErrorGrammar {
				return c.p.Err()
			} else if gt == css.EndRulesetGrammar {
				break
			} else if !hasRules {
				if err := c.minifySelectors(ruleset.Selectors); err != nil {
					return err
				}
				if err := c.write([]byte("{")); err != nil {
					return err
				}
				hasRules = true
			}
			if err := c.minifyRecursively(gt, m); err != nil {
				return err
			}
		}
		if hasRules {
			if err := c.write([]byte("}")); err != nil {
				return err
			}
			c.semicolonQueued = false
		}
	} else if rootGt == css.DeclarationGrammar {
		if err := c.minifyDeclaration(n.(*css.DeclarationNode)); err != nil {
			return err
		}
	} else if rootGt == css.TokenGrammar {
		if err := c.write(n.(*css.TokenNode).Data); err != nil {
			return err
		}
	}
	return nil
}

func (c *cssMinifier) minifyAtRuleNodes(atRuleNodes []css.Node) error {
	for i, atRuleNode := range atRuleNodes {
		if i != 0 {
			var t *css.TokenNode
			if k, ok := atRuleNodes[i-1].(*css.TokenNode); ok && len(k.Data) == 1 {
				t = k
			} else if k, ok := atRuleNodes[i].(*css.TokenNode); ok && len(k.Data) == 1 {
				t = k
			}
			if t == nil || t.Data[0] != ',' {
				if err := c.write([]byte(" ")); err != nil {
					return err
				}
			}
		} else {
			if err := c.write([]byte(" ")); err != nil {
				return err
			}
		}
		if err := atRuleNode.Serialize(c.w); err != nil {
			return err
		}
	}
	return nil
}

func (c *cssMinifier) minifySelectors(selectors []*css.SelectorNode) error {
	for i, selector := range selectors {
		if i != 0 {
			if err := c.write([]byte(",")); err != nil {
				return err
			}
		}
		inAttr := false
		isClass := false
		for _, elem := range selector.Elems {
			if !inAttr && elem.TokenType == css.LeftBracketToken {
				inAttr = true
			} else if inAttr && elem.TokenType == css.RightBracketToken {
				inAttr = false
			} else if inAttr && elem.TokenType == css.StringToken {
				s := elem.Data[1 : len(elem.Data)-1]
				if css.IsIdent([]byte(s)) {
					if err := c.write(s); err != nil {
						return err
					}
					continue
				}
			} else if !inAttr && elem.TokenType == css.DelimToken && elem.Data[0] == '.' {
				isClass = true
			} else if !inAttr && elem.TokenType == css.IdentToken {
				if !isClass {
					elem.Data = bytes.ToLower(elem.Data)
				}
				isClass = false
			}
			if err := c.write(elem.Data); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *cssMinifier) minifyDeclaration(decl *css.DeclarationNode) error {
	if err := c.write(decl.Prop.Data, []byte(":")); err != nil {
		return err
	}

	// shorten values
	progid := false
	for i, n := range decl.Vals {
		switch m := n.(type) {
		case *css.TokenNode:
			if !progid {
				if i == 0 && css.ToHash(m.Data) == css.Progid {
					progid = true
					continue
				}
				decl.Vals[i] = c.shortenToken(m)
			}
		case *css.FunctionNode:
			if !progid {
				m.Func.Data = bytes.ToLower(m.Func.Data)
			}
			decl.Vals[i] = c.shortenFunction(m)
		}
	}

	prop := css.ToHash(decl.Prop.Data)
	if prop == css.Margin || prop == css.Padding {
		tokens := make([]*css.TokenNode, 0, 4)
		for _, n := range decl.Vals {
			if m, ok := n.(*css.TokenNode); ok {
				tokens = append(tokens, m)
			} else {
				tokens = []*css.TokenNode{}
				break
			}
		}
		if len(tokens) == 2 {
			if bytes.Equal(tokens[0].Data, tokens[1].Data) {
				decl.Vals = []css.Node{decl.Vals[0]}
			}
		} else if len(tokens) == 3 {
			if bytes.Equal(tokens[0].Data, tokens[1].Data) && bytes.Equal(tokens[0].Data, tokens[2].Data) {
				decl.Vals = []css.Node{decl.Vals[0]}
			} else if bytes.Equal(tokens[0].Data, tokens[2].Data) {
				decl.Vals = []css.Node{decl.Vals[0], decl.Vals[1]}
			}
		} else if len(tokens) == 4 {
			if bytes.Equal(tokens[0].Data, tokens[1].Data) && bytes.Equal(tokens[0].Data, tokens[2].Data) && bytes.Equal(tokens[0].Data, tokens[3].Data) {
				decl.Vals = []css.Node{decl.Vals[0]}
			} else if bytes.Equal(tokens[0].Data, tokens[2].Data) && bytes.Equal(tokens[1].Data, tokens[3].Data) {
				decl.Vals = []css.Node{decl.Vals[0], decl.Vals[1]}
			} else if bytes.Equal(tokens[1].Data, tokens[3].Data) {
				decl.Vals = []css.Node{decl.Vals[0], decl.Vals[1], decl.Vals[2]}
			}
		}
	} else if prop == css.Font || prop == css.Font_Family || prop == css.Font_Weight {
		for i, n := range decl.Vals {
			if m, ok := n.(*css.TokenNode); ok {
				if m.TokenType == css.IdentToken && (prop == css.Font || prop == css.Font_Weight) {
					val := css.ToHash(m.Data)
					if val == css.Normal && prop == css.Font_Weight {
						// normal could also be specified for font-variant, not just font-weight
						decl.Vals[i] = css.NewToken(css.NumberToken, []byte("400"))
					} else if val == css.Bold {
						decl.Vals[i] = css.NewToken(css.NumberToken, []byte("700"))
					}
				} else if m.TokenType == css.StringToken && (prop == css.Font || prop == css.Font_Family) {
					m.Data = bytes.ToLower(m.Data)
					s := m.Data[1 : len(m.Data)-1]
					unquote := true
					for _, split := range bytes.Split(s, []byte(" ")) {
						val := css.ToHash(split)
						// if len is zero, it contains two consecutive spaces
						if val == css.Inherit || val == css.Serif || val == css.Sans_Serif || val == css.Monospace || val == css.Fantasy || val == css.Cursive || val == css.Initial || val == css.Default ||
							len(split) == 0 || !css.IsIdent(split) {
							unquote = false
							break
						}
					}
					if unquote {
						m.Data = s
					}
				}
			}
		}
	} else if (prop == css.Outline || prop == css.Background || prop == css.Border || prop == css.Border_Bottom || prop == css.Border_Left || prop == css.Border_Right || prop == css.Border_Top) && len(decl.Vals) == 1 {
		if n, ok := decl.Vals[0].(*css.TokenNode); ok && css.ToHash(n.Data) == css.None {
			decl.Vals[0] = css.NewToken(css.NumberToken, []byte("0"))
		}
	} else if prop == css.Filter && len(decl.Vals) == 7 {
		if n, ok := decl.Vals[6].(*css.FunctionNode); ok && bytes.Equal(n.Func.Data, []byte("Alpha")) {
			tokens := []byte{}
			for _, val := range decl.Vals[:len(decl.Vals)-1] {
				if m, ok := val.(*css.TokenNode); ok {
					tokens = append(tokens, m.Data...)
				} else {
					tokens = []byte{}
					break
				}
			}
			f := decl.Vals[6].(*css.FunctionNode)
			if bytes.Equal(tokens, []byte("progid:DXImageTransform.Microsoft.")) && len(f.Args) == 1 && len(f.Args[0].Vals) == 3 {
				if opacity, ok := f.Args[0].Vals[0].(*css.TokenNode); ok {
					opacity.Data = bytes.ToLower(opacity.Data)
					if is, ok := f.Args[0].Vals[1].(*css.TokenNode); ok && is.Data[0] == '=' && bytes.Equal(opacity.Data, []byte("opacity")) {
						newF := css.NewFunction(css.NewToken(css.FunctionToken, []byte("alpha(")))
						newF.Args = f.Args
						decl.Vals = []css.Node{newF}
					}
				}
			}
		}
	} else if len(decl.Vals) == 1 && bytes.Equal(decl.Prop.Data, []byte("-ms-filter")) {
		if n, ok := decl.Vals[0].(*css.TokenNode); ok {
			alpha := []byte("progid:DXImageTransform.Microsoft.Alpha(Opacity=")
			if n.TokenType == css.StringToken && bytes.HasPrefix(n.Data[1:len(n.Data)-1], alpha) {
				n.Data = append(append([]byte{n.Data[0]}, []byte("alpha(opacity=")...), n.Data[1+len(alpha):]...)
			}
		}
	}

	for i, m := range decl.Vals {
		if i != 0 {
			var t *css.TokenNode
			if k, ok := decl.Vals[i-1].(*css.TokenNode); ok && len(k.Data) == 1 {
				t = k
			} else if k, ok := decl.Vals[i].(*css.TokenNode); ok && len(k.Data) == 1 {
				t = k
			}
			if t == nil || (t.Data[0] != ',' && t.Data[0] != '/' && t.Data[0] != ':' && t.Data[0] != '.') {
				if err := c.write([]byte(" ")); err != nil {
					return err
				}
			}
		}
		if err := m.Serialize(c.w); err != nil {
			return err
		}
	}
	if decl.Important {
		if err := c.write([]byte("!important")); err != nil {
			return err
		}
	}
	c.semicolonQueued = true
	return nil
}

func (c *cssMinifier) shortenFunction(f *css.FunctionNode) css.Node {
	simpleFunction := true
	for j, arg := range f.Args {
		for k, val := range arg.Vals {
			if tVal, ok := val.(*css.TokenNode); ok {
				f.Args[j].Vals[k] = c.shortenToken(tVal)
				if k > 1 {
					simpleFunction = false
				}
			} else {
				simpleFunction = false
			}
		}
	}

	var n css.Node = f
	if simpleFunction {
		fun := css.ToHash(f.Func.Data)
		if fun == css.Rgba && len(f.Args) == 4 {
			d, _ := strconv.ParseFloat(string(f.Args[3].Vals[0].(*css.TokenNode).Data), 32)
			if math.Abs(d-1.0) < epsilon {
				f.Func = css.NewToken(css.FunctionToken, []byte("rgb"))
				f.Args = f.Args[:len(f.Args)-1]
				fun = css.Rgb
			}
		}
		if fun == css.Rgb && len(f.Args) == 3 {
			var err error
			rgb := make([]byte, 3)
			for j := 0; j < 3; j++ {
				v := f.Args[j].Vals[0].(*css.TokenNode)
				if v.TokenType == css.NumberToken {
					var d int64
					d, err = strconv.ParseInt(string(v.Data), 10, 32)
					if d < 0 {
						d = 0
					} else if d > 255 {
						d = 255
					}
					rgb[j] = byte(d)
				} else if v.TokenType == css.PercentageToken {
					var d float64
					d, err = strconv.ParseFloat(string(v.Data[:len(v.Data)-1]), 32)
					if d < 0.0 {
						d = 0.0
					} else if d > 100.0 {
						d = 100.0
					}
					rgb[j] = byte((d / 100.0 * 255.0) + 0.5)
				} else {
					err = errors.New("")
					break
				}
			}
			if err == nil {
				valHex := make([]byte, 6)
				hex.Encode(valHex, rgb)
				val := append([]byte("#"), bytes.ToLower(valHex)...)
				if s, ok := shortenColorHex[string(val)]; ok {
					n = css.NewToken(css.IdentToken, s)
				} else if len(val) == 7 && val[1] == val[2] && val[3] == val[4] && val[5] == val[6] {
					n = css.NewToken(css.HashToken, append([]byte("#"), val[1], val[3], val[5]))
				} else {
					n = css.NewToken(css.HashToken, val)
				}
			}
		}
	}
	return n
}

func (c *cssMinifier) shortenToken(t *css.TokenNode) *css.TokenNode {
	if t.TokenType == css.NumberToken || t.TokenType == css.DimensionToken || t.TokenType == css.PercentageToken {
		if len(t.Data) > 0 && t.Data[0] == '+' {
			t.Data = t.Data[1:]
		}
		num, dim := css.SplitNumberToken(t.Data)
		f, err := strconv.ParseFloat(string(num), 64)
		if err != nil {
			return t
		}
		if math.Abs(f) < epsilon {
			t.Data = []byte("0")
		} else if len(num) > 0 {
			if num[0] == '-' {
				num = num[1:]
				// trim 0 left
				for len(num) > 0 && num[0] == '0' {
					num = num[1:]
				}
				num = append([]byte{'-'}, num...)
			} else {
				// trim 0 left
				for len(num) > 0 && num[0] == '0' {
					num = num[1:]
				}
			}
			// trim 0 right
			for i, digit := range num {
				if digit == '.' {
					j := len(num)-1
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
				dim = bytes.ToLower(dim)
			}
			t.Data = append(num, dim...)
		}
	} else if t.TokenType == css.IdentToken {
		t.Data = bytes.ToLower(t.Data)
		if h, ok := shortenColorName[css.ToHash(t.Data)]; ok {
			t = css.NewToken(css.HashToken, h)
		}
	} else if t.TokenType == css.HashToken {
		val := bytes.ToLower(t.Data)
		if i, ok := shortenColorHex[string(val)]; ok {
			t = css.NewToken(css.IdentToken, i)
		} else if len(val) == 7 && val[1] == val[2] && val[3] == val[4] && val[5] == val[6] {
			t = css.NewToken(css.HashToken, append([]byte("#"), bytes.ToLower(append([]byte{val[1]}, val[3], val[5]))...))
		} else {
			t.Data = bytes.ToLower(t.Data)
		}
	} else if t.TokenType == css.StringToken {
		// remove any \\\r\n \\\r \\\n
		for i := 1; i < len(t.Data)-2; i++ {
			if t.Data[i] == '\\' && (t.Data[i+1] == '\n' || t.Data[i+1] == '\r') {
				// encountered first replacee, now start to move bytes to the front
				j := i + 2
				if t.Data[i+1] == '\r' && len(t.Data) > i+2 && t.Data[i+2] == '\n' {
					j++
				}
				for ; j < len(t.Data); j++ {
					if t.Data[j] == '\\' && len(t.Data) > j+1 && (t.Data[j+1] == '\n' || t.Data[j+1] == '\r') {
						if t.Data[j+1] == '\r' && len(t.Data) > j+2 && t.Data[j+2] == '\n' {
							j++
						}
						j++
					} else {
						t.Data[i] = t.Data[j]
						i++
					}
				}
				t.Data = t.Data[:i]
				break
			}
		}
	} else if t.TokenType == css.URLToken {
		t.Data = append([]byte("url"), t.Data[3:]...)
		if mediatype, originalData, ok := css.SplitDataURI(t.Data); ok {
			data := originalData
			minifiedBuffer := &bytes.Buffer{}
			if err := c.m.Minify(string(mediatype), minifiedBuffer, bytes.NewBuffer(data)); err == nil {
				data = minifiedBuffer.Bytes()
			}
			base64Len := base64.StdEncoding.EncodedLen(len(data))
			asciiLen := 7 + minifiedBuffer.Len()
			for _, c := range data {
				if 'A' <= c && c <= 'Z' || 'a' <= c && c <= 'z' || '0' <= c && c <= '9' || c == '-' || c == '_' || c == '.' || c == '~' {
					asciiLen += 2
				} else if c == '"' {
					asciiLen++
				}
				if asciiLen > base64Len {
					break
				}
			}
			if asciiLen > base64Len {
				encoded := make([]byte, base64Len)
				base64.StdEncoding.Encode(encoded, data)
				data = encoded
				mediatype = append(mediatype, []byte(";base64")...)
			} else {
				data = []byte(url.QueryEscape(string(data)))
				data = bytes.Replace(data, []byte("\""), []byte("\\\""), -1)
			}
			if len(data) < len(originalData) {
				t.Data = append(append(append(append([]byte("url(\"data:"), mediatype...), ','), data...), []byte("\")")...)
			}
		}
		s := t.Data[4 : len(t.Data)-1]
		if len(s) > 2 && (s[0] == '"' || s[0] == '\'') && css.IsUrlUnquoted([]byte(s[1:len(s)-1])) {
			t.Data = append(append([]byte("url("), s[1:len(s)-1]...), ')')
		}
	}
	return t
}

////////////////////////////////////////////////////////////////

func (c *cssMinifier) write(bs ...[]byte) error {
	for _, b := range bs {
		if _, err := c.w.Write(b); err != nil {
			return err
		}
	}
	return nil
}
