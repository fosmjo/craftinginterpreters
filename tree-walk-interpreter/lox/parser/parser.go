//go:generate go run ../tool/genast.go

package parser

import (
	"github.com/fosmjo/lox/scanner"
)

type Parser struct {
	tokens  []scanner.Token
	current int
	lox     loxer
}

type loxer interface {
	ErrorWithToken(token scanner.Token, msg string)
}

func NewParser(tokens []scanner.Token, lox loxer) *Parser {
	return &Parser{
		tokens:  tokens,
		lox:     lox,
		current: 0,
	}
}

func (p *Parser) Parse() (expr Expr, err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(ParseError); ok {
				err = e
			} else {
				panic(r)
			}
		}
	}()

	expr = p.expression()
	return
}

func (p *Parser) expression() Expr {
	return p.equality()
}

func (p *Parser) equality() Expr {
	expr := p.comparison()

	for p.match(scanner.BANG_EQUAL, scanner.EQUAL_EQUAL) {
		operator := p.previous()
		right := p.comparison()
		expr = NewBinaryExpr(expr, operator, right)
	}

	return expr
}

func (p *Parser) comparison() Expr {
	expr := p.term()

	for p.match(scanner.GREATER, scanner.GREATER_EQUAL, scanner.LESS, scanner.LESS_EQUAL) {
		operator := p.previous()
		right := p.term()
		expr = NewBinaryExpr(expr, operator, right)
	}

	return expr
}

func (p *Parser) term() Expr {
	expr := p.factor()

	for p.match(scanner.MINUS, scanner.PLUS) {
		operator := p.previous()
		right := p.factor()
		expr = NewBinaryExpr(expr, operator, right)
	}

	return expr
}

func (p *Parser) factor() Expr {
	expr := p.unary()

	for p.match(scanner.SLASH, scanner.STAR) {
		operator := p.previous()
		right := p.unary()
		expr = NewBinaryExpr(expr, operator, right)
	}

	return expr
}

func (p *Parser) unary() Expr {
	if p.match(scanner.MINUS, scanner.BANG) {
		operator := p.previous()
		right := p.unary()
		return NewUnaryExpr(operator, right)
	}

	return p.primary()
}

func (p *Parser) primary() Expr {
	switch {
	case p.match(scanner.FALSE):
		return NewLiteralExpr(false)
	case p.match(scanner.TRUE):
		return NewLiteralExpr(true)
	case p.match(scanner.NIL):
		return NewLiteralExpr(nil)
	case p.match(scanner.NUMBER, scanner.STRING):
		return NewLiteralExpr(p.previous().Literal)
	case p.match(scanner.LEFT_PAREN):
		expr := p.expression()
		p.consume(scanner.RIGHT_PAREN, "Expect ')' after expression.")
		return NewGroupingExpr(expr)
	default:
		err := p.error(p.peek(), "Expect expression.")
		panic(err)
	}
}

func (p *Parser) match(types ...scanner.TokenType) bool {
	for _, t := range types {
		if p.check(t) {
			p.advance()
			return true
		}
	}
	return false
}

func (p *Parser) consume(t scanner.TokenType, msg string) scanner.Token {
	if p.check(t) {
		return p.advance()
	}
	err := p.error(p.peek(), msg)
	panic(err)
}

func (p *Parser) error(token scanner.Token, msg string) ParseError {
	p.lox.ErrorWithToken(token, msg)
	return ParseError{msg: msg}
}

func (p *Parser) advance() scanner.Token {
	if !p.isAtEnd() {
		p.current++
	}
	return p.previous()
}

func (p *Parser) check(t scanner.TokenType) bool {
	if p.isAtEnd() {
		return false
	}

	return p.peek().Type == t
}

func (p *Parser) isAtEnd() bool {
	return p.peek().Type == scanner.EOF
}

func (p *Parser) peek() scanner.Token {
	return p.tokens[p.current]
}

func (p *Parser) previous() scanner.Token {
	return p.tokens[p.current-1]
}
