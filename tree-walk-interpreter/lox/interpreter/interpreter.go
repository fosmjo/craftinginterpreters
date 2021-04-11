package interpreter

import (
	"fmt"
	"strings"

	"github.com/fosmjo/lox/parser"
	"github.com/fosmjo/lox/scanner"
)

type Interpreter struct {
	lox loxer
}

type loxer interface {
	RuntimeError(err RuntimeError)
}

func New(lox loxer) *Interpreter {
	return &Interpreter{lox: lox}
}

func (i *Interpreter) Interpret(expr parser.Expr) {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(RuntimeError); ok {
				i.lox.RuntimeError(err)
			} else {
				panic(r)
			}
		}
	}()

	result := i.evaluate(expr)
	fmt.Println(i.stringify(result))
}

func (i *Interpreter) VisitBinaryExpr(expr parser.BinaryExpr) interface{} {
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

func (i *Interpreter) VisitGroupingExpr(expr parser.GroupingExpr) interface{} {
	return i.evaluate(expr.Expression)
}

func (i *Interpreter) VisitLiteralExpr(expr parser.LiteralExpr) interface{} {
	return expr.Value
}

func (i *Interpreter) VisitUnaryExpr(expr parser.UnaryExpr) interface{} {
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

func (i *Interpreter) evaluate(expr parser.Expr) interface{} {
	return expr.Accept(i)
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
