package interpreter

import (
	"time"

	"github.com/fosmjo/lox/parser"
)

type Callable interface {
	Arity() int
	Call(interpreter *Interpreter, arguments []interface{}) interface{}
	String() string
}

type clock struct{}

func (clock) Arity() int {
	return 0
}

func (clock) Call(interpreter *Interpreter, arguments []interface{}) interface{} {
	return time.Now()
}

func (clock) String() string {
	return "<native fn>"
}

type Function struct {
	closure     *Environment
	declaration parser.FunctionStmt
}

func NewFunction(declaration parser.FunctionStmt, closure *Environment) *Function {
	return &Function{declaration: declaration, closure: closure}
}

func (f *Function) Arity() int {
	return len(f.declaration.Params)
}

func (f *Function) Call(interpreter *Interpreter, arguments []interface{}) (ret interface{}) {
	env := NewEnvironment(WithEnclosing(f.closure))
	for i := 0; i < len(f.declaration.Params); i++ {
		env.Define(f.declaration.Params[i].Lexeme, arguments[i])
	}

	defer func() {
		if r := recover(); r != nil {
			if v, ok := r.(Return); ok {
				ret = v.value
			} else {
				panic(r)
			}
		}
	}()

	interpreter.executeBlock(f.declaration.Body, env)
	return
}

func (f *Function) String() string {
	return "<fn " + f.declaration.Name.Lexeme + ">"
}
