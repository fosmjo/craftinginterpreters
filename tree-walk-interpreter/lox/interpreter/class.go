package interpreter

type Class struct {
	name       string
	methods    map[string]*Function
	superclass *Class
}

func NewClass(name string, methods map[string]*Function, superclass *Class) *Class {
	return &Class{name: name, methods: methods, superclass: superclass}
}

func (c *Class) Arity() int {
	initializer := c.findMethod("init")
	if initializer == nil {
		return 0
	}
	return initializer.Arity()
}

func (c *Class) Call(interpreter *Interpreter, arguments []interface{}) interface{} {
	ins := NewInstance(c)
	initializer := c.findMethod("init")
	if initializer != nil {
		initializer.bind(ins).Call(interpreter, arguments)
	}
	return ins
}

func (c *Class) String() string {
	return c.name
}

func (c *Class) findMethod(name string) *Function {
	if m, ok := c.methods[name]; ok {
		return m
	}
	if c.superclass != nil {
		return c.superclass.findMethod(name)
	}
	return nil
}
