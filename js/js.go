// Package js minifies ECMAScript5.1 following the specifications at http://www.ecma-international.org/ecma-262/5.1/.
package js

import (
	"bytes"
	"io"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/js"
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

	buf, writerIsBuffer := w.(*bytes.Buffer)
	if !writerIsBuffer {
		buf = &bytes.Buffer{}
	}
	buf.Grow(z.Len())

	for _, item := range ast.List {
		o.minifyStmt(buf, item)
	}

	if !writerIsBuffer {
		_, err = buf.WriteTo(w)
		return err
	}
	return nil
}

func (o *Minifier) minifyStmt(w *bytes.Buffer, i js.IStmt) {
	switch stmt := i.(type) {
	case *js.ExprStmt:
		o.minifyExpr(w, stmt.Value)
	case *js.VarDecl:
		o.minifyVarDecl(w, *stmt)
	case *js.IfStmt:
		w.WriteString("if(")
		o.minifyExpr(w, stmt.Cond)
		w.WriteString(")")
		o.minifyStmt(w, stmt.Body)
		if stmt.Else != nil {
			w.WriteString("else")
			o.minifyStmt(w, stmt.Else)
		}
	case *js.BlockStmt:
		if len(stmt.List) == 1 {
			o.minifyStmt(w, stmt.List[0])
		} else {
			o.minifyBlockStmt(w, *stmt)
		}
	case *js.ReturnStmt:
		w.WriteString("return")
		if isIdentStartExpr(stmt.Value) {
			w.WriteString(" ")
		}
		o.minifyExpr(w, stmt.Value)
	case *js.LabelledStmt:
		w.Write(stmt.Token.Data)
		w.WriteString(":")
		o.minifyStmt(w, stmt.Value)
	case *js.BranchStmt:
		w.Write(stmt.Type.Bytes())
		if stmt.Name != nil {
			w.WriteString(" ")
			w.Write(stmt.Name.Data)
		}
	case *js.WithStmt:
		w.WriteString("with(")
		o.minifyExpr(w, stmt.Cond)
		w.WriteString(")")
		o.minifyStmt(w, stmt.Body)
	case *js.DoWhileStmt:
		w.WriteString("do")
		o.minifyStmt(w, stmt.Body)
		w.WriteString("while(")
		o.minifyExpr(w, stmt.Cond)
		w.WriteString(")")
	case *js.WhileStmt:
		w.WriteString("while(")
		o.minifyExpr(w, stmt.Cond)
		w.WriteString(")")
		o.minifyStmt(w, stmt.Body)
	case *js.ForStmt:
		w.WriteString("for(")
		o.minifyExpr(w, stmt.Init)
		w.WriteString(";")
		o.minifyExpr(w, stmt.Cond)
		w.WriteString(";")
		o.minifyExpr(w, stmt.Post)
		w.WriteString(")")
		o.minifyStmt(w, stmt.Body)
	case *js.ForInStmt:
		w.WriteString("for(")
		o.minifyExpr(w, stmt.Init)
		if isIdentEndExpr(stmt.Init) {
			w.WriteString(" ")
		}
		w.WriteString("in")
		if isIdentStartExpr(stmt.Value) {
			w.WriteString(" ")
		}
		o.minifyExpr(w, stmt.Value)
		w.WriteString(")")
		o.minifyStmt(w, stmt.Body)
	case *js.ForOfStmt:
		if stmt.Await {
			w.WriteString("for await(")
		} else {
			w.WriteString("for(")
		}
		o.minifyExpr(w, stmt.Init)
		if isIdentEndExpr(stmt.Init) {
			w.WriteString(" ")
		}
		w.WriteString("of")
		if isIdentStartExpr(stmt.Value) {
			w.WriteString(" ")
		}
		o.minifyExpr(w, stmt.Value)
		w.WriteString(")")
		o.minifyStmt(w, stmt.Body)
	case *js.SwitchStmt:
		w.WriteString("switch(")
		o.minifyExpr(w, stmt.Init)
		w.WriteString("){")
		for j, clause := range stmt.List {
			if j != 0 {
				w.WriteString(";")
			}
			w.Write(clause.TokenType.Bytes())
			if clause.Cond != nil {
				w.WriteString(" ")
				o.minifyExpr(w, clause.Cond)
			}
			w.WriteString(":")
			for i, item := range clause.List {
				if i != 0 {
					w.WriteString(";")
				}
				o.minifyStmt(w, item)
			}
		}
		w.WriteString("}")
	case *js.ThrowStmt:
		w.WriteString("throw")
		if isIdentStartExpr(stmt.Value) {
			w.WriteString(" ")
		}
		o.minifyExpr(w, stmt.Value)
	case *js.TryStmt:
		w.WriteString("try")
		o.minifyBlockStmt(w, stmt.Body)
		if len(stmt.Catch.List) != 0 || stmt.Binding != nil {
			w.WriteString("catch")
			if stmt.Binding != nil {
				w.WriteString("(")
				o.minifyBinding(w, stmt.Binding)
				w.WriteString(")")
			}
			o.minifyBlockStmt(w, stmt.Catch)
		}
		if len(stmt.Finally.List) != 0 {
			w.WriteString("finally")
			o.minifyBlockStmt(w, stmt.Finally)
		}
	case *js.FuncDecl:
		o.minifyFuncDecl(w, *stmt)
	case *js.ClassDecl:
		o.minifyClassDecl(w, *stmt)
	case *js.DebuggerStmt:
		w.WriteString("debugger")
	case *js.EmptyStmt:
	case *js.ImportStmt:
		w.WriteString("import")
		if stmt.Default != nil {
			w.WriteString(" ")
			w.Write(stmt.Default)
			if len(stmt.List) != 0 {
				w.WriteString(",")
			}
		}
		if len(stmt.List) == 1 {
			if stmt.Default == nil && isIdentStartAlias(stmt.List[0]) {
				w.WriteString(" ")
			}
			o.minifyAlias(w, stmt.List[0])
		} else if 1 < len(stmt.List) {
			w.WriteString("{")
			for i, item := range stmt.List {
				if i != 0 {
					w.WriteString(",")
				}
				o.minifyAlias(w, item)
			}
			w.WriteString("}")
		}
		if stmt.Default != nil || len(stmt.List) != 0 {
			if len(stmt.List) < 2 {
				w.WriteString(" ")
			}
			w.WriteString("from")
		}
		w.Write(stmt.Module)
	case *js.ExportStmt:
		w.WriteString("export")
		if stmt.Decl != nil {
			if stmt.Default {
				w.WriteString(" default ")
			} else {
				w.WriteString(" ")
			}
			o.minifyExpr(w, stmt.Decl)
		} else {
			if len(stmt.List) == 1 {
				if isIdentStartAlias(stmt.List[0]) {
					w.WriteString(" ")
				}
				o.minifyAlias(w, stmt.List[0])
			} else if 1 < len(stmt.List) {
				w.WriteString("{")
				for i, item := range stmt.List {
					if i != 0 {
						w.WriteString(",")
					}
					o.minifyAlias(w, item)
				}
				w.WriteString("}")
			}
			if stmt.Module != nil {
				if len(stmt.List) < 2 && (len(stmt.List) != 1 || isIdentEndAlias(stmt.List[0])) {
					w.WriteString(" ")
				}
				w.WriteString("from")
				w.Write(stmt.Module)
			}
		}
	}
}

