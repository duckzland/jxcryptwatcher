package core

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestDispatcherInit(t *testing.T) {
	d := &dispatcher{}
	d.Init()

	if d.queue == nil {
		t.Error("Expected queue to be initialized")
	}
	if d.buffer.Load() != 100 {
		t.Errorf("Expected default buffer size 100, got %d", d.buffer.Load())
	}
	if d.maxConcurrent.Load() != 4 {
		t.Errorf("Expected default maxConcurrent 4, got %d", d.maxConcurrent.Load())
	}
	if time.Duration(d.delay.Load()) != 16*time.Millisecond {
		t.Errorf("Expected default delay 16ms, got %v", time.Duration(d.delay.Load()))
	}
	if d.ctx.Load() == nil || d.cancel.Load() == nil {
		t.Error("Expected context and cancel to be initialized")
	}

	// cancellation propagation
	done := make(chan struct{})
	go func() {
		ctxPtr := d.ctx.Load().(*context.Context)
		select {
		case <-(*ctxPtr).Done():
			close(done)
		case <-time.After(100 * time.Millisecond):
			t.Error("Expected context to be canceled")
		}
	}()
	cancelPtr := d.cancel.Load().(*context.CancelFunc)
	(*cancelPtr)()

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Error("Context cancellation did not propagate")
	}
}

func TestNewDispatcher(t *testing.T) {
	d := NewDispatcher(10, 2, 50*time.Millisecond)

	if d.buffer.Load() != 10 {
		t.Errorf("Expected buffer size 10, got %d", d.buffer.Load())
	}
	if d.maxConcurrent.Load() != 2 {
		t.Errorf("Expected maxConcurrent 2, got %d", d.maxConcurrent.Load())
	}
	if time.Duration(d.delay.Load()) != 50*time.Millisecond {
		t.Errorf("Expected delay 50ms, got %v", time.Duration(d.delay.Load()))
	}
	if d.queue == nil || d.ctx.Load() == nil || d.cancel.Load() == nil {
		t.Error("Expected internal fields to be initialized")
	}
}

func TestRegisterAndUseDispatcher(t *testing.T) {
	d1 := RegisterDispatcher()
	d2 := UseDispatcher()

	if d1 != d2 {
		t.Error("Expected RegisterDispatcher and UseDispatcher to return the same instance")
	}
}

func TestDispatcherSubmitAndStart(t *testing.T) {
	d := NewDispatcher(5, 1, 10*time.Millisecond)
	d.Init()
	var mu sync.Mutex
	called := false

	d.Start()

	d.Submit(func() {
		mu.Lock()
		called = true
		mu.Unlock()
	})

	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	if !called {
		t.Error("Expected submitted function to be called")
	}
	mu.Unlock()
}

func TestDispatcherSetters(t *testing.T) {
	d := &dispatcher{}
	d.SetBufferSize(20)
	d.SetMaxConcurrent(3)
	d.SetDelayBetween(100 * time.Millisecond)

	if d.buffer.Load() != 20 {
		t.Errorf("Expected buffer size 20, got %d", d.buffer.Load())
	}
	if d.maxConcurrent.Load() != 3 {
		t.Errorf("Expected maxConcurrent 3, got %d", d.maxConcurrent.Load())
	}
	if d.getDelay() != 100*time.Millisecond {
		t.Errorf("Expected delay 100ms, got %v", d.getDelay())
	}
}

func TestDispatcherConcurrentAccess(t *testing.T) {
	d := NewDispatcher(10, 2, 10*time.Millisecond)

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			d.SetBufferSize(i + 1)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			d.SetMaxConcurrent(i + 1)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			_ = d.getDelay()
		}
	}()

	wg.Wait()
}

func TestDispatcherDrain(t *testing.T) {
	d := NewDispatcher(10, 1, 0)

	// Submit 5 dummy functions
	for i := 0; i < 5; i++ {
		d.Submit(func() {})
	}

	drained := false
	d.SetDrainer(func() {
		drained = true
	})

	d.Drain()

	if len(d.queue) != 0 {
		t.Errorf("Expected queue to be empty after drain, got %d", len(d.queue))
	}

	if !drained {
		t.Error("Expected drainer function to be called")
	}
}

func TestDispatcherDestroy(t *testing.T) {
	d := NewDispatcher(10, 2, 50*time.Millisecond)

	for i := 0; i < 3; i++ {
		d.Submit(func() {})
	}

	if d.queue == nil {
		t.Fatal("Expected queue to be initialized before destroy")
	}
	if d.ctx.Load() == nil || d.cancel.Load() == nil {
		t.Fatal("Expected context and cancel to be initialized before destroy")
	}

	d.Destroy()

	if d.queue != nil {
		t.Error("Expected queue to be nil after destroy")
	}

	if ctxPtr, ok := d.ctx.Load().(*context.Context); ok && ctxPtr != nil {
		t.Error("Expected ctx to be nil after destroy")
	}
	if cancelPtr, ok := d.cancel.Load().(*context.CancelFunc); ok && cancelPtr != nil {
		t.Error("Expected cancel to be nil after destroy")
	}
}
