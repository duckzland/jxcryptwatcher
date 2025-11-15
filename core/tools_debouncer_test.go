package core

import (
	"sync"
	"testing"
	"time"
)

func TestDebouncerInit(t *testing.T) {
	d := &debouncer{}
	d.Init()

	if d.generations == nil {
		t.Error("Expected generations map to be initialized")
	}
	if d.cancelMap == nil {
		t.Error("Expected cancelMap to be initialized")
	}
}

func TestRegisterDebouncerSingleton(t *testing.T) {
	d1 := RegisterDebouncer()
	d2 := RegisterDebouncer()

	if d1 != d2 {
		t.Error("Expected RegisterDebouncer to return the same instance")
	}
}

func TestUseDebouncerReturnsInstance(t *testing.T) {
	d1 := RegisterDebouncer()
	d2 := UseDebouncer()

	if d1 != d2 {
		t.Error("Expected UseDebouncer to return the registered instance")
	}
}

func TestDebouncerCall(t *testing.T) {
	d := RegisterDebouncer()
	var mu sync.Mutex
	called := false

	d.Call("test", 100*time.Millisecond, func() {
		mu.Lock()
		called = true
		mu.Unlock()
	})

	time.Sleep(150 * time.Millisecond)

	mu.Lock()
	if !called {
		t.Error("Expected function to be called after delay")
	}
	mu.Unlock()
}

func TestDebouncerCancel(t *testing.T) {
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
	if called {
		t.Error("Expected function to be canceled and not called")
	}
	mu.Unlock()
}

func TestDebouncerMultipleCalls(t *testing.T) {
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
	if callCount != 1 {
		t.Errorf("Expected function to be called once, got %d", callCount)
	}
	mu.Unlock()
}

func TestDebouncerDestroy(t *testing.T) {
	d := RegisterDebouncer()

	for i := 0; i < 3; i++ {
		d.Call("destroyTest", 50*time.Millisecond, func() {})
	}

	if d.generations == nil || d.cancelMap == nil || d.callbacks == nil {
		t.Fatal("Expected debouncer maps to be initialized before destroy")
	}

	d.Destroy()

	if d.generations != nil {
		t.Error("Expected generations map to be nil after destroy")
	}
	if d.cancelMap != nil {
		t.Error("Expected cancelMap to be nil after destroy")
	}
	if d.callbacks != nil {
		t.Error("Expected callbacks map to be nil after destroy")
	}
}
