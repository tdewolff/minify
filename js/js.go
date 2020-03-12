// Package js minifies ECMAScript5.1 following the specifications at http://www.ecma-international.org/ecma-262/5.1/.
package js

import (
	"bytes"
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
type Minifier struct {
	prevIdent bool
}

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
	for _, item := range ast.List {
		o.minifyStmt(w, item)
	}
	_, err = w.Write(nil)
	return err
}

func (o *Minifier) minifyStmt(w io.Writer, i js.IStmt) {
	switch stmt := i.(type) {
	case *js.ExprStmt:
		o.minifyExpr(w, stmt.Value)
	case *js.VarDecl:
		o.minifyVarDecl(w, *stmt)
	case *js.IfStmt:
		w.Write([]byte("if("))
		o.minifyExpr(w, stmt.Cond)
		w.Write([]byte(")"))
		o.minifyStmt(w, stmt.Body)
		if stmt.Else != nil {
			w.Write([]byte("else"))
			o.minifyStmt(w, stmt.Else)
		}
	case *js.BlockStmt:
		if len(stmt.List) == 1 {
			o.minifyStmt(w, stmt.List[0])
		} else {
			o.minifyBlockStmt(w, *stmt)
		}
	case *js.ReturnStmt:
		w.Write([]byte("return"))
		if isIdentStartExpr(stmt.Value) {
			w.Write([]byte(" "))
		}
		o.minifyExpr(w, stmt.Value)
	case *js.LabelledStmt:
		w.Write(stmt.Token.Data)
		w.Write([]byte(":"))
		o.minifyStmt(w, stmt.Value)
	case *js.BranchStmt:
		w.Write(stmt.Type.Bytes())
		if stmt.Name != nil {
			w.Write([]byte(" "))
			w.Write(stmt.Name.Data)
		}
	case *js.WithStmt:
		w.Write([]byte("with("))
		o.minifyExpr(w, stmt.Cond)
		w.Write([]byte(")"))
		o.minifyStmt(w, stmt.Body)
	case *js.DoWhileStmt:
		w.Write([]byte("do"))
		o.minifyStmt(w, stmt.Body)
		w.Write([]byte("while("))
		o.minifyExpr(w, stmt.Cond)
		w.Write([]byte(")"))
	case *js.WhileStmt:
		w.Write([]byte("while("))
		o.minifyExpr(w, stmt.Cond)
		w.Write([]byte(")"))
		o.minifyStmt(w, stmt.Body)
	case *js.ForStmt:
		w.Write([]byte("for("))
		o.minifyExpr(w, stmt.Init)
		w.Write([]byte(";"))
		o.minifyExpr(w, stmt.Cond)
		w.Write([]byte(";"))
		o.minifyExpr(w, stmt.Post)
		w.Write([]byte(")"))
		o.minifyStmt(w, stmt.Body)
	case *js.ForInStmt:
		w.Write([]byte("for("))
		o.minifyExpr(w, stmt.Init)
		if isIdentEndExpr(stmt.Init) {
			w.Write([]byte(" "))
		}
		w.Write([]byte("in"))
		if isIdentStartExpr(stmt.Value) {
			w.Write([]byte(" "))
		}
		o.minifyExpr(w, stmt.Value)
		w.Write([]byte(")"))
		o.minifyStmt(w, stmt.Body)
	case *js.ForOfStmt:
		if stmt.Await {
			w.Write([]byte("for await("))
		} else {
			w.Write([]byte("for("))
		}
		o.minifyExpr(w, stmt.Init)
		if isIdentEndExpr(stmt.Init) {
			w.Write([]byte(" "))
		}
		w.Write([]byte("of"))
		if isIdentStartExpr(stmt.Value) {
			w.Write([]byte(" "))
		}
		o.minifyExpr(w, stmt.Value)
		w.Write([]byte(")"))
		o.minifyStmt(w, stmt.Body)
	case *js.SwitchStmt:
		w.Write([]byte("switch("))
		o.minifyExpr(w, stmt.Init)
		w.Write([]byte("){"))
		for j, clause := range stmt.List {
			if j != 0 {
				w.Write([]byte(";"))
			}
			w.Write(clause.TokenType.Bytes())
			if clause.Cond != nil {
				w.Write([]byte(" "))
				o.minifyExpr(w, clause.Cond)
			}
			w.Write([]byte(":"))
			for i, item := range clause.List {
				if i != 0 {
					w.Write([]byte(";"))
				}
				o.minifyStmt(w, item)
			}
		}
		w.Write([]byte("}"))
	case *js.ThrowStmt:
		w.Write([]byte("throw"))
		if isIdentStartExpr(stmt.Value) {
			w.Write([]byte(" "))
		}
		o.minifyExpr(w, stmt.Value)
	case *js.TryStmt:
		w.Write([]byte("try"))
		o.minifyBlockStmt(w, stmt.Body)
		if len(stmt.Catch.List) != 0 || stmt.Binding != nil {
			w.Write([]byte("catch"))
			if stmt.Binding != nil {
				w.Write([]byte("("))
				o.minifyBinding(w, stmt.Binding)
				w.Write([]byte(")"))
			}
			o.minifyBlockStmt(w, stmt.Catch)
		}
		if len(stmt.Finally.List) != 0 {
			w.Write([]byte("finally"))
			o.minifyBlockStmt(w, stmt.Finally)
		}
	case *js.FuncDecl:
		o.minifyFuncDecl(w, *stmt)
	case *js.ClassDecl:
		o.minifyClassDecl(w, *stmt)
	case *js.DebuggerStmt:
		w.Write([]byte("debugger"))
	case *js.EmptyStmt:
	case *js.ImportStmt:
		w.Write([]byte("import"))
		if stmt.Default != nil {
			w.Write([]byte(" "))
			w.Write(stmt.Default)
			if len(stmt.List) != 0 {
				w.Write([]byte(","))
			}
		}
		if len(stmt.List) == 1 {
			if stmt.Default == nil && isIdentStartAlias(stmt.List[0]) {
				w.Write([]byte(" "))
			}
			o.minifyAlias(w, stmt.List[0])
		} else if 1 < len(stmt.List) {
			w.Write([]byte("{"))
			for i, item := range stmt.List {
				if i != 0 {
					w.Write([]byte(","))
				}
				o.minifyAlias(w, item)
			}
			w.Write([]byte("}"))
		}
		if stmt.Default != nil || len(stmt.List) != 0 {
			if len(stmt.List) < 2 {
				w.Write([]byte(" "))
			}
			w.Write([]byte("from"))
		}
		w.Write(stmt.Module)
	case *js.ExportStmt:
		w.Write([]byte("export"))
		if stmt.Decl != nil {
			if stmt.Default {
				w.Write([]byte(" default "))
			} else {
				w.Write([]byte(" "))
			}
			o.minifyExpr(w, stmt.Decl)
		} else {
			if len(stmt.List) == 1 {
				if isIdentStartAlias(stmt.List[0]) {
					w.Write([]byte(" "))
				}
				o.minifyAlias(w, stmt.List[0])
			} else if 1 < len(stmt.List) {
				w.Write([]byte("{"))
				for i, item := range stmt.List {
					if i != 0 {
						w.Write([]byte(","))
					}
					o.minifyAlias(w, item)
				}
				w.Write([]byte("}"))
			}
			if stmt.Module != nil {
				if len(stmt.List) < 2 && (len(stmt.List) != 1 || isIdentEndAlias(stmt.List[0])) {
					w.Write([]byte(" "))
				}
				w.Write([]byte("from"))
				w.Write(stmt.Module)
			}
		}
	}
}

