package resolver

type stack struct {
	scopes []scope
}

type scope map[string]bool

func NewStack() *stack {
	scopes := make([]scope, 0)
	return &stack{scopes: scopes}
}

func (s *stack) Push(sc scope) {
	s.scopes = append(s.scopes, sc)
}

func (s *stack) Pop() {
	s.scopes = s.scopes[:len(s.scopes)-1]
}

func (s *stack) Peek() scope {
	return s.Get(len(s.scopes) - 1)
}

func (s *stack) Get(index int) scope {
	if index >= s.Size() {
		panic("index out of range")
	}
	return s.scopes[index]
}

func (s *stack) Size() int {
	return len(s.scopes)
}

func (s *stack) IsEmpty() bool {
	return s.Size() == 0
}
