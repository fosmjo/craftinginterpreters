package astprinter

import (
	"fmt"
	"strings"

	"github.com/fosmjo/lox/parser"
)

// 中序遍历 -> S 表达式
type AstPrinter struct {
}

func (ap *AstPrinter) Print(expr parser.Expr) string {
	return expr.Accept(ap).(string)
}

func (ap *AstPrinter) VisitBinaryExpr(expr parser.BinaryExpr) interface{} {
	return ap.parenthesize(expr.Operator.Lexeme, expr.Left, expr.Right)
}

func (ap *AstPrinter) VisitGroupingExpr(expr parser.GroupingExpr) interface{} {
	return ap.parenthesize("group", expr.Expression)
}

func (ap *AstPrinter) VisitLiteralExpr(expr parser.LiteralExpr) interface{} {
	if expr.Value == nil {
		return "nil"
	}
	if e, ok := expr.Value.(interface{ String() string }); ok {
		return e.String()
	}
	return fmt.Sprintf("%#v", expr.Value)
}

func (ap *AstPrinter) VisitUnaryExpr(expr parser.UnaryExpr) interface{} {
	return ap.parenthesize(expr.Operator.Lexeme, expr.Right)
}

func (ap *AstPrinter) parenthesize(name string, exprs ...parser.Expr) string {
	var builder strings.Builder

	builder.WriteString("(")
	builder.WriteString(name)
	for _, expr := range exprs {
		builder.WriteString(" ")
		builder.WriteString(expr.Accept(ap).(string))
	}
	builder.WriteString(")")

	return builder.String()
}
