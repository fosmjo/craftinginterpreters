//go:generate go run ../tool/genast.go

package parser

import (
	"fmt"

	"github.com/fosmjo/lox/scanner"
)

const maxArgumentCount = 255

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

func (p *Parser) Parse() (stmts []Stmt, err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(ParseError); ok {
				err = e
			} else {
				panic(r)
			}
		}
	}()

	stmts = make([]Stmt, 0)
	for !p.isAtEnd() {
		stmts = append(stmts, p.declaration())
	}
	return
}

func (p *Parser) expression() Expr {
	return p.assignment()
}

func (p *Parser) declaration() Stmt {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(ParseError); ok {
				p.synchronize()
			} else {
				panic(r)
			}
		}
	}()

	switch {
	case p.match(scanner.CLASS):
		return p.classDeclaration()
	case p.match(scanner.FUN):
		return p.function("function")
	case p.match(scanner.VAR):
		return p.varDeclaration()
	}

	return p.statement()
}

func (p *Parser) classDeclaration() Stmt {
	name := p.consume(scanner.IDENTIFIER, "Expect class name.")

	var superclass VariableExpr
	if p.match(scanner.LESS) {
		p.consume(scanner.IDENTIFIER, "Expect superclass name.")
		superclass = NewVariableExpr(p.previous())
	}

	p.consume(scanner.LEFT_BRACE, "Expect '{' before class body.")

	methods := make([]FunctionStmt, 0)
	for !p.check(scanner.RIGHT_BRACE) && !p.isAtEnd() {
		methods = append(methods, p.function("method"))
	}

	p.consume(scanner.RIGHT_BRACE, "Expect '}' after class body.")
	return NewClassStmt(name, superclass, methods)
}

func (p *Parser) statement() Stmt {
	switch {
	case p.match(scanner.FOR):
		return p.forStatement()
	case p.match(scanner.IF):
		return p.ifStatement()
	case p.match(scanner.PRINT):
		return p.printStatement()
	case p.match(scanner.RETURN):
		return p.returnStatement()
	case p.match(scanner.WHILE):
		return p.whileStatement()
	case p.match(scanner.LEFT_BRACE):
		return NewBlockStmt(p.block())
	default:
		return p.expressionStatement()
	}
}

func (p *Parser) forStatement() Stmt {
	p.consume(scanner.LEFT_PAREN, "Expect '(' after 'for'.")

	var initializer Stmt
	switch {
	case p.match(scanner.SEMICOLON):
		initializer = nil
	case p.match(scanner.VAR):
		initializer = p.varDeclaration()
	default:
		initializer = p.expressionStatement()
	}

	var condition Expr
	if !p.check(scanner.SEMICOLON) {
		condition = p.expression()
	}
	p.consume(scanner.SEMICOLON, "Expect ';' after loop condition.")

	var increment Expr
	if !p.check(scanner.RIGHT_PAREN) {
		increment = p.expression()
	}
	p.consume(scanner.RIGHT_PAREN, "Expect ')' after for clauses.")

	body := p.statement()

	if increment != nil {
		body = NewBlockStmt([]Stmt{body, NewExpressionStmt(increment)})
	}

	if condition == nil {
		condition = NewLiteralExpr(true)
	}
	body = NewWhileStmt(condition, body)

	if initializer != nil {
		body = NewBlockStmt([]Stmt{initializer, body})
	}

	return body
}

func (p *Parser) ifStatement() Stmt {
	p.consume(scanner.LEFT_PAREN, "Expect '(' after 'if'.")
	condition := p.expression()
	p.consume(scanner.RIGHT_PAREN, "Expect ')' after if condition.")

	thenBranch := p.statement()
	var elseBranch Stmt
	if p.match(scanner.ELSE) {
		elseBranch = p.statement()
	}

	return NewIfStmt(condition, thenBranch, elseBranch)
}

func (p *Parser) printStatement() Stmt {
	expr := p.expression()
	p.consume(scanner.SEMICOLON, "Expect ';' after value.")
	return NewExpressionStmt(expr)
}

func (p *Parser) returnStatement() Stmt {
	keyword := p.previous()
	var value Expr
	if !p.check(scanner.SEMICOLON) {
		value = p.expression()
	}
	p.consume(scanner.SEMICOLON, "Expect ';' after return value.")

	return NewReturnStmt(keyword, value)
}

func (p *Parser) varDeclaration() Stmt {
	name := p.consume(scanner.IDENTIFIER, "Expect variable name.")

	var initializer Expr
	if p.match(scanner.EQUAL) {
		initializer = p.expression()
	}

	p.consume(scanner.SEMICOLON, "Expect ';' after variable declaration.")
	return NewVarStmt(name, initializer)
}

func (p *Parser) whileStatement() Stmt {
	p.consume(scanner.LEFT_PAREN, "Expect '(' after 'while'.")
	condition := p.expression()
	p.consume(scanner.RIGHT_PAREN, "Expect '(' after 'while'.")
	body := p.statement()

	return NewWhileStmt(condition, body)
}

