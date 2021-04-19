package interpreter

import "github.com/fosmjo/lox/parser"

type Function struct {
	closure       *Environment
	declaration   *parser.FunctionStmt
	isInitializer bool
}

func NewFunction(declaration *parser.FunctionStmt, closure *Environment, isInitializer bool) *Function {
	return &Function{declaration: declaration, closure: closure, isInitializer: isInitializer}
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
				if f.isInitializer {
					ret = f.closure.GetAt(0, "this")
					return
				}

				ret = v.value
			} else {
				panic(r)
			}
		}
	}()

	interpreter.executeBlock(f.declaration.Body, env)
	if f.isInitializer {
		ret = f.closure.GetAt(0, "this")
	}
	return
}

func (f *Function) String() string {
	return "<fn " + f.declaration.Name.Lexeme + ">"
}

func (f *Function) bind(i *Instance) *Function {
	env := NewEnvironment(WithEnclosing(f.closure))
	env.Define("this", i)
	return NewFunction(f.declaration, env, f.isInitializer)
}
