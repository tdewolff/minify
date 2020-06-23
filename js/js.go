// Package js minifies ECMAScript5.1 following the specifications at http://www.ecma-international.org/ecma-262/5.1/.
package js

// TODO: remove dead code, such as in if (false) or statements after return statement, difficulty with var decls
// TODO: move var declaration or expr statement into for loop init (var only if for has var decl)
// TODO: don't minify variable names in with statement, what todo with eval? Don't minify any variable name?

import (
	"bytes"
	"io"
	"sort"

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
type Minifier struct {
	Precision     int // number of significant digits
	Deterministic bool
}

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
		renamer: newRenamer(ast.Scope, o.Deterministic),
	}
	ast.List = m.mergeStmtList(ast.List)
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
	renamer        *renamer2
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
	i = m.stmtToExpr(i)

	switch stmt := i.(type) {
	case *js.ExprStmt:
		// prefix ! to function or group to class to remain expressions
		expr := stmt.Value
		commaExpr, ok := expr.(*js.BinaryExpr)
		for ok && commaExpr.Op == js.CommaToken {
			expr = commaExpr.X
			commaExpr, ok = expr.(*js.BinaryExpr)
		}
		if group, isGroup := expr.(*js.GroupExpr); isGroup {
			if _, isFunc := group.X.(*js.FuncDecl); isFunc {
				m.write([]byte("!"))
			} else if _, isClass := group.X.(*js.ClassDecl); isClass {
				m.write([]byte("!"))
			} else if call, isCall := group.X.(*js.CallExpr); isCall {
				if _, isFunc := call.X.(*js.FuncDecl); isFunc {
					m.write([]byte("!"))
				}
			}
		} else if call, isCall := expr.(*js.CallExpr); isCall {
			if group, isGroup := call.X.(*js.GroupExpr); isGroup {
				if _, isFunc := group.X.(*js.FuncDecl); isFunc {
					m.write([]byte("!"))
				}
			}
		}
		m.minifyExpr(stmt.Value, js.OpExpr)
		m.requireSemicolon()
	case *js.VarDecl:
		m.minifyVarDecl(*stmt)
		m.requireSemicolon()
	case *js.IfStmt:
		hasIf := !isEmptyStmt(stmt.Body)
		hasElse := !isEmptyStmt(stmt.Else)

		m.write([]byte("if("))
		m.minifyExpr(stmt.Cond, js.OpExpr)
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
			if !hasReturnThrowStmt(stmt.Body) {
				m.write([]byte("else"))
				m.writeSpaceBeforeIdent()
				m.minifyStmt(stmt.Else)
			} else if block, ok := stmt.Else.(*js.BlockStmt); ok {
				for _, item := range block.List {
					m.writeSemicolon()
					m.minifyStmt(item)
				}
			} else {
				m.minifyStmt(stmt.Else)
			}
		}
	case *js.BlockStmt:
		m.minifyBlockStmt(*stmt, true)
	case *js.ReturnStmt:
		m.write([]byte("return"))
		m.writeSpaceBeforeIdent()
		m.minifyExpr(stmt.Value, js.OpExpr)
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
		m.minifyExpr(stmt.Cond, js.OpExpr)
		m.write([]byte(")"))
		m.minifyStmt(stmt.Body)
	case *js.DoWhileStmt:
		m.write([]byte("do"))
		m.writeSpaceBeforeIdent()
		m.minifyStmt(stmt.Body)
		m.writeSemicolon()
		m.write([]byte("while("))
		m.minifyExpr(stmt.Cond, js.OpExpr)
		m.write([]byte(")"))
		m.requireSemicolon()
	case *js.WhileStmt:
		m.write([]byte("while("))
		m.minifyExpr(stmt.Cond, js.OpExpr)
		m.write([]byte(")"))
		m.minifyStmt(stmt.Body)
	case *js.ForStmt:
		m.write([]byte("for("))
		m.minifyExpr(stmt.Init, js.OpExpr)
		m.write([]byte(";"))
		m.minifyExpr(stmt.Cond, js.OpExpr)
		m.write([]byte(";"))
		m.minifyExpr(stmt.Post, js.OpExpr)
		m.write([]byte(")"))
		m.minifyStmt(stmt.Body)
	case *js.ForInStmt:
		m.write([]byte("for("))
		m.minifyExpr(stmt.Init, js.OpLHS)
		m.writeSpaceAfterIdent()
		m.write([]byte("in"))
		m.writeSpaceBeforeIdent()
		m.minifyExpr(stmt.Value, js.OpExpr)
		m.write([]byte(")"))
		m.minifyStmt(stmt.Body)
	case *js.ForOfStmt:
		if stmt.Await {
			m.write([]byte("for await("))
		} else {
			m.write([]byte("for("))
		}
		m.minifyExpr(stmt.Init, js.OpLHS)
		m.writeSpaceAfterIdent()
		m.write([]byte("of"))
		m.writeSpaceBeforeIdent()
		m.minifyExpr(stmt.Value, js.OpAssign)
		m.write([]byte(")"))
		m.minifyStmt(stmt.Body)
	case *js.SwitchStmt:
		m.write([]byte("switch("))
		m.minifyExpr(stmt.Init, js.OpExpr)
		m.write([]byte("){"))
		m.needsSemicolon = false
		for _, clause := range stmt.List {
			m.writeSemicolon()
			m.write(clause.TokenType.Bytes())
			if clause.Cond != nil {
				m.write([]byte(" "))
				m.minifyExpr(clause.Cond, js.OpExpr)
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
		m.minifyExpr(stmt.Value, js.OpExpr)
		m.requireSemicolon()
	case *js.TryStmt:
		m.write([]byte("try"))
		m.minifyBlockStmt(stmt.Body, true)
		if len(stmt.Catch.List) != 0 || stmt.Binding != nil {
			m.write([]byte("catch"))
			m.renamer.enterScope(stmt.Catch.Scope, true)
			if stmt.Binding != nil {
				m.write([]byte("("))
				m.minifyBinding(stmt.Binding)
				m.write([]byte(")"))
			}
			m.minifyBlockStmt(stmt.Catch, false)
			m.renamer.exitScope()
		}
		if len(stmt.Finally.List) != 0 {
			m.write([]byte("finally"))
			m.minifyBlockStmt(stmt.Finally, true)
		}
	case *js.FuncDecl:
		m.minifyFuncDecl(*stmt, false)
	case *js.ClassDecl:
		m.minifyClassDecl(*stmt)
	case *js.DebuggerStmt:
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
				m.write([]byte(" default"))
			}
			m.writeSpaceBeforeIdent()
			m.minifyExpr(stmt.Decl, js.OpAssign)
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

func (m *jsMinifier) stmtToExpr(i js.IStmt) js.IStmt {
	if stmt, ok := i.(*js.IfStmt); ok {
		if unaryExpr, ok := stmt.Cond.(*js.UnaryExpr); ok && unaryExpr.Op == js.NotToken {
			stmt.Cond = unaryExpr.X
			stmt.Body, stmt.Else = stmt.Else, stmt.Body
		}
		hasIf := !isEmptyStmt(stmt.Body)
		hasElse := !isEmptyStmt(stmt.Else)
		if !hasIf && !hasElse {
			return &js.ExprStmt{&js.GroupExpr{stmt.Cond}}
		} else if hasIf && !hasElse {
			stmt.Body = m.stmtToExpr(stmt.Body)
			if X, isExprBody := stmt.Body.(*js.ExprStmt); isExprBody {
				return &js.ExprStmt{&js.BinaryExpr{js.AndToken, &js.GroupExpr{stmt.Cond}, &js.GroupExpr{X.Value}}}
			}
		} else if !hasIf && hasElse {
			stmt.Else = m.stmtToExpr(stmt.Else)
			if X, isExprElse := stmt.Else.(*js.ExprStmt); isExprElse {
				return &js.ExprStmt{&js.BinaryExpr{js.OrToken, &js.GroupExpr{stmt.Cond}, &js.GroupExpr{X.Value}}}
			}
		} else if hasIf && hasElse {
			stmt.Body = m.stmtToExpr(stmt.Body)
			stmt.Else = m.stmtToExpr(stmt.Else)
			XExpr, isExprBody := stmt.Body.(*js.ExprStmt)
			YExpr, isExprElse := stmt.Else.(*js.ExprStmt)
			if isExprBody && isExprElse {
				return &js.ExprStmt{&js.CondExpr{&js.GroupExpr{stmt.Cond}, &js.GroupExpr{XExpr.Value}, &js.GroupExpr{YExpr.Value}}}
			}
			XReturn, isReturnBody := stmt.Body.(*js.ReturnStmt)
			YReturn, isReturnElse := stmt.Else.(*js.ReturnStmt)
			if isReturnBody && isReturnElse {
				if XReturn.Value == nil {
					XReturn.Value = &js.UnaryExpr{js.VoidToken, &js.LiteralExpr{js.NumericToken, []byte("0")}}
				}
				if YReturn.Value == nil {
					YReturn.Value = &js.UnaryExpr{js.VoidToken, &js.LiteralExpr{js.NumericToken, []byte("0")}}
				}
				return &js.ReturnStmt{&js.CondExpr{&js.GroupExpr{stmt.Cond}, &js.GroupExpr{XReturn.Value}, &js.GroupExpr{YReturn.Value}}}
			}
			XThrow, isThrowBody := stmt.Body.(*js.ThrowStmt)
			YThrow, isThrowElse := stmt.Else.(*js.ThrowStmt)
			if isThrowBody && isThrowElse {
				return &js.ThrowStmt{&js.CondExpr{&js.GroupExpr{stmt.Cond}, &js.GroupExpr{XThrow.Value}, &js.GroupExpr{YThrow.Value}}}
			}
		}
	} else if stmt, ok := i.(*js.BlockStmt); ok {
		// merge body and remove braces if possible from independent blocks
		stmt.List = m.mergeStmtList(stmt.List)
		if len(stmt.List) == 1 {
			varDecl, isVarDecl := stmt.List[0].(*js.VarDecl)
			_, isClassDecl := stmt.List[0].(*js.ClassDecl)
			if !isClassDecl && (!isVarDecl || varDecl.TokenType == js.VarToken) {
				return m.stmtToExpr(stmt.List[0])
			}
		}
		return js.IStmt(stmt)
	}
	return i
}

func (m *jsMinifier) mergeStmtList(list []js.IStmt) []js.IStmt {
	if len(list) < 2 {
		return list
	}
	list[0] = m.stmtToExpr(list[0])
	j := 0
	for i, _ := range list[:len(list)-1] {
		list[i+1] = m.stmtToExpr(list[i+1])
		j++
		if left, ok := list[i].(*js.ExprStmt); ok {
			// merge expression statements with expression, return, and throw statements
			if right, ok := list[i+1].(*js.ExprStmt); ok {
				right.Value = &js.BinaryExpr{js.CommaToken, left.Value, right.Value}
				j--
			} else if returnStmt, ok := list[i+1].(*js.ReturnStmt); ok && returnStmt.Value != nil {
				returnStmt.Value = &js.BinaryExpr{js.CommaToken, left.Value, returnStmt.Value}
				j--
			} else if throwStmt, ok := list[i+1].(*js.ThrowStmt); ok {
				throwStmt.Value = &js.BinaryExpr{js.CommaToken, left.Value, throwStmt.Value}
				j--
			}
		} else if left, ok := list[i].(*js.VarDecl); ok {
			// merge var, const, let declarations
			if right, ok := list[i+1].(*js.VarDecl); ok && left.TokenType == right.TokenType {
				right.List = append(left.List, right.List...)
				j--
			} else if left.TokenType == js.VarToken {
				if forStmt, ok := list[i+1].(*js.ForStmt); ok {
					if init, ok := forStmt.Init.(*js.VarDecl); ok && init.TokenType == js.VarToken {
						init.List = append(left.List, init.List...)
						j--
					}
				} else if whileStmt, ok := list[i+1].(*js.WhileStmt); ok {
					list[i+1] = &js.ForStmt{left, whileStmt.Cond, nil, whileStmt.Body}
					j--
				}
			}
		}
		list[j] = list[i+1]
		if 0 < j {
			// merge if/else with return/throw when followed by return/throw
			if ifStmt, ok := list[j-1].(*js.IfStmt); ok && isEmptyStmt(ifStmt.Body) != isEmptyStmt(ifStmt.Else) {
				if returnStmt, ok := list[j].(*js.ReturnStmt); ok && returnStmt.Value != nil {
					if left, ok := ifStmt.Body.(*js.ReturnStmt); ok && left.Value != nil {
						returnStmt.Value = &js.CondExpr{&js.GroupExpr{ifStmt.Cond}, &js.GroupExpr{left.Value}, &js.GroupExpr{returnStmt.Value}}
						list[j-1] = returnStmt
						j--
					} else if left, ok := ifStmt.Else.(*js.ReturnStmt); ok && left.Value != nil {
						returnStmt.Value = &js.CondExpr{&js.GroupExpr{ifStmt.Cond}, &js.GroupExpr{returnStmt.Value}, &js.GroupExpr{left.Value}}
						list[j-1] = returnStmt
						j--
					}
				} else if throwStmt, ok := list[j].(*js.ThrowStmt); ok {
					if left, ok := ifStmt.Body.(*js.ThrowStmt); ok {
						throwStmt.Value = &js.CondExpr{&js.GroupExpr{ifStmt.Cond}, &js.GroupExpr{left.Value}, &js.GroupExpr{throwStmt.Value}}
						list[j-1] = throwStmt
						j--
					} else if left, ok := ifStmt.Else.(*js.ThrowStmt); ok {
						throwStmt.Value = &js.CondExpr{&js.GroupExpr{ifStmt.Cond}, &js.GroupExpr{throwStmt.Value}, &js.GroupExpr{left.Value}}
						list[j-1] = throwStmt
						j--
					}
				}
			}
		}
	}
	return list[:j+1]
}

func (m *jsMinifier) minifyBlockStmt(stmt js.BlockStmt, enterScope bool) {
	stmt.List = m.mergeStmtList(stmt.List)
	m.write([]byte("{"))
	m.needsSemicolon = false
	if enterScope {
		m.renamer.enterScope(stmt.Scope, true)
	}
	for _, item := range stmt.List {
		m.writeSemicolon()
		m.minifyStmt(item)
	}
	if enterScope {
		m.renamer.exitScope()
	}
	m.write([]byte("}"))
	m.needsSemicolon = false
}

func (m *jsMinifier) minifyFuncBody(stmt js.BlockStmt) {
	stmt.List = m.mergeStmtList(stmt.List)
	m.write([]byte("{"))
	m.needsSemicolon = false
	for _, item := range stmt.List {
		if returnStmt, ok := item.(*js.ReturnStmt); ok {
			if returnStmt.Value != nil && !m.isUndefined(returnStmt.Value) {
				m.writeSemicolon()
				m.minifyStmt(item)
			}
			break
		} else {
			m.writeSemicolon()
			m.minifyStmt(item)
		}
	}
	m.write([]byte("}"))
	m.needsSemicolon = false
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
		m.minifyBinding(params.Rest)
	}
	m.write([]byte(")"))
}

