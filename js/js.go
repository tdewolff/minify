// Package js minifies ECMAScript5.1 following the specifications at http://www.ecma-international.org/ecma-262/5.1/.
package js

import (
	"io"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/parse/v2/js"
)

var (
	spaceBytes   = []byte(" ")
	newlineBytes = []byte("\n")
)

////////////////////////////////////////////////////////////////

// DefaultMinifier is the default minifier.
var DefaultMinifier = &Minifier{}

// Minifier is a JS minifier.
type Minifier struct{}

// Minify minifies JS data, it reads from r and writes to w.
func Minify(m *minify.M, w io.Writer, r io.Reader, params map[string]string) error {
	return DefaultMinifier.Minify(m, w, r, params)
}

// Minify minifies JS data, it reads from r and writes to w.
func (o *Minifier) Minify(_ *minify.M, w io.Writer, r io.Reader, _ map[string]string) error {
	ast, err := js.Parse(r)
	if err != nil {
		return err
	}
	o.minifyNode(w, ast)
	_, err = w.Write(nil)
	return err
}

func (o *Minifier) minifyNode(w io.Writer, n js.Node) {
	switch n.GrammarType {
	case js.TokenGrammar:
		o.minifyToken(w, n)
	case js.ModuleGrammar:
		for i, node := range n.Nodes {
			if i != 0 {
				w.Write([]byte(";"))
			}
			o.minifyNode(w, node)
		}
	case js.BindingGrammar:
		for _, node := range n.Nodes {
			o.minifyNode(w, node)
		}
	case js.ClauseGrammar:
		w.Write(n.Nodes[0].Data)
		if n.Nodes[0].TokenType == js.CaseToken {
			w.Write([]byte(" "))
			o.minifyNode(w, n.Nodes[1])
			w.Write([]byte(":"))
			for _, node := range n.Nodes[2:] {
				o.minifyNode(w, node)
			}
		} else {
			w.Write([]byte(":"))
			for _, node := range n.Nodes[1:] {
				o.minifyNode(w, node)
			}
		}
	case js.ParamsGrammar:
		w.Write([]byte("("))
		for i, node := range n.Nodes {
			if i != 0 && n.Nodes[i-1].TokenType != js.EllipsisToken {
				w.Write([]byte(","))
			}
			o.minifyNode(w, node)
		}
		w.Write([]byte(")"))
	case js.MethodGrammar:
		prevAsterisk := true
		for _, node := range n.Nodes[:len(n.Nodes)-2] {
			if !prevAsterisk && node.TokenType != js.MulToken {
				w.Write([]byte(" "))
			}
			o.minifyNode(w, node)
			prevAsterisk = node.TokenType == js.MulToken
		}
		o.minifyNode(w, n.Nodes[len(n.Nodes)-2])
		o.minifyNode(w, n.Nodes[len(n.Nodes)-1])
	case js.ExprGrammar:
		var ttPrevPrev, ttPrev, tt js.TokenType
		for _, node := range n.Nodes {
			if node.GrammarType == js.TokenGrammar {
				tt = node.TokenType
				if ((ttPrev == js.AddToken || ttPrev == js.PosToken) && (tt == js.AddToken || tt == js.PosToken || tt == js.PreIncrToken)) || ((ttPrev == js.SubToken || ttPrev == js.NegToken) && (tt == js.SubToken || tt == js.NegToken || tt == js.PreDecrToken)) || (ttPrev == js.PostDecrToken && tt == js.GtToken) || (ttPrevPrev == js.LtToken && ttPrev == js.NotToken && tt == js.PreDecrToken) {
					w.Write([]byte(" "))
				} else if (js.IsIdentifier(ttPrev) || ttPrev == js.RegExpToken) && js.IsIdentifier(tt) {
					w.Write([]byte(" "))
				}
				o.minifyToken(w, node)
			} else if node.GrammarType == js.ParamsGrammar && ttPrev != js.FunctionToken && ttPrevPrev != js.FunctionToken && len(node.Nodes) == 1 && len(node.Nodes[0].Nodes) == 1 && node.Nodes[0].Nodes[0].TokenType == js.IdentifierToken {
				w.Write(node.Nodes[0].Nodes[0].Data) // (a)=>{a++} --> a=>{a++}
			} else if ttPrev == js.ArrowToken && node.GrammarType == js.StmtGrammar && len(node.Nodes) == 3 {
				o.minifyNode(w, node.Nodes[1])
			} else {
				o.minifyNode(w, node)
			}
			ttPrevPrev = ttPrev
			ttPrev = tt
			tt = 0
		}
	case js.StmtGrammar:
		o.minifyStmt(w, n)
	case js.ErrorGrammar:
		panic("should not happen")
	}
}

