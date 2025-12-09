package utils

import (
	"testing"
)

func TestNewFifo(t *testing.T) {
	size := 5
	f := NewFifo[int](size)
	if f.Cap != size {
		t.Errorf("Expected Cap %d, got %d", size, f.Cap)
	}
	if len(f.fifo) != size {
		t.Errorf("Expected fifo length %d, got %d", size, len(f.fifo))
	}
	if f.head != 0 {
		t.Errorf("Expected head to be 0, got %d", f.head)
	}
}

func check(t testing.TB, val int, expected int) {
	t.Helper()
	if val != expected {
		t.Errorf("Expected %d, got %v", expected, val)
	}
}

func TestPushWithinSize(t *testing.T) {
	f := NewFifo[int](3)
	f.Push(1)
	f.Push(2)
	f.Push(3)

	// Let's test Get.
	val, _ := f.Get()
	check(t, val, 1)
	val, _ = f.Get()
	check(t, val, 2)
	val, _ = f.Get()
	check(t, val, 3)
	_, ok := f.Get()
	if ok {
		t.Errorf("Expected false, got true")
	}
}

func TestPushBelowSize(t *testing.T) {
	f := NewFifo[int](3)
	f.Push(1)
	f.Push(2)

	// Let's test Get.
	val, _ := f.Get()
	check(t, val, 1)
	val, _ = f.Get()
	check(t, val, 2)
	_, ok := f.Get()
	if ok {
		t.Errorf("Expected false, got true")
	}
	f.Push(3)
	f.Push(4)
	val, _ = f.Get()
	check(t, val, 3)
	val, _ = f.Get()
	check(t, val, 4)
	_, ok = f.Get()
	if ok {
		t.Errorf("Expected false, got true")
	}
}

func TestPushOverflow(t *testing.T) {
	f := NewFifo[int](3)
	f.Push(1)
	f.Push(2)
	f.Push(3)
	f.Push(4) // Should overwrite 1, now fifo=[4,2,3], head=1, tail=1? Wait.

	val, _ := f.Get()
	check(t, val, 2)
	val, _ = f.Get()
	check(t, val, 3)
	val, _ = f.Get()
	check(t, val, 4)
	_, ok := f.Get()
	if ok {
		t.Errorf("Expected false, got true")
	}
}

func TestGetFromEmpty(t *testing.T) {
	f := NewFifo[int](3)
	_, ok := f.Get()
	if ok {
		t.Errorf("Expected false, got true")
	}
}
