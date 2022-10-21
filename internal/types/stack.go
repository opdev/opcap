package types

import "errors"

type Stack[T any] struct {
	stack *element[T]
}

type element[T any] struct {
	previous *element[T]
	val      T
}

var StackEmptyError = errors.New("stack empty")

func (s *Stack[T]) Push(v T) {
	e := &element[T]{
		previous: s.stack,
		val:      v,
	}
	s.stack = e
}

func (s *Stack[T]) Pop() (T, error) {
	if s.stack == nil {
		var r T
		return r, StackEmptyError
	}
	e := *s.stack
	s.stack = e.previous

	return e.val, nil
}

func (s *Stack[T]) Empty() bool {
	return s.stack == nil
}