func (o *Minifier) minifyAlias(w *bytes.Buffer, alias js.Alias) {
	if alias.Name != nil {
		w.Write(alias.Name)
		if !bytes.Equal(alias.Name, []byte("*")) {
			w.WriteString(" ")
		}
		w.WriteString("as ")
	}
	w.Write(alias.Binding)
}

func (o *Minifier) minifyBlockStmt(w *bytes.Buffer, stmt js.BlockStmt) {
	w.WriteString("{")
	for i, item := range stmt.List {
		if i != 0 {
			w.WriteString(";")
		}
		o.minifyStmt(w, item)
	}
	w.WriteString("}")
}

func (o *Minifier) minifyParams(w *bytes.Buffer, params js.Params) {
	w.WriteString("(")
	for i, item := range params.List {
		if i != 0 {
			w.WriteString(",")
		}
		o.minifyBindingElement(w, item)
	}
	if params.Rest != nil {
		if len(params.List) != 0 {
			w.WriteString(",")
		}
		w.WriteString("...")
		o.minifyBindingElement(w, *params.Rest)
	}
	w.WriteString(")")
}

func (o *Minifier) minifyArguments(w *bytes.Buffer, args js.Arguments) {
	w.WriteString("(")
	for i, item := range args.List {
		if i != 0 {
			w.WriteString(",")
		}
		o.minifyExpr(w, item)
	}
	if args.Rest != nil {
		if len(args.List) != 0 {
			w.WriteString(",")
		}
		w.WriteString("...")
		o.minifyExpr(w, args.Rest)
	}
	w.WriteString(")")
}

