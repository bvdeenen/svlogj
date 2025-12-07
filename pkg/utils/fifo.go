package utils

type Fifo[T interface{}] struct {
	fifo []T
	tail int
	head int
	cap  int
	fill int
}

func NewFifo[T interface{}](s int) Fifo[T] {
	f := Fifo[T]{
		fifo: make([]T, s),
		tail: 0,
		head: 0,
		cap:  s,
		fill: 0,
	}
	return f
}

func (f *Fifo[T]) Push(i T) {
	f.fifo[f.head] = i
	f.fill += 1
	f.head = (f.head + 1) % f.cap
	if f.fill > f.cap {
		f.tail = (f.tail + 1) % f.cap
		f.fill -= 1
	}
}

func (f *Fifo[T]) Get() *T {
	if f.fill == 0 {
		return nil
	}
	f.fill -= 1
	v := f.fifo[f.tail]
	f.tail = (f.tail + 1) % f.cap
	return &v
}
