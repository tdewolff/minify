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
	"encoding/hex"
	"errors"
	"io"
	"math"
	"strconv"
	"strings"

	"github.com/tdewolff/css"
)

var epsilon = 0.00001

var shortenColorHex = map[string]string{
	"#000080": "navy",
	"#008000": "green",
	"#008080": "teal",
	"#4B0082": "indigo",
	"#800000": "maroon",
	"#800080": "purple",
	"#808000": "olive",
	"#808080": "gray",
	"#A0522D": "sienna",
	"#A52A2A": "brown",
	"#C0C0C0": "silver",
	"#CD853F": "peru",
	"#D2B48C": "tan",
	"#DA70D6": "orchid",
	"#DDA0DD": "plum",
	"#EE82EE": "violet",
	"#F0E68C": "khaki",
	"#F0FFFF": "azure",
	"#F5DEB3": "wheat",
	"#F5F5DC": "beige",
	"#FA8072": "salmon",
	"#FAF0E6": "linen",
	"#FF6347": "tomato",
	"#FF7F50": "coral",
	"#FFA500": "orange",
	"#FFC0CB": "pink",
	"#FFD700": "gold",
	"#FFE4C4": "bisque",
	"#FFFAFA": "snow",
	"#FFFFF0": "ivory",
	"#FF0000": "red",
	"#F00":    "red",
}

var shortenColorName = map[string]string{
	"black":                "#000",
	"darkblue":             "#00008B",
	"mediumblue":           "#0000CD",
	"darkgreen":            "#006400",
	"darkcyan":             "#008B8B",
	"deepskyblue":          "#00BFFF",
	"darkturquoise":        "#00CED1",
	"mediumspringgreen":    "#00FA9A",
	"springgreen":          "#00FF7F",
	"midnightblue":         "#191970",
	"dodgerblue":           "#1E90FF",
	"lightseagreen":        "#20B2AA",
	"forestgreen":          "#228B22",
	"seagreen":             "#2E8B57",
	"darkslategray":        "#2F4F4F",
	"limegreen":            "#32CD32",
	"mediumseagreen":       "#3CB371",
	"turquoise":            "#40E0D0",
	"royalblue":            "#4169E1",
	"steelblue":            "#4682B4",
	"darkslateblue":        "#483D8B",
	"mediumturquoise":      "#48D1CC",
	"darkolivegreen":       "#556B2F",
	"cadetblue":            "#5F9EA0",
	"cornflowerblue":       "#6495ED",
	"mediumaquamarine":     "#66CDAA",
	"slateblue":            "#6A5ACD",
	"olivedrab":            "#6B8E23",
	"slategray":            "#708090",
	"lightslateblue":       "#789",
	"mediumslateblue":      "#7B68EE",
	"lawngreen":            "#7CFC00",
	"chartreuse":           "#7FFF00",
	"aquamarine":           "#7FFFD4",
	"lightskyblue":         "#87CEFA",
	"blueviolet":           "#8A2BE2",
	"darkmagenta":          "#8B008B",
	"saddlebrown":          "#8B4513",
	"darkseagreen":         "#8FBC8F",
	"lightgreen":           "#90EE90",
	"mediumpurple":         "#9370DB",
	"darkviolet":           "#9400D3",
	"palegreen":            "#98FB98",
	"darkorchid":           "#9932CC",
	"yellowgreen":          "#9ACD32",
	"darkgray":             "#A9A9A9",
	"lightblue":            "#ADD8E6",
	"greenyellow":          "#ADFF2F",
	"paleturquoise":        "#AFEEEE",
	"lightsteelblue":       "#B0C4DE",
	"powderblue":           "#B0E0E6",
	"firebrick":            "#B22222",
	"darkgoldenrod":        "#B8860B",
	"mediumorchid":         "#BA55D3",
	"rosybrown":            "#BC8F8F",
	"darkkhaki":            "#BDB76B",
	"mediumvioletred":      "#C71585",
	"indianred":            "#CD5C5C",
	"chocolate":            "#D2691E",
	"lightgray":            "#D3D3D3",
	"goldenrod":            "#DAA520",
	"palevioletred":        "#DB7093",
	"gainsboro":            "#DCDCDC",
	"burlywood":            "#DEB887",
	"lightcyan":            "#E0FFFF",
	"lavender":             "#E6E6FA",
	"darksalmon":           "#E9967A",
	"palegoldenrod":        "#EEE8AA",
	"lightcoral":           "#F08080",
	"aliceblue":            "#F0F8FF",
	"honeydew":             "#F0FFF0",
	"sandybrown":           "#F4A460",
	"whitesmoke":           "#F5F5F5",
	"mintcream":            "#F5FFFA",
	"ghostwhite":           "#F8F8FF",
	"antiquewhite":         "#FAEBD7",
	"lightgoldenrodyellow": "#FAFAD2",
	"fuchsia":              "#F0F",
	"magenta":              "#F0F",
	"deeppink":             "#FF1493",
	"orangered":            "#FF4500",
	"darkorange":           "#FF8C00",
	"lightsalmon":          "#FFA07A",
	"lightpink":            "#FFB6C1",
	"peachpuff":            "#FFDAB9",
	"navajowhite":          "#FFDEAD",
	"moccasin":             "#FFE4B5",
	"mistyrose":            "#FFE4E1",
	"blanchedalmond":       "#FFEBCD",
	"papayawhip":           "#FFEFD5",
	"lavenderblush":        "#FFF0F5",
	"seashell":             "#FFF5EE",
	"cornsilk":             "#FFF8DC",
	"lemonchiffon":         "#FFFACD",
	"floralwhite":          "#FFFAF0",
	"yellow":               "#FF0",
	"lightyellow":          "#FFFFE0",
	"white":                "#FFF",
}

