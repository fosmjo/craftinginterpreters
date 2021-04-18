package interpreter

import "github.com/fosmjo/lox/scanner"

type Instance struct {
	class  *Class
	fields map[string]interface{}
}

func NewInstance(class *Class) *Instance {
	fields := make(map[string]interface{})
	return &Instance{class: class, fields: fields}
}

func (i *Instance) Get(name scanner.Token) interface{} {
	if v, ok := i.fields[name.Lexeme]; ok {
		return v
	}

	method := i.class.findMethod(name.Lexeme)
	if method != nil {
		return method.bind(i)
	}

	err := RuntimeError{token: name, msg: "Undefined property '" + name.Lexeme + "'."}
	panic(err)
}

func (i *Instance) Set(name scanner.Token, value interface{}) {
	i.fields[name.Lexeme] = value
}

func (i *Instance) String() string {
	return i.class.name + " instance"
}