func (o *Minifier) minifyAlias(w io.Writer, alias js.Alias) {
	if alias.Name != nil {
		w.Write(alias.Name)
		if !bytes.Equal(alias.Name, []byte("*")) {
			w.Write([]byte(" "))
		}
		w.Write([]byte("as "))
	}
	w.Write(alias.Binding)
}

func (o *Minifier) minifyBlockStmt(w io.Writer, stmt js.BlockStmt) {
	w.Write([]byte("{"))
	for i, item := range stmt.List {
		if i != 0 {
			w.Write([]byte(";"))
		}
		o.minifyStmt(w, item)
	}
	w.Write([]byte("}"))
}

func (o *Minifier) minifyParams(w io.Writer, params js.Params) {
	w.Write([]byte("("))
	for i, item := range params.List {
		if i != 0 {
			w.Write([]byte(","))
		}
		o.minifyBindingElement(w, item)
	}
	if params.Rest != nil {
		if len(params.List) != 0 {
			w.Write([]byte(","))
		}
		w.Write([]byte("..."))
		o.minifyBindingElement(w, *params.Rest)
	}
	w.Write([]byte(")"))
}

func (o *Minifier) minifyArguments(w io.Writer, args js.Arguments) {
	w.Write([]byte("("))
	for i, item := range args.List {
		if i != 0 {
			w.Write([]byte(","))
		}
		o.minifyExpr(w, item)
	}
	if args.Rest != nil {
		if len(args.List) != 0 {
			w.Write([]byte(","))
		}
		w.Write([]byte("..."))
		o.minifyExpr(w, args.Rest)
	}
	w.Write([]byte(")"))
}