func (o *Minifier) minifyVarDecl(w *bytes.Buffer, decl js.VarDecl) {
	w.Write(decl.TokenType.Bytes())
	w.WriteString(" ")
	for i, item := range decl.List {
		if i != 0 {
			w.WriteString(",")
		}
		o.minifyBindingElement(w, item)
	}
}

func (o *Minifier) minifyFuncDecl(w *bytes.Buffer, decl js.FuncDecl) {
	if decl.Async {
		w.WriteString("async")
	}
	w.WriteString("function")
	if decl.Generator {
		w.WriteString("*")
	}
	if decl.Name != nil {
		if !decl.Generator {
			w.WriteString(" ")
		}
		w.Write(decl.Name)
	}
	o.minifyParams(w, decl.Params)
	o.minifyBlockStmt(w, decl.Body)
}

func (o *Minifier) minifyMethodDecl(w *bytes.Buffer, decl js.MethodDecl) {
	if decl.Static {
		w.WriteString("static ")
	}
	if decl.Async {
		w.WriteString("async")
		if decl.Generator {
			w.WriteString("*")
		}
	} else if decl.Generator {
		w.WriteString("*")
	} else if decl.Get {
		w.WriteString("get ")
	} else if decl.Set {
		w.WriteString("set ")
	}
	o.minifyPropertyName(w, decl.Name)
	o.minifyParams(w, decl.Params)
	o.minifyBlockStmt(w, decl.Body)
}

