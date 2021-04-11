package parser

type ParseError struct {
	msg string
}

func (pe ParseError) Error() string {
	return pe.msg
}
