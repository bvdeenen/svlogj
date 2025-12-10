package utils

import (
	"iter"
	"maps"
)

// Set logical set type.
type Set[T comparable] struct {
	set map[T]struct{}
}

func NewSet[T comparable]() Set[T] {
	return Set[T]{
		set: make(map[T]struct{}),
	}
}

func (s *Set[T]) Entries() iter.Seq[T] {
	return maps.Keys(s.set)
}

// Get returns true if an entry is present in the set
func (s *Set[T]) Get(i T) bool {
	_, ok := s.set[i]
	return ok
}
func (s *Set[T]) Delete(i T) {
	delete(s.set, i)
}

func (s *Set[T]) Add(i T) {
	s.set[i] = struct{}{}
}
func (s *Set[T]) AddMultiple(i []T) {
	for _, v := range i {
		s.set[v] = struct{}{}
	}
}
func (s *Set[T]) Union(other Set[T]) {
	for v := range other.Entries() {
		s.set[v] = struct{}{}
	}
}

// Sub subtract another set from this one.
func (s *Set[T]) Sub(other Set[T]) {
	for v := range other.Entries() {
		delete(s.set, v)
	}
}

// Intersect returns a new set with only the entries in both arguments
func Intersect[T comparable](first Set[T], second Set[T]) Set[T] {
	result := NewSet[T]()
	for v := range first.Entries() {
		if second.Get(v) {
			result.Add(v)
		}
	}
	return result
}
