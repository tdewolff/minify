// Package js minifies ECMAScript5.1 following the specifications at http://www.ecma-international.org/ecma-262/5.1/.
package js

import (
	"fmt"
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
	fmt.Println(ast)
	if err != nil {
		return err
	}
	o.minifyNode(w, ast)
	_, err = w.Write(nil)
	return err
}

func (o *Minifier) minifyNode(w io.Writer, n js.Node) {
	semicolon := false
	switch n.GrammarType {
	case js.TokenGrammar:
		w.Write(n.Data)
	case js.ModuleGrammar:
		for _, node := range n.Nodes {
			o.minifyNode(w, node)
		}
	case js.BindingGrammar:
		for _, node := range n.Nodes {
			o.minifyNode(w, node)
		}
	case js.ClauseGrammar:
		for _, node := range n.Nodes {
			o.minifyNode(w, node)
		}
	case js.MethodGrammar:
		for _, node := range n.Nodes {
			o.minifyNode(w, node)
		}
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
				w.Write(node.Data)
			} else {
				o.minifyNode(w, node)
			}
			ttPrevPrev = ttPrev
			ttPrev = tt
			tt = 0
		}
	case js.StmtGrammar:
		if semicolon {
			w.Write([]byte(";"))
		}
		switch n.Nodes[0].TokenType {
		case js.VarToken, js.LetToken, js.ConstToken:
			w.Write(n.Nodes[0].Data)
			w.Write([]byte(" "))
			for _, node := range n.Nodes[1:] {
				o.minifyNode(w, node)
			}
		case js.FunctionToken:
			w.Write(n.Nodes[0].Data)
			d := 1
			if n.Nodes[1].TokenType == js.MulToken {
				w.Write(n.Nodes[1].Data)
				d++
				if n.Nodes[2].TokenType == js.IdentifierToken {
					w.Write(n.Nodes[2].Data)
					d++
				}
			} else if n.Nodes[1].TokenType == js.IdentifierToken {
				w.Write([]byte(" "))
				w.Write(n.Nodes[1].Data)
				d++
			}
			w.Write([]byte("("))
			for _, node := range n.Nodes[d : len(n.Nodes)-1] {
				o.minifyNode(w, node)
			}
			w.Write([]byte(")"))
			o.minifyNode(w, n.Nodes[len(n.Nodes)-1])
		case js.IfToken:
			w.Write(n.Nodes[0].Data)
			w.Write([]byte("("))
			o.minifyNode(w, n.Nodes[1])
			w.Write([]byte(")"))
			o.minifyNode(w, n.Nodes[2])
			if 3 < len(n.Nodes) {
				w.Write(n.Nodes[3].Data) // else
				o.minifyNode(w, n.Nodes[4])
			}
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
		default:
			for _, node := range n.Nodes {
				o.minifyNode(w, node)
			}
		}
		semicolon = true
	case js.ErrorGrammar:
		panic("should not happen")
	}
}
