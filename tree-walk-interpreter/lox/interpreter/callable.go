package interpreter

import (
	"time"
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
