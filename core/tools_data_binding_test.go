package core

import (
	"sync"
	"testing"
	"time"

	"fyne.io/fyne/v2/data/binding"
)

type testListener struct {
	called int
	mu     sync.Mutex
}

func (t *testListener) DataChanged() {
	t.mu.Lock()
	t.called++
	t.mu.Unlock()
}

func (t *testListener) Count() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.called
}

func TestNewDataBinding(t *testing.T) {
	db := NewDataBinding("hello", 10)

	if db.GetData() != "hello" {
		t.Fatalf("expected 'hello', got '%s'", db.GetData())
	}

	if db.GetStatus() != 10 {
		t.Fatalf("expected 10, got %d", db.GetStatus())
	}
}

func TestSetDataNotifies(t *testing.T) {
	db := NewDataBinding("", 0)

	l := &testListener{}
	db.AddListener(binding.NewDataListener(l.DataChanged))

	db.SetData("abc")
	time.Sleep(10 * time.Millisecond)

	if l.Count() != 1 {
		t.Fatalf("listener fired %d times", l.Count())
	}

	if db.GetData() != "abc" {
		t.Fatalf("expected 'abc', got '%s'", db.GetData())
	}
}

func TestSetStatusNotifies(t *testing.T) {
	db := NewDataBinding("", 0)

	l := &testListener{}
	db.AddListener(binding.NewDataListener(l.DataChanged))

	db.SetStatus(99)
	time.Sleep(10 * time.Millisecond)

	if l.Count() != 1 {
		t.Fatalf("listener fired %d times", l.Count())
	}

	if db.GetStatus() != 99 {
		t.Fatalf("expected 99, got %d", db.GetStatus())
	}
}

func TestRemoveListener(t *testing.T) {
	db := NewDataBinding("", 0)

	l1 := &testListener{}
	l2 := &testListener{}

	w1 := binding.NewDataListener(l1.DataChanged)
	w2 := binding.NewDataListener(l2.DataChanged)

	db.AddListener(w1)
	db.AddListener(w2)
	db.RemoveListener(w1)

	db.SetData("x")
	time.Sleep(10 * time.Millisecond)

	if l1.Count() != 0 {
		t.Fatalf("removed listener fired")
	}

	if l2.Count() != 1 {
		t.Fatalf("remaining listener fired %d times", l2.Count())
	}
}

func TestConcurrentAccess(t *testing.T) {
	db := NewDataBinding("start", 1)

	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			db.SetData("val")
			db.SetStatus(i)
		}(i)
	}

	wg.Wait()
}
