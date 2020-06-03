// Package js minifies ECMAScript5.1 following the specifications at http://www.ecma-international.org/ecma-262/5.1/.
package js

import (
	"bytes"
	"io"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/js"
)

var (
	spaceBytes     = []byte(" ")
	starBytes      = []byte("*")
	semicolonBytes = []byte(";")
)

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
	z := parse.NewInput(r)
	ast, err := js.Parse(z)
	if err != nil {
		return err
	}

	m := &jsMinifier{
		o:       o,
		w:       w,
		renamer: newRenamer(ast.Unbound),
	}
	for _, item := range ast.List {
		m.writeSemicolon()
		m.minifyStmt(item)
	}

	if _, err := w.Write(nil); err != nil {
		return err
	}
	return nil
}

type jsMinifier struct {
	o *Minifier
	w io.Writer

	prev           []byte
	needsSemicolon bool
	needsSpace     bool
	*renamer
}

func (m *jsMinifier) write(b []byte) {
	if m.needsSpace && js.IsIdentifierStart(b) {
		m.w.Write(spaceBytes)
	}
	m.w.Write(b)
	m.prev = b
	m.needsSpace = false
}

func (m *jsMinifier) writeSpaceAfterIdent() {
	if js.IsIdentifierEnd(m.prev) || 1 < len(m.prev) && m.prev[0] == '/' {
		m.w.Write(spaceBytes)
	}
}

func (m *jsMinifier) writeSpaceBeforeIdent() {
	m.needsSpace = true
}

func (m *jsMinifier) requireSemicolon() {
	m.needsSemicolon = true
}

func (m *jsMinifier) writeSemicolon() {
	if m.needsSemicolon {
		m.w.Write(semicolonBytes)
		m.needsSemicolon = false
		m.needsSpace = false
	}
}

