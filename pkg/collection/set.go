package collection

// ArraySet is a set implementation that uses an array to maintain order.
type ArraySet[T comparable] struct {
	list []T
	m    map[T]struct{}
}

// NewArraySet creates a new ArraySet instance.
func NewArraySet[T comparable]() *ArraySet[T] {
	return &ArraySet[T]{
		m: make(map[T]struct{}),
	}
}

// Add adds an element to the set if it doesn't already exist.
func (s *ArraySet[T]) Add(data T) {
	if s.m == nil {
		s.m = make(map[T]struct{})
	}

	if _, ok := s.m[data]; ok {
		return
	}

	// Add the element to the set and the underlying array.
	s.m[data] = struct{}{}
	s.list = append(s.list, data)
}

// List returns a slice of the set's elements in the order they were added.
func (s *ArraySet[T]) List() []T {
	return s.list
}
