package core

import (
	"sync"
	"testing"
)

func TestStateManagerInit(t *testing.T) {
	m := NewStateManager(5)
	if m.Get() != 5 {
		t.Fatalf("expected 5, got %d", m.Get())
	}
}

func TestStateManagerChange(t *testing.T) {
	m := NewStateManager(1)
	m.Change(9)
	if m.Get() != 9 {
		t.Fatalf("expected 9, got %d", m.Get())
	}
}

func TestStateManagerIs(t *testing.T) {
	m := NewStateManager(3)
	if !m.Is(3) {
		t.Fatalf("expected Is(3) true")
	}
	if m.Is(4) {
		t.Fatalf("expected Is(4) false")
	}
}

func TestStateManagerCompareAndChange(t *testing.T) {
	m := NewStateManager(10)

	if !m.CompareAndChange(10, 20) {
		t.Fatalf("expected CAS success")
	}
	if m.Get() != 20 {
		t.Fatalf("expected 20, got %d", m.Get())
	}

	if m.CompareAndChange(10, 30) {
		t.Fatalf("expected CAS fail")
	}
	if m.Get() != 20 {
		t.Fatalf("expected 20, got %d", m.Get())
	}
}

func TestStateManagerConcurrent(t *testing.T) {
	m := NewStateManager(0)
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.Change(m.Get() + 1)
		}()
	}

	wg.Wait()
}
