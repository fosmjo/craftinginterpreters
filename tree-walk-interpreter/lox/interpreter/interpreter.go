package interpreter

import (
	"fmt"
	"strings"

	"github.com/fosmjo/lox/parser"
	"github.com/fosmjo/lox/scanner"
)

type Interpreter struct {
	globals *Environment
	env     *Environment
	locals  map[parser.Expr]int
	lox     loxer
}

type loxer interface {
	RuntimeError(err RuntimeError)
}

func NewInterpreter(lox loxer) *Interpreter {
	globals := NewEnvironment()
	globals.Define("clock", clock{})
	locals := make(map[parser.Expr]int)

	return &Interpreter{
		globals: globals,
		env:     globals,
		locals:  locals,
		lox:     lox,
	}
}

func (i *Interpreter) Interpret(stmts []parser.Stmt) {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(RuntimeError); ok {
				i.lox.RuntimeError(err)
			} else {
				panic(r)
			}
		}
	}()

	for _, stmt := range stmts {
		i.execute(stmt)
	}
}

func (i *Interpreter) VisitBinaryExpr(expr *parser.BinaryExpr) interface{} {
	left := i.evaluate(expr.Left)
	right := i.evaluate(expr.Right)

	switch expr.Operator.Type {
	case scanner.GREATER:
		i.checkNumberOperands(expr.Operator, left, right)
		return left.(float64) > right.(float64)
	case scanner.GREATER_EQUAL:
		i.checkNumberOperands(expr.Operator, left, right)
		return left.(float64) >= right.(float64)
	case scanner.LESS:
		i.checkNumberOperands(expr.Operator, left, right)
		return left.(float64) < right.(float64)
	case scanner.LESS_EQUAL:
		i.checkNumberOperands(expr.Operator, left, right)
		return left.(float64) <= right.(float64)
	case scanner.EQUAL_EQUAL:
		return i.isEqual(left, right)
	case scanner.BANG_EQUAL:
		return !i.isEqual(left, right)
	case scanner.MINUS:
		i.checkNumberOperands(expr.Operator, left, right)
		return left.(float64) - right.(float64)
	case scanner.PLUS:
		value1, ok1 := left.(float64)
		value2, ok2 := right.(float64)
		if ok1 && ok2 {
			return value1 + value2
		}

		value3, ok3 := left.(string)
		value4, ok4 := right.(string)
		if ok3 && ok4 {
			return value3 + value4
		}

		err := RuntimeError{token: expr.Operator, msg: "Operands must be two numbers or two strings."}
		panic(err)
	case scanner.SLASH:
		i.checkNumberOperands(expr.Operator, left, right)
		return left.(float64) / right.(float64)
	case scanner.STAR:
		i.checkNumberOperands(expr.Operator, left, right)
		return left.(float64) * right.(float64)
	}

	return nil
}

func (i *Interpreter) VisitCallExpr(expr *parser.CallExpr) interface{} {
	callee := i.evaluate(expr.Callee)

	arguments := make([]interface{}, 0)
	for _, arg := range expr.Arguments {
		arguments = append(arguments, i.evaluate(arg))
	}

	function, ok := callee.(Callable)
	if !ok {
		err := RuntimeError{token: expr.Paren, msg: "Can only call functions and classes."}
		panic(err)
	}
	if len(arguments) != function.Arity() {
		err := RuntimeError{token: expr.Paren, msg: fmt.Sprintf("Expected %d arguments but got %d.", function.Arity(), len(arguments))}
		panic(err)
	}
	return function.Call(i, arguments)
}

func (i *Interpreter) VisitGetExpr(expr *parser.GetExpr) interface{} {
	object := i.evaluate(expr.Object)

	instance, ok := object.(*Instance)
	if !ok {
		err := RuntimeError{token: expr.Name, msg: "Only instances have properties."}
		panic(err)
	}

	return instance.Get(expr.Name)
}