func (m *jsMinifier) minifyStmt(i js.IStmt) {
	switch stmt := i.(type) {
	case *js.ExprStmt:
		m.minifyExpr(stmt.Value, false)
		m.requireSemicolon()
	case *js.VarDecl:
		m.minifyVarDecl(*stmt)
		m.requireSemicolon()
	case *js.IfStmt:
		hasIf := !isEmptyStmt(stmt.Body)
		hasElse := !isEmptyStmt(stmt.Else)

		m.write([]byte("if("))
		m.minifyExpr(stmt.Cond, false)
		m.write([]byte(")"))

		if !hasIf && hasElse {
			m.requireSemicolon()
		} else if hasIf {
			if block, ok := stmt.Body.(*js.BlockStmt); ok && len(block.List) == 1 {
				stmt.Body = block.List[0]
			}
			if ifStmt, ok := stmt.Body.(*js.IfStmt); ok && isEmptyStmt(ifStmt.Else) {
				m.write([]byte("{"))
				m.minifyStmt(stmt.Body)
				m.write([]byte("}"))
				m.needsSemicolon = false
			} else {
				m.minifyStmt(stmt.Body)
			}
		}
		if hasElse {
			m.writeSemicolon()
			m.write([]byte("else"))
			m.writeSpaceBeforeIdent()
			m.minifyStmt(stmt.Else)
		}
	case *js.BlockStmt:
		if len(stmt.List) == 1 {
			m.minifyStmt(stmt.List[0])
		} else {
			m.minifyBlockStmt(*stmt)
		}
	case *js.ReturnStmt:
		m.write([]byte("return"))
		m.writeSpaceBeforeIdent()
		m.minifyExpr(stmt.Value, false)
		m.requireSemicolon()
	case *js.LabelledStmt:
		m.write(stmt.Token.Data)
		m.write([]byte(":"))
		m.minifyStmt(stmt.Value)
	case *js.BranchStmt:
		m.write(stmt.Type.Bytes())
		if stmt.Name != nil {
			m.write([]byte(" "))
			m.write(stmt.Name.Data)
		}
		m.requireSemicolon()
	case *js.WithStmt:
		m.write([]byte("with("))
		m.minifyExpr(stmt.Cond, false)
		m.write([]byte(")"))
		m.minifyStmt(stmt.Body)
	case *js.DoWhileStmt:
		m.write([]byte("do"))
		m.writeSpaceBeforeIdent()
		m.minifyStmt(stmt.Body)
		m.writeSemicolon()
		m.write([]byte("while("))
		m.minifyExpr(stmt.Cond, false)
		m.write([]byte(")"))
		m.requireSemicolon()
	case *js.WhileStmt:
		m.write([]byte("while("))
		m.minifyExpr(stmt.Cond, false)
		m.write([]byte(")"))
		m.minifyStmt(stmt.Body)
	case *js.ForStmt:
		m.write([]byte("for("))
		m.minifyExpr(stmt.Init, false)
		m.write([]byte(";"))
		m.minifyExpr(stmt.Cond, false)
		m.write([]byte(";"))
		m.minifyExpr(stmt.Post, false)
		m.write([]byte(")"))
		m.minifyStmt(stmt.Body)
	case *js.ForInStmt:
		m.write([]byte("for("))
		m.minifyExpr(stmt.Init, false)
		m.writeSpaceAfterIdent()
		m.write([]byte("in"))
		m.writeSpaceBeforeIdent()
		m.minifyExpr(stmt.Value, false)
		m.write([]byte(")"))
		m.minifyStmt(stmt.Body)
	case *js.ForOfStmt:
		if stmt.Await {
			m.write([]byte("for await("))
		} else {
			m.write([]byte("for("))
		}
		m.minifyExpr(stmt.Init, false)
		m.writeSpaceAfterIdent()
		m.write([]byte("of"))
		m.writeSpaceBeforeIdent()
		m.minifyExpr(stmt.Value, false)
		m.write([]byte(")"))
		m.minifyStmt(stmt.Body)
	case *js.SwitchStmt:
		m.write([]byte("switch("))
		m.minifyExpr(stmt.Init, false)
		m.write([]byte("){"))
		m.needsSemicolon = false
		for _, clause := range stmt.List {
			m.writeSemicolon()
			m.write(clause.TokenType.Bytes())
			if clause.Cond != nil {
				m.write([]byte(" "))
				m.minifyExpr(clause.Cond, false)
			}
			m.write([]byte(":"))
			for _, item := range clause.List {
				m.writeSemicolon()
				m.minifyStmt(item)
			}
		}
		m.write([]byte("}"))
	case *js.ThrowStmt:
		m.write([]byte("throw"))
		m.writeSpaceBeforeIdent()
		m.minifyExpr(stmt.Value, false)
		m.requireSemicolon()
	case *js.TryStmt:
		m.write([]byte("try"))
		m.minifyBlockStmt(stmt.Body)
		if len(stmt.Catch.List) != 0 || stmt.Binding != nil {
			m.write([]byte("catch"))
			if stmt.Binding != nil {
				m.write([]byte("("))
				m.minifyBinding(stmt.Binding)
				m.write([]byte(")"))
			}
			m.minifyBlockStmt(stmt.Catch)
		}
		if len(stmt.Finally.List) != 0 {
			m.write([]byte("finally"))
			m.minifyBlockStmt(stmt.Finally)
		}
	case *js.FuncDecl:
		m.minifyFuncDecl(*stmt)
	case *js.ClassDecl:
		m.minifyClassDecl(*stmt)
	case *js.DebuggerStmt:
		m.write([]byte("debugger"))
		m.requireSemicolon()
	case *js.EmptyStmt:
	case *js.ImportStmt:
		m.write([]byte("import"))
		if stmt.Default != nil {
			m.write([]byte(" "))
			m.write(stmt.Default)
			if len(stmt.List) != 0 {
				m.write([]byte(","))
			}
		}
		if len(stmt.List) == 1 {
			m.writeSpaceBeforeIdent()
			m.minifyAlias(stmt.List[0])
		} else if 1 < len(stmt.List) {
			m.write([]byte("{"))
			for i, item := range stmt.List {
				if i != 0 {
					m.write([]byte(","))
				}
				m.minifyAlias(item)
			}
			m.write([]byte("}"))
		}
		if stmt.Default != nil || len(stmt.List) != 0 {
			if len(stmt.List) < 2 {
				m.write([]byte(" "))
			}
			m.write([]byte("from"))
		}
		m.write(stmt.Module)
		m.requireSemicolon()
	case *js.ExportStmt:
		m.write([]byte("export"))
		if stmt.Decl != nil {
			if stmt.Default {
				m.write([]byte(" default "))
			} else {
				m.write([]byte(" "))
			}
			m.minifyExpr(stmt.Decl, false)
			_, isHoistable := stmt.Decl.(*js.FuncDecl)
			_, isClass := stmt.Decl.(*js.ClassDecl)
			if !isHoistable && !isClass {
				m.requireSemicolon()
			}
		} else {
			if len(stmt.List) == 1 {
				m.writeSpaceBeforeIdent()
				m.minifyAlias(stmt.List[0])
			} else if 1 < len(stmt.List) {
				m.write([]byte("{"))
				for i, item := range stmt.List {
					if i != 0 {
						m.write([]byte(","))
					}
					m.minifyAlias(item)
				}
				m.write([]byte("}"))
			}
			if stmt.Module != nil {
				if len(stmt.List) < 2 && (len(stmt.List) != 1 || isIdentEndAlias(stmt.List[0])) {
					m.write([]byte(" "))
				}
				m.write([]byte("from"))
				m.write(stmt.Module)
			}
			m.requireSemicolon()
		}
	}
}

