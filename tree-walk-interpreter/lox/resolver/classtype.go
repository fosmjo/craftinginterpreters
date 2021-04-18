package resolver

//go:generate stringer -type=ClassType

type ClassType int

const (
	ClassTypeNone ClassType = iota
	ClassTypeClass
)