func (m *jsMinifier) minifyArguments(args js.Arguments) {
	m.write([]byte("("))
	for i, item := range args.List {
		if i != 0 {
			m.write([]byte(","))
		}
		m.minifyExpr(item, js.OpExpr)
	}
	if args.Rest != nil {
		if len(args.List) != 0 {
			m.write([]byte(","))
		}
		m.write([]byte("..."))
		m.minifyExpr(args.Rest, js.OpExpr)
	}
	m.write([]byte(")"))
}

func (m *jsMinifier) minifyVarDecl(decl js.VarDecl) {
	m.write(decl.TokenType.Bytes())
	m.writeSpaceBeforeIdent()
	for i, item := range decl.List {
		if i != 0 {
			m.write([]byte(","))
		}
		m.minifyBindingElement(item)
	}
}

func (m *jsMinifier) minifyFuncDecl(decl js.FuncDecl, inExpr bool) {
	if decl.Async {
		m.write([]byte("async "))
	}
	m.write([]byte("function"))
	if decl.Generator {
		m.write([]byte("*"))
	}
	if inExpr {
		m.renamer.enterScope(decl.Scope, false)
	}
	if decl.Name != nil {
		if !decl.Generator {
			m.write([]byte(" "))
		}
		m.write(m.renamer.rename(decl.Name))
	}
	if !inExpr {
		m.renamer.enterScope(decl.Scope, false)
	}
	m.minifyParams(decl.Params)
	m.minifyFuncBody(decl.Body)
	m.renamer.exitScope()
}