func (m *jsMinifier) minifyAlias(alias js.Alias) {
	if alias.Name != nil {
		m.write(alias.Name)
		if !bytes.Equal(alias.Name, starBytes) {
			m.write([]byte(" "))
		}
		m.write([]byte("as "))
	}
	m.write(alias.Binding)
}

func (m *jsMinifier) minifyBlockStmt(stmt js.BlockStmt) {
	m.write([]byte("{"))
	m.needsSemicolon = false
	for _, item := range stmt.List {
		m.writeSemicolon()
		m.minifyStmt(item)
	}
	m.write([]byte("}"))
	m.needsSemicolon = false
}

func (m *jsMinifier) minifyParams(params js.Params) {
	m.write([]byte("("))
	for i, item := range params.List {
		if i != 0 {
			m.write([]byte(","))
		}
		m.minifyBindingElement(item)
	}
	if params.Rest != nil {
		if len(params.List) != 0 {
			m.write([]byte(","))
		}
		m.write([]byte("..."))
		m.minifyBindingElement(*params.Rest)
	}
	m.write([]byte(")"))
}

func (m *jsMinifier) minifyArguments(args js.Arguments) {
	m.write([]byte("("))
	for i, item := range args.List {
		if i != 0 {
			m.write([]byte(","))
		}
		m.minifyExpr(item, false)
	}
	if args.Rest != nil {
		if len(args.List) != 0 {
			m.write([]byte(","))
		}
		m.write([]byte("..."))
		m.minifyExpr(args.Rest, false)
	}
	m.write([]byte(")"))
}

func (m *jsMinifier) minifyVarDecl(decl js.VarDecl) {
	m.write(decl.TokenType.Bytes())
	m.write([]byte(" "))
	for i, item := range decl.List {
		if i != 0 {
			m.write([]byte(","))
		}
		m.minifyBindingElement(item)
	}
}

func (m *jsMinifier) minifyFuncDecl(decl js.FuncDecl) {
	if decl.Async {
		m.write([]byte("async"))
	}
	m.write([]byte("function"))
	if decl.Generator {
		m.write([]byte("*"))
	}
	if decl.Name != nil {
		if !decl.Generator {
			m.write([]byte(" "))
		}
		m.write(m.renamer.add(decl.Name))
	}
	m.renamer.openScope()
	m.minifyParams(decl.Params)
	m.minifyBlockStmt(decl.Body)
	m.renamer.closeScope()
}

func (m *jsMinifier) minifyMethodDecl(decl js.MethodDecl) {
	if decl.Static {
		m.write([]byte("static "))
	}
	if decl.Async {
		m.write([]byte("async"))
		if decl.Generator {
			m.write([]byte("*"))
		}
	} else if decl.Generator {
		m.write([]byte("*"))
	} else if decl.Get {
		m.write([]byte("get "))
	} else if decl.Set {
		m.write([]byte("set "))
	}
	m.minifyPropertyName(decl.Name)
	m.minifyParams(decl.Params)
	m.minifyBlockStmt(decl.Body)
}

