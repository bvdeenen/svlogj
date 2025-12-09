package utils

import "errors"

var empty = errors.New("empty")

type Fifo[T interface{}] struct {
	fifo []T
	tail int
	head int
	Cap  int
	Fill int
}

func NewFifo[T interface{}](s int) Fifo[T] {
	f := Fifo[T]{
		fifo: make([]T, s),
		tail: 0,
		head: 0,
		Cap:  s,
		Fill: 0,
	}
	return f
}

func (f *Fifo[T]) Push(i T) {
	f.fifo[f.head] = i
	f.Fill += 1
	f.head = (f.head + 1) % f.Cap
	if f.Fill > f.Cap {
		f.tail = (f.tail + 1) % f.Cap
		f.Fill -= 1
	}
}

func (f *Fifo[T]) Get() (T, bool) {
	if f.Fill == 0 {
		var result T
		return result, false
	}
	f.Fill -= 1
	v := f.fifo[f.tail]
	f.tail = (f.tail + 1) % f.Cap
	return v, true
}
