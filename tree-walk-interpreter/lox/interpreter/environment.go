package interpreter

import "github.com/fosmjo/lox/scanner"

type Environment struct {
	vars      map[string]interface{}
	enclosing *Environment
}

type Option func(*Environment)

func NewEnvironment(options ...Option) *Environment {
	vars := make(map[string]interface{})
	env := &Environment{vars: vars}

	for _, option := range options {
		option(env)
	}

	return env
}

func WithEnclosing(enclosing *Environment) Option {
	return func(env *Environment) {
		env.enclosing = enclosing
	}
}

func (e *Environment) Define(name string, value interface{}) {
	e.vars[name] = value
}

func (e *Environment) Get(name scanner.Token) interface{} {
	if v, ok := e.vars[name.Lexeme]; ok {
		return v
	}

	if e.enclosing != nil {
		return e.enclosing.Get(name)
	}

	err := RuntimeError{token: name, msg: "Undefined variable '" + name.Lexeme + "'."}
	panic(err)
}

func (e *Environment) Assign(name scanner.Token, value interface{}) {
	if _, ok := e.vars[name.Lexeme]; ok {
		e.vars[name.Lexeme] = value
		return
	}

	if e.enclosing != nil {
		e.enclosing.Assign(name, value)
		return
	}

	err := RuntimeError{token: name, msg: "Undefined variable '" + name.Lexeme + "'."}
	panic(err)
}
