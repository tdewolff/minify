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
	spaceBytes                 = []byte(" ")
	starBytes                  = []byte("*")
	colonBytes                 = []byte(":")
	semicolonBytes             = []byte(";")
	commaBytes                 = []byte(",")
	dotBytes                   = []byte(".")
	ellipsisBytes              = []byte("...")
	openBraceBytes             = []byte("{")
	closeBraceBytes            = []byte("}")
	openParenBytes             = []byte("(")
	closeParenBytes            = []byte(")")
	openBracketBytes           = []byte("[")
	closeBracketBytes          = []byte("]")
	openParenBracketBytes      = []byte("({")
	closeBracketParenBytes     = []byte("})")
	closeParenOpenBracketBytes = []byte("){")
	notBytes                   = []byte("!")
	questionBytes              = []byte("?")
	equalBytes                 = []byte("=")
	notNotBytes                = []byte("!!")
	andBytes                   = []byte("&&")
	orBytes                    = []byte("||")
	optChainBytes              = []byte("?.")
	arrowBytes                 = []byte("=>")
	zeroBytes                  = []byte("0")
	oneBytes                   = []byte("1")
	letBytes                   = []byte("let")
	getBytes                   = []byte("get")
	setBytes                   = []byte("set")
	asyncBytes                 = []byte("async")
	functionBytes              = []byte("function")
	staticBytes                = []byte("static")
	ifOpenBytes                = []byte("if(")
	elseBytes                  = []byte("else")
	withOpenBytes              = []byte("with(")
	doBytes                    = []byte("do")
	whileOpenBytes             = []byte("while(")
	forOpenBytes               = []byte("for(")
	forAwaitOpenBytes          = []byte("for await(")
	inBytes                    = []byte("in")
	ofBytes                    = []byte("of")
	switchOpenBytes            = []byte("switch(")
	throwBytes                 = []byte("throw")
	tryBytes                   = []byte("try")
	catchBytes                 = []byte("catch")
	finallyBytes               = []byte("finally")
	importBytes                = []byte("import")
	exportBytes                = []byte("export")
	fromBytes                  = []byte("from")
	returnBytes                = []byte("return")
	classBytes                 = []byte("class")
	asSpaceBytes               = []byte("as ")
	asyncSpaceBytes            = []byte("async ")
	spaceDefaultBytes          = []byte(" default")
	spaceExtendsBytes          = []byte(" extends")
	yieldBytes                 = []byte("yield")
	newBytes                   = []byte("new")
	openNewBytes               = []byte("(new")
	newTargetBytes             = []byte("new.target")
	importMetaBytes            = []byte("import.meta")
	undefinedBytes             = []byte("undefined")
	infinityBytes              = []byte("Infinity")
	voidZeroBytes              = []byte("void 0")
	groupedVoidZeroBytes       = []byte("(void 0)")
	oneDivZeroBytes            = []byte("1/0")
	groupedOneDivZeroBytes     = []byte("(1/0)")
	notZeroBytes               = []byte("!0")
	groupedNotZeroBytes        = []byte("(!0)")
	notOneBytes                = []byte("!1")
	groupedNotOneBytes         = []byte("(!1)")
)

// DefaultMinifier is the default minifier.
var DefaultMinifier = &Minifier{}

// Minifier is a JS minifier.
type Minifier struct {
	Precision    int // number of significant digits
	KeepVarNames bool
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
		ctx:     ast.Ctx,
		renamer: newRenamer(ast.Ctx, ast.Undeclared, !o.KeepVarNames),
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
	needsSemicolon bool // write a semicolon if required
	needsSpace     bool // write a space if next token is an identifier
	expectStmt     bool // avoid ambiguous syntax such as an expression starting with function
	ctx            *js.VarCtx
	renamer        *renamer
}

