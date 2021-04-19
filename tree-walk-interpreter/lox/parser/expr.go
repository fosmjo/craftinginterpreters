// Code generated by `go run ../tool/genast.go`; DO NOT EDIT.

package parser

import "github.com/fosmjo/lox/scanner"

type Expr interface {
    Accept(ExprVisitor) interface{}
}

type ExprVisitor interface {
    VisitAssignExpr(*AssignExpr) interface{}
    VisitBinaryExpr(*BinaryExpr) interface{}
    VisitCallExpr(*CallExpr) interface{}
    VisitGetExpr(*GetExpr) interface{}
    VisitGroupingExpr(*GroupingExpr) interface{}
    VisitLiteralExpr(*LiteralExpr) interface{}
    VisitLogicalExpr(*LogicalExpr) interface{}
    VisitSetExpr(*SetExpr) interface{}
    VisitSuperExpr(*SuperExpr) interface{}
    VisitThisExpr(*ThisExpr) interface{}
    VisitUnaryExpr(*UnaryExpr) interface{}
    VisitVariableExpr(*VariableExpr) interface{}
}

type AssignExpr struct {
    Name scanner.Token
    Value Expr
}

func NewAssignExpr(name scanner.Token, value Expr) *AssignExpr {
    return &AssignExpr{
        Name: name,
        Value: value,
    }
}

func (expr *AssignExpr) Accept(visitor ExprVisitor) interface{} {
    return visitor.VisitAssignExpr(expr)
}

type BinaryExpr struct {
    Left Expr
    Operator scanner.Token
    Right Expr
}

func NewBinaryExpr(left Expr, operator scanner.Token, right Expr) *BinaryExpr {
    return &BinaryExpr{
        Left: left,
        Operator: operator,
        Right: right,
    }
}

func (expr *BinaryExpr) Accept(visitor ExprVisitor) interface{} {
    return visitor.VisitBinaryExpr(expr)
}

type CallExpr struct {
    Callee Expr
    Paren scanner.Token
    Arguments []Expr
}

func NewCallExpr(callee Expr, paren scanner.Token, arguments []Expr) *CallExpr {
    return &CallExpr{
        Callee: callee,
        Paren: paren,
        Arguments: arguments,
    }
}

func (expr *CallExpr) Accept(visitor ExprVisitor) interface{} {
    return visitor.VisitCallExpr(expr)
}

type GetExpr struct {
    Object Expr
    Name scanner.Token
}

func NewGetExpr(object Expr, name scanner.Token) *GetExpr {
    return &GetExpr{
        Object: object,
        Name: name,
    }
}

func (expr *GetExpr) Accept(visitor ExprVisitor) interface{} {
    return visitor.VisitGetExpr(expr)
}

type GroupingExpr struct {
    Expression Expr
}

func NewGroupingExpr(expression Expr) *GroupingExpr {
    return &GroupingExpr{
        Expression: expression,
    }
}

func (expr *GroupingExpr) Accept(visitor ExprVisitor) interface{} {
    return visitor.VisitGroupingExpr(expr)
}

type LiteralExpr struct {
    Value interface{}
}

func NewLiteralExpr(value interface{}) *LiteralExpr {
    return &LiteralExpr{
        Value: value,
    }
}

func (expr *LiteralExpr) Accept(visitor ExprVisitor) interface{} {
    return visitor.VisitLiteralExpr(expr)
}

type LogicalExpr struct {
    Left Expr
    Operator scanner.Token
    Right Expr
}

func NewLogicalExpr(left Expr, operator scanner.Token, right Expr) *LogicalExpr {
    return &LogicalExpr{
        Left: left,
        Operator: operator,
        Right: right,
    }
}

func (expr *LogicalExpr) Accept(visitor ExprVisitor) interface{} {
    return visitor.VisitLogicalExpr(expr)
}

type SetExpr struct {
    Object Expr
    Name scanner.Token
    Value Expr
}

func NewSetExpr(object Expr, name scanner.Token, value Expr) *SetExpr {
    return &SetExpr{
        Object: object,
        Name: name,
        Value: value,
    }
}

func (expr *SetExpr) Accept(visitor ExprVisitor) interface{} {
    return visitor.VisitSetExpr(expr)
}

type SuperExpr struct {
    Keyword scanner.Token
    Method scanner.Token
}

func NewSuperExpr(keyword scanner.Token, method scanner.Token) *SuperExpr {
    return &SuperExpr{
        Keyword: keyword,
        Method: method,
    }
}

func (expr *SuperExpr) Accept(visitor ExprVisitor) interface{} {
    return visitor.VisitSuperExpr(expr)
}

type ThisExpr struct {
    Keyword scanner.Token
}

func NewThisExpr(keyword scanner.Token) *ThisExpr {
    return &ThisExpr{
        Keyword: keyword,
    }
}

func (expr *ThisExpr) Accept(visitor ExprVisitor) interface{} {
    return visitor.VisitThisExpr(expr)
}

type UnaryExpr struct {
    Operator scanner.Token
    Right Expr
}

func NewUnaryExpr(operator scanner.Token, right Expr) *UnaryExpr {
    return &UnaryExpr{
        Operator: operator,
        Right: right,
    }
}

func (expr *UnaryExpr) Accept(visitor ExprVisitor) interface{} {
    return visitor.VisitUnaryExpr(expr)
}

type VariableExpr struct {
    Name scanner.Token
}

func NewVariableExpr(name scanner.Token) *VariableExpr {
    return &VariableExpr{
        Name: name,
    }
}

func (expr *VariableExpr) Accept(visitor ExprVisitor) interface{} {
    return visitor.VisitVariableExpr(expr)
}

