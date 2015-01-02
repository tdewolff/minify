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
	"#4B0082": []byte("indigo"),
	"#800000": []byte("maroon"),
	"#800080": []byte("purple"),
	"#808000": []byte("olive"),
	"#808080": []byte("gray"),
	"#A0522D": []byte("sienna"),
	"#A52A2A": []byte("brown"),
	"#C0C0C0": []byte("silver"),
	"#CD853F": []byte("peru"),
	"#D2B48C": []byte("tan"),
	"#DA70D6": []byte("orchid"),
	"#DDA0DD": []byte("plum"),
	"#EE82EE": []byte("violet"),
	"#F0E68C": []byte("khaki"),
	"#F0FFFF": []byte("azure"),
	"#F5DEB3": []byte("wheat"),
	"#F5F5DC": []byte("beige"),
	"#FA8072": []byte("salmon"),
	"#FAF0E6": []byte("linen"),
	"#FF6347": []byte("tomato"),
	"#FF7F50": []byte("coral"),
	"#FFA500": []byte("orange"),
	"#FFC0CB": []byte("pink"),
	"#FFD700": []byte("gold"),
	"#FFE4C4": []byte("bisque"),
	"#FFFAFA": []byte("snow"),
	"#FFFFF0": []byte("ivory"),
	"#FF0000": []byte("red"),
	"#F00":    []byte("red"),
}