func (o *Minifier) minifyVarDecl(w io.Writer, decl js.VarDecl) {
	w.Write(decl.TokenType.Bytes())
	w.Write([]byte(" "))
	for i, item := range decl.List {
		if i != 0 {
			w.Write([]byte(","))
		}
		o.minifyBindingElement(w, item)
	}
}

func (o *Minifier) minifyFuncDecl(w io.Writer, decl js.FuncDecl) {
	if decl.Async {
		w.Write([]byte("async"))
	}
	w.Write([]byte("function"))
	if decl.Generator {
		w.Write([]byte("*"))
	}
	if decl.Name != nil {
		if !decl.Generator {
			w.Write([]byte(" "))
		}
		w.Write(decl.Name)
	}
	o.minifyParams(w, decl.Params)
	o.minifyBlockStmt(w, decl.Body)
}

func (o *Minifier) minifyMethodDecl(w io.Writer, decl js.MethodDecl) {
	if decl.Static {
		w.Write([]byte("static "))
	}
	if decl.Async {
		w.Write([]byte("async"))
		if decl.Generator {
			w.Write([]byte("*"))
		}
	} else if decl.Generator {
		w.Write([]byte("*"))
	} else if decl.Get {
		w.Write([]byte("get "))
	} else if decl.Set {
		w.Write([]byte("set "))
	}
	o.minifyPropertyName(w, decl.Name)
	o.minifyParams(w, decl.Params)
	o.minifyBlockStmt(w, decl.Body)
}

func (o *Minifier) minifyArrowFuncDecl(w io.Writer, decl js.ArrowFuncDecl) {
	if decl.Async {
		w.Write([]byte("async"))
	}
	if decl.Params.Rest == nil && len(decl.Params.List) == 1 && decl.Params.List[0].Default == nil {
		if decl.Async && isIdentStartBindingElement(decl.Params.List[0]) {
			w.Write([]byte(" "))
		}
		o.minifyBindingElement(w, decl.Params.List[0])
	} else {
		o.minifyParams(w, decl.Params)
	}
	w.Write([]byte("=>"))
	if len(decl.Body.List) == 1 {
		if stmt, ok := decl.Body.List[0].(*js.ExprStmt); ok {
			o.minifyExpr(w, stmt.Value)
		} else {
			o.minifyBlockStmt(w, decl.Body)
		}
	} else {
		o.minifyBlockStmt(w, decl.Body)
	}
}

func (o *Minifier) minifyClassDecl(w io.Writer, decl js.ClassDecl) {
	w.Write([]byte("class"))
	if decl.Name != nil {
		w.Write([]byte(" "))
		w.Write(decl.Name)
	}
	if decl.Extends != nil {
		w.Write([]byte(" extends "))
		o.minifyExpr(w, decl.Extends)
	}
	w.Write([]byte("{"))
	for _, item := range decl.Methods {
		o.minifyMethodDecl(w, item)
	}
	w.Write([]byte("}"))
}

func (o *Minifier) minifyPropertyName(w io.Writer, name js.PropertyName) {
	if name.Computed != nil {
		o.minifyExpr(w, name.Computed)
	} else {
		w.Write(name.Literal.Data)
	}
}

func (o *Minifier) minifyProperty(w io.Writer, property js.Property) {
	if property.Key != nil {
		o.minifyPropertyName(w, *property.Key)
		w.Write([]byte(":"))
	} else if property.Spread {
		w.Write([]byte("..."))
	}
	o.minifyExpr(w, property.Value)
	if property.Init != nil {
		w.Write([]byte("="))
		o.minifyExpr(w, property.Init)
	}
}

func (o *Minifier) minifyBindingElement(w io.Writer, element js.BindingElement) {
	if element.Binding != nil {
		o.minifyBinding(w, element.Binding)
		if element.Default != nil {
			w.Write([]byte("="))
			o.minifyExpr(w, element.Default)
		}
	}
}