func (m *jsMinifier) minifyMethodDecl(decl js.MethodDecl) {
	if decl.Static {
		m.write([]byte("static"))
		m.writeSpaceBeforeIdent()
	}
	if decl.Async {
		m.write([]byte("async"))
		if decl.Generator {
			m.write([]byte("*"))
		} else {
			m.writeSpaceBeforeIdent()
		}
	} else if decl.Generator {
		m.write([]byte("*"))
	} else if decl.Get {
		m.write([]byte("get"))
		m.writeSpaceBeforeIdent()
	} else if decl.Set {
		m.write([]byte("set"))
		m.writeSpaceBeforeIdent()
	}
	m.minifyPropertyName(decl.Name)
	m.renamer.enterScope(decl.Scope, false)
	m.minifyParams(decl.Params)
	m.minifyFuncBody(decl.Body)
	m.renamer.exitScope()
}

func (m *jsMinifier) minifyArrowFunc(decl js.ArrowFunc) {
	m.renamer.enterScope(decl.Scope, false)
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
	removeBraces := false
	if 0 < len(decl.Body.List) {
		returnStmt, isReturn := decl.Body.List[len(decl.Body.List)-1].(*js.ReturnStmt)
		if isReturn && returnStmt.Value != nil {
			// merge expression statements to final return statement, remove function body braces
			var list []js.IExpr
			removeBraces = true
			for _, item := range decl.Body.List[:len(decl.Body.List)-1] {
				if expr, isExpr := item.(*js.ExprStmt); isExpr {
					list = append(list, expr.Value)
				} else {
					removeBraces = false
					break
				}
			}
			if removeBraces {
				list = append(list, returnStmt.Value)
				expr := list[0]
				for _, right := range list[1:] {
					expr = &js.BinaryExpr{js.CommaToken, expr, right}
				}
				m.minifyExpr(expr, js.OpAssign)
			}
		} else if isReturn && returnStmt.Value == nil {
			// remove empty return
			decl.Body.List = decl.Body.List[:len(decl.Body.List)-1]
		}
	}
	if !removeBraces {
		m.minifyBlockStmt(decl.Body, false)
	}
	m.renamer.exitScope()
}

