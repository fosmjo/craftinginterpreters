package resolver

import (
	"github.com/fosmjo/lox/interpreter"
	"github.com/fosmjo/lox/parser"
	"github.com/fosmjo/lox/scanner"
)

type Resolver struct {
	interpreter     *interpreter.Interpreter
	scopes          *stack
	currentFunction FunctionType
	currentClass    ClassType
	lox             loxer
}

type loxer interface {
	ErrorWithToken(token scanner.Token, msg string)
}

func NewResolver(interpreter *interpreter.Interpreter, lox loxer) *Resolver {
	stack := NewStack()
	return &Resolver{
		interpreter:     interpreter,
		scopes:          stack,
		currentFunction: FunctionTypeNone,
		currentClass:    ClassTypeNone,
		lox:             lox,
	}
}

func (r *Resolver) VisitBlockStmt(stmt parser.BlockStmt) interface{} {
	r.beginScope()
	r.Resolve(stmt.Statements)
	r.endScope()
	return nil
}

func (r *Resolver) VisitClassStmt(stmt parser.ClassStmt) interface{} {
	enclosingClass := r.currentClass
	r.currentClass = ClassTypeClass

	r.declare(stmt.Name)
	r.define(stmt.Name)

	r.beginScope()
	r.scopes.Peek()["this"] = true

	for _, m := range stmt.Methods {
		funType := FunctionTypeMethod
		if m.Name.Lexeme == "init" {
			funType = FunctionTypeInitializer
		}
		r.resolveFunction(m, funType)
	}

	r.endScope()
	r.currentClass = enclosingClass
	return nil
}

func (r *Resolver) VisitExpressionStmt(stmt parser.ExpressionStmt) interface{} {
	r.resolveExpr(stmt.Expression)
	return nil
}

func (r *Resolver) VisitFunctionStmt(stmt parser.FunctionStmt) interface{} {
	r.declare(stmt.Name)
	r.define(stmt.Name)
	r.resolveFunction(stmt, FunctionTypeFunction)
	return nil
}

func (r *Resolver) VisitIfStmt(stmt parser.IfStmt) interface{} {
	r.resolveExpr(stmt.Condition)
	r.resolveStmt(stmt.ThenBranch)
	if stmt.ElseBranch != nil {
		r.resolveStmt(stmt.ElseBranch)
	}
	return nil
}

func (r *Resolver) VisitPrintStmt(stmt parser.PrintStmt) interface{} {
	r.resolveExpr(stmt.Expression)
	return nil
}

func (r *Resolver) VisitReturnStmt(stmt parser.ReturnStmt) interface{} {
	if r.currentFunction == FunctionTypeNone {
		r.lox.ErrorWithToken(stmt.Keyword, "Can't return from top-level code.")
	}

	if stmt.Value != nil {
		if r.currentFunction == FunctionTypeInitializer {
			r.lox.ErrorWithToken(stmt.Keyword, "Can't return a value from an initializer.")
		}
		r.resolveExpr(stmt.Value)
	}
	return nil
}

func (r *Resolver) VisitVarStmt(stmt parser.VarStmt) interface{} {
	r.declare(stmt.Name)
	if stmt.Initializer != nil {
		r.resolveExpr(stmt.Initializer)
	}
	r.define(stmt.Name)
	return nil
}

func (r *Resolver) VisitWhileStmt(stmt parser.WhileStmt) interface{} {
	r.resolveExpr(stmt.Condition)
	r.resolveStmt(stmt.Body)
	return nil
}

func (r *Resolver) VisitAssignExpr(expr parser.AssignExpr) interface{} {
	r.resolveExpr(expr.Value)
	r.resolveLocal(expr, expr.Name)
	return nil
}

func (r *Resolver) VisitBinaryExpr(expr parser.BinaryExpr) interface{} {
	r.resolveExpr(expr.Left)
	r.resolveExpr(expr.Right)
	return nil
}

func (r *Resolver) VisitCallExpr(expr parser.CallExpr) interface{} {
	r.resolveExpr(expr.Callee)
	for _, arg := range expr.Arguments {
		r.resolveExpr(arg)
	}
	return nil
}

func (r *Resolver) VisitGetExpr(expr parser.GetExpr) interface{} {
	r.resolveExpr(expr.Object)
	return nil
}

func (r *Resolver) VisitGroupingExpr(expr parser.GroupingExpr) interface{} {
	r.resolveExpr(expr.Expression)
	return nil
}

func (r *Resolver) VisitLiteralExpr(expr parser.LiteralExpr) interface{} {
	return nil
}

func (r *Resolver) VisitLogicalExpr(expr parser.LogicalExpr) interface{} {
	r.resolveExpr(expr.Left)
	r.resolveExpr(expr.Right)
	return nil
}

func (r *Resolver) VisitSetExpr(expr parser.SetExpr) interface{} {
	r.resolveExpr(expr.Value)
	r.resolveExpr(expr.Object)
	return nil
}

func (r *Resolver) VisitThisExpr(expr parser.ThisExpr) interface{} {
	if r.currentClass == ClassTypeNone {
		r.lox.ErrorWithToken(expr.Keyword, "Can't use 'this' outside of a class.")
		return nil
	}

	r.resolveLocal(expr, expr.Keyword)
	return nil
}

func (r *Resolver) VisitUnaryExpr(expr parser.UnaryExpr) interface{} {
	r.resolveExpr(expr.Right)
	return nil
}

func (r *Resolver) VisitVariableExpr(expr parser.VariableExpr) interface{} {
	if !r.scopes.IsEmpty() {
		if v, ok := r.scopes.Peek()[expr.Name.Lexeme]; ok && !v {
			r.lox.ErrorWithToken(expr.Name, "Can't read local variable in its own initializer.")
		}
	}
	r.resolveLocal(expr, expr.Name)
	return nil
}

func (r *Resolver) beginScope() {
	r.scopes.Push(make(scope))
}

func (r *Resolver) endScope() {
	r.scopes.Pop()
}

func (r *Resolver) declare(name scanner.Token) {
	if r.scopes.IsEmpty() {
		return
	}

	scope := r.scopes.Peek()
	if _, ok := scope[name.Lexeme]; ok {
		r.lox.ErrorWithToken(name, "Already variable with this name in this scope.")
	}

	scope[name.Lexeme] = false
}

func (r *Resolver) define(name scanner.Token) {
	if r.scopes.IsEmpty() {
		return
	}

	scope := r.scopes.Peek()
	scope[name.Lexeme] = true
}

func (r *Resolver) resolveLocal(expr parser.Expr, name scanner.Token) {
	for i := r.scopes.Size() - 1; i >= 0; i-- {
		if _, ok := r.scopes.Get(i)[name.Lexeme]; ok {
			r.interpreter.Resolve(expr, r.scopes.Size()-1-i)
			return
		}
	}
}

func (r *Resolver) Resolve(stmts []parser.Stmt) {
	for _, stmt := range stmts {
		r.resolveStmt(stmt)
	}
}

func (r *Resolver) resolveFunction(stmt parser.FunctionStmt, funType FunctionType) {
	enclosingFunction := r.currentFunction
	r.currentFunction = funType

	r.beginScope()
	for _, param := range stmt.Params {
		r.declare(param)
		r.define(param)
	}
	r.Resolve(stmt.Body)
	r.endScope()
	r.currentFunction = enclosingFunction
}

func (r *Resolver) resolveStmt(stmt parser.Stmt) {
	stmt.Accept(r)
}

func (r *Resolver) resolveExpr(expr parser.Expr) {
	expr.Accept(r)
}