func (p *Parser) expressionStatement() Stmt {
	expr := p.expression()
	p.consume(scanner.SEMICOLON, "Expect ';' after value.")
	return NewPrintStmt(expr)
}

func (p *Parser) function(kind string) FunctionStmt {
	name := p.consume(scanner.IDENTIFIER, "Expect "+kind+" name.")

	p.consume(scanner.LEFT_PAREN, "Expect '(' after "+kind+" name.")
	parameters := make([]scanner.Token, 0)
	if !p.check(scanner.RIGHT_PAREN) {
		parameters = p.addFunctionParameter(parameters)

		for p.match(scanner.COMMA) {
			parameters = p.addFunctionParameter(parameters)
		}
	}
	p.consume(scanner.RIGHT_PAREN, "Expect ')' after parameters.")

	p.consume(scanner.LEFT_BRACE, "Expect '{' before "+kind+" body.")
	body := p.block()
	return NewFunctionStmt(name, parameters, body)
}

func (p *Parser) addFunctionParameter(parameters []scanner.Token) []scanner.Token {
	if len(parameters) >= maxArgumentCount {
		_ = p.error(p.peek(), fmt.Sprintf("Can't have more than %d parameters.", maxArgumentCount))
	}

	param := p.consume(scanner.IDENTIFIER, "Expect parameter name.")
	return append(parameters, param)
}

func (p *Parser) block() []Stmt {
	stmts := make([]Stmt, 0)

	for !p.check(scanner.RIGHT_BRACE) && !p.isAtEnd() {
		stmts = append(stmts, p.declaration())
	}

	p.consume(scanner.RIGHT_BRACE, "Expect '}' after block.")
	return stmts
}

func (p *Parser) assignment() Expr {
	expr := p.or()

	if p.match(scanner.EQUAL) {
		equals := p.previous()
		value := p.assignment()

		if varExpr, ok := expr.(VariableExpr); ok {
			return NewAssignExpr(varExpr.Name, value)
		}

		if getExpr, ok := expr.(GetExpr); ok {
			return NewSetExpr(getExpr.Object, getExpr.Name, value)
		}

		_ = p.error(equals, "Invalid assignment target.")
	}

	return expr
}

func (p *Parser) or() Expr {
	expr := p.and()

	for p.match(scanner.OR) {
		operator := p.previous()
		right := p.and()
		expr = NewLogicalExpr(expr, operator, right)
	}

	return expr
}

func (p *Parser) and() Expr {
	expr := p.equality()

	for p.match(scanner.AND) {
		operator := p.previous()
		right := p.equality()
		expr = NewLogicalExpr(expr, operator, right)
	}

	return expr
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

	return p.call()
}

func (p *Parser) call() Expr {
	expr := p.primary()

	for {
		if p.match(scanner.LEFT_PAREN) {
			expr = p.finishCall(expr)
		} else if p.match(scanner.DOT) {
			name := p.consume(scanner.IDENTIFIER, "Excpect property name after '.'.")
			expr = NewGetExpr(expr, name)
		} else {
			break
		}
	}

	return expr
}

func (p *Parser) finishCall(callee Expr) Expr {
	arguments := make([]Expr, 0)

	if !p.check(scanner.RIGHT_PAREN) {
		arguments = p.addCallArgument(arguments)

		for p.match(scanner.COMMA) {
			arguments = p.addCallArgument(arguments)
		}
	}

	paren := p.consume(scanner.RIGHT_PAREN, "Expect ')' after arguments.")
	return NewCallExpr(callee, paren, arguments)
}

func (p *Parser) addCallArgument(arguments []Expr) []Expr {
	if len(arguments) > maxArgumentCount {
		_ = p.error(p.peek(), fmt.Sprintf("Can't have more than %d arguments.", maxArgumentCount))
	}

	return append(arguments, p.expression())
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
	case p.match(scanner.SUPER):
		keyword := p.previous()
		p.consume(scanner.DOT, "Expect '.' after 'super'.")
		method := p.consume(scanner.IDENTIFIER, "Expect superclass method name.")
		return NewSuperExpr(keyword, method)
	case p.match(scanner.THIS):
		return NewThisExpr(p.previous())
	case p.match(scanner.IDENTIFIER):
		return NewVariableExpr(p.previous())
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

func (p *Parser) synchronize() {
	p.advance()

	for !p.isAtEnd() {
		if p.previous().Type == scanner.SEMICOLON {
			return
		}

		switch p.peek().Type {
		case scanner.CLASS,
			scanner.FUN,
			scanner.VAR,
			scanner.FOR,
			scanner.IF,
			scanner.WHILE,
			scanner.PRINT,
			scanner.RETURN:
			return
		}

		p.advance()
	}
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