func (m *jsMinifier) minifyClassDecl(decl js.ClassDecl) {
	m.write([]byte("class"))
	if decl.Name != nil {
		m.write([]byte(" "))
		m.write(m.renamer.rename(decl.Name))
	}
	if decl.Extends != nil {
		m.write([]byte(" extends"))
		m.writeSpaceBeforeIdent()
		m.minifyExpr(decl.Extends, js.OpLHS)
	}
	m.write([]byte("{"))
	for _, item := range decl.Methods {
		m.minifyMethodDecl(item)
	}
	m.write([]byte("}"))
}

func (m *jsMinifier) minifyPropertyName(name js.PropertyName) {
	if name.Computed != nil {
		m.write([]byte("["))
		m.minifyExpr(name.Computed, js.OpAssign)
		m.write([]byte("]"))
	} else if name.Literal.TokenType == js.StringToken {
		if _, ok := js.IsIdentifierName(name.Literal.Data[1 : len(name.Literal.Data)-1]); ok {
			m.write(name.Literal.Data[1 : len(name.Literal.Data)-1])
		} else if _, ok := js.IsNumericLiteral(name.Literal.Data[1 : len(name.Literal.Data)-1]); ok {
			m.write(name.Literal.Data[1 : len(name.Literal.Data)-1])
		} else {
			m.write(name.Literal.Data)
		}
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
	} else if lit, ok := property.Value.(*js.LiteralExpr); ok && lit.TokenType == js.IdentifierToken && !m.renamer.inGlobalScope() {
		// add 'old-name:' before BindingName as the latter will be renamed
		m.write(lit.Data)
		m.write([]byte(":"))
	}
	m.minifyExpr(property.Value, js.OpAssign)
	if property.Init != nil {
		m.write([]byte("="))
		m.minifyExpr(property.Init, js.OpAssign)
	}
}

