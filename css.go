package minify

/* TODO: (non-exhaustive)
- remove space before !important
- collapse margin/padding/border/background/list/etc. definitions into one
- remove duplicate selector blocks
- remove quotes within url()?
*/

/*
Uses http://www.w3.org/TR/2010/PR-css3-color-20101028/ for colors
*/

import (
	"bytes"
	"encoding/hex"
	"errors"
	"io"
	"math"
	"strconv"

	"github.com/tdewolff/css"
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

var shortenColorName = map[string][]byte{
	"black":                []byte("#000"),
	"darkblue":             []byte("#00008b"),
	"mediumblue":           []byte("#0000cd"),
	"darkgreen":            []byte("#006400"),
	"darkcyan":             []byte("#008b8b"),
	"deepskyblue":          []byte("#00bfff"),
	"darkturquoise":        []byte("#00ced1"),
	"mediumspringgreen":    []byte("#00fa9a"),
	"springgreen":          []byte("#00ff7f"),
	"midnightblue":         []byte("#191970"),
	"dodgerblue":           []byte("#1e90ff"),
	"lightseagreen":        []byte("#20b2aa"),
	"forestgreen":          []byte("#228b22"),
	"seagreen":             []byte("#2e8b57"),
	"darkslategray":        []byte("#2f4f4f"),
	"limegreen":            []byte("#32cd32"),
	"mediumseagreen":       []byte("#3cb371"),
	"turquoise":            []byte("#40e0d0"),
	"royalblue":            []byte("#4169e1"),
	"steelblue":            []byte("#4682b4"),
	"darkslateblue":        []byte("#483d8b"),
	"mediumturquoise":      []byte("#48d1cc"),
	"darkolivegreen":       []byte("#556b2f"),
	"cadetblue":            []byte("#5f9ea0"),
	"cornflowerblue":       []byte("#6495ed"),
	"mediumaquamarine":     []byte("#66cdaa"),
	"slateblue":            []byte("#6a5acd"),
	"olivedrab":            []byte("#6b8e23"),
	"slategray":            []byte("#708090"),
	"lightslateblue":       []byte("#789"),
	"mediumslateblue":      []byte("#7b68ee"),
	"lawngreen":            []byte("#7cfc00"),
	"chartreuse":           []byte("#7fff00"),
	"aquamarine":           []byte("#7fffd4"),
	"lightskyblue":         []byte("#87cefa"),
	"blueviolet":           []byte("#8a2be2"),
	"darkmagenta":          []byte("#8b008b"),
	"saddlebrown":          []byte("#8b4513"),
	"darkseagreen":         []byte("#8fbc8f"),
	"lightgreen":           []byte("#90ee90"),
	"mediumpurple":         []byte("#9370db"),
	"darkviolet":           []byte("#9400d3"),
	"palegreen":            []byte("#98fb98"),
	"darkorchid":           []byte("#9932cc"),
	"yellowgreen":          []byte("#9acd32"),
	"darkgray":             []byte("#a9a9a9"),
	"lightblue":            []byte("#add8e6"),
	"greenyellow":          []byte("#adff2f"),
	"paleturquoise":        []byte("#afeeee"),
	"lightsteelblue":       []byte("#b0c4de"),
	"powderblue":           []byte("#b0e0e6"),
	"firebrick":            []byte("#b22222"),
	"darkgoldenrod":        []byte("#b8860b"),
	"mediumorchid":         []byte("#ba55d3"),
	"rosybrown":            []byte("#bc8f8f"),
	"darkkhaki":            []byte("#bdb76b"),
	"mediumvioletred":      []byte("#c71585"),
	"indianred":            []byte("#cd5c5c"),
	"chocolate":            []byte("#d2691e"),
	"lightgray":            []byte("#d3d3d3"),
	"goldenrod":            []byte("#daa520"),
	"palevioletred":        []byte("#db7093"),
	"gainsboro":            []byte("#dcdcdc"),
	"burlywood":            []byte("#deb887"),
	"lightcyan":            []byte("#e0ffff"),
	"lavender":             []byte("#e6e6fa"),
	"darksalmon":           []byte("#e9967a"),
	"palegoldenrod":        []byte("#eee8aa"),
	"lightcoral":           []byte("#f08080"),
	"aliceblue":            []byte("#f0f8ff"),
	"honeydew":             []byte("#f0fff0"),
	"sandybrown":           []byte("#f4a460"),
	"whitesmoke":           []byte("#f5f5f5"),
	"mintcream":            []byte("#f5fffa"),
	"ghostwhite":           []byte("#f8f8ff"),
	"antiquewhite":         []byte("#faebd7"),
	"lightgoldenrodyellow": []byte("#fafad2"),
	"fuchsia":              []byte("#f0f"),
	"magenta":              []byte("#f0f"),
	"deeppink":             []byte("#ff1493"),
	"orangered":            []byte("#ff4500"),
	"darkorange":           []byte("#ff8c00"),
	"lightsalmon":          []byte("#ffa07a"),
	"lightpink":            []byte("#ffb6c1"),
	"peachpuff":            []byte("#ffdab9"),
	"navajowhite":          []byte("#ffdead"),
	"moccasin":             []byte("#ffe4b5"),
	"mistyrose":            []byte("#ffe4e1"),
	"blanchedalmond":       []byte("#ffebcd"),
	"papayawhip":           []byte("#ffefd5"),
	"lavenderblush":        []byte("#fff0f5"),
	"seashell":             []byte("#fff5ee"),
	"cornsilk":             []byte("#fff8dc"),
	"lemonchiffon":         []byte("#fffacd"),
	"floralwhite":          []byte("#fffaf0"),
	"yellow":               []byte("#ff0"),
	"lightyellow":          []byte("#ffffe0"),
	"white":                []byte("#fff"),
}

// CSS minifies CSS files, it reads from r and writes to w.
func (m Minifier) CSS(w io.Writer, r io.Reader) error {
	stylesheet, err := css.Parse(r)
	if err != nil {
		return err
	}
	shortenNodes(stylesheet.Nodes)
	return writeNodes(w, stylesheet.Nodes)
}

////////////////////////////////////////////////////////////////

func shortenNodes(nodes []css.Node) {
	inHeader := true
	for i, n := range nodes {
		if inHeader && n.Type() != css.AtRuleNode {
			inHeader = false
		}
		switch n.Type() {
		case css.DeclarationNode:
			shortenDecl(n.(*css.NodeDeclaration))
		case css.RulesetNode:
			ruleset := n.(*css.NodeRuleset)
			for _, selGroup := range ruleset.SelGroups {
				for _, sel := range selGroup.Selectors {
					shortenSelector(sel)
				}
			}
			for _, decl := range ruleset.Decls {
				shortenDecl(decl)
			}
		case css.AtRuleNode:
			atRule := n.(*css.NodeAtRule)
			if !inHeader && (bytes.Equal(atRule.At.Data, []byte("@charset")) || bytes.Equal(atRule.At.Data, []byte("@import"))) {
				nodes[i] = css.NewToken(css.DelimToken, []byte(""))
			}
			if n.(*css.NodeAtRule).Block != nil {
				shortenNodes(n.(*css.NodeAtRule).Block.Nodes)
			}
		case css.BlockNode:
			shortenNodes(n.(*css.NodeBlock).Nodes)
		}
	}
}

func shortenSelector(sel *css.NodeSelector) {
	class := false
	for i, n := range sel.Nodes {
		if n.Type() == css.TokenNode {
			token := n.(*css.NodeToken)
			if token.TokenType == css.DelimToken && token.Data[0] == '.' {
				class = true
			} else if token.TokenType == css.IdentToken {
				if !class {
					token.Data = bytes.ToLower(token.Data)
				}
				class = false
			}
		} else if n.Type() == css.AttributeSelectorNode {
			attr := n.(*css.NodeAttributeSelector)
			for j, val := range attr.Vals {
				if val.TokenType == css.StringToken {
					s := val.Data[1 : len(val.Data)-1]
					if css.IsIdent([]byte(s)) {
						attr.Vals[j] = css.NewToken(css.IdentToken, s)
					}
				}
			}
			if (bytes.Equal(attr.Key.Data, []byte("id")) || bytes.Equal(attr.Key.Data, []byte("class"))) && attr.Op != nil && bytes.Equal(attr.Op.Data, []byte("=")) && len(attr.Vals) == 1 && css.IsIdent(attr.Vals[0].Data) {
				if bytes.Equal(attr.Key.Data, []byte("id")) {
					sel.Nodes[i] = css.NewToken(css.HashToken, append([]byte("#"), attr.Vals[0].Data...))
				} else {
					sel.Nodes[i] = css.NewToken(css.DelimToken, []byte("."))
					sel.Nodes = append(append(sel.Nodes[:i+1], css.NewToken(css.IdentToken, attr.Vals[0].Data)), sel.Nodes[i+1:]...)
					class = true
				}
			}
		}
	}
}

func shortenDecl(decl *css.NodeDeclaration) {
	// shorten zeros
	progid := false
	for i, val := range decl.Vals {
		if val.Type() == css.TokenNode && !progid {
			if i == 0 && bytes.Equal(val.(*css.NodeToken).Data, []byte("progid")) {
				progid = true
				continue
			}
			decl.Vals[i] = shortenToken(val.(*css.NodeToken))
		} else if val.Type() == css.FunctionNode {
			f := val.(*css.NodeFunction)
			if !progid {
				f.Func.Data = bytes.ToLower(f.Func.Data)
			}
			for j, arg := range f.Args {
				f.Args[j].Val = shortenToken(arg.Val)
			}
		}
	}

	decl.Prop.Data = bytes.ToLower(decl.Prop.Data)
	prop := decl.Prop.Data
	if bytes.Equal(prop, []byte("margin")) || bytes.Equal(prop, []byte("padding")) {
		if len(decl.Vals) == 2 && decl.Vals[0].Type() == css.TokenNode && decl.Vals[1].Type() == css.TokenNode {
			if bytes.Equal(decl.Vals[0].(*css.NodeToken).Data, decl.Vals[1].(*css.NodeToken).Data) {
				decl.Vals = []css.Node{decl.Vals[0]}
			}
		} else if len(decl.Vals) == 3 && decl.Vals[0].Type() == css.TokenNode && decl.Vals[1].Type() == css.TokenNode && decl.Vals[2].Type() == css.TokenNode {
			if bytes.Equal(decl.Vals[0].(*css.NodeToken).Data, decl.Vals[1].(*css.NodeToken).Data) && bytes.Equal(decl.Vals[0].(*css.NodeToken).Data, decl.Vals[2].(*css.NodeToken).Data) {
				decl.Vals = []css.Node{decl.Vals[0]}
			} else if bytes.Equal(decl.Vals[0].(*css.NodeToken).Data, decl.Vals[2].(*css.NodeToken).Data) {
				decl.Vals = []css.Node{decl.Vals[0], decl.Vals[1]}
			}
		} else if len(decl.Vals) == 4 && decl.Vals[0].Type() == css.TokenNode && decl.Vals[1].Type() == css.TokenNode && decl.Vals[2].Type() == css.TokenNode && decl.Vals[3].Type() == css.TokenNode {
			if bytes.Equal(decl.Vals[0].(*css.NodeToken).Data, decl.Vals[1].(*css.NodeToken).Data) &&
			   bytes.Equal(decl.Vals[0].(*css.NodeToken).Data, decl.Vals[2].(*css.NodeToken).Data) &&
			   bytes.Equal(decl.Vals[0].(*css.NodeToken).Data, decl.Vals[3].(*css.NodeToken).Data) {
				decl.Vals = []css.Node{decl.Vals[0]}
			} else if bytes.Equal(decl.Vals[0].(*css.NodeToken).Data, decl.Vals[2].(*css.NodeToken).Data) &&bytes.Equal(decl.Vals[1].(*css.NodeToken).Data, decl.Vals[3].(*css.NodeToken).Data) {
				decl.Vals = []css.Node{decl.Vals[0], decl.Vals[1]}
			} else if bytes.Equal(decl.Vals[1].(*css.NodeToken).Data, decl.Vals[3].(*css.NodeToken).Data) {
				decl.Vals = []css.Node{decl.Vals[0], decl.Vals[1], decl.Vals[2]}
			}
		}
	} else if bytes.HasPrefix(prop, []byte("font")) {
		for i, val := range decl.Vals {
			if val.Type() == css.TokenNode {
				if val.(*css.NodeToken).TokenType == css.IdentToken {
					if bytes.Equal(val.(*css.NodeToken).Data, []byte("normal")) && bytes.Equal(prop, []byte("font-weight")) { // normal could also be specified for font-variant, not just font-weight
						decl.Vals[i] = css.NewToken(css.NumberToken, []byte("400"))
					} else if bytes.Equal(val.(*css.NodeToken).Data, []byte("bold")) {
						decl.Vals[i] = css.NewToken(css.NumberToken, []byte("700"))
					}
				} else if val.(*css.NodeToken).TokenType == css.StringToken {
					n := val.(*css.NodeToken)
					n.Data = bytes.ToLower(n.Data)
					s := n.Data[1 : len(n.Data)-1]
					unquote := true
					for _, fontName := range bytes.Split(s, []byte(" ")) {
						// if len is zero, it contains two consecutive spaces
						if len(fontName) == 0 || !css.IsIdent(fontName) || bytes.Equal(fontName, []byte("inherit")) || bytes.Equal(fontName, []byte("serif")) || bytes.Equal(fontName, []byte("sans-serif")) || bytes.Equal(fontName, []byte("monospace")) ||
						   bytes.Equal(fontName, []byte("fantasy")) || bytes.Equal(fontName, []byte("cursive")) || bytes.Equal(fontName, []byte("initial")) || bytes.Equal(fontName, []byte("default")) {
							unquote = false
							break
						}
					}
					if unquote {
						n.Data = s
					}
				}
			}
		}
	} else if bytes.Equal(prop, []byte("filter")) {
		if len(decl.Vals) == 7 && decl.Vals[6].Type() == css.FunctionNode && bytes.Equal(decl.Vals[6].(*css.NodeFunction).Func.Data, []byte("Alpha(")) {
			tokens := []byte{}
			for _, val := range decl.Vals[:len(decl.Vals)-1] {
				if val.Type() == css.TokenNode {
					tokens = append(tokens, val.(*css.NodeToken).Data...)
				} else {
					tokens = []byte{}
					break
				}
			}
			f := decl.Vals[6].(*css.NodeFunction)
			if bytes.Equal(tokens, []byte("progid:DXImageTransform.Microsoft.")) && len(f.Args) == 1 && bytes.Equal(f.Args[0].Key.Data, []byte("Opacity")) {
				newF := css.NewFunction(css.NewToken(css.FunctionToken, []byte("alpha(")))
				newF.Args = f.Args
				newF.Args[0].Key.Data = bytes.ToLower(newF.Args[0].Key.Data)
				decl.Vals = []css.Node{newF}
			}
		}
	} else if bytes.Equal(prop, []byte("-ms-filter")) {
		if len(decl.Vals) == 1 && decl.Vals[0].Type() == css.TokenNode {
			n := decl.Vals[0].(*css.NodeToken)
			alpha := []byte("progid:DXImageTransform.Microsoft.Alpha(Opacity=")
			if n.TokenType == css.StringToken && bytes.HasPrefix(n.Data[1:len(n.Data)-1], alpha) {
				n.Data = append(append([]byte{n.Data[0]}, []byte("alpha(opacity=")...), n.Data[1+len(alpha):]...)
			}
		}
	} else {
		if bytes.HasPrefix(prop, []byte("outline")) || bytes.HasPrefix(prop, []byte("background")) || bytes.HasPrefix(prop, []byte("border")) {
			if len(decl.Vals) == 1 && decl.Vals[0].Type() == css.TokenNode && bytes.Equal(bytes.ToLower(decl.Vals[0].(*css.NodeToken).Data), []byte("none")) {
				decl.Vals[0] = css.NewToken(css.NumberToken, []byte("0"))
			}
		}

		for i, val := range decl.Vals {
			if val.Type() == css.FunctionNode {
				f := val.(*css.NodeFunction)
				if bytes.Equal(f.Func.Data, []byte("rgba(")) && len(f.Args) == 4 {
					d, _ := strconv.ParseFloat(string(f.Args[3].Val.Data), 32)
					if math.Abs(d-1.0) < epsilon {
						f.Func = css.NewToken(css.FunctionToken, []byte("rgb("))
						f.Args = f.Args[:len(f.Args)-1]
					}
				}
				if bytes.Equal(f.Func.Data, []byte("rgb(")) && len(f.Args) == 3 {
					var err error
					rgb := make([]byte, 3)
					for j := 0; j < 3; j++ {
						v := f.Args[j].Val
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
							decl.Vals[i] = css.NewToken(css.IdentToken, s)
						} else if len(val) == 7 && val[1] == val[2] && val[3] == val[4] && val[5] == val[6] {
							decl.Vals[i] = css.NewToken(css.HashToken, append([]byte("#"), val[1], val[3], val[5]))
						} else {
							decl.Vals[i] = css.NewToken(css.HashToken, val)
						}
					}
				}
			} else if val.Type() == css.TokenNode {
				n := val.(*css.NodeToken)
				if n.TokenType == css.URLToken {
					s := n.Data[4 : len(n.Data)-1]
					if s[0] == '"' || s[0] == '\'' {
						if css.IsUrlUnquoted([]byte(s[1 : len(s)-1])) {
							s = s[1 : len(s)-1]
						}
					}
					n.Data = append(append([]byte("url("), s...), ')')
				}
			}
		}
	}
}

