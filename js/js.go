// Package js minifies ECMAScript5.1 following the specifications at http://www.ecma-international.org/ecma-262/5.1/.
package js

// TODO: remove dead code (if(false) or after return/throw/break/continue), difficulty with var/func decls
// TODO: move var declaration or expr statement into for loop init (var only if for has var decl)
// TODO: don't minify variable names in with statement, what todo with eval? Don't minify any variable name?

import (
	"bytes"
	"io"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/js"
)

type blockType int

const (
	defaultBlock blockType = iota
	functionBlock
	iterationBlock
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

	if 3 < len(ast.Comment) && ast.Comment[1] == '*' && ast.Comment[2] == '!' {
		w.Write(ast.Comment) // license comment
	}

	m := &jsMinifier{
		o:       o,
		w:       w,
		ast:     ast,
		renamer: newRenamer(ast, ast.Undeclared, !o.KeepVarNames),
	}
	m.hoistVars(&ast.BlockStmt)
	ast.List = m.optimizeStmtList(ast.List, functionBlock)
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
	groupedStmt    bool // avoid ambiguous syntax by grouping the expression statement
	spaceBefore    byte
	varsHoisted    bool // whether variables are hoisted to the top for this function scope

	ast     *js.AST
	renamer *renamer
}

func (m *jsMinifier) write(b []byte) {
	// 0 < len(b)
	if m.needsSpace && js.IsIdentifierContinue(b) || m.spaceBefore == b[0] {
		m.w.Write(spaceBytes)
	}
	m.w.Write(b)
	m.prev = b
	m.needsSpace = false
	m.expectStmt = false
	m.spaceBefore = 0
}

func (m *jsMinifier) writeSpaceAfterIdent() {
	if js.IsIdentifierEnd(m.prev) || 1 < len(m.prev) && m.prev[0] == '/' {
		m.w.Write(spaceBytes)
	}
}

func (m *jsMinifier) writeSpaceBeforeIdent() {
	m.needsSpace = true
}