func (m *jsMinifier) minifyBindingElement(element js.BindingElement) {
	if element.Binding != nil {
		m.minifyBinding(element.Binding)
		if element.Default != nil {
			m.write([]byte("="))
			m.minifyExpr(element.Default, js.OpAssign)
		}
	}
}

func (m *jsMinifier) minifyBinding(i js.IBinding) {
	switch binding := i.(type) {
	case *js.BindingName:
		m.write(m.renamer.rename(binding.Data))
	case *js.BindingArray:
		m.write([]byte("["))
		for i, item := range binding.List {
			if i != 0 {
				m.write([]byte(","))
			}
			m.minifyBindingElement(item)
		}
		if binding.Rest != nil {
			if 0 < len(binding.List) {
				m.write([]byte(","))
			}
			m.write([]byte("..."))
			m.minifyBinding(binding.Rest)
		} else if 0 < len(binding.List) && binding.List[len(binding.List)-1].Binding == nil {
			m.write([]byte(","))
		}
		m.write([]byte("]"))
	case *js.BindingObject:
		m.write([]byte("{"))
		for i, item := range binding.List {
			if i != 0 {
				m.write([]byte(","))
			}
			if item.Key != nil {
				m.minifyPropertyName(*item.Key)
				m.write([]byte(":"))
			} else if name, ok := item.Value.Binding.(*js.BindingName); ok && !bytes.Equal(name.Data, m.renamer.rename(name.Data)) {
				// add 'old-name:' before BindingName as the latter will be renamed
				m.write(name.Data)
				m.write([]byte(":"))
			}
			m.minifyBindingElement(item.Value)
		}
		if binding.Rest != nil {
			if 0 < len(binding.List) {
				m.write([]byte(","))
			}
			m.write([]byte("..."))
			m.write(binding.Rest.Data)
		}
		m.write([]byte("}"))
	}
}

