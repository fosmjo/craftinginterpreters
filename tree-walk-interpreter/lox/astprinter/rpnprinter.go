package astprinter

import (
	"fmt"
	"strings"

	"github.com/fosmjo/lox/parser"
)

// 后续遍历 -> 逆波兰式
type RPNAstPrinter struct {
}

func (ap *RPNAstPrinter) Print(expr parser.Expr) string {
	return expr.Accept(ap).(string)
}

func (ap *RPNAstPrinter) VisitBinaryExpr(expr parser.BinaryExpr) interface{} {
	return ap.visit(expr.Left, expr.Right) + expr.Operator.Lexeme
}

func (ap *RPNAstPrinter) VisitGroupingExpr(expr parser.GroupingExpr) interface{} {
	return ap.visit(expr.Expression)
}

func (ap *RPNAstPrinter) VisitLiteralExpr(expr parser.LiteralExpr) interface{} {
	if expr.Value == nil {
		return "nil"
	}
	if e, ok := expr.Value.(fmt.Stringer); ok {
		return e.String()
	}

	return fmt.Sprintf("%#v", expr.Value)
}

func (ap *RPNAstPrinter) VisitUnaryExpr(expr parser.UnaryExpr) interface{} {
	return ap.visit(expr.Right) + expr.Operator.Lexeme
}

func (ap *RPNAstPrinter) visit(exprs ...parser.Expr) string {
	var builder strings.Builder

	for _, expr := range exprs {
		builder.WriteString(expr.Accept(ap).(string))
		if len(exprs) > 1 {
			builder.WriteString(" ")
		}
	}

	return builder.String()
}