func (m *jsMinifier) writeSpaceBefore(c byte) {
	m.spaceBefore = c
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

func (m *jsMinifier) minifyStmtOrBlock(i js.IStmt, blockType blockType) {
	if blockStmt, ok := i.(*js.BlockStmt); ok {
		blockStmt.List = m.optimizeStmtList(blockStmt.List, blockType)
		if 1 < len(blockStmt.List) {
			m.minifyBlockStmt(*blockStmt)
		} else if len(blockStmt.List) == 1 {
			m.minifyStmt(blockStmt.List[0])
		} else {
			m.write(semicolonBytes)
		}
	} else if _, ok := i.(*js.EmptyStmt); ok {
		m.write(semicolonBytes)
	} else {
		m.minifyStmt(i)
	}
}

func (m *jsMinifier) minifyStmt(i js.IStmt) {
	switch stmt := i.(type) {
	case *js.ExprStmt:
		m.expectStmt = true
		m.minifyExpr(stmt.Value, js.OpExpr)
		if m.groupedStmt {
			m.write(closeParenBytes)
			m.groupedStmt = false
		}
		m.requireSemicolon()
	case *js.VarDecl:
		m.minifyVarDecl(*stmt, false, false)
		m.requireSemicolon()
	case *js.IfStmt:
		hasIf := !m.isEmptyStmt(stmt.Body)
		hasElse := !m.isEmptyStmt(stmt.Else)

		m.write(ifOpenBytes)
		m.minifyExpr(stmt.Cond, js.OpExpr)
		m.write(closeParenBytes)

		if !hasIf && hasElse {
			m.requireSemicolon()
		} else if hasIf {
			if ifStmt, ok := stmt.Body.(*js.IfStmt); ok && m.isEmptyStmt(ifStmt.Else) {
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
			m.write(elseBytes)
			m.writeSpaceBeforeIdent()
			m.minifyStmt(stmt.Else)
		}
	case *js.BlockStmt:
		m.renamer.renameScope(stmt.Scope)
		m.minifyBlockStmt(*stmt)
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
		m.minifyStmtOrBlock(stmt.Body, defaultBlock)
	case *js.DoWhileStmt:
		m.write(doBytes)
		m.writeSpaceBeforeIdent()
		m.minifyStmtOrBlock(stmt.Body, iterationBlock)
		m.writeSemicolon()
		m.write(whileOpenBytes)
		m.minifyExpr(stmt.Cond, js.OpExpr)
		m.write(closeParenBytes)
		m.requireSemicolon()
	case *js.WhileStmt:
		m.write(whileOpenBytes)
		m.minifyExpr(stmt.Cond, js.OpExpr)
		m.write(closeParenBytes)
		m.minifyStmtOrBlock(stmt.Body, iterationBlock)
	case *js.ForStmt:
		m.write(forOpenBytes)
		if decl, ok := stmt.Init.(*js.VarDecl); ok {
			m.minifyVarDecl(*decl, false, true)
		} else {
			m.minifyExpr(stmt.Init, js.OpLHS)
		}
		m.write(semicolonBytes)
		m.minifyExpr(stmt.Cond, js.OpExpr)
		m.write(semicolonBytes)
		m.minifyExpr(stmt.Post, js.OpExpr)
		m.write(closeParenBytes)
		m.minifyStmtOrBlock(stmt.Body, iterationBlock)
	case *js.ForInStmt:
		m.write(forOpenBytes)
		if decl, ok := stmt.Init.(*js.VarDecl); ok {
			m.minifyVarDecl(*decl, false, true)
		} else {
			m.minifyExpr(stmt.Init, js.OpLHS)
		}
		m.writeSpaceAfterIdent()
		m.write(inBytes)
		m.writeSpaceBeforeIdent()
		m.minifyExpr(stmt.Value, js.OpExpr)
		m.write(closeParenBytes)
		m.minifyStmtOrBlock(stmt.Body, iterationBlock)
	case *js.ForOfStmt:
		if stmt.Await {
			m.write(forAwaitOpenBytes)
		} else {
			m.write(forOpenBytes)
		}
		if decl, ok := stmt.Init.(*js.VarDecl); ok {
			m.minifyVarDecl(*decl, false, true)
		} else {
			m.minifyExpr(stmt.Init, js.OpLHS)
		}
		m.writeSpaceAfterIdent()
		m.write(ofBytes)
		m.writeSpaceBeforeIdent()
		m.minifyExpr(stmt.Value, js.OpAssign)
		m.write(closeParenBytes)
		m.minifyStmtOrBlock(stmt.Body, iterationBlock)
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
			clause.List = m.optimizeStmtList(clause.List, defaultBlock)
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
		m.renamer.renameScope(stmt.Body.Scope)
		stmt.Body.List = m.optimizeStmtList(stmt.Body.List, defaultBlock)
		m.minifyBlockStmt(stmt.Body)
		if stmt.Catch != nil {
			m.write(catchBytes)
			m.renamer.renameScope(stmt.Catch.Scope)
			if stmt.Binding != nil {
				m.write(openParenBytes)
				m.minifyBinding(stmt.Binding)
				m.write(closeParenBytes)
			}
			stmt.Catch.List = m.optimizeStmtList(stmt.Catch.List, defaultBlock)
			m.minifyBlockStmt(*stmt.Catch)
		}
		if stmt.Finally != nil {
			m.write(finallyBytes)
			m.renamer.renameScope(stmt.Finally.Scope)
			stmt.Finally.List = m.optimizeStmtList(stmt.Finally.List, defaultBlock)
			m.minifyBlockStmt(*stmt.Finally)
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
				if len(stmt.List) < 2 && (len(stmt.List) != 1 || !bytes.Equal(stmt.List[0].Binding, starBytes)) {
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

func (m *jsMinifier) optimizeStmt(i js.IStmt) js.IStmt {
	// convert if/else into expression statement, and optimize blocks
	if ifStmt, ok := i.(*js.IfStmt); ok {
		if unaryExpr, ok := ifStmt.Cond.(*js.UnaryExpr); ok && unaryExpr.Op == js.NotToken {
			ifStmt.Cond = unaryExpr.X
			ifStmt.Body, ifStmt.Else = ifStmt.Else, ifStmt.Body
		}
		hasIf := !m.isEmptyStmt(ifStmt.Body)
		hasElse := !m.isEmptyStmt(ifStmt.Else)
		if !hasIf && !hasElse {
			return &js.ExprStmt{ifStmt.Cond}
		} else if hasIf && !hasElse {
			ifStmt.Body = m.optimizeStmt(ifStmt.Body)
			if X, isExprBody := ifStmt.Body.(*js.ExprStmt); isExprBody {
				left := groupExpr(ifStmt.Cond, binaryLeftPrecMap[js.AndToken])
				right := groupExpr(X.Value, binaryRightPrecMap[js.AndToken])
				return &js.ExprStmt{&js.BinaryExpr{js.AndToken, left, right}}
			}
		} else if !hasIf && hasElse {
			ifStmt.Else = m.optimizeStmt(ifStmt.Else)
			if X, isExprElse := ifStmt.Else.(*js.ExprStmt); isExprElse {
				left := groupExpr(ifStmt.Cond, binaryLeftPrecMap[js.OrToken])
				right := groupExpr(X.Value, binaryRightPrecMap[js.OrToken])
				return &js.ExprStmt{&js.BinaryExpr{js.OrToken, left, right}}
			}
		} else if hasIf && hasElse {
			ifStmt.Body = m.optimizeStmt(ifStmt.Body)
			ifStmt.Else = m.optimizeStmt(ifStmt.Else)
			XExpr, isExprBody := ifStmt.Body.(*js.ExprStmt)
			YExpr, isExprElse := ifStmt.Else.(*js.ExprStmt)
			if isExprBody && isExprElse {
				return &js.ExprStmt{condExpr(ifStmt.Cond, XExpr.Value, YExpr.Value)}
			}
			XReturn, isReturnBody := ifStmt.Body.(*js.ReturnStmt)
			YReturn, isReturnElse := ifStmt.Else.(*js.ReturnStmt)
			if isReturnBody && isReturnElse {
				if XReturn.Value == nil && YReturn.Value == nil {
					return &js.ReturnStmt{&js.BinaryExpr{js.CommaToken, ifStmt.Cond, &js.UnaryExpr{js.VoidToken, &js.LiteralExpr{js.NumericToken, zeroBytes}}}}
				} else if XReturn.Value != nil && YReturn.Value != nil {
					return &js.ReturnStmt{condExpr(ifStmt.Cond, XReturn.Value, YReturn.Value)}
				}
				return ifStmt
			}
			XThrow, isThrowBody := ifStmt.Body.(*js.ThrowStmt)
			YThrow, isThrowElse := ifStmt.Else.(*js.ThrowStmt)
			if isThrowBody && isThrowElse {
				return &js.ThrowStmt{condExpr(ifStmt.Cond, XThrow.Value, YThrow.Value)}
			}
		}
	} else if decl, ok := i.(*js.VarDecl); ok && m.varsHoisted {
		for _, item := range decl.List {
			if item.Default != nil {
				return &js.ExprStmt{decl}
			}
		}
		return &js.EmptyStmt{}
	} else if blockStmt, ok := i.(*js.BlockStmt); ok {
		// merge body and remove braces if it is not a lexical declaration
		blockStmt.List = m.optimizeStmtList(blockStmt.List, defaultBlock)
		if len(blockStmt.List) == 1 {
			varDecl, isVarDecl := blockStmt.List[0].(*js.VarDecl)
			_, isClassDecl := blockStmt.List[0].(*js.ClassDecl)
			if !isClassDecl && (!isVarDecl || varDecl.TokenType == js.VarToken) {
				return m.optimizeStmt(blockStmt.List[0])
			}
		}
		return blockStmt
	}
	return i
}

func (m *jsMinifier) flattenStmt(list []js.IStmt, i int) []js.IStmt {
	if ifStmt, ok := list[i].(*js.IfStmt); ok && !m.isEmptyStmt(ifStmt.Else) && isFlowStmt(lastStmt(ifStmt.Body)) {
		// if body ends in flow statement (return, throw, break, continue), so we can remove the else statement and put its body in the current scope
		if blockStmt, ok := ifStmt.Else.(*js.BlockStmt); ok {
			list = append(append(list[:i+1], blockStmt.List...), list[i+1:]...)
		} else {
			list = append(append(list[:i+1], ifStmt.Else), list[i+1:]...)
		}
		ifStmt.Else = nil
	}
	return list
}

func (m *jsMinifier) optimizeStmtList(list []js.IStmt, blockType blockType) []js.IStmt {
	// merge expression statements as well as if/else statements followed by flow control statements
	if len(list) == 0 {
		return list
	}
	j := 0
	if !m.varsHoisted || blockType != functionBlock {
		// optimizeStmt only if we're not hoisting vars
		list[0] = m.optimizeStmt(list[0])
		if _, ok := list[0].(*js.EmptyStmt); ok {
			j--
		} else {
			list = m.flattenStmt(list, 0)
		}
	}
	for i, _ := range list[:len(list)-1] {
		// probe at every i allowing one lookahead to i+1, write to position j <= i
		j++
		list[i+1] = m.optimizeStmt(list[i+1])
		if _, ok := list[i+1].(*js.EmptyStmt); ok {
			j--
			continue
		} else {
			list = m.flattenStmt(list, i+1)
		}

		// merge expression statements with expression, return, and throw statements
		if left, ok := list[i].(*js.ExprStmt); ok {
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
			// merge const, let declarations
			if right, ok := list[i+1].(*js.VarDecl); ok && left.TokenType == right.TokenType {
				right.List = append(left.List, right.List...)
				j--
				//} else if left.TokenType == js.VarToken {
				//	if forStmt, ok := list[i+1].(*js.ForStmt); ok {
				//		if init, ok := forStmt.Init.(*js.VarDecl); ok && init.TokenType == js.VarToken {
				//			init.List = append(left.List, init.List...)
				//			j--
				//		}
				//	} else if whileStmt, ok := list[i+1].(*js.WhileStmt); ok {
				//		list[i+1] = &js.ForStmt{left, whileStmt.Cond, nil, whileStmt.Body}
				//		j--
				//	}
			}
		}
		list[j] = list[i+1]

		// merge if/else with return/throw when followed by return/throw
		if 0 < j {
			if ifStmt, ok := list[j-1].(*js.IfStmt); ok && m.isEmptyStmt(ifStmt.Body) != m.isEmptyStmt(ifStmt.Else) {
				// either the if body is empty or the else body is empty. In case where both bodies have return/throw, we already rewrote that if statement to an return/throw statement
				if returnStmt, ok := list[j].(*js.ReturnStmt); ok {
					if returnStmt.Value == nil {
						if left, ok := ifStmt.Body.(*js.ReturnStmt); ok && left.Value == nil {
							list[j-1] = &js.ExprStmt{ifStmt.Cond}
						} else if left, ok := ifStmt.Else.(*js.ReturnStmt); ok && left.Value == nil {
							list[j-1] = &js.ExprStmt{ifStmt.Cond}
						}
					} else {
						if left, ok := ifStmt.Body.(*js.ReturnStmt); ok && left.Value != nil {
							returnStmt.Value = condExpr(ifStmt.Cond, left.Value, returnStmt.Value)
							list[j-1] = returnStmt
							j--
						} else if left, ok := ifStmt.Else.(*js.ReturnStmt); ok && left.Value != nil {
							returnStmt.Value = condExpr(ifStmt.Cond, returnStmt.Value, left.Value)
							list[j-1] = returnStmt
							j--
						}
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

	// remove superfluous return or continue
	if blockType == functionBlock {
		if returnStmt, ok := list[j].(*js.ReturnStmt); ok {
			if returnStmt.Value == nil || m.isUndefined(returnStmt.Value) {
				j--
			} else if binaryExpr, ok := returnStmt.Value.(*js.BinaryExpr); ok && binaryExpr.Op == js.CommaToken && m.isUndefined(binaryExpr.Y) {
				// rewrite function f(){return a,void 0} => function f(){a}
				list[j] = &js.ExprStmt{binaryExpr.X}
			}
		}
	} else if blockType == iterationBlock {
		if branchStmt, ok := list[j].(*js.BranchStmt); ok && branchStmt.Type == js.ContinueToken && branchStmt.Label == nil {
			j--
		}
	}
	return list[:j+1]
}

func (m *jsMinifier) minifyBlockStmt(stmt js.BlockStmt) {
	m.write(openBraceBytes)
	m.needsSemicolon = false
	for _, item := range stmt.List {
		m.writeSemicolon()
		m.minifyStmt(item)
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

func (m *jsMinifier) minifyVarDecl(decl js.VarDecl, onlyDefines, inExpr bool) {
	if inExpr && m.varsHoisted {
		// remove 'var' when hoisting variables
		first := true
		for _, item := range decl.List {
			if item.Default != nil || !onlyDefines {
				if !first {
					m.write(commaBytes)
				}
				m.minifyBindingElement(item)
				first = false
			}
		}
	} else {
		m.write(decl.TokenType.Bytes())
		m.writeSpaceBeforeIdent()
		for i, item := range decl.List {
			if i != 0 {
				m.write(commaBytes)
			}
			m.minifyBindingElement(item)
		}
	}
}

func (m *jsMinifier) hoistVars(body *js.BlockStmt) bool {
	// Hoist all variable declarations in the current module/function scope to the top.
	// If the first statement is a var declaration, expand it. Otherwise prepend a new var declaration.
	// Except for the first var declaration, all others are converted to expressions. This is possible because an ArrayBindingPattern and ObjectBindingPattern can be converted to an ArrayLiteral or ObjectLiteral respectively, as they are supersets of the BindingPatterns.
	parentVarsHoisted := m.varsHoisted
	if 1 < body.Scope.Count(js.VariableDecl) {
		if decl, ok := body.List[0].(*js.VarDecl); ok && decl.TokenType == js.VarToken {
			// original declarations
			refs := []js.VarRef{}
			for _, item := range decl.List {
				refs = append(refs, bindingRefs(item.Binding)...)
			}

			// hoist other variable declarations in this function scope but don't initialize yet
		DeclaredLoop:
			for _, v := range body.Scope.Declared {
				if v.Decl == js.VariableDecl {
					for _, ref := range refs {
						if ref == v.Ref {
							continue DeclaredLoop
						}
					}
					decl.List = append(decl.List, js.BindingElement{v.Ref, nil})
				}
			}
		} else {
			decl := &js.VarDecl{js.VarToken, nil}
			for _, v := range body.Scope.Declared {
				if v.Decl == js.VariableDecl {
					decl.List = append(decl.List, js.BindingElement{v.Ref, nil})
				}
			}
			body.List = append([]js.IStmt{decl}, body.List...)
		}
		m.varsHoisted = true
	}
	return parentVarsHoisted
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
		m.renamer.renameScope(decl.Body.Scope)
	}
	if decl.Name != 0 && (!inExpr || 1 < decl.Name.Var(m.ast).Uses) {
		if !decl.Generator {
			m.write(spaceBytes)
		}
		m.write(decl.Name.Name(m.ast))
	}
	if !inExpr {
		m.renamer.renameScope(decl.Body.Scope)
	}
	m.minifyParams(decl.Params)

	parentVarsHoisted := m.hoistVars(&decl.Body)
	decl.Body.List = m.optimizeStmtList(decl.Body.List, functionBlock)
	m.minifyBlockStmt(decl.Body)
	m.varsHoisted = parentVarsHoisted
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
	m.renamer.renameScope(decl.Body.Scope)
	m.minifyParams(decl.Params)

	parentVarsHoisted := m.hoistVars(&decl.Body)
	decl.Body.List = m.optimizeStmtList(decl.Body.List, functionBlock)
	m.minifyBlockStmt(decl.Body)
	m.varsHoisted = parentVarsHoisted
}

func (m *jsMinifier) minifyArrowFunc(decl js.ArrowFunc) {
	m.renamer.renameScope(decl.Body.Scope)
	if decl.Async {
		m.write(asyncBytes)
	}
	if decl.Params.Rest == nil && len(decl.Params.List) == 1 && decl.Params.List[0].Default == nil {
		if decl.Async && decl.Params.List[0].Binding != nil {
			// add space after async in: async a => ...
			if _, ok := decl.Params.List[0].Binding.(js.VarRef); ok {
				m.write(spaceBytes)
			}
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
				if 0 < len(list) {
					expr = &js.GroupExpr{expr}
				}
				m.minifyExpr(expr, js.OpAssign)
			}
		} else if isReturn && returnStmt.Value == nil {
			// remove empty return
			decl.Body.List = decl.Body.List[:len(decl.Body.List)-1]
		}
	}
	if !removeBraces {
		parentVarsHoisted := m.hoistVars(&decl.Body)
		decl.Body.List = m.optimizeStmtList(decl.Body.List, functionBlock)
		m.minifyBlockStmt(decl.Body)
		m.varsHoisted = parentVarsHoisted
	}
}

func (m *jsMinifier) minifyClassDecl(decl js.ClassDecl) {
	m.write(classBytes)
	if decl.Name != 0 {
		m.write(spaceBytes)
		m.write(decl.Name.Name(m.ast))
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
	} else if ref, ok := property.Value.(js.VarRef); property.Name != nil && (!ok || !property.Name.IsIdent(ref.Name(m.ast))) {
		// add 'old-name:' before BindingName as the latter will be renamed
		m.minifyPropertyName(*property.Name)
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

func (m *jsMinifier) minifyBinding(ibinding js.IBinding) {
	switch binding := ibinding.(type) {
	case js.VarRef:
		m.write(binding.Name(m.ast))
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
				m.minifyPropertyName(*item.Key)
				m.write(colonBytes)
			} else if ref, ok := item.Value.Binding.(js.VarRef); !ok || !item.Key.IsIdent(ref.Name(m.ast)) {
				// add 'old-name:' before BindingName as the latter will be renamed
				m.minifyPropertyName(*item.Key)
				m.write(colonBytes)
			}
			m.minifyBindingElement(item.Value)
		}
		if binding.Rest != 0 {
			if 0 < len(binding.List) {
				m.write(commaBytes)
			}
			m.write(ellipsisBytes)
			m.write(binding.Rest.Name(m.ast))
		}
		m.write(closeBraceBytes)
	}
}

func (m *jsMinifier) minifyExpr(i js.IExpr, prec js.OpPrec) {
	switch expr := i.(type) {
	case js.VarRef:
		data := expr.Name(m.ast)
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
		} else if expr.TokenType == js.HexadecimalToken {
			m.write(hexadecimalNumber(expr.Data))
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
			if expr.Op == js.GtToken && m.prev[len(m.prev)-1] == '-' {
				m.write(spaceBytes)
			}
			m.write(expr.Op.Bytes())
			if expr.Op == js.AddToken {
				// +++  =>  + ++
				m.writeSpaceBefore('+')
			} else if expr.Op == js.SubToken {
				// ---  =>  - --
				m.writeSpaceBefore('-')
			} else if expr.Op == js.DivToken {
				// //  =>  / /
				m.writeSpaceBefore('/')
			}
		}
		m.minifyExpr(expr.Y, binaryRightPrecMap[expr.Op])
	case *js.UnaryExpr:
		if expr.Op == js.PostIncrToken || expr.Op == js.PostDecrToken {
			m.minifyExpr(expr.X, unaryPrecMap[expr.Op])
			m.write(expr.Op.Bytes())
		} else {
			isLtNot := expr.Op == js.NotToken && len(m.prev) == 1 && m.prev[0] == '<'
			m.write(expr.Op.Bytes())
			if expr.Op == js.DeleteToken || expr.Op == js.VoidToken || expr.Op == js.TypeofToken || expr.Op == js.AwaitToken {
				m.writeSpaceBeforeIdent()
			} else if expr.Op == js.PosToken {
				// +++  =>  + ++
				m.writeSpaceBefore('+')
			} else if expr.Op == js.NegToken || isLtNot {
				// ---  =>  - --
				// <!--  =>  <! --
				m.writeSpaceBefore('-')
			} else if expr.Op == js.NotToken {
				if lit, ok := expr.X.(*js.LiteralExpr); ok && (lit.TokenType == js.StringToken || lit.TokenType == js.RegExpToken) {
					// !"string"  =>  !1
					m.write(oneBytes)
					break
				} else if ok && lit.TokenType == js.DecimalToken {
					// !123  =>  !1 (except for !0)
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
		m.minifyExpr(expr.X, js.OpLHS)
		m.write(dotBytes)
		m.write(expr.Y.Data)
	case *js.GroupExpr:
		precInside := exprPrec(expr.X)
		if prec <= precInside || precInside == js.OpCoalesce && prec == js.OpBitOr {
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
		if expr.Args == nil && js.OpLHS < prec && prec != js.OpNew {
			m.write(openNewBytes)
			m.writeSpaceBeforeIdent()
			m.minifyExpr(expr.X, js.OpNew)
			m.write(closeParenBytes)
		} else {
			m.write(newBytes)
			m.writeSpaceBeforeIdent()
			if expr.Args != nil {
				m.minifyExpr(expr.X, js.OpMember)
				m.minifyArguments(*expr.Args)
			} else {
				m.minifyExpr(expr.X, js.OpNew)
			}
		}
	case *js.NewTargetExpr:
		m.write(newTargetBytes)
		m.writeSpaceBeforeIdent()
	case *js.ImportMetaExpr:
		if m.expectStmt {
			m.write(openParenBytes)
			m.groupedStmt = true
		}
		m.write(importMetaBytes)
		m.writeSpaceBeforeIdent()
	case *js.YieldExpr:
		m.write(yieldBytes)
		m.writeSpaceBeforeIdent()
		if expr.X != nil {
			if expr.Generator {
				m.write(starBytes)
				m.minifyExpr(expr.X, js.OpAssign)
			} else if ref, ok := expr.X.(js.VarRef); !ok || !bytes.Equal(ref.Name(m.ast), undefinedBytes) { // TODO: only if not bound
				m.minifyExpr(expr.X, js.OpAssign)
			}
		}
	case *js.CallExpr:
		m.minifyExpr(expr.X, js.OpCall)
		m.minifyArguments(expr.Args)
	case *js.IndexExpr:
		if m.expectStmt {
			if ref, ok := expr.X.(js.VarRef); ok && bytes.Equal(ref.Name(m.ast), letBytes) {
				m.write(notBytes)
			}
		}
		m.minifyExpr(expr.X, js.OpLHS)
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

		if truthy, ok := m.isTruthy(expr.Cond); truthy && ok {
			// if condition is truthy
			m.minifyExpr(expr.X, prec)
		} else if !truthy && ok {
			// if condition is falsy
			m.minifyExpr(expr.Y, prec)
		} else if m.isEqualExpr(expr.Cond, expr.X) && prec <= js.OpOr && (exprPrec(expr.X) < js.OpAssign || binaryLeftPrecMap[js.OrToken] <= exprPrec(expr.X)) && (exprPrec(expr.Y) < js.OpAssign || binaryRightPrecMap[js.OrToken] <= exprPrec(expr.Y)) {
			// if condition is equal to true body
			// for higher prec we need to add group parenthesis, and for lower prec we have parenthesis anyways. This only is shorter if len(expr.X) >= 3. isEqualExpr only checks for literal variables, which is a name will be minified to a one or two character name.
			m.minifyExpr(expr.X, binaryLeftPrecMap[js.OrToken])
			m.write(orBytes)
			m.minifyExpr(expr.Y, binaryRightPrecMap[js.OrToken])
		} else if m.isEqualExpr(expr.X, expr.Y) {
			// if true and false bodies are equal
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
			// shorten when true and false bodies are true and false
			trueX, falseX := m.isTrue(expr.X), m.isFalse(expr.X)
			trueY, falseY := m.isTrue(expr.Y), m.isFalse(expr.Y)
			if trueX && falseY || falseX && trueY {
				m.minifyBooleanExpr(expr.Cond, falseX, prec)
			} else if trueX || trueY {
				// trueX != trueY
				m.minifyBooleanExpr(expr.Cond, trueY, binaryLeftPrecMap[js.OrToken])
				m.write(orBytes)
				if trueY {
					m.minifyExpr(&js.GroupExpr{expr.X}, binaryRightPrecMap[js.OrToken])
				} else {
					m.minifyExpr(&js.GroupExpr{expr.Y}, binaryRightPrecMap[js.OrToken])
				}
			} else if falseX || falseY {
				// falseX != falseY
				m.minifyBooleanExpr(expr.Cond, falseX, binaryLeftPrecMap[js.AndToken])
				m.write(andBytes)
				if falseX {
					m.minifyExpr(&js.GroupExpr{expr.Y}, binaryRightPrecMap[js.AndToken])
				} else {
					m.minifyExpr(&js.GroupExpr{expr.X}, binaryRightPrecMap[js.AndToken])
				}
			} else if condExpr, ok := expr.X.(*js.CondExpr); ok && m.isEqualExpr(expr.Y, condExpr.Y) {
				// nested conditional expression with same false bodies
				m.minifyExpr(&js.GroupExpr{expr.Cond}, binaryLeftPrecMap[js.AndToken])
				m.write(andBytes)
				m.minifyExpr(&js.GroupExpr{condExpr.Cond}, binaryRightPrecMap[js.AndToken])
				m.write(questionBytes)
				m.minifyExpr(condExpr.X, js.OpAssign)
				m.write(colonBytes)
				m.minifyExpr(expr.Y, js.OpAssign)
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
		if callExpr, ok := expr.Y.(*js.CallExpr); ok {
			m.minifyArguments(callExpr.Args)
		} else if indexExpr, ok := expr.Y.(*js.IndexExpr); ok {
			m.write(openBracketBytes)
			m.minifyExpr(indexExpr.Index, js.OpExpr)
			m.write(closeBracketBytes)
		} else {
			m.minifyExpr(expr.Y, js.OpLHS)
		}
	case *js.VarDecl:
		m.minifyVarDecl(*expr, true, true) // happens in when vars were hoisted
	case *js.FuncDecl:
		if m.expectStmt {
			m.write(notBytes)
		}
		m.minifyFuncDecl(*expr, true)
	case *js.ArrowFunc:
		m.minifyArrowFunc(*expr)
	case *js.MethodDecl:
		m.minifyMethodDecl(*expr) // only happens in object literal
	case *js.ClassDecl:
		if m.expectStmt {
			m.write(notBytes)
		}
		m.minifyClassDecl(*expr)
	}
}