var shortenColorName = map[string][]byte{
	"black":                []byte("#000"),
	"darkblue":             []byte("#00008B"),
	"mediumblue":           []byte("#0000CD"),
	"darkgreen":            []byte("#006400"),
	"darkcyan":             []byte("#008B8B"),
	"deepskyblue":          []byte("#00BFFF"),
	"darkturquoise":        []byte("#00CED1"),
	"mediumspringgreen":    []byte("#00FA9A"),
	"springgreen":          []byte("#00FF7F"),
	"midnightblue":         []byte("#191970"),
	"dodgerblue":           []byte("#1E90FF"),
	"lightseagreen":        []byte("#20B2AA"),
	"forestgreen":          []byte("#228B22"),
	"seagreen":             []byte("#2E8B57"),
	"darkslategray":        []byte("#2F4F4F"),
	"limegreen":            []byte("#32CD32"),
	"mediumseagreen":       []byte("#3CB371"),
	"turquoise":            []byte("#40E0D0"),
	"royalblue":            []byte("#4169E1"),
	"steelblue":            []byte("#4682B4"),
	"darkslateblue":        []byte("#483D8B"),
	"mediumturquoise":      []byte("#48D1CC"),
	"darkolivegreen":       []byte("#556B2F"),
	"cadetblue":            []byte("#5F9EA0"),
	"cornflowerblue":       []byte("#6495ED"),
	"mediumaquamarine":     []byte("#66CDAA"),
	"slateblue":            []byte("#6A5ACD"),
	"olivedrab":            []byte("#6B8E23"),
	"slategray":            []byte("#708090"),
	"lightslateblue":       []byte("#789"),
	"mediumslateblue":      []byte("#7B68EE"),
	"lawngreen":            []byte("#7CFC00"),
	"chartreuse":           []byte("#7FFF00"),
	"aquamarine":           []byte("#7FFFD4"),
	"lightskyblue":         []byte("#87CEFA"),
	"blueviolet":           []byte("#8A2BE2"),
	"darkmagenta":          []byte("#8B008B"),
	"saddlebrown":          []byte("#8B4513"),
	"darkseagreen":         []byte("#8FBC8F"),
	"lightgreen":           []byte("#90EE90"),
	"mediumpurple":         []byte("#9370DB"),
	"darkviolet":           []byte("#9400D3"),
	"palegreen":            []byte("#98FB98"),
	"darkorchid":           []byte("#9932CC"),
	"yellowgreen":          []byte("#9ACD32"),
	"darkgray":             []byte("#A9A9A9"),
	"lightblue":            []byte("#ADD8E6"),
	"greenyellow":          []byte("#ADFF2F"),
	"paleturquoise":        []byte("#AFEEEE"),
	"lightsteelblue":       []byte("#B0C4DE"),
	"powderblue":           []byte("#B0E0E6"),
	"firebrick":            []byte("#B22222"),
	"darkgoldenrod":        []byte("#B8860B"),
	"mediumorchid":         []byte("#BA55D3"),
	"rosybrown":            []byte("#BC8F8F"),
	"darkkhaki":            []byte("#BDB76B"),
	"mediumvioletred":      []byte("#C71585"),
	"indianred":            []byte("#CD5C5C"),
	"chocolate":            []byte("#D2691E"),
	"lightgray":            []byte("#D3D3D3"),
	"goldenrod":            []byte("#DAA520"),
	"palevioletred":        []byte("#DB7093"),
	"gainsboro":            []byte("#DCDCDC"),
	"burlywood":            []byte("#DEB887"),
	"lightcyan":            []byte("#E0FFFF"),
	"lavender":             []byte("#E6E6FA"),
	"darksalmon":           []byte("#E9967A"),
	"palegoldenrod":        []byte("#EEE8AA"),
	"lightcoral":           []byte("#F08080"),
	"aliceblue":            []byte("#F0F8FF"),
	"honeydew":             []byte("#F0FFF0"),
	"sandybrown":           []byte("#F4A460"),
	"whitesmoke":           []byte("#F5F5F5"),
	"mintcream":            []byte("#F5FFFA"),
	"ghostwhite":           []byte("#F8F8FF"),
	"antiquewhite":         []byte("#FAEBD7"),
	"lightgoldenrodyellow": []byte("#FAFAD2"),
	"fuchsia":              []byte("#F0F"),
	"magenta":              []byte("#F0F"),
	"deeppink":             []byte("#FF1493"),
	"orangered":            []byte("#FF4500"),
	"darkorange":           []byte("#FF8C00"),
	"lightsalmon":          []byte("#FFA07A"),
	"lightpink":            []byte("#FFB6C1"),
	"peachpuff":            []byte("#FFDAB9"),
	"navajowhite":          []byte("#FFDEAD"),
	"moccasin":             []byte("#FFE4B5"),
	"mistyrose":            []byte("#FFE4E1"),
	"blanchedalmond":       []byte("#FFEBCD"),
	"papayawhip":           []byte("#FFEFD5"),
	"lavenderblush":        []byte("#FFF0F5"),
	"seashell":             []byte("#FFF5EE"),
	"cornsilk":             []byte("#FFF8DC"),
	"lemonchiffon":         []byte("#FFFACD"),
	"floralwhite":          []byte("#FFFAF0"),
	"yellow":               []byte("#FF0"),
	"lightyellow":          []byte("#FFFFE0"),
	"white":                []byte("#FFF"),
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

func shortenNodes(nodes []css.Node) {
	for _, n := range nodes {
		switch n.Type() {
		case css.DeclarationNode:
			shortenDecl(n.(*css.NodeDeclaration))
		case css.RulesetNode:
			for _, selGroup := range n.(*css.NodeRuleset).SelGroups {
				for _, sel := range selGroup.Selectors {
					shortenSelector(sel)
				}
			}
			for _, decl := range n.(*css.NodeRuleset).Decls {
				shortenDecl(decl)
			}
		case css.AtRuleNode:
			if n.(*css.NodeAtRule).Block != nil {
				shortenNodes(n.(*css.NodeAtRule).Block.Nodes)
			}
		case css.BlockNode:
			shortenNodes(n.(*css.NodeBlock).Nodes)
		}
	}
}

func shortenSelector(sel *css.NodeSelector) {
	for _, n := range sel.Nodes {
		if n.TokenType == css.StringToken {
			s := n.Data[1 : len(n.Data)-1]
			if css.IsIdent([]byte(s)) {
				n.Data = s
			}
		}
	}
}

func shortenDecl(decl *css.NodeDeclaration) {
	// shorten zeros
	for i, val := range decl.Vals {
		if val.Type() == css.TokenNode {
			decl.Vals[i] = shortenToken(val.(*css.NodeToken))
		} else if val.Type() == css.FunctionNode {
			for j, arg := range val.(*css.NodeFunction).Args {
				val.(*css.NodeFunction).Args[j].Val = shortenToken(arg.Val)
			}
		}
	}

	prop := bytes.ToLower(decl.Prop.Data)
	if bytes.Equal(prop, []byte("outline")) || bytes.Equal(prop, []byte("font-weight")) {
		if len(decl.Vals) == 1 && decl.Vals[0].Type() == css.TokenNode {
			val := bytes.ToLower(decl.Vals[0].(*css.NodeToken).Data)
			if bytes.Equal(prop, []byte("outline")) && bytes.Equal(val, []byte("none")) {
				decl.Vals[0] = css.NewToken(css.NumberToken, []byte("0"))
			} else if bytes.Equal(prop, []byte("font-weight")) {
				if bytes.Equal(val, []byte("normal")) {
					decl.Vals[0] = css.NewToken(css.NumberToken, []byte("400"))
				} else if bytes.Equal(val, []byte("bold")) {
					decl.Vals[0] = css.NewToken(css.NumberToken, []byte("700"))
				}
			}
		}
	} else if bytes.Equal(prop, []byte("margin")) || bytes.Equal(prop, []byte("padding")) {
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
			if bytes.Equal(decl.Vals[0].(*css.NodeToken).Data, decl.Vals[1].(*css.NodeToken).Data) && bytes.Equal(decl.Vals[0].(*css.NodeToken).Data, decl.Vals[2].(*css.NodeToken).Data) && bytes.Equal(decl.Vals[0].(*css.NodeToken).Data, decl.Vals[3].(*css.NodeToken).Data) {
				decl.Vals = []css.Node{decl.Vals[0]}
			} else if bytes.Equal(decl.Vals[0].(*css.NodeToken).Data, decl.Vals[2].(*css.NodeToken).Data) && bytes.Equal(decl.Vals[1].(*css.NodeToken).Data, decl.Vals[3].(*css.NodeToken).Data) {
				decl.Vals = []css.Node{decl.Vals[0], decl.Vals[1]}
			} else if bytes.Equal(decl.Vals[1].(*css.NodeToken).Data, decl.Vals[3].(*css.NodeToken).Data) {
				decl.Vals = []css.Node{decl.Vals[0], decl.Vals[1], decl.Vals[2]}
			}
		}
	} else if bytes.Equal(prop, []byte("font-family")) {
		for _, val := range decl.Vals {
			if val.Type() == css.TokenNode && val.(*css.NodeToken).TokenType == css.StringToken {
				n := val.(*css.NodeToken)
				s := n.Data[1 : len(n.Data)-1]
				unquote := true
				for _, fontName := range bytes.Split(s, []byte(" ")) {
					if !css.IsIdent([]byte(fontName)) {
						unquote = false
						break
					}
				}
				if unquote {
					n.Data = s
				}
			}
		}
	} else {
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
						val := append([]byte("#"), bytes.ToUpper(valHex)...)
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
						s = s[1 : len(s)-1]
						if css.IsUrlUnquoted([]byte(s)) {
							n.Data = append([]byte("url("), append(s, ')')...)
						}
					}
				}
			}
		}
	}
}