func (o *Minifier) minifyBinding(w io.Writer, i js.IBinding) {
	switch binding := i.(type) {
	case *js.BindingName:
		w.Write(binding.Data)
	case *js.BindingArray:
		w.Write([]byte("["))
		for _, item := range binding.List {
			o.minifyBindingElement(w, item)
		}
		if binding.Rest != nil {
			w.Write([]byte("..."))
			o.minifyBinding(w, binding.Rest)
		}
		w.Write([]byte("]"))
	case *js.BindingObject:
		w.Write([]byte("{"))
		for _, item := range binding.List {
			if item.Key != nil {
				o.minifyPropertyName(w, *item.Key)
				w.Write([]byte(":"))
			}
			o.minifyBindingElement(w, item.Value)
		}
		if binding.Rest != nil {
			w.Write([]byte("..."))
			w.Write(binding.Rest.Data)
		}
		w.Write([]byte("}"))
	}
}

func (o *Minifier) minifyExpr(w io.Writer, i js.IExpr) {
	switch expr := i.(type) {
	case *js.LiteralExpr:
		if expr.TokenType == js.DecimalToken {
			w.Write(minify.Number(expr.Data, 0))
		} else {
			w.Write(expr.Data)
		}
	case *js.BinaryExpr:
		o.minifyExpr(w, expr.X)
		if expr.Op == js.InstanceofToken || expr.Op == js.InToken {
			if isIdentEndExpr(expr.X) {
				w.Write([]byte(" "))
			}
			w.Write(expr.Op.Bytes())
			if isIdentStartExpr(expr.Y) {
				w.Write([]byte(" "))
			}
		} else {
			w.Write(expr.Op.Bytes())
		}
		o.minifyExpr(w, expr.Y)
	case *js.UnaryExpr:
		if expr.Op == js.PostIncrToken || expr.Op == js.PostDecrToken {
			o.minifyExpr(w, expr.X)
			w.Write(expr.Op.Bytes())
		} else {
			w.Write(expr.Op.Bytes())
			if expr.Op == js.PosToken {
				if unary, ok := expr.X.(*js.UnaryExpr); ok && (unary.Op == js.PosToken || unary.Op == js.PreIncrToken) {
					w.Write([]byte(" "))
				}
			} else if expr.Op == js.NegToken {
				if unary, ok := expr.X.(*js.UnaryExpr); ok && (unary.Op == js.NegToken || unary.Op == js.PreDecrToken) {
					w.Write([]byte(" "))
				}
			}
			o.minifyExpr(w, expr.X)
		}
	case *js.DotExpr:
		o.minifyExpr(w, expr.X)
		w.Write([]byte("."))
		w.Write(expr.Y.Data)
	case *js.GroupExpr:
		w.Write([]byte("("))
		for i, item := range expr.List {
			if i != 0 {
				w.Write([]byte(","))
			}
			o.minifyExpr(w, item)
		}
		if expr.Rest != nil {
			if len(expr.List) != 0 {
				w.Write([]byte(","))
			}
			w.Write([]byte("..."))
			o.minifyBinding(w, expr.Rest)
		}
		w.Write([]byte(")"))
	case *js.ArrayExpr:
		w.Write([]byte("["))
		for i, item := range expr.List {
			if i != 0 {
				w.Write([]byte(","))
			}
			o.minifyExpr(w, item)
		}
		if expr.Rest != nil {
			if len(expr.List) != 0 {
				w.Write([]byte(","))
			}
			w.Write([]byte("..."))
			o.minifyExpr(w, expr.Rest)
		}
		w.Write([]byte("]"))
	case *js.ObjectExpr:
		w.Write([]byte("{"))
		for i, item := range expr.List {
			if i != 0 {
				w.Write([]byte(","))
			}
			o.minifyProperty(w, item)
		}
		w.Write([]byte("}"))
	case *js.TemplateExpr:
		if expr.Tag != nil {
			o.minifyExpr(w, expr.Tag)
		}
		for _, item := range expr.List {
			w.Write(item.Value)
			o.minifyExpr(w, item.Expr)
		}
		w.Write(expr.Tail)
	case *js.NewExpr:
		w.Write([]byte("new"))
		o.minifyExpr(w, expr.X)
	case *js.NewTargetExpr:
		w.Write([]byte("new.target"))
	case *js.YieldExpr:
		w.Write([]byte("yield"))
		if expr.X != nil {
			if expr.Generator {
				w.Write([]byte("*"))
			}
			o.minifyExpr(w, expr.X)
		}
	case *js.CallExpr:
		o.minifyExpr(w, expr.X)
		o.minifyArguments(w, expr.Args)
	case *js.IndexExpr:
		o.minifyExpr(w, expr.X)
		w.Write([]byte("["))
		o.minifyExpr(w, expr.Index)
		w.Write([]byte("]"))
	case *js.ConditionalExpr:
		o.minifyExpr(w, expr.X)
		w.Write([]byte("?"))
		o.minifyExpr(w, expr.Y)
		w.Write([]byte(":"))
		o.minifyExpr(w, expr.Z)
	case *js.OptChainExpr:
		o.minifyExpr(w, expr.X)
		w.Write([]byte("?."))
		o.minifyExpr(w, expr.Y)
	case *js.VarDecl:
		o.minifyVarDecl(w, *expr)
	case *js.FuncDecl:
		o.minifyFuncDecl(w, *expr)
	case *js.ArrowFuncDecl:
		o.minifyArrowFuncDecl(w, *expr)
	case *js.MethodDecl:
		o.minifyMethodDecl(w, *expr)
	case *js.ClassDecl:
		o.minifyClassDecl(w, *expr)
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

func isIdentEndBindingElement(element js.BindingElement) bool {
	if element.Binding != nil {
		if element.Default != nil {
			return isIdentEndExpr(element.Default)
		}
		if _, ok := element.Binding.(*js.BindingName); ok {
			return true
		}
	}
	return false
}

func isIdentStartAlias(alias js.Alias) bool {
	return alias.Name != nil && !bytes.Equal(alias.Name, []byte("*")) || alias.Name == nil && !bytes.Equal(alias.Binding, []byte("*"))
}

func isIdentEndAlias(alias js.Alias) bool {
	return !bytes.Equal(alias.Binding, []byte("*"))
}

func isIdentStartExpr(i js.IExpr) bool {
	switch expr := i.(type) {
	case *js.LiteralExpr:
		return expr.TokenType != js.StringToken && (expr.TokenType != js.DecimalToken || expr.Data[0] != '.')
	case *js.BinaryExpr:
		return isIdentStartExpr(expr.X)
	case *js.UnaryExpr:
		return (expr.Op == js.PostIncrToken || expr.Op == js.PostDecrToken) && isIdentStartExpr(expr.X)
	case *js.ConditionalExpr:
		return isIdentStartExpr(expr.X)
	case *js.OptChainExpr:
		return isIdentStartExpr(expr.X)
	case *js.IndexExpr:
		return isIdentStartExpr(expr.X)
	case *js.DotExpr:
		return isIdentStartExpr(expr.X)
	case *js.CallExpr:
		return isIdentStartExpr(expr.X)
	case *js.GroupExpr:
		return false
	case *js.ArrayExpr:
		return false
	case *js.ObjectExpr:
		return false
	case *js.TemplateExpr:
		return expr.Tag != nil && isIdentStartExpr(expr.Tag)
	case *js.ArrowFuncDecl:
		return expr.Async || expr.Params.Rest == nil && len(expr.Params.List) == 1 && expr.Params.List[0].Default == nil && isIdentStartBindingElement(expr.Params.List[0])
	case *js.MethodDecl:
		return !expr.Generator || expr.Async
	}
	return true
}

func isIdentEndExpr(i js.IExpr) bool {
	switch expr := i.(type) {
	case *js.LiteralExpr:
		return expr.TokenType != js.StringToken
	case *js.BinaryExpr:
		return isIdentEndExpr(expr.Y)
	case *js.UnaryExpr:
		return expr.Op != js.PostIncrToken && expr.Op != js.PostDecrToken && isIdentEndExpr(expr.X)
	case *js.ConditionalExpr:
		return isIdentEndExpr(expr.Z)
	case *js.OptChainExpr:
		return isIdentEndExpr(expr.Y)
	case *js.IndexExpr:
		return false
	case *js.DotExpr:
		return isIdentEndExpr(expr.Y)
	case *js.YieldExpr:
		return expr.X == nil || isIdentEndExpr(expr.X)
	case *js.NewExpr:
		return isIdentEndExpr(expr.X)
	case *js.CallExpr:
		return false
	case *js.GroupExpr:
		return false
	case *js.ArrayExpr:
		return false
	case *js.ObjectExpr:
		return false
	case *js.TemplateExpr:
		return false
	case *js.ArrowFuncDecl:
		// TODO
	case *js.MethodDecl:
		return false
	}
	return true
}