func (m *jsMinifier) minifyExpr(i js.IExpr, prec js.OpPrec) {
	switch expr := i.(type) {
	case *js.LiteralExpr:
		if expr.TokenType == js.DecimalToken {
			m.write(minify.Number(expr.Data, 0))
		} else if expr.TokenType == js.TrueToken {
			if js.OpUnary < prec {
				m.write([]byte("(!0)"))
			} else {
				m.write([]byte("!0"))
			}
		} else if expr.TokenType == js.FalseToken {
			if js.OpUnary < prec {
				m.write([]byte("(!1)"))
			} else {
				m.write([]byte("!1"))
			}
		} else if expr.TokenType == js.IdentifierToken && bytes.Equal(expr.Data, []byte("undefined")) && !m.renamer.exists(expr.Data) {
			if js.OpUnary < prec {
				m.write([]byte("(void 0)"))
			} else {
				m.write([]byte("void 0"))
			}
		} else if expr.TokenType == js.IdentifierToken {
			m.write(m.renamer.rename(expr.Data))
		} else if expr.TokenType == js.StringToken {
			m.write(minifyString(expr.Data))
		} else {
			m.write(expr.Data)
		}
	case *js.BinaryExpr:
		m.minifyExpr(expr.X, binaryLeftPrecMap[expr.Op])
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
		m.minifyExpr(expr.Y, binaryRightPrecMap[expr.Op])
	case *js.UnaryExpr:
		if expr.Op == js.PostIncrToken || expr.Op == js.PostDecrToken {
			m.minifyExpr(expr.X, unaryPrecMap[expr.Op])
			m.write(expr.Op.Bytes())
		} else {
			m.write(expr.Op.Bytes())
			if expr.Op == js.DeleteToken || expr.Op == js.VoidToken || expr.Op == js.TypeofToken || expr.Op == js.AwaitToken {
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
			m.minifyExpr(expr.X, unaryPrecMap[expr.Op])
		}
	case *js.DotExpr:
		m.minifyExpr(expr.X, js.OpMember)
		m.write([]byte("."))
		m.write(expr.Y.Data)
	case *js.GroupExpr:
		precInside := exprPrec(expr.X)
		if prec <= precInside {
			m.minifyExpr(expr.X, prec)
		} else {
			m.write([]byte("("))
			m.minifyExpr(expr.X, js.OpExpr)
			m.write([]byte(")"))
		}
	case *js.ArrayExpr:
		m.write([]byte("["))
		for i, item := range expr.List {
			if i != 0 {
				m.write([]byte(","))
			}
			if item.Spread {
				m.write([]byte("..."))
			}
			m.minifyExpr(item.Value, js.OpAssign)
		}
		if 0 < len(expr.List) && expr.List[len(expr.List)-1].Value == nil {
			m.write([]byte(","))
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
			m.minifyExpr(expr.Tag, js.OpLHS)
		}
		for _, item := range expr.List {
			m.write(item.Value)
			m.minifyExpr(item.Expr, js.OpExpr)
		}
		m.write(expr.Tail)
	case *js.NewExpr:
		m.write([]byte("new"))
		m.writeSpaceBeforeIdent()
		m.minifyExpr(expr.X, js.OpMember)
		if expr.Args != nil {
			m.minifyArguments(*expr.Args)
		}
	case *js.NewTargetExpr:
		m.write([]byte("new.target"))
		m.writeSpaceBeforeIdent()
	case *js.ImportMetaExpr:
		m.write([]byte("import.meta"))
		m.writeSpaceBeforeIdent()
	case *js.YieldExpr:
		m.write([]byte("yield"))
		m.writeSpaceBeforeIdent()
		if expr.X != nil {
			if expr.Generator {
				m.write([]byte("*"))
			}
			m.minifyExpr(expr.X, js.OpAssign)
		}
	case *js.CallExpr:
		m.minifyExpr(expr.X, js.OpMember)
		m.minifyArguments(expr.Args)
	case *js.IndexExpr:
		m.minifyExpr(expr.X, js.OpMember)
		if lit, ok := expr.Index.(*js.LiteralExpr); ok && lit.TokenType == js.StringToken {
			if _, ok := js.IsIdentifierName(lit.Data[1 : len(lit.Data)-1]); ok {
				m.write([]byte("."))
				m.write(lit.Data[1 : len(lit.Data)-1])
				break
			} else if tt, ok := js.IsNumericLiteral(lit.Data[1 : len(lit.Data)-1]); ok {
				lit.TokenType = tt
				lit.Data = lit.Data[1 : len(lit.Data)-1]
			}
		}
		m.write([]byte("["))
		m.minifyExpr(expr.Index, js.OpExpr)
		m.write([]byte("]"))
	case *js.CondExpr:
		// remove double negative !! in condition, or switch cases for single negative !
		if unary1, ok := expr.Cond.(*js.UnaryExpr); ok && unary1.Op == js.NotToken {
			if unary2, ok := unary1.X.(*js.UnaryExpr); ok && unary2.Op == js.NotToken {
				if isBooleanExpr(unary2.X) {
					expr.Cond = unary2.X
				}
			} else {
				expr.Cond = unary1.X
				expr.X, expr.Y = expr.Y, expr.X
			}
		}
		// if value is truthy or falsy, remove false case
		// if condition and true case are equal, or true and false case, simplify
		if truthy, ok := m.isTruthy(expr.Cond); truthy && ok {
			m.minifyExpr(expr.X, prec)
		} else if !truthy && ok {
			m.minifyExpr(expr.Y, prec)
		} else if isEqualExpr(expr.Cond, expr.X) && prec <= js.OpOr && (exprPrec(expr.X) < js.OpAssign || binaryLeftPrecMap[js.OrToken] <= exprPrec(expr.X)) && (exprPrec(expr.Y) < js.OpAssign || binaryRightPrecMap[js.OrToken] <= exprPrec(expr.Y)) {
			// for higher prec we need to add group parenthesis, and for lower prec we have parenthesis anyways. This only is shorter if len(expr.X) >= 3. isEqualExpr only checks for literal variables, which is a name will be minified to a one or two character name.
			m.minifyExpr(expr.X, binaryLeftPrecMap[js.OrToken])
			m.write([]byte("||"))
			m.minifyExpr(expr.Y, binaryRightPrecMap[js.OrToken])
		} else if isEqualExpr(expr.X, expr.Y) {
			if prec <= js.OpExpr {
				m.minifyExpr(expr.Cond, binaryLeftPrecMap[js.CommaToken])
				m.write([]byte(","))
				m.minifyExpr(expr.X, binaryRightPrecMap[js.CommaToken])
			} else {
				m.write([]byte("("))
				m.minifyExpr(expr.Cond, binaryLeftPrecMap[js.CommaToken])
				m.write([]byte(","))
				m.minifyExpr(expr.X, binaryRightPrecMap[js.CommaToken])
				m.write([]byte(")"))
			}
		} else {
			// shorten if cases are true and false
			trueX, falseX := isTrue(expr.X), isFalse(expr.X)
			trueY, falseY := isTrue(expr.Y), isFalse(expr.Y)
			if trueX && falseY || falseX && trueY {
				m.minifyBooleanExpr(expr.Cond, falseX, prec)
			} else if trueX || trueY {
				// trueX != trueY
				m.minifyBooleanExpr(expr.Cond, trueY, binaryLeftPrecMap[js.OrToken])
				m.write([]byte("||"))
				if trueY {
					m.minifyExpr(expr.X, binaryRightPrecMap[js.OrToken])
				} else {
					m.minifyExpr(expr.Y, binaryRightPrecMap[js.OrToken])
				}
			} else if falseX || falseY {
				// falseX != falseY
				m.minifyBooleanExpr(expr.Cond, falseX, binaryLeftPrecMap[js.AndToken])
				m.write([]byte("&&"))
				if falseX {
					m.minifyExpr(expr.Y, binaryRightPrecMap[js.AndToken])
				} else {
					m.minifyExpr(expr.X, binaryRightPrecMap[js.AndToken])
				}
			} else {
				// regular conditional expression
				m.minifyExpr(expr.Cond, js.OpCoalesce)
				m.write([]byte("?"))
				m.minifyExpr(expr.X, js.OpAssign)
				m.write([]byte(":"))
				m.minifyExpr(expr.Y, js.OpAssign)
			}
		}
	case *js.OptChainExpr:
		m.minifyExpr(expr.X, js.OpLHS)
		m.write([]byte("?."))
		m.minifyExpr(expr.Y, js.OpMember)
	case *js.VarDecl:
		m.minifyVarDecl(*expr) // only happens for init in for statement
	case *js.FuncDecl:
		m.minifyFuncDecl(*expr, true)
	case *js.ArrowFunc:
		m.minifyArrowFunc(*expr)
	case *js.MethodDecl:
		m.minifyMethodDecl(*expr)
	case *js.ClassDecl:
		m.minifyClassDecl(*expr)
	}
}

func (m *jsMinifier) minifyBooleanExpr(expr js.IExpr, invert bool, prec js.OpPrec) {
	if invert {
		unaryExpr, isUnary := expr.(*js.UnaryExpr)
		binaryExpr, isBinary := expr.(*js.BinaryExpr)
		if isUnary && unaryExpr.Op == js.NotToken && isBooleanExpr(unaryExpr.X) {
			m.minifyExpr(&js.GroupExpr{expr}, prec)
		} else if isBinary && binaryOpPrecMap[binaryExpr.Op] == js.OpEquals {
			if binaryExpr.Op == js.EqEqToken {
				binaryExpr.Op = js.NotEqToken
			} else if binaryExpr.Op == js.NotEqToken {
				binaryExpr.Op = js.EqEqToken
			} else if binaryExpr.Op == js.EqEqToken {
				binaryExpr.Op = js.NotEqEqToken
			} else if binaryExpr.Op == js.NotEqEqToken {
				binaryExpr.Op = js.EqEqEqToken
			}
			m.minifyExpr(expr, prec)
		} else {
			m.write([]byte("!"))
			m.minifyExpr(&js.GroupExpr{expr}, js.OpUnary)
		}
	} else if isBooleanExpr(expr) {
		m.minifyExpr(&js.GroupExpr{expr}, prec)
	} else {
		m.write([]byte("!!"))
		m.minifyExpr(&js.GroupExpr{expr}, js.OpUnary)
	}
}

func (m *jsMinifier) isUndefined(i js.IExpr) bool {
	if lit, ok := i.(*js.LiteralExpr); ok && lit.TokenType == js.IdentifierToken && bytes.Equal(lit.Data, []byte("undefined")) && !m.renamer.exists(lit.Data) {
		return true
	} else if unary, ok := i.(*js.UnaryExpr); ok && unary.Op == js.VoidToken {
		return true
	}
	return false
}

func (m *jsMinifier) isTruthy(i js.IExpr) (bool, bool) {
	if falsy, ok := m.isFalsy(i); ok {
		return !falsy, true
	}
	return false, false
}

func (m *jsMinifier) isFalsy(i js.IExpr) (bool, bool) {
	negated := false
	group, isGroup := i.(*js.GroupExpr)
	unary, isUnary := i.(*js.UnaryExpr)
	for isGroup || isUnary && unary.Op == js.NotToken {
		if isGroup {
			i = group.X
		} else {
			i = unary.X
			negated = !negated
		}
		group, isGroup = i.(*js.GroupExpr)
		unary, isUnary = i.(*js.UnaryExpr)
	}
	if lit, ok := i.(*js.LiteralExpr); ok {
		if lit.TokenType == js.FalseToken || lit.TokenType == js.NullToken ||
			lit.TokenType == js.StringToken && len(lit.Data) == 0 ||
			lit.TokenType == js.DecimalToken && (len(lit.Data) == 1 && lit.Data[0] == '0' || len(lit.Data) == 2 && lit.Data[0] == '.' && lit.Data[1] == '0') ||
			(lit.TokenType == js.BinaryToken || lit.TokenType == js.OctalToken || lit.TokenType == js.HexadecimalToken) && len(lit.Data) == 3 && lit.Data[2] == '0' ||
			lit.TokenType == js.BigIntToken && len(lit.Data) == 2 && lit.Data[0] == '0' {
			return !negated, true // false
		} else if lit.TokenType == js.TrueToken || lit.TokenType == js.StringToken || lit.TokenType == js.DecimalToken || lit.TokenType == js.BinaryToken || lit.TokenType == js.OctalToken || lit.TokenType == js.HexadecimalToken || lit.TokenType == js.BigIntToken {
			return negated, true // true
		}
	} else if m.isUndefined(i) {
		return !negated, true // false
	}
	return false, false // unknown
}

type renamer2 struct {
	renames       []map[string][]byte
	reserved      []map[string]bool
	lastRename    [][]byte
	globals       map[string]bool
	deterministic bool
}

func newRenamer(scope js.Scope, deterministic bool) *renamer2 {
	reserved := map[string]bool{}
	for name, count := range scope.Unbound {
		if 0 < count {
			reserved[name] = true
		}
	}
	return &renamer2{
		renames:       []map[string][]byte{map[string][]byte{}},
		reserved:      []map[string]bool{reserved},
		lastRename:    [][]byte{[]byte("`")},
		globals:       scope.Bound,
		deterministic: deterministic,
	}
}

func (r *renamer2) enterScope(scope js.Scope, isBlock bool) {
	renames := map[string][]byte{}
	var reserved map[string]bool
	var rename []byte

	n := len(r.renames)
	parentRenames := r.renames[n-1]
	if isBlock {
		reserved = r.reserved[n-1]
		rename = r.lastRename[n-1]
		for name, rename := range parentRenames {
			renames[name] = rename
		}
	} else {
		reserved = map[string]bool{}
		rename = []byte("`") // so that the next is 'a'
		for name, _ := range js.Keywords {
			reserved[name] = true
		}
		for name, count := range scope.Unbound {
			if 0 < count {
				if rename, ok := parentRenames[name]; ok {
					renames[name] = rename
					reserved[string(rename)] = true
				} else {
					reserved[name] = true
				}
			}
		}
	}
	if r.deterministic {
		bound := []string{}
		for name, ok := range scope.Bound {
			if ok {
				bound = append(bound, name)
			}
		}
		sort.Strings(bound)
		for _, name := range bound {
			rename = r.next(rename)
			for reserved[string(rename)] {
				rename = r.next(rename)
			}
			renames[name] = parse.Copy(rename)
		}
	} else {
		for name, ok := range scope.Bound {
			if ok {
				rename = r.next(rename)
				for reserved[string(rename)] {
					rename = r.next(rename)
				}
				renames[name] = parse.Copy(rename)
			}
		}
	}
	r.renames = append(r.renames, renames)
	r.reserved = append(r.reserved, reserved)
	r.lastRename = append(r.lastRename, rename)
}

func (r *renamer2) exitScope() {
	n := len(r.renames)
	r.renames = r.renames[:n-1]
	r.reserved = r.reserved[:n-1]
	r.lastRename = r.lastRename[:n-1]
}

func (r *renamer2) next(name []byte) []byte {
	if name[len(name)-1] == 'z' {
		name[len(name)-1] = 'A'
	} else if name[len(name)-1] == 'Z' {
		name[len(name)-1] = '_'
	} else if name[len(name)-1] == '_' {
		name[len(name)-1] = '$'
	} else if name[len(name)-1] == '$' {
		isLast := true
		for i := len(name) - 2; 0 <= i; i-- {
			if name[i] != 'Z' {
				if name[i] == 'z' {
					name[i] = 'A'
				} else {
					name[i]++
				}
				for j := i + 1; j < len(name); j++ {
					name[j] = 'a'
				}
				isLast = false
			}
		}
		if isLast {
			for j := 0; j < len(name); j++ {
				name[j] = 'a'
			}
			name = append(name, 'a')
		}
	} else {
		name[len(name)-1]++
	}
	return name
}

func (r *renamer2) rename(name []byte) []byte {
	if r.inGlobalScope() {
		return name
	}
	if rename, ok := r.renames[len(r.renames)-1][string(name)]; ok {
		return rename
	}
	return name
}

func (r *renamer2) exists(name []byte) bool {
	if _, ok := r.globals[string(name)]; ok {
		return true
	}
	for j := len(r.renames) - 1; 0 <= j; j-- {
		if _, ok := r.renames[j][string(name)]; ok {
			return true
		}
	}
	return false
}

func (r *renamer2) inGlobalScope() bool {
	return len(r.renames) == 1
}

type renamer struct {
	unbound  map[string]bool
	reserved map[string]bool
	renames  [][]byte         // list of renames, only ever grows as we find new variables that do not collide
	idx      int              // index into renames, decreases when we leave scope
	scopes   []map[string]int // index into renames, ordered from outer scope to inner
}

func newRenamerOld(unboundCount map[string]int) *renamer {
	unbound := map[string]bool{}
	for name, count := range unboundCount {
		if count != 0 {
			unbound[name] = true
		}
	}
	reserved := map[string]bool{}
	for name, _ := range js.Keywords {
		reserved[name] = true
	}
	for name := range unbound {
		reserved[name] = true
	}
	return &renamer{
		unbound:  unbound,
		reserved: reserved,
	}
}

func (r *renamer) next(name []byte) []byte {
	if name[len(name)-1] == 'z' {
		name[len(name)-1] = 'A'
	} else if name[len(name)-1] == 'Z' {
		name[len(name)-1] = '_'
	} else if name[len(name)-1] == '_' {
		name[len(name)-1] = '$'
	} else if name[len(name)-1] == '$' {
		isLast := true
		for i := len(name) - 2; 0 <= i; i-- {
			if name[i] != 'Z' {
				if name[i] == 'z' {
					name[i] = 'A'
				} else {
					name[i]++
				}
				for j := i + 1; j < len(name); j++ {
					name[j] = 'a'
				}
				isLast = false
			}
		}
		if isLast {
			for j := 0; j < len(name); j++ {
				name[j] = 'a'
			}
			name = append(name, 'a')
		}
	} else {
		name[len(name)-1]++
	}
	return name
}

func (r *renamer) add(src []byte) []byte {
	if r.inGlobal() {
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
	if !r.reserved[string(name)] {
		return r.add(name)
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

func (r *renamer) inGlobal() bool {
	return len(r.scopes) == 0
}