func shortenToken(token *css.NodeToken) *css.NodeToken {
	val := token.Data
	if token.TokenType == css.NumberToken || token.TokenType == css.DimensionToken || token.TokenType == css.PercentageToken {
		if token.TokenType == css.PercentageToken {
			val = val[:len(val)-1]
		} else if token.TokenType == css.DimensionToken {
			val, _ = css.SplitDimensionToken(val)
		}

		f, err := strconv.ParseFloat(string(val), 64)
		if err != nil {
			return token
		}
		if math.Abs(f) < epsilon {
			token.Data = []byte("0")
			if token.TokenType == css.PercentageToken {
				token.Data = append(token.Data, '%')
			}
		} else if len(token.Data) > 2 && bytes.Equal(token.Data[:2], []byte("0.")) {
			token.Data = token.Data[1:]
		} else if len(token.Data) > 3 && bytes.Equal(token.Data[:3], []byte("-0.")) {
			token.Data = append([]byte("-"), token.Data[2:]...)
		}
	} else if token.TokenType == css.IdentToken {
		if h, ok := shortenColorName[string(val)]; ok {
			token = css.NewToken(css.HashToken, h)
		}
	} else if token.TokenType == css.HashToken {
		if i, ok := shortenColorHex[string(bytes.ToUpper(val))]; ok {
			token = css.NewToken(css.IdentToken, i)
		} else if len(val) == 7 && val[1] == val[2] && val[3] == val[4] && val[5] == val[6] {
			token = css.NewToken(css.HashToken, append([]byte("#"), bytes.ToUpper(append([]byte{val[1]}, val[3], val[5]))...))
		} else {
			token.Data = bytes.ToUpper(token.Data)
		}
	} else if token.TokenType == css.StringToken {
		token.Data = bytes.Replace(token.Data, []byte("\\\r\n"), []byte(""), -1)
		token.Data = bytes.Replace(token.Data, []byte("\\\r"), []byte(""), -1)
		token.Data = bytes.Replace(token.Data, []byte("\\\n"), []byte(""), -1)
	}
	return token
}

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
					if len(sel.Nodes) == 1 {
						tt := sel.Nodes[0].TokenType
						op := sel.Nodes[0].Data
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
						if _, err := w.Write(node.Data); err != nil {
							return ErrWrite
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