func shortenToken(token *css.NodeToken) *css.NodeToken {
	if token.TokenType == css.NumberToken || token.TokenType == css.DimensionToken || token.TokenType == css.PercentageToken {
		token.Data = bytes.ToLower(token.Data)
		if len(token.Data) > 0 && token.Data[0] == '+' {
			token.Data = token.Data[1:]
		}

		num, dim := css.SplitNumberToken(token.Data)
		f, err := strconv.ParseFloat(string(num), 64)
		if err != nil {
			return token
		}
		if math.Abs(f) < epsilon {
			token.Data = []byte("0")
		} else {
			if len(num) > 0 && num[0] == '-' {
				num = append([]byte{'-'}, bytes.TrimLeft(num[1:], "0")...)
			} else {
				num = bytes.TrimLeft(num, "0")
			}
			if bytes.Index(num, []byte(".")) != -1 {
				num = bytes.TrimRight(num, "0")
			}
			token.Data = append(num, dim...)
		}
	} else if token.TokenType == css.IdentToken {
		token.Data = bytes.ToLower(token.Data)
		if h, ok := shortenColorName[string(token.Data)]; ok {
			token = css.NewToken(css.HashToken, h)
		}
	} else if token.TokenType == css.HashToken {
		val := bytes.ToLower(token.Data)
		if i, ok := shortenColorHex[string(val)]; ok {
			token = css.NewToken(css.IdentToken, i)
		} else if len(val) == 7 && val[1] == val[2] && val[3] == val[4] && val[5] == val[6] {
			token = css.NewToken(css.HashToken, append([]byte("#"), bytes.ToLower(append([]byte{val[1]}, val[3], val[5]))...))
		} else {
			token.Data = bytes.ToLower(token.Data)
		}
	} else if token.TokenType == css.StringToken {
		token.Data = bytes.Replace(token.Data, []byte("\\\r\n"), []byte(""), -1)
		token.Data = bytes.Replace(token.Data, []byte("\\\r"), []byte(""), -1)
		token.Data = bytes.Replace(token.Data, []byte("\\\n"), []byte(""), -1)
	}
	return token
}

