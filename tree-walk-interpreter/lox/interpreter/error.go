package interpreter

import "github.com/fosmjo/lox/scanner"

type RuntimeError struct {
	token scanner.Token
	msg   string
}

func (re RuntimeError) Token() scanner.Token {
	return re.token
}

func (re RuntimeError) Error() string {
	return re.msg + re.token.Lexeme
}