func (i *Interpreter) VisitGroupingExpr(expr *parser.GroupingExpr) interface{} {
	return i.evaluate(expr.Expression)
}

func (i *Interpreter) VisitLiteralExpr(expr *parser.LiteralExpr) interface{} {
	return expr.Value
}

func (i *Interpreter) VisitLogicalExpr(expr *parser.LogicalExpr) interface{} {
	left := i.evaluate(expr.Left)

	if expr.Operator.Type == scanner.OR {
		if i.isTruthy(left) {
			return left
		}
	} else {
		if !i.isTruthy(left) {
			return left
		}
	}

	return i.evaluate(expr.Right)
}

func (i *Interpreter) VisitSetExpr(expr *parser.SetExpr) interface{} {
	object := i.evaluate(expr.Object)

	instance, ok := object.(*Instance)
	if !ok {
		err := RuntimeError{token: expr.Name, msg: "Only instances have fields"}
		panic(err)
	}

	value := i.evaluate(expr.Value)
	instance.Set(expr.Name, value)
	return value
}

func (i *Interpreter) VisitSuperExpr(expr *parser.SuperExpr) interface{} {
	distance := i.locals[expr]
	superclass := i.env.GetAt(distance, "super").(*Class)
	object := i.env.GetAt(distance-1, "this").(*Instance)

	method := superclass.findMethod(expr.Method.Lexeme)
	if method == nil {
		err := RuntimeError{token: expr.Method, msg: "Undefined property '" + expr.Method.Lexeme + "'."}
		panic(err)
	}

	return method.bind(object)
}

func (i *Interpreter) VisitThisExpr(expr *parser.ThisExpr) interface{} {
	return i.lookUpVariable(expr.Keyword, expr)
}

func (i *Interpreter) VisitUnaryExpr(expr *parser.UnaryExpr) interface{} {
	right := i.evaluate(expr.Right)

	switch expr.Operator.Type {
	case scanner.BANG:
		return !i.isTruthy(right)
	case scanner.MINUS:
		return -right.(float64)
	default:
		return nil
	}
}

func (i *Interpreter) VisitVariableExpr(expr *parser.VariableExpr) interface{} {
	return i.lookUpVariable(expr.Name, expr)
}

func (i *Interpreter) VisitAssignExpr(expr *parser.AssignExpr) interface{} {
	value := i.evaluate(expr.Value)

	distance, ok := i.locals[expr]
	if ok {
		i.env.AssignAt(distance, expr.Name, value)
	} else {
		i.globals.Assign(expr.Name, value)
	}

	return value
}

func (i *Interpreter) VisitExpressionStmt(stmt *parser.ExpressionStmt) interface{} {
	i.evaluate(stmt.Expression)
	return nil
}

func (i *Interpreter) VisitFunctionStmt(stmt *parser.FunctionStmt) interface{} {
	function := NewFunction(stmt, i.env, false)
	i.env.Define(stmt.Name.Lexeme, function)
	return nil
}

func (i *Interpreter) VisitIfStmt(stmt *parser.IfStmt) interface{} {
	cond := i.evaluate(stmt.Condition)
	if i.isTruthy(cond) {
		i.execute(stmt.ThenBranch)
	} else if stmt.ElseBranch != nil {
		i.execute(stmt.ElseBranch)
	}
	return nil
}

func (i *Interpreter) VisitPrintStmt(stmt *parser.PrintStmt) interface{} {
	ret := i.evaluate(stmt.Expression)
	fmt.Println(i.stringify(ret))
	return nil
}

func (i *Interpreter) VisitReturnStmt(stmt *parser.ReturnStmt) interface{} {
	var value interface{}
	if stmt.Value != nil {
		value = i.evaluate(stmt.Value)
	}

	ret := Return{value: value}
	panic(ret)
}

func (i *Interpreter) VisitVarStmt(stmt *parser.VarStmt) interface{} {
	var value interface{}
	if stmt.Initializer != nil {
		value = i.evaluate(stmt.Initializer)
	}
	i.env.Define(stmt.Name.Lexeme, value)
	return nil
}