// CSS minifies CSS files, it reads from r and writes to w.
// It does a mediocre job of minifying CSS files and should be improved in the future.
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
			shortenNodes(n.(*css.NodeAtRule).Block)
		}
	}
}

func shortenSelector(sel *css.NodeSelector) {
	for _, n := range sel.Nodes {
		if n.TokenType == css.StringToken {
			s := n.Data[1:len(n.Data)-1]
			if css.IsIdent(s) {
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
				val.(*css.NodeFunction).Args[j] = shortenToken(arg)
			}
		}
	}

	prop := strings.ToLower(decl.Prop.String())
	if prop == "outline" || prop == "font-weight" {
		if len(decl.Vals) == 1 && decl.Vals[0].Type() == css.TokenNode {
			val := strings.ToLower(decl.Vals[0].String())
			if prop == "outline" && val == "none" {
				decl.Vals[0] = css.NewToken(css.NumberToken, "0")
			} else if prop == "font-weight" {
				if val == "normal" {
					decl.Vals[0] = css.NewToken(css.NumberToken, "400")
				} else if val == "bold" {
					decl.Vals[0] = css.NewToken(css.NumberToken, "700")
				}
			}
		}
	} else if prop == "margin" || prop == "padding" {
		if len(decl.Vals) == 2 {
			if decl.Vals[0].String() == decl.Vals[1].String() {
				decl.Vals = []css.Node{decl.Vals[0]}
			}
		} else if len(decl.Vals) == 3 {
			if decl.Vals[0].String() == decl.Vals[1].String() && decl.Vals[0].String() == decl.Vals[2].String() {
				decl.Vals = []css.Node{decl.Vals[0]}
			} else if decl.Vals[0].String() == decl.Vals[2].String() {
				decl.Vals = []css.Node{decl.Vals[0], decl.Vals[1]}
			}
		} else if len(decl.Vals) == 4 {
			if decl.Vals[0].String() == decl.Vals[1].String() && decl.Vals[0].String() == decl.Vals[2].String() && decl.Vals[0].String() == decl.Vals[3].String() {
				decl.Vals = []css.Node{decl.Vals[0]}
			} else if decl.Vals[0].String() == decl.Vals[2].String() && decl.Vals[1].String() == decl.Vals[3].String() {
				decl.Vals = []css.Node{decl.Vals[0], decl.Vals[1]}
			} else if decl.Vals[1].String() == decl.Vals[3].String() {
				decl.Vals = []css.Node{decl.Vals[0], decl.Vals[1], decl.Vals[2]}
			}
		}
	} else if prop == "font-family" {
		for _, val := range decl.Vals {
			if val.Type() == css.TokenNode && val.(*css.NodeToken).TokenType == css.StringToken {
				n := val.(*css.NodeToken)
				s := n.Data[1:len(n.Data)-1]
				unquote := true
				for _, fontName := range strings.Split(s, " ") {
					if !css.IsIdent(fontName) {
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
				if f.Func.String() == "rgba(" && len(f.Args) == 4 {
					d, _ := strconv.ParseFloat(f.Args[3].Data, 32)
					if math.Abs(d-1.0) < epsilon {
						f.Func = css.NewToken(css.FunctionToken, "rgb(")
						f.Args = f.Args[:len(f.Args)-1]
					}
				}
				if f.Func.String() == "rgb(" && len(f.Args) == 3 {
					var err error
					rgb := make([]byte, 3)
					for j := 0; j < 3; j++ {
						if f.Args[j].TokenType == css.NumberToken {
							var d int64
							d, err = strconv.ParseInt(f.Args[j].Data, 10, 32)
							if d < 0 {
								d = 0
							} else if d > 255 {
								d = 255
							}
							rgb[j] = byte(d)
						} else if f.Args[j].TokenType == css.PercentageToken {
							var d float64
							d, err = strconv.ParseFloat(f.Args[j].Data[:len(f.Args[j].Data)-1], 32)
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
						val := "#" + strings.ToUpper(hex.EncodeToString(rgb))
						if s, ok := shortenColorHex[val]; ok {
							decl.Vals[i] = css.NewToken(css.IdentToken, s)
						} else if len(val) == 7 && val[1] == val[2] && val[3] == val[4] && val[5] == val[6] {
							decl.Vals[i] = css.NewToken(css.HashToken, "#"+string(val[1])+string(val[3])+string(val[5]))
						} else {
							decl.Vals[i] = css.NewToken(css.HashToken, val)
						}
					}
				}
			} else if val.Type() == css.TokenNode {
				n := val.(*css.NodeToken)
				if n.TokenType == css.URLToken {
					s := n.Data[4:len(n.Data)-1]
					if s[0] == '"' || s[0] == '\'' {
						s = s[1:len(s)-1]
						if css.IsUrlUnquoted(s) {
							n.Data = "url("+s+")"
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

		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return token
		}
		if math.Abs(f) < epsilon {
			token.Data = "0"
			if token.TokenType == css.PercentageToken {
				token.Data += "%"
			}
		} else if len(token.Data) > 2 && token.Data[:2] == "0." {
			token.Data = token.Data[1:]
		} else if len(token.Data) > 3 && token.Data[:3] == "-0." {
			token.Data = "-"+token.Data[2:]
		}
	} else if token.TokenType == css.IdentToken {
		if h, ok := shortenColorName[val]; ok {
			token = css.NewToken(css.HashToken, h)
		}
	} else if token.TokenType == css.HashToken {
		if i, ok := shortenColorHex[strings.ToUpper(val)]; ok {
			token = css.NewToken(css.IdentToken, i)
		} else if len(val) == 7 && val[1] == val[2] && val[3] == val[4] && val[5] == val[6] {
			token = css.NewToken(css.HashToken, "#"+strings.ToUpper(string(val[1])+string(val[3])+string(val[5])))
		} else {
			token.Data = strings.ToUpper(token.Data)
		}
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
						op := sel.Nodes[0].String()
						if tt == css.DelimToken && (op == ">" || op == "+" || op == "~") || tt == css.IncludeMatchToken || tt == css.DashMatchToken ||
							tt == css.PrefixMatchToken || tt == css.SuffixMatchToken || tt == css.SubstringMatchToken {
							if _, err := w.Write([]byte(op)); err != nil {
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
					if _, err := w.Write([]byte(sel.String())); err != nil {
						return ErrWrite
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
			if _, err := w.Write([]byte(atRule.At.String())); err != nil {
				return ErrWrite
			}
			for _, node := range atRule.Nodes {
				if _, err := w.Write([]byte(" " + node.String())); err != nil {
					return ErrWrite
				}
			}
			if len(atRule.Block) > 0 {
				if _, err := w.Write([]byte("{")); err != nil {
					return ErrWrite
				}
				if err := writeNodes(w, atRule.Block); err != nil {
					return err
				}
				if _, err := w.Write([]byte("}")); err != nil {
					return ErrWrite
				}
			} else {
				semicolonQueued = true
			}
		default:
			if _, err := w.Write([]byte(n.String())); err != nil {
				return ErrWrite
			}
		}
	}
	return nil
}

func writeDecl(w io.Writer, decl *css.NodeDeclaration) error {
	if _, err := w.Write([]byte(decl.Prop.String() + ":")); err != nil {
		return ErrWrite
	}
	prevDelim := false
	for j, val := range decl.Vals {
		currDelim := (val.Type() == css.TokenNode && (val.(*css.NodeToken).TokenType == css.DelimToken || val.(*css.NodeToken).TokenType == css.CommaToken))
		if j > 0 && !currDelim && !prevDelim {
			if _, err := w.Write([]byte(" ")); err != nil {
				return ErrWrite
			}
		}
		if _, err := w.Write([]byte(val.String())); err != nil {
			return ErrWrite
		}
		prevDelim = currDelim
	}
	return nil
}