func (o *Minifier) minifyArrowFuncDecl(w *bytes.Buffer, decl js.ArrowFuncDecl) {
	if decl.Async {
		w.WriteString("async")
	}
	if decl.Params.Rest == nil && len(decl.Params.List) == 1 && decl.Params.List[0].Default == nil {
		if decl.Async && isIdentStartBindingElement(decl.Params.List[0]) {
			w.WriteString(" ")
		}
		o.minifyBindingElement(w, decl.Params.List[0])
	} else {
		o.minifyParams(w, decl.Params)
	}
	w.WriteString("=>")
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

func (o *Minifier) minifyClassDecl(w *bytes.Buffer, decl js.ClassDecl) {
	w.WriteString("class")
	if decl.Name != nil {
		w.WriteString(" ")
		w.Write(decl.Name)
	}
	if decl.Extends != nil {
		w.WriteString(" extends ")
		o.minifyExpr(w, decl.Extends)
	}
	w.WriteString("{")
	for _, item := range decl.Methods {
		o.minifyMethodDecl(w, item)
	}
	w.WriteString("}")
}

func (o *Minifier) minifyPropertyName(w *bytes.Buffer, name js.PropertyName) {
	if name.Computed != nil {
		o.minifyExpr(w, name.Computed)
	} else {
		w.Write(name.Literal.Data)
	}
}

func (o *Minifier) minifyProperty(w *bytes.Buffer, property js.Property) {
	if property.Key != nil {
		o.minifyPropertyName(w, *property.Key)
		w.WriteString(":")
	} else if property.Spread {
		w.WriteString("...")
	}
	o.minifyExpr(w, property.Value)
	if property.Init != nil {
		w.WriteString("=")
		o.minifyExpr(w, property.Init)
	}
}

func (o *Minifier) minifyBindingElement(w *bytes.Buffer, element js.BindingElement) {
	if element.Binding != nil {
		o.minifyBinding(w, element.Binding)
		if element.Default != nil {
			w.WriteString("=")
			o.minifyExpr(w, element.Default)
		}
	}
}

func (o *Minifier) minifyBinding(w *bytes.Buffer, i js.IBinding) {
	switch binding := i.(type) {
	case *js.BindingName:
		w.Write(binding.Data)
	case *js.BindingArray:
		w.WriteString("[")
		for _, item := range binding.List {
			o.minifyBindingElement(w, item)
		}
		if binding.Rest != nil {
			w.WriteString("...")
			o.minifyBinding(w, binding.Rest)
		}
		w.WriteString("]")
	case *js.BindingObject:
		w.WriteString("{")
		for _, item := range binding.List {
			if item.Key != nil {
				o.minifyPropertyName(w, *item.Key)
				w.WriteString(":")
			}
			o.minifyBindingElement(w, item.Value)
		}
		if binding.Rest != nil {
			w.WriteString("...")
			w.Write(binding.Rest.Data)
		}
		w.WriteString("}")
	}
}

func (o *Minifier) minifyExpr(w *bytes.Buffer, i js.IExpr) {
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
				w.WriteString(" ")
			}
			w.Write(expr.Op.Bytes())
			if isIdentStartExpr(expr.Y) {
				w.WriteString(" ")
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
					w.WriteString(" ")
				}
			} else if expr.Op == js.NegToken {
				if unary, ok := expr.X.(*js.UnaryExpr); ok && (unary.Op == js.NegToken || unary.Op == js.PreDecrToken) {
					w.WriteString(" ")
				}
			}
			o.minifyExpr(w, expr.X)
		}
	case *js.DotExpr:
		o.minifyExpr(w, expr.X)
		w.WriteString(".")
		w.Write(expr.Y.Data)
	case *js.GroupExpr:
		w.WriteString("(")
		for i, item := range expr.List {
			if i != 0 {
				w.WriteString(",")
			}
			o.minifyExpr(w, item)
		}
		if expr.Rest != nil {
			if len(expr.List) != 0 {
				w.WriteString(",")
			}
			w.WriteString("...")
			o.minifyBinding(w, expr.Rest)
		}
		w.WriteString(")")
	case *js.ArrayExpr:
		w.WriteString("[")
		for i, item := range expr.List {
			if i != 0 {
				w.WriteString(",")
			}
			o.minifyExpr(w, item)
		}
		if expr.Rest != nil {
			if len(expr.List) != 0 {
				w.WriteString(",")
			}
			w.WriteString("...")
			o.minifyExpr(w, expr.Rest)
		}
		w.WriteString("]")
	case *js.ObjectExpr:
		w.WriteString("{")
		for i, item := range expr.List {
			if i != 0 {
				w.WriteString(",")
			}
			o.minifyProperty(w, item)
		}
		w.WriteString("}")
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
		w.WriteString("new")
		o.minifyExpr(w, expr.X)
	case *js.NewTargetExpr:
		w.WriteString("new.target")
	case *js.YieldExpr:
		w.WriteString("yield")
		if expr.X != nil {
			if expr.Generator {
				w.WriteString("*")
			}
			o.minifyExpr(w, expr.X)
		}
	case *js.CallExpr:
		o.minifyExpr(w, expr.X)
		o.minifyArguments(w, expr.Args)
	case *js.IndexExpr:
		o.minifyExpr(w, expr.X)
		w.WriteString("[")
		o.minifyExpr(w, expr.Index)
		w.WriteString("]")
	case *js.ConditionalExpr:
		o.minifyExpr(w, expr.X)
		w.WriteString("?")
		o.minifyExpr(w, expr.Y)
		w.WriteString(":")
		o.minifyExpr(w, expr.Z)
	case *js.OptChainExpr:
		o.minifyExpr(w, expr.X)
		w.WriteString("?.")
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