func (i *Interpreter) VisitWhileStmt(stmt *parser.WhileStmt) interface{} {
	for i.isTruthy(i.evaluate(stmt.Condition)) {
		i.execute(stmt.Body)
	}
	return nil
}

func (i *Interpreter) VisitBlockStmt(stmt *parser.BlockStmt) interface{} {
	env := NewEnvironment(WithEnclosing(i.env))
	i.executeBlock(stmt.Statements, env)
	return nil
}

func (i *Interpreter) VisitClassStmt(stmt *parser.ClassStmt) interface{} {
	var superclass *Class

	if stmt.Superclass != nil {
		class := i.evaluate(stmt.Superclass)
		var ok bool
		superclass, ok = class.(*Class)
		if !ok {
			err := RuntimeError{token: stmt.Superclass.Name, msg: "Superclass must be a class."}
			panic(err)
		}
	}

	i.env.Define(stmt.Name.Lexeme, nil)

	if stmt.Superclass != nil {
		i.env = NewEnvironment(WithEnclosing(i.env))
		i.env.Define("super", superclass)
	}

	methods := make(map[string]*Function)
	for _, m := range stmt.Methods {
		fn := NewFunction(m, i.env, m.Name.Lexeme == "init")
		methods[m.Name.Lexeme] = fn
	}

	class := NewClass(stmt.Name.Lexeme, methods, superclass)

	if stmt.Superclass != nil {
		i.env = i.env.enclosing
	}

	i.env.Assign(stmt.Name, class)
	return nil
}

func (i *Interpreter) Resolve(expr parser.Expr, depth int) {
	i.locals[expr] = depth
}

func (i *Interpreter) lookUpVariable(name scanner.Token, expr parser.Expr) interface{} {
	distance, ok := i.locals[expr]
	if ok {
		return i.env.GetAt(distance, name.Lexeme)
	} else {
		return i.globals.Get(name)
	}
}

func (i *Interpreter) evaluate(expr parser.Expr) interface{} {
	return expr.Accept(i)
}

func (i *Interpreter) execute(stmt parser.Stmt) {
	stmt.Accept(i)
}

func (i *Interpreter) executeBlock(stmts []parser.Stmt, env *Environment) {
	preEnv := i.env
	defer func() { i.env = preEnv }()

	i.env = env
	for _, stmt := range stmts {
		i.execute(stmt)
	}
}

func (i *Interpreter) checkNumberOperand(operator scanner.Token, object interface{}) {
	if _, ok := object.(float64); !ok {
		err := RuntimeError{token: operator, msg: "Operand must be a number."}
		panic(err)
	}
}

func (i *Interpreter) checkNumberOperands(operator scanner.Token, left, right interface{}) {
	_, ok1 := left.(float64)
	v2, ok2 := right.(float64)

	if !ok1 || !ok2 {
		err := RuntimeError{token: operator, msg: "Operands must be numbers."}
		panic(err)
	}

	if operator.Type == scanner.SLASH && v2 == 0 {
		err := RuntimeError{token: operator, msg: "Division by zero."}
		panic(err)
	}
}

func (i *Interpreter) isTruthy(object interface{}) bool {
	if object == nil {
		return false
	}

	if ret, ok := object.(bool); ok {
		return ret
	}

	return true
}

func (i *Interpreter) isEqual(a, b interface{}) bool {
	return a == b
}

func (i *Interpreter) stringify(object interface{}) string {
	if object == nil {
		return "nil"
	}

	if v, ok := object.(float64); ok {
		s := fmt.Sprintf("%f", v)
		if strings.HasSuffix(s, ".000000") {
			s = s[0 : len(s)-7]
		}
		return s
	}

	if s, ok := object.(interface{ String() string }); ok {
		return s.String()
	}

	return fmt.Sprintf("%#v", object)
}
