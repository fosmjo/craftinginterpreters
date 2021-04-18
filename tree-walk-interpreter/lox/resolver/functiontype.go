package resolver

//go:generate stringer -type=FunctionType

type FunctionType int

const (
	NONE FunctionType = iota
	FUNCTION
)