func (m *jsMinifier) minifyArrowFuncDecl(decl js.ArrowFuncDecl) {
	if decl.Async {
		m.write([]byte("async"))
	}
	if decl.Params.Rest == nil && len(decl.Params.List) == 1 && decl.Params.List[0].Default == nil {
		if decl.Async && isIdentStartBindingElement(decl.Params.List[0]) {
			m.write([]byte(" "))
		}
		m.minifyBindingElement(decl.Params.List[0])
	} else {
		m.minifyParams(decl.Params)
	}
	m.write([]byte("=>"))
	if len(decl.Body.List) == 1 {
		if stmt, ok := decl.Body.List[0].(*js.ExprStmt); ok {
			m.minifyExpr(stmt.Value, false)
		} else {
			m.minifyBlockStmt(decl.Body)
		}
	} else {
		m.minifyBlockStmt(decl.Body)
	}
}

func (m *jsMinifier) minifyClassDecl(decl js.ClassDecl) {
	m.write([]byte("class"))
	if decl.Name != nil {
		m.write([]byte(" "))
		m.write(decl.Name)
	}
	if decl.Extends != nil {
		m.write([]byte(" extends"))
		m.writeSpaceBeforeIdent()
		m.minifyExpr(decl.Extends, true)
	}
	m.write([]byte("{"))
	for _, item := range decl.Methods {
		m.minifyMethodDecl(item)
	}
	m.write([]byte("}"))
}

func (m *jsMinifier) minifyPropertyName(name js.PropertyName) {
	if name.Computed != nil {
		m.minifyExpr(name.Computed, false)
	} else {
		m.write(name.Literal.Data)
	}
}

func (m *jsMinifier) minifyProperty(property js.Property) {
	if property.Key != nil {
		m.minifyPropertyName(*property.Key)
		m.write([]byte(":"))
	} else if property.Spread {
		m.write([]byte("..."))
	}
	m.minifyExpr(property.Value, false)
	if property.Init != nil {
		m.write([]byte("="))
		m.minifyExpr(property.Init, false)
	}
}

func (m *jsMinifier) minifyBindingElement(element js.BindingElement) {
	if element.Binding != nil {
		m.minifyBinding(element.Binding)
		if element.Default != nil {
			m.write([]byte("="))
			m.minifyExpr(element.Default, false)
		}
	}
}

func (m *jsMinifier) minifyBinding(i js.IBinding) {
	switch binding := i.(type) {
	case *js.BindingName:
		m.write(m.renamer.add(binding.Data))
	case *js.BindingArray:
		m.write([]byte("["))
		for _, item := range binding.List {
			m.minifyBindingElement(item)
		}
		if binding.Rest != nil {
			m.write([]byte("..."))
			m.minifyBinding(binding.Rest)
		}
		m.write([]byte("]"))
	case *js.BindingObject:
		m.write([]byte("{"))
		for _, item := range binding.List {
			if item.Key != nil {
				m.minifyPropertyName(*item.Key)
				m.write([]byte(":"))
			}
			m.minifyBindingElement(item.Value)
		}
		if binding.Rest != nil {
			m.write([]byte("..."))
			m.write(binding.Rest.Data)
		}
		m.write([]byte("}"))
	}
}

