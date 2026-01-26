package core

import (
	"sync"
	"testing"
	"time"
)

func resetDebouncer() {
	coreDebouncer = nil
}

func TestDebouncerInit(t *testing.T) {
	resetDebouncer()
	d := &debouncer{}
	d.Init()

	// sync.Map is always initialized, so just check destroyed flag
	if d.destroyed.Load() {
		t.Error("Expected debouncer not destroyed after init")
	}
}

func TestRegisterDebouncerSingleton(t *testing.T) {
	resetDebouncer()
	d1 := RegisterDebouncer()
	d2 := RegisterDebouncer()
	if d1 != d2 {
		t.Error("Expected RegisterDebouncer to return the same instance")
	}
}

func TestUseDebouncerReturnsInstance(t *testing.T) {
	resetDebouncer()
	d1 := RegisterDebouncer()
	d2 := UseDebouncer()
	if d1 != d2 {
		t.Error("Expected UseDebouncer to return the registered instance")
	}
}

func TestDebouncerCallFires(t *testing.T) {
	resetDebouncer()
	d := RegisterDebouncer()
	var mu sync.Mutex
	called := false

	d.Call("test", 50*time.Millisecond, func() {
		mu.Lock()
		called = true
		mu.Unlock()
	})

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if !called {
		t.Error("Expected function to be called after delay")
	}
	// cancelMap should be cleaned
	if _, ok := d.cancelMap.Load("test"); ok {
		t.Error("Expected cancelMap entry cleaned after callback")
	}
}

func TestDebouncerCancelPreventsCallback(t *testing.T) {
	resetDebouncer()
	d := RegisterDebouncer()
	var mu sync.Mutex
	called := false

	d.Call("cancelTest", 100*time.Millisecond, func() {
		mu.Lock()
		called = true
		mu.Unlock()
	})

	d.Cancel("cancelTest")
	time.Sleep(150 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if called {
		t.Error("Expected function to be canceled and not called")
	}
	if _, ok := d.cancelMap.Load("cancelTest"); ok {
		t.Error("Expected cancelMap entry cleaned after cancel")
	}
}

func TestDebouncerMultipleCallsSuppressesEarlier(t *testing.T) {
	resetDebouncer()
	d := RegisterDebouncer()
	var mu sync.Mutex
	callCount := 0

	for i := 0; i < 3; i++ {
		d.Call("multiTest", 100*time.Millisecond, func() {
			mu.Lock()
			callCount++
			mu.Unlock()
		})
		time.Sleep(50 * time.Millisecond)
	}

	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if callCount > 1 {
		t.Errorf("Expected at most one callback, got %d", callCount)
	}
}

func TestDebouncerDestroyCleansMaps(t *testing.T) {
	resetDebouncer()
	d := RegisterDebouncer()

	d.Call("destroyTest", 50*time.Millisecond, func() {})

	d.Destroy()

	// After destroy, destroyed flag should be true and maps emptied
	if !d.destroyed.Load() {
		t.Error("Expected debouncer destroyed flag set")
	}
	// Check maps are empty
	d.cancelMap.Range(func(k, v any) bool {
		t.Error("Expected cancelMap to be empty after destroy")
		return false
	})
	d.callbacks.Range(func(k, v any) bool {
		t.Error("Expected callbacks to be empty after destroy")
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

	d.Call("afterDestroy", 10*time.Millisecond, func() { t.Error("Callback should not fire after destroy") })
	d.Cancel("afterDestroy")

	time.Sleep(20 * time.Millisecond)

	// Maps should remain empty
	d.cancelMap.Range(func(k, v any) bool {
		t.Error("Expected cancelMap to remain empty after destroy")
		return false
	})
	d.callbacks.Range(func(k, v any) bool {
		t.Error("Expected callbacks to remain empty after destroy")
		return false
	})
	d.generations.Range(func(k, v any) bool {
		t.Error("Expected generations to remain empty after destroy")
		return false
	})
}