func (m *jsMinifier) write(b []byte) {
	if m.needsSpace && js.IsIdentifierStart(b) {
		m.w.Write(spaceBytes)
	}
	m.w.Write(b)
	m.prev = b
	m.needsSpace = false
	m.expectStmt = false
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
		m.expectStmt = true
		m.minifyExpr(stmt.Value, js.OpExpr)
		m.requireSemicolon()
	case *js.VarDecl:
		m.minifyVarDecl(*stmt)
		m.requireSemicolon()
	case *js.IfStmt:
		hasIf := !isEmptyStmt(stmt.Body)
		hasElse := !isEmptyStmt(stmt.Else)

		m.write(ifOpenBytes)
		m.minifyExpr(stmt.Cond, js.OpExpr)
		m.write(closeParenBytes)

		if !hasIf && hasElse {
			m.requireSemicolon()
		} else if hasIf {
			if block, ok := stmt.Body.(*js.BlockStmt); ok && len(block.List) == 1 {
				stmt.Body = block.List[0]
			}
			if ifStmt, ok := stmt.Body.(*js.IfStmt); ok && isEmptyStmt(ifStmt.Else) {
				m.write(openBraceBytes)
				m.minifyStmt(stmt.Body)
				m.write(closeBraceBytes)
				m.needsSemicolon = false
			} else {
				m.minifyStmt(stmt.Body)
			}
		}
		if hasElse {
			m.writeSemicolon()
			if !hasReturnThrowStmt(stmt.Body) {
				m.write(elseBytes)
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
		m.write(returnBytes)
		m.writeSpaceBeforeIdent()
		m.minifyExpr(stmt.Value, js.OpExpr)
		m.requireSemicolon()
	case *js.LabelledStmt:
		m.write(stmt.Label)
		m.write(colonBytes)
		m.minifyStmt(stmt.Value)
	case *js.BranchStmt:
		m.write(stmt.Type.Bytes())
		if stmt.Label != nil {
			m.write(spaceBytes)
			m.write(stmt.Label)
		}
		m.requireSemicolon()
	case *js.WithStmt:
		m.write(withOpenBytes)
		m.minifyExpr(stmt.Cond, js.OpExpr)
		m.write(closeParenBytes)
		m.minifyStmt(stmt.Body)
	case *js.DoWhileStmt:
		m.write(doBytes)
		m.writeSpaceBeforeIdent()
		m.minifyStmt(stmt.Body)
		m.writeSemicolon()
		m.write(whileOpenBytes)
		m.minifyExpr(stmt.Cond, js.OpExpr)
		m.write(closeParenBytes)
		m.requireSemicolon()
	case *js.WhileStmt:
		m.write(whileOpenBytes)
		m.minifyExpr(stmt.Cond, js.OpExpr)
		m.write(closeParenBytes)
		m.minifyStmt(stmt.Body)
	case *js.ForStmt:
		m.write(forOpenBytes)
		m.minifyExpr(stmt.Init, js.OpExpr)
		m.write(semicolonBytes)
		m.minifyExpr(stmt.Cond, js.OpExpr)
		m.write(semicolonBytes)
		m.minifyExpr(stmt.Post, js.OpExpr)
		m.write(closeParenBytes)
		m.minifyStmt(stmt.Body)
	case *js.ForInStmt:
		m.write(forOpenBytes)
		m.minifyExpr(stmt.Init, js.OpLHS)
		m.writeSpaceAfterIdent()
		m.write(inBytes)
		m.writeSpaceBeforeIdent()
		m.minifyExpr(stmt.Value, js.OpExpr)
		m.write(closeParenBytes)
		m.minifyStmt(stmt.Body)
	case *js.ForOfStmt:
		if stmt.Await {
			m.write(forAwaitOpenBytes)
		} else {
			m.write(forOpenBytes)
		}
		m.minifyExpr(stmt.Init, js.OpLHS)
		m.writeSpaceAfterIdent()
		m.write(ofBytes)
		m.writeSpaceBeforeIdent()
		m.minifyExpr(stmt.Value, js.OpAssign)
		m.write(closeParenBytes)
		m.minifyStmt(stmt.Body)
	case *js.SwitchStmt:
		m.write(switchOpenBytes)
		m.minifyExpr(stmt.Init, js.OpExpr)
		m.write(closeParenOpenBracketBytes)
		m.needsSemicolon = false
		for _, clause := range stmt.List {
			m.writeSemicolon()
			m.write(clause.TokenType.Bytes())
			if clause.Cond != nil {
				m.write(spaceBytes)
				m.minifyExpr(clause.Cond, js.OpExpr)
			}
			m.write(colonBytes)
			for _, item := range clause.List {
				m.writeSemicolon()
				m.minifyStmt(item)
			}
		}
		m.write(closeBraceBytes)
	case *js.ThrowStmt:
		m.write(throwBytes)
		m.writeSpaceBeforeIdent()
		m.minifyExpr(stmt.Value, js.OpExpr)
		m.requireSemicolon()
	case *js.TryStmt:
		m.write(tryBytes)
		m.minifyBlockStmt(stmt.Body, true)
		if len(stmt.Catch.List) != 0 || stmt.Binding != nil {
			m.write(catchBytes)
			m.renamer.renameScope(stmt.Catch.Scope)
			if stmt.Binding != nil {
				m.write(openParenBytes)
				m.minifyBinding(stmt.Binding)
				m.write(closeParenBytes)
			}
			m.minifyBlockStmt(stmt.Catch, false)
		}
		if len(stmt.Finally.List) != 0 {
			m.write(finallyBytes)
			m.minifyBlockStmt(stmt.Finally, true)
		}
	case *js.FuncDecl:
		m.minifyFuncDecl(*stmt, false)
	case *js.ClassDecl:
		m.minifyClassDecl(*stmt)
	case *js.DebuggerStmt:
	case *js.EmptyStmt:
	case *js.ImportStmt:
		m.write(importBytes)
		if stmt.Default != nil {
			m.write(spaceBytes)
			m.write(stmt.Default)
			if len(stmt.List) != 0 {
				m.write(commaBytes)
			}
		}
		if len(stmt.List) == 1 {
			m.writeSpaceBeforeIdent()
			m.minifyAlias(stmt.List[0])
		} else if 1 < len(stmt.List) {
			m.write(openBraceBytes)
			for i, item := range stmt.List {
				if i != 0 {
					m.write(commaBytes)
				}
				m.minifyAlias(item)
			}
			m.write(closeBraceBytes)
		}
		if stmt.Default != nil || len(stmt.List) != 0 {
			if len(stmt.List) < 2 {
				m.write(spaceBytes)
			}
			m.write(fromBytes)
		}
		m.write(stmt.Module)
		m.requireSemicolon()
	case *js.ExportStmt:
		m.write(exportBytes)
		if stmt.Decl != nil {
			if stmt.Default {
				m.write(spaceDefaultBytes)
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
				m.write(openBraceBytes)
				for i, item := range stmt.List {
					if i != 0 {
						m.write(commaBytes)
					}
					m.minifyAlias(item)
				}
				m.write(closeBraceBytes)
			}
			if stmt.Module != nil {
				if len(stmt.List) < 2 && (len(stmt.List) != 1 || isIdentEndAlias(stmt.List[0])) {
					m.write(spaceBytes)
				}
				m.write(fromBytes)
				m.write(stmt.Module)
			}
			m.requireSemicolon()
		}
	}
}

func groupExpr(i js.IExpr, prec js.OpPrec) js.IExpr {
	if exprPrec(i) < prec {
		return &js.GroupExpr{i}
	}
	return i
}

func condExpr(x, y, z js.IExpr) js.IExpr {
	return &js.CondExpr{groupExpr(x, js.OpCoalesce), groupExpr(y, js.OpAssign), groupExpr(z, js.OpAssign)}
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
			return &js.ExprStmt{stmt.Cond}
		} else if hasIf && !hasElse {
			stmt.Body = m.stmtToExpr(stmt.Body)
			if X, isExprBody := stmt.Body.(*js.ExprStmt); isExprBody {
				left := groupExpr(stmt.Cond, binaryLeftPrecMap[js.AndToken])
				right := groupExpr(X.Value, binaryRightPrecMap[js.AndToken])
				return &js.ExprStmt{&js.BinaryExpr{js.AndToken, left, right}}
			}
		} else if !hasIf && hasElse {
			stmt.Else = m.stmtToExpr(stmt.Else)
			if X, isExprElse := stmt.Else.(*js.ExprStmt); isExprElse {
				left := groupExpr(stmt.Cond, binaryLeftPrecMap[js.OrToken])
				right := groupExpr(X.Value, binaryRightPrecMap[js.OrToken])
				return &js.ExprStmt{&js.BinaryExpr{js.OrToken, left, right}}
			}
		} else if hasIf && hasElse {
			stmt.Body = m.stmtToExpr(stmt.Body)
			stmt.Else = m.stmtToExpr(stmt.Else)
			XExpr, isExprBody := stmt.Body.(*js.ExprStmt)
			YExpr, isExprElse := stmt.Else.(*js.ExprStmt)
			if isExprBody && isExprElse {
				return &js.ExprStmt{condExpr(stmt.Cond, XExpr.Value, YExpr.Value)}
			}
			XReturn, isReturnBody := stmt.Body.(*js.ReturnStmt)
			YReturn, isReturnElse := stmt.Else.(*js.ReturnStmt)
			if isReturnBody && isReturnElse {
				if XReturn.Value == nil {
					XReturn.Value = &js.UnaryExpr{js.VoidToken, &js.LiteralExpr{js.NumericToken, zeroBytes}}
				}
				if YReturn.Value == nil {
					YReturn.Value = &js.UnaryExpr{js.VoidToken, &js.LiteralExpr{js.NumericToken, zeroBytes}}
				}
				return &js.ReturnStmt{condExpr(stmt.Cond, XReturn.Value, YReturn.Value)}
			}
			XThrow, isThrowBody := stmt.Body.(*js.ThrowStmt)
			YThrow, isThrowElse := stmt.Else.(*js.ThrowStmt)
			if isThrowBody && isThrowElse {
				return &js.ThrowStmt{condExpr(stmt.Cond, XThrow.Value, YThrow.Value)}
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
			} else if forStmt, ok := list[i+1].(*js.ForStmt); ok {
				if forStmt.Init == nil {
					forStmt.Init = left.Value
					j--
				} else if _, ok := forStmt.Init.(*js.VarDecl); !ok {
					forStmt.Init = &js.BinaryExpr{js.CommaToken, left.Value, forStmt.Init}
					j--
				}
			} else if whileStmt, ok := list[i+1].(*js.WhileStmt); ok {
				list[i+1] = &js.ForStmt{left.Value, whileStmt.Cond, nil, whileStmt.Body}
				j--
			} else if switchStmt, ok := list[i+1].(*js.SwitchStmt); ok {
				switchStmt.Init = &js.BinaryExpr{js.CommaToken, left.Value, switchStmt.Init}
				j--
			} else if withStmt, ok := list[i+1].(*js.WithStmt); ok {
				withStmt.Cond = &js.BinaryExpr{js.CommaToken, left.Value, withStmt.Cond}
				j--
			} else if ifStmt, ok := list[i+1].(*js.IfStmt); ok {
				ifStmt.Cond = &js.BinaryExpr{js.CommaToken, left.Value, ifStmt.Cond}
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
						returnStmt.Value = condExpr(ifStmt.Cond, left.Value, returnStmt.Value)
						list[j-1] = returnStmt
						j--
					} else if left, ok := ifStmt.Else.(*js.ReturnStmt); ok && left.Value != nil {
						returnStmt.Value = condExpr(ifStmt.Cond, returnStmt.Value, left.Value)
						list[j-1] = returnStmt
						j--
					}
				} else if throwStmt, ok := list[j].(*js.ThrowStmt); ok {
					if left, ok := ifStmt.Body.(*js.ThrowStmt); ok {
						throwStmt.Value = condExpr(ifStmt.Cond, left.Value, throwStmt.Value)
						list[j-1] = throwStmt
						j--
					} else if left, ok := ifStmt.Else.(*js.ThrowStmt); ok {
						throwStmt.Value = condExpr(ifStmt.Cond, throwStmt.Value, left.Value)
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
	m.write(openBraceBytes)
	m.needsSemicolon = false
	if enterScope {
		m.renamer.renameScope(stmt.Scope)
	}
	for _, item := range stmt.List {
		m.writeSemicolon()
		m.minifyStmt(item)
		// TODO
		//if _, ok := item.(*js.ReturnStmt); ok {
		//	break
		//} else if _, ok := item.(*js.BranchStmt); ok {
		//	break
		//}
	}
	m.write(closeBraceBytes)
	m.needsSemicolon = false
}

func (m *jsMinifier) minifyAlias(alias js.Alias) {
	if alias.Name != nil {
		m.write(alias.Name)
		if !bytes.Equal(alias.Name, starBytes) {
			m.write(spaceBytes)
		}
		m.write(asSpaceBytes)
	}
	m.write(alias.Binding)
}

func (m *jsMinifier) minifyParams(params js.Params) {
	m.write(openParenBytes)
	for i, item := range params.List {
		if i != 0 {
			m.write(commaBytes)
		}
		m.minifyBindingElement(item)
	}
	if params.Rest != nil {
		if len(params.List) != 0 {
			m.write(commaBytes)
		}
		m.write(ellipsisBytes)
		m.minifyBinding(params.Rest)
	}
	m.write(closeParenBytes)
}

func (m *jsMinifier) minifyArguments(args js.Arguments) {
	m.write(openParenBytes)
	for i, item := range args.List {
		if i != 0 {
			m.write(commaBytes)
		}
		m.minifyExpr(item, js.OpExpr)
	}
	if args.Rest != nil {
		if len(args.List) != 0 {
			m.write(commaBytes)
		}
		m.write(ellipsisBytes)
		m.minifyExpr(args.Rest, js.OpExpr)
	}
	m.write(closeParenBytes)
}

func (m *jsMinifier) minifyVarDecl(decl js.VarDecl) {
	m.write(decl.TokenType.Bytes())
	m.writeSpaceBeforeIdent()
	for i, item := range decl.List {
		if i != 0 {
			m.write(commaBytes)
		}
		m.minifyBindingElement(item)
	}
}

func (m *jsMinifier) minifyFuncDecl(decl js.FuncDecl, inExpr bool) {
	if decl.Async {
		m.write(asyncSpaceBytes)
	}
	m.write(functionBytes)
	if decl.Generator {
		m.write(starBytes)
	}
	if inExpr {
		m.renamer.renameScope(decl.Scope)
	}
	if decl.Name != nil {
		if !decl.Generator {
			m.write(spaceBytes)
		}
		m.write(decl.Name.Get(m.ctx).Name)
	}
	if !inExpr {
		m.renamer.renameScope(decl.Scope)
	}
	m.minifyParams(decl.Params)
	m.minifyBlockStmt(decl.Body, false)
}

func (m *jsMinifier) minifyMethodDecl(decl js.MethodDecl) {
	if decl.Static {
		m.write(staticBytes)
		m.writeSpaceBeforeIdent()
	}
	if decl.Async {
		m.write(asyncBytes)
		if decl.Generator {
			m.write(starBytes)
		} else {
			m.writeSpaceBeforeIdent()
		}
	} else if decl.Generator {
		m.write(starBytes)
	} else if decl.Get {
		m.write(getBytes)
		m.writeSpaceBeforeIdent()
	} else if decl.Set {
		m.write(setBytes)
		m.writeSpaceBeforeIdent()
	}
	m.minifyPropertyName(decl.Name)
	m.renamer.renameScope(decl.Scope)
	m.minifyParams(decl.Params)
	m.minifyBlockStmt(decl.Body, false)
}

func (m *jsMinifier) minifyArrowFunc(decl js.ArrowFunc) {
	m.renamer.renameScope(decl.Scope)
	if decl.Async {
		m.write(asyncBytes)
	}
	if decl.Params.Rest == nil && len(decl.Params.List) == 1 && decl.Params.List[0].Default == nil {
		if decl.Async && isIdentStartBindingElement(decl.Params.List[0]) {
			m.write(spaceBytes)
		}
		m.minifyBindingElement(decl.Params.List[0])
	} else {
		m.minifyParams(decl.Params)
	}
	m.write(arrowBytes)
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
}

func (m *jsMinifier) minifyClassDecl(decl js.ClassDecl) {
	m.write(classBytes)
	if decl.Name != nil {
		m.write(spaceBytes)
		m.write(decl.Name.Get(m.ctx).Name)
	}
	if decl.Extends != nil {
		m.write(spaceExtendsBytes)
		m.writeSpaceBeforeIdent()
		m.minifyExpr(decl.Extends, js.OpLHS)
	}
	m.write(openBraceBytes)
	for _, item := range decl.Methods {
		m.minifyMethodDecl(item)
	}
	m.write(closeBraceBytes)
}

func (m *jsMinifier) minifyPropertyName(name js.PropertyName) {
	if name.IsComputed() {
		m.write(openBracketBytes)
		m.minifyExpr(name.Computed, js.OpAssign)
		m.write(closeBracketBytes)
	} else {
		m.write(name.Literal.Data)
	}
}

func (m *jsMinifier) minifyProperty(property js.Property) {
	// property.Name is always set in ObjectLiteral
	if property.Spread {
		m.write(ellipsisBytes)
	} else if ref, ok := property.Value.(*js.VarRef); !ok || !property.Name.IsIdent(ref.Get(m.ctx).Name) {
		// add 'old-name:' before BindingName as the latter will be renamed
		m.minifyPropertyName(property.Name)
		m.write(colonBytes)
	}
	m.minifyExpr(property.Value, js.OpAssign)
	if property.Init != nil {
		m.write(equalBytes)
		m.minifyExpr(property.Init, js.OpAssign)
	}
}

func (m *jsMinifier) minifyBindingElement(element js.BindingElement) {
	if element.Binding != nil {
		m.minifyBinding(element.Binding)
		if element.Default != nil {
			m.write(equalBytes)
			m.minifyExpr(element.Default, js.OpAssign)
		}
	}
}

func (m *jsMinifier) minifyBinding(i js.IBinding) {
	switch binding := i.(type) {
	case *js.VarRef:
		m.write(binding.Get(m.ctx).Name)
	case *js.BindingArray:
		m.write(openBracketBytes)
		for i, item := range binding.List {
			if i != 0 {
				m.write(commaBytes)
			}
			m.minifyBindingElement(item)
		}
		if binding.Rest != nil {
			if 0 < len(binding.List) {
				m.write(commaBytes)
			}
			m.write(ellipsisBytes)
			m.minifyBinding(binding.Rest)
		} else if 0 < len(binding.List) && binding.List[len(binding.List)-1].Binding == nil {
			m.write(commaBytes)
		}
		m.write(closeBracketBytes)
	case *js.BindingObject:
		m.write(openBraceBytes)
		for i, item := range binding.List {
			if i != 0 {
				m.write(commaBytes)
			}
			// item.Key is always set
			if item.Key.IsComputed() {
				m.minifyPropertyName(item.Key)
				m.write(colonBytes)
			} else if ref, ok := item.Value.Binding.(*js.VarRef); !ok || !item.Key.IsIdent(ref.Get(m.ctx).Name) {
				// add 'old-name:' before BindingName as the latter will be renamed
				m.minifyPropertyName(item.Key)
				m.write(colonBytes)
			}
			m.minifyBindingElement(item.Value)
		}
		if binding.Rest != nil {
			if 0 < len(binding.List) {
				m.write(commaBytes)
			}
			m.write(ellipsisBytes)
			m.write(binding.Rest.Get(m.ctx).Name)
		}
		m.write(closeBraceBytes)
	}
}

func (m *jsMinifier) minifyExpr(i js.IExpr, prec js.OpPrec) {
	switch expr := i.(type) {
	case *js.VarRef:
		data := expr.Get(m.ctx).Name
		if bytes.Equal(data, undefinedBytes) { // TODO: only if not defined
			if js.OpUnary < prec {
				m.write(groupedVoidZeroBytes)
			} else {
				m.write(voidZeroBytes)
			}
		} else if bytes.Equal(data, infinityBytes) { // TODO: only if not defined
			if js.OpMul < prec {
				m.write(groupedOneDivZeroBytes)
			} else {
				m.write(oneDivZeroBytes)
			}
		} else {
			m.write(data)
		}
	case *js.LiteralExpr:
		if expr.TokenType == js.DecimalToken {
			m.write(minify.Number(expr.Data, 0))
		} else if expr.TokenType == js.BinaryToken {
			m.write(binaryNumber(expr.Data))
		} else if expr.TokenType == js.OctalToken {
			m.write(octalNumber(expr.Data))
		} else if expr.TokenType == js.TrueToken {
			if js.OpUnary < prec {
				m.write(groupedNotZeroBytes)
			} else {
				m.write(notZeroBytes)
			}
		} else if expr.TokenType == js.FalseToken {
			if js.OpUnary < prec {
				m.write(groupedNotOneBytes)
			} else {
				m.write(notOneBytes)
			}
		} else if expr.TokenType == js.StringToken {
			m.write(minifyString(expr.Data))
		} else {
			m.write(expr.Data)
		}
	case *js.BinaryExpr:
		precLeft := binaryLeftPrecMap[expr.Op]
		// convert (a,b)&&c into a,b&&c but not a=(b,c)&&d into a=(b,c&&d)
		if prec <= js.OpExpr {
			if group, ok := expr.X.(*js.GroupExpr); ok {
				if binary, ok := group.X.(*js.BinaryExpr); ok && binary.Op == js.CommaToken {
					expr.X = group.X
					precLeft = js.OpExpr
				}
			}
		}
		m.minifyExpr(expr.X, precLeft)
		if expr.Op == js.InstanceofToken || expr.Op == js.InToken {
			m.writeSpaceAfterIdent()
			m.write(expr.Op.Bytes())
			m.writeSpaceBeforeIdent()
		} else {
			if expr.Op == js.GtToken {
				if unary, ok := expr.X.(*js.UnaryExpr); ok && unary.Op == js.PostDecrToken {
					m.write(spaceBytes)
				}
			}
			m.write(expr.Op.Bytes())
			if expr.Op == js.AddToken {
				if unary, ok := expr.Y.(*js.UnaryExpr); ok && (unary.Op == js.PosToken || unary.Op == js.PreIncrToken) {
					m.write(spaceBytes)
				}
			} else if expr.Op == js.NegToken {
				if unary, ok := expr.X.(*js.UnaryExpr); ok && (unary.Op == js.NegToken || unary.Op == js.PreDecrToken) {
					m.write(spaceBytes)
				}
			} else if expr.Op == js.LtToken {
				if unary, ok := expr.Y.(*js.UnaryExpr); ok && unary.Op == js.NotToken {
					if unary2, ok2 := unary.X.(*js.UnaryExpr); ok2 && unary2.Op == js.PreDecrToken {
						m.write(spaceBytes)
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
					m.write(spaceBytes)
				}
			} else if expr.Op == js.NegToken {
				if unary, ok := expr.X.(*js.UnaryExpr); ok && (unary.Op == js.NegToken || unary.Op == js.PreDecrToken) {
					m.write(spaceBytes)
				}
			} else if expr.Op == js.NotToken {
				if lit, ok := expr.X.(*js.LiteralExpr); ok && (lit.TokenType == js.StringToken || lit.TokenType == js.RegExpToken) {
					m.write(oneBytes)
					break
				} else if ok && lit.TokenType == js.DecimalToken {
					if num := minify.Number(lit.Data, 0); len(num) == 1 && num[0] == '0' {
						m.write(zeroBytes)
					} else {
						m.write(oneBytes)
					}
					break
				}
			}
			m.minifyExpr(expr.X, unaryPrecMap[expr.Op])
		}
	case *js.DotExpr:
		if group, ok := expr.X.(*js.GroupExpr); ok {
			if lit, ok := group.X.(*js.LiteralExpr); ok && lit.TokenType == js.DecimalToken {
				num := minify.Number(lit.Data, 0)
				isInt := true
				for _, c := range num {
					if c == '.' || c == 'e' || c == 'E' {
						isInt = false
						break
					}
				}
				if isInt {
					m.write(num)
					m.write(dotBytes)
				} else {
					m.write(num)
				}
				m.write(dotBytes)
				m.write(expr.Y.Data)
				break
			}
		}
		m.minifyExpr(expr.X, js.OpMember)
		m.write(dotBytes)
		m.write(expr.Y.Data)
	case *js.GroupExpr:
		precInside := exprPrec(expr.X)
		if prec <= precInside {
			m.minifyExpr(expr.X, prec)
		} else {
			m.write(openParenBytes)
			m.minifyExpr(expr.X, js.OpExpr)
			m.write(closeParenBytes)
		}
	case *js.ArrayExpr:
		m.write(openBracketBytes)
		for i, item := range expr.List {
			if i != 0 {
				m.write(commaBytes)
			}
			if item.Spread {
				m.write(ellipsisBytes)
			}
			m.minifyExpr(item.Value, js.OpAssign)
		}
		if 0 < len(expr.List) && expr.List[len(expr.List)-1].Value == nil {
			m.write(commaBytes)
		}
		m.write(closeBracketBytes)
	case *js.ObjectExpr:
		expectStmt := m.expectStmt
		if expectStmt {
			m.write(openParenBracketBytes)
		} else {
			m.write(openBraceBytes)
		}
		for i, item := range expr.List {
			if i != 0 {
				m.write(commaBytes)
			}
			m.minifyProperty(item)
		}
		if expectStmt {
			m.write(closeBracketParenBytes)
		} else {
			m.write(closeBraceBytes)
		}
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
		if expr.Args == nil && js.OpMember <= prec {
			m.write(openNewBytes)
			m.writeSpaceBeforeIdent()
			m.minifyExpr(expr.X, js.OpMember)
			m.write(closeParenBytes)
		} else {
			m.write(newBytes)
			m.writeSpaceBeforeIdent()
			m.minifyExpr(expr.X, js.OpMember)
			if expr.Args != nil {
				m.minifyArguments(*expr.Args)
			}
		}
	case *js.NewTargetExpr:
		m.write(newTargetBytes)
		m.writeSpaceBeforeIdent()
	case *js.ImportMetaExpr:
		m.write(importMetaBytes)
		m.writeSpaceBeforeIdent()
	case *js.YieldExpr:
		m.write(yieldBytes)
		m.writeSpaceBeforeIdent()
		if expr.X != nil {
			if expr.Generator {
				m.write(starBytes)
				m.minifyExpr(expr.X, js.OpAssign)
			} else if ref, ok := expr.X.(*js.VarRef); !ok || !bytes.Equal(ref.Get(m.ctx).Name, undefinedBytes) { // TODO: only if not bound
				m.minifyExpr(expr.X, js.OpAssign)
			}
		}
	case *js.CallExpr:
		m.minifyExpr(expr.X, js.OpMember)
		m.minifyArguments(expr.Args)
	case *js.IndexExpr:
		if m.expectStmt {
			if ref, ok := expr.X.(*js.VarRef); ok && bytes.Equal(ref.Get(m.ctx).Name, letBytes) {
				m.write(notBytes)
			}
		}
		m.minifyExpr(expr.X, js.OpMember)
		if lit, ok := expr.Index.(*js.LiteralExpr); ok && lit.TokenType == js.StringToken {
			if _, ok := js.ParseIdentifierName(lit.Data[1 : len(lit.Data)-1]); ok {
				m.write(dotBytes)
				m.write(lit.Data[1 : len(lit.Data)-1])
				break
			} else if _, ok := js.ParseNumericLiteral(lit.Data[1 : len(lit.Data)-1]); ok {
				m.write(openBracketBytes)
				m.write(lit.Data[1 : len(lit.Data)-1])
				m.write(closeBracketBytes)
				break
			}
		}
		m.write(openBracketBytes)
		m.minifyExpr(expr.Index, js.OpExpr)
		m.write(closeBracketBytes)
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
		} else if m.isEqualExpr(expr.Cond, expr.X) && prec <= js.OpOr && (exprPrec(expr.X) < js.OpAssign || binaryLeftPrecMap[js.OrToken] <= exprPrec(expr.X)) && (exprPrec(expr.Y) < js.OpAssign || binaryRightPrecMap[js.OrToken] <= exprPrec(expr.Y)) {
			// for higher prec we need to add group parenthesis, and for lower prec we have parenthesis anyways. This only is shorter if len(expr.X) >= 3. isEqualExpr only checks for literal variables, which is a name will be minified to a one or two character name.
			m.minifyExpr(expr.X, binaryLeftPrecMap[js.OrToken])
			m.write(orBytes)
			m.minifyExpr(expr.Y, binaryRightPrecMap[js.OrToken])
		} else if m.isEqualExpr(expr.X, expr.Y) {
			if prec <= js.OpExpr {
				m.minifyExpr(expr.Cond, binaryLeftPrecMap[js.CommaToken])
				m.write(commaBytes)
				m.minifyExpr(expr.X, binaryRightPrecMap[js.CommaToken])
			} else {
				m.write(openParenBytes)
				m.minifyExpr(expr.Cond, binaryLeftPrecMap[js.CommaToken])
				m.write(commaBytes)
				m.minifyExpr(expr.X, binaryRightPrecMap[js.CommaToken])
				m.write(closeParenBytes)
			}
		} else {
			// shorten if cases are true and false
			trueX, falseX := m.isTrue(expr.X), m.isFalse(expr.X)
			trueY, falseY := m.isTrue(expr.Y), m.isFalse(expr.Y)
			if trueX && falseY || falseX && trueY {
				m.minifyBooleanExpr(expr.Cond, falseX, prec)
			} else if trueX || trueY {
				// trueX != trueY
				m.minifyBooleanExpr(expr.Cond, trueY, binaryLeftPrecMap[js.OrToken])
				m.write(orBytes)
				if trueY {
					m.minifyExpr(expr.X, binaryRightPrecMap[js.OrToken])
				} else {
					m.minifyExpr(expr.Y, binaryRightPrecMap[js.OrToken])
				}
			} else if falseX || falseY {
				// falseX != falseY
				m.minifyBooleanExpr(expr.Cond, falseX, binaryLeftPrecMap[js.AndToken])
				m.write(andBytes)
				if falseX {
					m.minifyExpr(expr.Y, binaryRightPrecMap[js.AndToken])
				} else {
					m.minifyExpr(expr.X, binaryRightPrecMap[js.AndToken])
				}
			} else {
				// regular conditional expression
				m.minifyExpr(expr.Cond, js.OpCoalesce)
				m.write(questionBytes)
				m.minifyExpr(expr.X, js.OpAssign)
				m.write(colonBytes)
				m.minifyExpr(expr.Y, js.OpAssign)
			}
		}
	case *js.OptChainExpr:
		m.minifyExpr(expr.X, js.OpLHS)
		m.write(optChainBytes)
		m.minifyExpr(expr.Y, js.OpMember)
	case *js.VarDecl:
		m.minifyVarDecl(*expr) // only happens for init in for statement
	case *js.FuncDecl:
		if m.expectStmt {
			m.write(notBytes)
		}
		m.minifyFuncDecl(*expr, true)
	case *js.ArrowFunc:
		m.minifyArrowFunc(*expr)
	case *js.MethodDecl:
		m.minifyMethodDecl(*expr)
	case *js.ClassDecl:
		if m.expectStmt {
			m.write(notBytes)
		}
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
			m.write(notBytes)
			m.minifyExpr(&js.GroupExpr{expr}, js.OpUnary)
		}
	} else if isBooleanExpr(expr) {
		m.minifyExpr(&js.GroupExpr{expr}, prec)
	} else {
		m.write(notNotBytes)
		m.minifyExpr(&js.GroupExpr{expr}, js.OpUnary)
	}
}

type renaming struct {
	src, dst []byte
}

type renamer struct {
	ctx      *js.VarCtx
	reserved map[string]struct{}
	rename   bool
}

func newRenamer(ctx *js.VarCtx, undeclared js.VarArray, rename bool) *renamer {
	reserved := make(map[string]struct{}, len(js.Keywords)+len(js.Globals)+len(undeclared))
	for name, _ := range js.Keywords {
		reserved[name] = struct{}{}
	}
	for name, _ := range js.Globals {
		reserved[name] = struct{}{}
	}
	for _, v := range undeclared {
		if 0 < v.Uses {
			reserved[string(v.Name)] = struct{}{}
		}
	}
	// TODO: sort variable names on highest usage throughout the file, right now lower scopes can have high usage but are forced to use two-character names as the one-character names are depleted
	return &renamer{
		ctx:      ctx,
		reserved: reserved,
		rename:   rename,
	}
}

func (r *renamer) renameScope(scope js.Scope) {
	if !r.rename {
		return
	}

	rename := []byte("`") // so that the next is 'a'
	sort.Sort(scope.Declared)
	for _, v := range scope.Declared {
		rename = r.next(rename)
		for r.isReserved(rename, scope.Undeclared) {
			rename = r.next(rename)
		}
		v.Name = parse.Copy(rename)
	}
}

func (r *renamer) isReserved(name []byte, undeclared js.VarArray) bool {
	if _, ok := r.reserved[string(name)]; ok {
		return true
	}
	for _, v := range undeclared {
		if bytes.Equal(name, v.Name) { // TODO: check if decl != NoDecl (already in r.reserved)
			return true
		}
	}
	return false
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
