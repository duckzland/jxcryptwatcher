package core

import (
	"testing"
	"time"
)

func resetDebouncer() {
	coreDebouncer = nil
}

func TestDebouncerCallFires(t *testing.T) {
	resetDebouncer()
	d := RegisterDebouncer()

	fired := make(chan struct{}, 1)
	d.Call("test", 20*time.Millisecond, func() { fired <- struct{}{} })

	select {
	case <-fired:
		// success
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected function to be called after delay")
	}

	if _, ok := d.registry.Get("test"); ok {
		t.Error("Expected cancelRegistry entry cleaned after callback")
	}
}

func TestDebouncerCancelPreventsCallback(t *testing.T) {
	resetDebouncer()
	d := RegisterDebouncer()

	fired := make(chan struct{}, 1)
	d.Call("cancelTest", 50*time.Millisecond, func() { fired <- struct{}{} })
	d.Cancel("cancelTest")

	select {
	case <-fired:
		t.Error("Expected function to be canceled and not called")
	case <-time.After(100 * time.Millisecond):
		// success
	}

	if _, ok := d.registry.Get("cancelTest"); ok {
		t.Error("Expected cancelRegistry entry cleaned after cancel")
	}
}

func TestDebouncerMultipleCallsSuppressesEarlier(t *testing.T) {
	resetDebouncer()
	d := RegisterDebouncer()

	fired := make(chan int, 3)
	for i := 1; i <= 3; i++ {
		val := i
		d.Call("multiTest", 50*time.Millisecond, func() { fired <- val })
		time.Sleep(10 * time.Millisecond)
	}

	select {
	case v := <-fired:
		if v != 3 {
			t.Errorf("Expected only last callback to fire, got %d", v)
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("Expected last callback to fire")
	}

	if len(fired) > 1 {
		t.Errorf("Expected at most one callback, got %d", len(fired))
	}
}

func TestDebouncerDestroyCleansMaps(t *testing.T) {
	resetDebouncer()
	d := RegisterDebouncer()

	d.Call("destroyTest", 20*time.Millisecond, func() {})
	d.Destroy()

	if !d.destroyed.Load() {
		t.Error("Expected debouncer destroyed flag set")
	}

	d.registry.data.Range(func(k, v any) bool {
		t.Error("Expected cancelRegistry to be empty after destroy")
		return false
	})

	d.generations.Range(func(k, v any) bool {
		t.Error("Expected generations to be empty after destroy")
		return false
	})
}

func TestDebouncerIgnoresAfterDestroy(t *testing.T) {
	resetDebouncer()
	d := RegisterDebouncer()
	d.Destroy()

	fired := make(chan struct{}, 1)
	d.Call("afterDestroy", 10*time.Millisecond, func() { fired <- struct{}{} })
	d.Cancel("afterDestroy")

	select {
	case <-fired:
		t.Error("Callback should not fire after destroy")
	case <-time.After(50 * time.Millisecond):
		// success
	}

	d.registry.data.Range(func(k, v any) bool {
		t.Error("Expected cancelRegistry to remain empty after destroy")
		return false
	})

	d.generations.Range(func(k, v any) bool {
		t.Error("Expected generations to remain empty after destroy")
		return false
	})
}
