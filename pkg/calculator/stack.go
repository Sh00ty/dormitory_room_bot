package calculator

type stack[t any] []t

func (s stack[t]) Top() t {
	return s[len(s)-1]
}

func (s *stack[t]) Pop() {
	*s = (*s)[:len(*s)-1]
}

func (s *stack[t]) Push(item t) {
	*s = append(*s, item)
}