func (o *Minifier) minifyStmt(w io.Writer, n js.Node) {
	if len(n.Nodes) == 0 || n.Nodes[0].TokenType == js.SemicolonToken {
		return
	}

	switch n.Nodes[0].TokenType {
	case js.OpenBraceToken:
		w.Write([]byte("{"))
		for i, node := range n.Nodes[1 : len(n.Nodes)-1] {
			if i != 0 {
				w.Write([]byte(";"))
			}
			o.minifyNode(w, node)
		}
		w.Write([]byte("}"))
	case js.VarToken, js.LetToken, js.ConstToken:
		w.Write(n.Nodes[0].Data)
		w.Write([]byte(" "))
		for i, node := range n.Nodes[1:] {
			if i != 0 {
				w.Write([]byte(","))
			}
			o.minifyNode(w, node)
		}
	case js.BreakToken, js.ContinueToken:
		w.Write(n.Nodes[0].Data)
		if 1 < len(n.Nodes) {
			w.Write([]byte(" "))
			w.Write(n.Nodes[1].Data)
		}
	case js.ReturnToken, js.ThrowToken:
		w.Write(n.Nodes[0].Data)
		if 1 < len(n.Nodes) {
			if o.isIdentifierStart(n.Nodes[1]) {
				w.Write([]byte(" "))
			}
			o.minifyNode(w, n.Nodes[1])
		}
	case js.SwitchToken:
		w.Write(n.Nodes[0].Data)
		w.Write([]byte("("))
		o.minifyNode(w, n.Nodes[1])
		w.Write([]byte("){"))
		for i, node := range n.Nodes[2:] {
			if i != 0 {
				w.Write([]byte(";"))
			}
			o.minifyNode(w, node)
		}
		w.Write([]byte("}"))
	case js.FunctionToken:
		w.Write(n.Nodes[0].Data)
		if n.Nodes[1].TokenType == js.MulToken {
			w.Write(n.Nodes[1].Data)
			if n.Nodes[2].TokenType == js.IdentifierToken {
				w.Write(n.Nodes[2].Data)
			}
		} else if n.Nodes[1].TokenType == js.IdentifierToken {
			w.Write([]byte(" "))
			w.Write(n.Nodes[1].Data)
		}
		o.minifyNode(w, n.Nodes[len(n.Nodes)-2])
		o.minifyNode(w, n.Nodes[len(n.Nodes)-1])
	case js.IfToken:
		w.Write(n.Nodes[0].Data)
		w.Write([]byte("("))
		o.minifyNode(w, n.Nodes[1])
		w.Write([]byte(")"))

		ifStmt := n.Nodes[2]
		if len(ifStmt.Nodes) == 3 && ifStmt.Nodes[0].TokenType == js.OpenBraceToken {
			if ifStmt.Nodes[1].Nodes[0].TokenType != js.IfToken || 3 < len(ifStmt.Nodes[1].Nodes) {
				n.Nodes[2] = n.Nodes[2].Nodes[1]
				ifStmt = ifStmt.Nodes[1] // block with one statement, but not if statement without else
			}
		}
		o.minifyNode(w, ifStmt)

		if 3 < len(n.Nodes) && n.Nodes[4].Nodes[0].TokenType != js.SemicolonToken {
			if o.needsSemicolon(ifStmt) {
				w.Write([]byte(";"))
			}
			w.Write(n.Nodes[3].Data) // else

			elseStmt := n.Nodes[4]
			if len(elseStmt.Nodes) == 3 && elseStmt.Nodes[0].TokenType == js.OpenBraceToken {
				n.Nodes[4] = n.Nodes[4].Nodes[1]
				elseStmt = elseStmt.Nodes[1] // block with one statement
			}
			if o.isIdentifierStart(elseStmt) {
				w.Write([]byte(" "))
			}
			o.minifyNode(w, elseStmt)
		}
	case js.WithToken:
		w.Write(n.Nodes[0].Data)
		w.Write([]byte("("))
		o.minifyNode(w, n.Nodes[1])
		w.Write([]byte(")"))
		if len(n.Nodes[2].Nodes) == 3 && n.Nodes[2].Nodes[0].TokenType == js.OpenBraceToken {
			n.Nodes[2] = n.Nodes[2].Nodes[1] // block with one statement
		}
		o.minifyNode(w, n.Nodes[2])
	case js.ForToken:
		w.Write(n.Nodes[0].Data)
		d := 1
		if n.Nodes[1].TokenType == js.AwaitToken {
			w.Write([]byte(" "))
			w.Write(n.Nodes[1].Data)
			d++
		}
		w.Write([]byte("("))
		prevIdentifier := false
		for _, node := range n.Nodes[d : len(n.Nodes)-1] {
			identifier := o.isIdentifierStart(node)
			if prevIdentifier && identifier {
				w.Write([]byte(" "))
			}
			prevIdentifier = identifier
			o.minifyNode(w, node)
		}
		w.Write([]byte(")"))
		last := len(n.Nodes) - 1
		if len(n.Nodes[last].Nodes) == 3 && n.Nodes[last].Nodes[0].TokenType == js.OpenBraceToken {
			n.Nodes[last] = n.Nodes[last].Nodes[1] // block with one statement
		}
		o.minifyNode(w, n.Nodes[last])
	case js.WhileToken:
		w.Write(n.Nodes[0].Data)
		w.Write([]byte("("))
		o.minifyNode(w, n.Nodes[1])
		w.Write([]byte(")"))
		if len(n.Nodes[2].Nodes) == 3 && n.Nodes[2].Nodes[0].TokenType == js.OpenBraceToken {
			n.Nodes[2] = n.Nodes[2].Nodes[1] // block with one statement
		}
		o.minifyNode(w, n.Nodes[2])
	case js.DoToken:
		w.Write(n.Nodes[0].Data)
		if o.isIdentifierStart(n.Nodes[1]) {
			w.Write([]byte("{"))
			o.minifyNode(w, n.Nodes[1])
			w.Write([]byte("}"))
		} else {
			o.minifyNode(w, n.Nodes[1])
			if o.needsSemicolon(n.Nodes[1]) {
				w.Write([]byte(";"))
			}
		}
		o.minifyNode(w, n.Nodes[2])
		w.Write([]byte("("))
		o.minifyNode(w, n.Nodes[3])
		w.Write([]byte(")"))
	case js.ClassToken:
		w.Write(n.Nodes[0].Data)
		d := 1
		if 1 < len(n.Nodes) && n.Nodes[1].GrammarType != js.MethodGrammar {
			w.Write([]byte(" "))
			w.Write(n.Nodes[1].Data)
			d = 2
			if 2 < len(n.Nodes) && n.Nodes[2].GrammarType != js.MethodGrammar {
				w.Write([]byte(" "))
				w.Write(n.Nodes[2].Data)
				w.Write([]byte(" "))
				o.minifyNode(w, n.Nodes[3])
				d = 4
			}
		}
		w.Write([]byte("{"))
		for _, node := range n.Nodes[d:] {
			o.minifyNode(w, node)
		}
		w.Write([]byte("}"))
	case js.ImportToken:
		prevIdentifier := false
		for _, node := range n.Nodes {
			if prevIdentifier && js.IsIdentifier(node.TokenType) {
				w.Write([]byte(" "))
			}
			w.Write(node.Data)
			prevIdentifier = js.IsIdentifier(node.TokenType)
		}
	case js.ExportToken:
		if n.Nodes[1].TokenType == js.MulToken || n.Nodes[1].TokenType == js.OpenBraceToken {
			prevIdentifier := false
			for _, node := range n.Nodes {
				if prevIdentifier && js.IsIdentifier(node.TokenType) {
					w.Write([]byte(" "))
				}
				o.minifyNode(w, node)
				prevIdentifier = js.IsIdentifier(node.TokenType)
			}
		} else {
			w.Write(n.Nodes[0].Data)
			w.Write([]byte(" "))
			if n.Nodes[1].TokenType == js.DefaultToken {
				w.Write(n.Nodes[1].Data)
				w.Write([]byte(" "))
				o.minifyNode(w, n.Nodes[2])
			} else {
				o.minifyNode(w, n.Nodes[1])
			}
		}
	case js.TryToken:
		w.Write(n.Nodes[0].Data)
		o.minifyNode(w, n.Nodes[1])
		if 2 < len(n.Nodes) {
			if n.Nodes[2].TokenType == js.CatchToken {
				w.Write(n.Nodes[2].Data)
				if len(n.Nodes) == 4 {
					o.minifyNode(w, n.Nodes[3])
				} else {
					w.Write([]byte("("))
					o.minifyNode(w, n.Nodes[3])
					w.Write([]byte(")"))
					o.minifyNode(w, n.Nodes[4])
					if 5 < len(n.Nodes) {
						w.Write(n.Nodes[5].Data)
						o.minifyNode(w, n.Nodes[6])
					}
				}
			} else {
				w.Write(n.Nodes[2].Data)
				o.minifyNode(w, n.Nodes[3])
			}
		}
	default:
		for _, node := range n.Nodes {
			o.minifyNode(w, node)
		}
	}
}

func (o *Minifier) minifyToken(w io.Writer, n js.Node) {
	if n.TokenType == js.DecimalToken {
		w.Write(minify.Number(n.Data, 0))
	} else {
		w.Write(n.Data)
	}
}

func (o *Minifier) isIdentifierStart(n js.Node) bool {
	if n.GrammarType == js.TokenGrammar {
		return js.IsIdentifier(n.TokenType) || js.IsNumeric(n.TokenType) && n.Data[0] != '.'
	}
	return o.isIdentifierStart(n.Nodes[0])
}

func (o *Minifier) needsSemicolon(n js.Node) bool {
	if n.GrammarType == js.ExprGrammar || n.TokenType == js.SemicolonToken {
		return true
	} else if n.GrammarType == js.StmtGrammar {
		//if len(n.Nodes) == 3 && n.Nodes[0].TokenType == js.OpenBraceToken {
		//		return o.needsSemicolon(n.Nodes[1])
		//	}
		return o.needsSemicolon(n.Nodes[len(n.Nodes)-1])
	}
	return false
}