////////////////////////////////////////////////////////////////

func writeNodes(w io.Writer, nodes []css.Node) error {
	semicolonQueued := false
	for _, n := range nodes {
		if semicolonQueued && n.Type() != css.TokenNode { // it is only TokenNode for CDO and CDC (<!-- and --> respectively)
			if _, err := w.Write([]byte(";")); err != nil {
				return ErrWrite
			}
			semicolonQueued = false
		}

		switch n.Type() {
		case css.DeclarationNode:
			if err := writeDecl(w, n.(*css.NodeDeclaration)); err != nil {
				return err
			}
			semicolonQueued = true
		case css.RulesetNode:
			ruleset := n.(*css.NodeRuleset)
			for i, selGroup := range ruleset.SelGroups {
				if i > 0 {
					if _, err := w.Write([]byte(",")); err != nil {
						return ErrWrite
					}
				}
				prevOperator := false
				for j, sel := range selGroup.Selectors {
					if len(sel.Nodes) == 1 && sel.Nodes[0].Type() == css.TokenNode {
						tt := sel.Nodes[0].(*css.NodeToken).TokenType
						op := sel.Nodes[0].(*css.NodeToken).Data
						// TODO: check if clause
						if tt == css.DelimToken && len(op) == 1 && (op[0] == '>' || op[0] == '+' || op[0] == '~') || tt == css.IncludeMatchToken || tt == css.DashMatchToken ||
							tt == css.PrefixMatchToken || tt == css.SuffixMatchToken || tt == css.SubstringMatchToken {
							if _, err := w.Write(op); err != nil {
								return ErrWrite
							}
							prevOperator = true
							continue
						}
					}
					if j > 0 && !prevOperator {
						if _, err := w.Write([]byte(" ")); err != nil {
							return ErrWrite
						}
					}
					for _, node := range sel.Nodes {
						if node.Type() == css.TokenNode {
							if _, err := w.Write(node.(*css.NodeToken).Data); err != nil {
								return ErrWrite
							}
						} else if node.Type() == css.AttributeSelectorNode {
							attr := node.(*css.NodeAttributeSelector)
							if _, err := w.Write(append([]byte("["), attr.Key.Data...)); err != nil {
								return ErrWrite
							}
							if attr.Op != nil {
								if _, err := w.Write(attr.Op.Data); err != nil {
									return ErrWrite
								}
								for _, val := range attr.Vals {
									if _, err := w.Write(val.Data); err != nil {
										return ErrWrite
									}
								}
							}
							if _, err := w.Write([]byte("]")); err != nil {
								return ErrWrite
							}
						}
					}
					prevOperator = false
				}
			}
			if _, err := w.Write([]byte("{")); err != nil {
				return ErrWrite
			}
			for i, decl := range ruleset.Decls {
				if i > 0 {
					if _, err := w.Write([]byte(";")); err != nil {
						return ErrWrite
					}
				}
				if err := writeDecl(w, decl); err != nil {
					return err
				}
			}
			if _, err := w.Write([]byte("}")); err != nil {
				return ErrWrite
			}
		case css.AtRuleNode:
			atRule := n.(*css.NodeAtRule)
			if len(atRule.Nodes) == 0 && atRule.Block == nil {
				break
			}
			if _, err := w.Write(atRule.At.Data); err != nil {
				return ErrWrite
			}
			for _, node := range atRule.Nodes {
				if _, err := w.Write([]byte(" ")); err != nil {
					return ErrWrite
				}
				if _, err := w.Write(node.Data); err != nil {
					return ErrWrite
				}
			}
			if atRule.Block != nil {
				if _, err := w.Write(atRule.Block.Open.Data); err != nil {
					return ErrWrite
				}
				if err := writeNodes(w, atRule.Block.Nodes); err != nil {
					return err
				}
				if _, err := w.Write(atRule.Block.Close.Data); err != nil {
					return ErrWrite
				}
			} else {
				semicolonQueued = true
			}
		case css.BlockNode:
			block := n.(*css.NodeBlock)
			if _, err := w.Write(block.Open.Data); err != nil {
				return ErrWrite
			}
			if err := writeNodes(w, block.Nodes); err != nil {
				return err
			}
			if _, err := w.Write(block.Close.Data); err != nil {
				return ErrWrite
			}
		case css.TokenNode:
			token := n.(*css.NodeToken)
			if _, err := w.Write(token.Data); err != nil {
				return ErrWrite
			}
		}
	}
	return nil
}