func (m *jsMinifier) minifyExpr(i js.IExpr, precGtEqPrefix bool) {
	switch expr := i.(type) {
	case *js.LiteralExpr:
		if expr.TokenType == js.DecimalToken {
			m.write(minify.Number(expr.Data, 0))
		} else if expr.TokenType == js.TrueToken {
			if precGtEqPrefix {
				m.write([]byte("(!0)"))
			} else {
				m.write([]byte("!0"))
			}
		} else if expr.TokenType == js.FalseToken {
			if precGtEqPrefix {
				m.write([]byte("(!1)"))
			} else {
				m.write([]byte("!1"))
			}
		} else if expr.TokenType == js.IdentifierToken && bytes.Equal(expr.Data, []byte("undefined")) && !m.renamer.exists(expr.Data) {
			if precGtEqPrefix {
				m.write([]byte("(void 0)"))
			} else {
				m.write([]byte("void 0"))
			}
		} else {
			m.write(m.renamer.name(expr.Data))
		}
	case *js.BinaryExpr:
		m.minifyExpr(expr.X, false)
		if expr.Op == js.InstanceofToken || expr.Op == js.InToken {
			m.writeSpaceAfterIdent()
			m.write(expr.Op.Bytes())
			m.writeSpaceBeforeIdent()
		} else {
			if expr.Op == js.GtToken {
				if unary, ok := expr.X.(*js.UnaryExpr); ok && unary.Op == js.PostDecrToken {
					m.write([]byte(" "))
				}
			}
			m.write(expr.Op.Bytes())
			if expr.Op == js.AddToken {
				if unary, ok := expr.Y.(*js.UnaryExpr); ok && (unary.Op == js.PosToken || unary.Op == js.PreIncrToken) {
					m.write([]byte(" "))
				}
			} else if expr.Op == js.NegToken {
				if unary, ok := expr.X.(*js.UnaryExpr); ok && (unary.Op == js.NegToken || unary.Op == js.PreDecrToken) {
					m.write([]byte(" "))
				}
			} else if expr.Op == js.LtToken {
				if unary, ok := expr.Y.(*js.UnaryExpr); ok && unary.Op == js.NotToken {
					if unary2, ok2 := unary.X.(*js.UnaryExpr); ok2 && unary2.Op == js.PreDecrToken {
						m.write([]byte(" "))
					}
				}
			}
		}
		m.minifyExpr(expr.Y, false)
	case *js.UnaryExpr:
		if expr.Op == js.PostIncrToken || expr.Op == js.PostDecrToken {
			m.minifyExpr(expr.X, false)
			m.write(expr.Op.Bytes())
		} else {
			m.write(expr.Op.Bytes())
			if expr.Op == js.DeleteToken || expr.Op == js.VoidToken || expr.Op == js.TypeofToken {
				m.writeSpaceBeforeIdent()
			} else if expr.Op == js.PosToken {
				if unary, ok := expr.X.(*js.UnaryExpr); ok && (unary.Op == js.PosToken || unary.Op == js.PreIncrToken) {
					m.write([]byte(" "))
				}
			} else if expr.Op == js.NegToken {
				if unary, ok := expr.X.(*js.UnaryExpr); ok && (unary.Op == js.NegToken || unary.Op == js.PreDecrToken) {
					m.write([]byte(" "))
				}
			}
			m.minifyExpr(expr.X, false)
		}
	case *js.DotExpr:
		m.minifyExpr(expr.X, false)
		m.write([]byte("."))
		m.write(expr.Y.Data)
	case *js.GroupExpr:
		m.write([]byte("("))
		for i, item := range expr.List {
			if i != 0 {
				m.write([]byte(","))
			}
			m.minifyExpr(item, false)
		}
		if expr.Rest != nil {
			if len(expr.List) != 0 {
				m.write([]byte(","))
			}
			m.write([]byte("..."))
			m.minifyBinding(expr.Rest)
		}
		m.write([]byte(")"))
	case *js.ArrayExpr:
		m.write([]byte("["))
		for i, item := range expr.List {
			if i != 0 {
				m.write([]byte(","))
			}
			m.minifyExpr(item, false)
		}
		if expr.Rest != nil {
			if len(expr.List) != 0 {
				m.write([]byte(","))
			}
			m.write([]byte("..."))
			m.minifyExpr(expr.Rest, false)
		}
		m.write([]byte("]"))
	case *js.ObjectExpr:
		m.write([]byte("{"))
		for i, item := range expr.List {
			if i != 0 {
				m.write([]byte(","))
			}
			m.minifyProperty(item)
		}
		m.write([]byte("}"))
	case *js.TemplateExpr:
		if expr.Tag != nil {
			m.minifyExpr(expr.Tag, false)
		}
		for _, item := range expr.List {
			m.write(item.Value)
			m.minifyExpr(item.Expr, false)
		}
		m.write(expr.Tail)
	case *js.NewExpr:
		m.write([]byte("new"))
		m.writeSpaceBeforeIdent()
		m.minifyExpr(expr.X, true)
	case *js.NewTargetExpr:
		m.write([]byte("new.target"))
		m.writeSpaceBeforeIdent()
	case *js.YieldExpr:
		m.write([]byte("yield"))
		m.writeSpaceBeforeIdent()
		if expr.X != nil {
			if expr.Generator {
				m.write([]byte("*"))
			}
			m.minifyExpr(expr.X, false)
		}
	case *js.CallExpr:
		m.minifyExpr(expr.X, true)
		m.minifyArguments(expr.Args)
	case *js.IndexExpr:
		m.minifyExpr(expr.X, false)
		m.write([]byte("["))
		m.minifyExpr(expr.Index, false)
		m.write([]byte("]"))
	case *js.ConditionalExpr:
		m.minifyExpr(expr.X, false)
		m.write([]byte("?"))
		m.minifyExpr(expr.Y, false)
		m.write([]byte(":"))
		m.minifyExpr(expr.Z, false)
	case *js.OptChainExpr:
		m.minifyExpr(expr.X, false)
		m.write([]byte("?."))
		m.minifyExpr(expr.Y, false)
	case *js.VarDecl:
		m.minifyVarDecl(*expr)
	case *js.FuncDecl:
		m.minifyFuncDecl(*expr)
	case *js.ArrowFuncDecl:
		m.minifyArrowFuncDecl(*expr)
	case *js.MethodDecl:
		m.minifyMethodDecl(*expr)
	case *js.ClassDecl:
		m.minifyClassDecl(*expr)
	}
}

