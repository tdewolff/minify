package js

import "github.com/tdewolff/parse/v2/js"

func (m *jsMinifier) optimizeStmt(i js.IStmt) js.IStmt {
	// convert if/else into expression statement, and optimize blocks
	if ifStmt, ok := i.(*js.IfStmt); ok {
		hasIf := !m.isEmptyStmt(ifStmt.Body)
		hasElse := !m.isEmptyStmt(ifStmt.Else)
		if unaryExpr, ok := ifStmt.Cond.(*js.UnaryExpr); ok && unaryExpr.Op == js.NotToken && hasElse {
			ifStmt.Cond = unaryExpr.X
			ifStmt.Body, ifStmt.Else = ifStmt.Else, ifStmt.Body
			hasIf, hasElse = hasElse, hasIf
		}
		if !hasIf && !hasElse {
			return &js.ExprStmt{ifStmt.Cond}
		} else if hasIf && !hasElse {
			ifStmt.Body = m.optimizeStmt(ifStmt.Body)
			if X, isExprBody := ifStmt.Body.(*js.ExprStmt); isExprBody {
				if unaryExpr, ok := ifStmt.Cond.(*js.UnaryExpr); ok && unaryExpr.Op == js.NotToken {
					left := groupExpr(unaryExpr.X, binaryLeftPrecMap[js.OrToken])
					right := groupExpr(X.Value, binaryRightPrecMap[js.OrToken])
					return &js.ExprStmt{&js.BinaryExpr{js.OrToken, left, right}}
				} else {
					left := groupExpr(ifStmt.Cond, binaryLeftPrecMap[js.AndToken])
					right := groupExpr(X.Value, binaryRightPrecMap[js.AndToken])
					return &js.ExprStmt{&js.BinaryExpr{js.AndToken, left, right}}
				}
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
	} else if decl, ok := i.(*js.VarDecl); ok && m.varsHoisted != nil && decl != m.varsHoisted {
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
		} else if len(blockStmt.List) == 0 {
			return &js.EmptyStmt{}
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
	list[0] = m.optimizeStmt(list[0])
	if _, ok := list[0].(*js.EmptyStmt); ok {
		j--
	} else {
		list = m.flattenStmt(list, 0)
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
				var body js.BlockStmt
				if blockStmt, ok := whileStmt.Body.(*js.BlockStmt); ok {
					body = *blockStmt
				} else {
					body.List = []js.IStmt{whileStmt.Body}
				}
				list[i+1] = &js.ForStmt{left.Value, whileStmt.Cond, nil, body}
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
		} else if left, ok := list[i].(*js.VarDecl); ok && left.TokenType != js.VarToken {
			// merge const, let declarations
			if right, ok := list[i+1].(*js.VarDecl); ok && left.TokenType == right.TokenType {
				right.List = append(left.List, right.List...)
				j--
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
	if 0 <= j {
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
	}
	return list[:j+1]
}