func writeDecl(w io.Writer, decl *css.NodeDeclaration) error {
	if _, err := w.Write(decl.Prop.Data); err != nil {
		return ErrWrite
	}
	if _, err := w.Write([]byte(":")); err != nil {
		return ErrWrite
	}
	prevDelim := false
	for j, val := range decl.Vals {
		currDelim := (val.Type() == css.TokenNode && (val.(*css.NodeToken).TokenType == css.DelimToken || val.(*css.NodeToken).TokenType == css.CommaToken || val.(*css.NodeToken).TokenType == css.ColonToken))
		if j > 0 && !currDelim && !prevDelim {
			if _, err := w.Write([]byte(" ")); err != nil {
				return ErrWrite
			}
		}
		if val.Type() == css.TokenNode {
			if _, err := w.Write(val.(*css.NodeToken).Data); err != nil {
				return ErrWrite
			}
		} else if val.Type() == css.FunctionNode {
			if err := writeFunc(w, val.(*css.NodeFunction)); err != nil {
				return err
			}
		}
		prevDelim = currDelim
	}
	return nil
}

func writeFunc(w io.Writer, f *css.NodeFunction) error {
	if _, err := w.Write(f.Func.Data); err != nil {
		return ErrWrite
	}
	for j, arg := range f.Args {
		if j > 0 {
			if _, err := w.Write([]byte(",")); err != nil {
				return ErrWrite
			}
		}
		if arg.Key != nil {
			if _, err := w.Write(arg.Key.Data); err != nil {
				return ErrWrite
			}
			if _, err := w.Write([]byte("=")); err != nil {
				return ErrWrite
			}
		}
		if _, err := w.Write(arg.Val.Data); err != nil {
			return ErrWrite
		}
	}
	if _, err := w.Write([]byte(")")); err != nil {
		return ErrWrite
	}
	return nil
}