func isIdentStartBindingElement(element js.BindingElement) bool {
	if element.Binding != nil {
		if _, ok := element.Binding.(*js.BindingName); ok {
			return true
		}
	}
	return false
}

func isIdentEndAlias(alias js.Alias) bool {
	return !bytes.Equal(alias.Binding, starBytes)
}

func isEmptyStmt(stmt js.IStmt) bool {
	if stmt == nil {
		return true
	} else if _, ok := stmt.(*js.EmptyStmt); ok {
		return true
	} else if block, ok := stmt.(*js.BlockStmt); ok {
		for _, item := range block.List {
			if ok := isEmptyStmt(item); !ok {
				return false
			}
		}
		return true
	}
	return false
}

type renamer struct {
	unbound  map[string]bool
	reserved map[string]bool
	renames  [][]byte         // list of renames, only ever grows as we find new variables that do not collide
	idx      int              // index into renames, decreases when we leave scope
	scopes   []map[string]int // index into renames, ordered from outer scope to inner
}

func newRenamer(unbound []string) *renamer {
	unboundMap := map[string]bool{}
	for _, name := range unbound {
		unboundMap[name] = true
	}
	reserved := map[string]bool{}
	for name, _ := range js.Keywords {
		reserved[name] = true
	}
	for _, name := range unbound {
		reserved[name] = true
	}
	return &renamer{
		unbound:  unboundMap,
		reserved: reserved,
	}
}

func (r *renamer) next(name []byte) []byte {
	if name[len(name)-1] == 'z' {
		name[len(name)-1] = 'A'
	} else if name[len(name)-1] == 'Z' {
		name = append(name, 'a')
	} else {
		name[len(name)-1]++
	}
	return name
}

func (r *renamer) add(src []byte) []byte {
	if len(r.scopes) == 0 {
		r.reserved[string(src)] = true // top-level variables
		return src
	} else if r.idx < len(r.renames) {
		dst := r.renames[r.idx]
		r.scopes[len(r.scopes)-1][string(src)] = r.idx
		r.idx++
		return dst
	}
	var dst []byte
	if len(r.renames) == 0 {
		dst = []byte("a")
	} else {
		dst = parse.Copy(r.renames[len(r.renames)-1])
		dst = r.next(dst)
	}
	for r.reserved[string(dst)] {
		dst = r.next(dst)
	}
	r.renames = append(r.renames, dst)
	r.scopes[len(r.scopes)-1][string(src)] = r.idx
	r.idx++
	return dst
}

func (r *renamer) name(name []byte) []byte {
	for j := len(r.scopes) - 1; 0 <= j; j-- {
		if i, ok := r.scopes[j][string(name)]; ok {
			return r.renames[i]
		}
	}
	return name
}

func (r *renamer) exists(name []byte) bool {
	if r.unbound[string(name)] {
		return false
	}
	if r.reserved[string(name)] {
		return true
	}
	for j := len(r.scopes) - 1; 0 <= j; j-- {
		if _, ok := r.scopes[j][string(name)]; ok {
			return true
		}
	}
	return false
}

func (r *renamer) openScope() {
	r.scopes = append(r.scopes, map[string]int{})
}

func (r *renamer) closeScope() {
	last := len(r.scopes) - 1
	r.idx -= len(r.scopes[last])
	r.scopes = r.scopes[:last]
}
