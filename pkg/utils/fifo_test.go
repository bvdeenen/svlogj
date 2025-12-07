package utils

import (
	"testing"
)

func TestNewFifo(t *testing.T) {
	size := 5
	f := NewFifo[int](size)
	if f.cap != size {
		t.Errorf("Expected cap %d, got %d", size, f.cap)
	}
	if len(f.fifo) != size {
		t.Errorf("Expected fifo length %d, got %d", size, len(f.fifo))
	}
	if f.head != 0 {
		t.Errorf("Expected head to be 0, got %d", f.head)
	}
}

func check(t testing.TB, val *int, expected int) {
	t.Helper()
	if val == nil {
		t.Errorf("Expected %d, got nil", expected)
	} else if *val != expected {
		t.Errorf("Expected %d, got %v", expected, *val)
	}
}
func check_nil(t testing.TB, val *int) {
	t.Helper()
	if val != nil {
		t.Errorf("Expected nil, got %v", *val)
	}
}
func TestPushWithinSize(t *testing.T) {
	f := NewFifo[int](3)
	f.Push(1)
	f.Push(2)
	f.Push(3)

	// Let's test Get.
	val := f.Get()
	check(t, val, 1)
	val = f.Get()
	check(t, val, 2)
	val = f.Get()
	check(t, val, 3)
	val = f.Get()
	check_nil(t, val)
}

func TestPushBelowSize(t *testing.T) {
	f := NewFifo[int](3)
	f.Push(1)
	f.Push(2)

	// Let's test Get.
	check(t, f.Get(), 1)
	check(t, f.Get(), 2)
	check_nil(t, f.Get())
	f.Push(3)
	f.Push(4)
	check(t, f.Get(), 3)
	check(t, f.Get(), 4)
	check_nil(t, f.Get())
}

func TestPushOverflow(t *testing.T) {
	f := NewFifo[int](3)
	f.Push(1)
	f.Push(2)
	f.Push(3)
	f.Push(4) // Should overwrite 1, now fifo=[4,2,3], head=1, tail=1? Wait.

	check(t, f.Get(), 2)
	check(t, f.Get(), 3)
	check(t, f.Get(), 4)
	check_nil(t, f.Get())
}

func TestGetFromEmpty(t *testing.T) {
	f := NewFifo[int](3)
	check_nil(t, f.Get())
}
