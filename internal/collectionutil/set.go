package collectionutil

type Set[T comparable] struct {
	elements map[T]struct{}
}

// NewSet creates a new set
func NewSet[T comparable](initial ...T) *Set[T] {
	s := &Set[T]{
		elements: make(map[T]struct{}),
	}

	s.Add(initial...)

	return s
}

// Add inserts an element into the set
func (s *Set[T]) Add(values ...T) {
	for _, v := range values {
		s.elements[v] = struct{}{}
	}
}

func (s *Set[T]) AddFromSet(other *Set[T]) {
	for k, _ := range other.elements {
		s.elements[k] = struct{}{}
	}
}

// Remove deletes an element from the set
func (s *Set[T]) Remove(value T) {
	delete(s.elements, value)
}

// Contains checks if an element is in the set
func (s *Set[T]) Contains(value T) bool {
	_, found := s.elements[value]
	return found
}

// Len returns the number of elements in the set
func (s *Set[T]) Len() int {
	return len(s.elements)
}

// List returns all elements in the set as a slice
func (s *Set[T]) List() []T {
	keys := make([]T, 0, len(s.elements))
	for key := range s.elements {
		keys = append(keys, key)
	}
	return keys
}
