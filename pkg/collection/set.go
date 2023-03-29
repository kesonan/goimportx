package collection

type ArraySet[T comparable] struct {
	list []T
	m    map[T]struct{}
}

func NewArraySet[T comparable]() *ArraySet[T] {
	return &ArraySet[T]{
		m: make(map[T]struct{}),
	}
}

func (s *ArraySet[T]) Add(data T) {
	if s.m == nil {
		s.m = make(map[T]struct{})
	}

	if _, ok := s.m[data]; ok {
		return
	}
	s.m[data] = struct{}{}
	s.list = append(s.list, data)
}

func (s *ArraySet[T]) List() []T {
	return s.list
}
