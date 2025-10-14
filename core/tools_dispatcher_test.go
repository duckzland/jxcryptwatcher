package core

import (
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
	if d.buffer != 1000 {
		t.Errorf("Expected default buffer size 1000, got %d", d.buffer)
	}
	if d.maxConcurrent != 4 {
		t.Errorf("Expected default maxConcurrent 4, got %d", d.maxConcurrent)
	}
	if d.delay != 16*time.Millisecond {
		t.Errorf("Expected default delay 16ms, got %v", d.delay)
	}
	if d.ctx == nil || d.cancel == nil {
		t.Error("Expected context and cancel to be initialized")
	}

	done := make(chan struct{})
	go func() {
		select {
		case <-d.ctx.Done():
			close(done)
		case <-time.After(100 * time.Millisecond):
			t.Error("Expected context to be canceled")
		}
	}()
	d.cancel()

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Error("Context cancellation did not propagate")
	}
}

func TestNewDispatcher(t *testing.T) {
	d := NewDispatcher(10, 2, 50*time.Millisecond)

	if d.buffer != 10 {
		t.Errorf("Expected buffer size 10, got %d", d.buffer)
	}
	if d.maxConcurrent != 2 {
		t.Errorf("Expected maxConcurrent 2, got %d", d.maxConcurrent)
	}
	if d.delay != 50*time.Millisecond {
		t.Errorf("Expected delay 50ms, got %v", d.delay)
	}
	if d.queue == nil || d.ctx == nil || d.cancel == nil {
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
	var mu sync.Mutex
	called := false

	d.Submit(func() {
		mu.Lock()
		called = true
		mu.Unlock()
	})
	d.Start()

	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	if !called {
		t.Error("Expected submitted function to be called")
	}
	mu.Unlock()
}

func TestDispatcherPauseResume(t *testing.T) {
	d := NewDispatcher(5, 1, 10*time.Millisecond)
	d.Pause()

	if !d.IsPaused() {
		t.Error("Expected dispatcher to be paused")
	}

	d.Resume()

	if d.IsPaused() {
		t.Error("Expected dispatcher to be resumed")
	}
}

func TestDispatcherSetters(t *testing.T) {
	d := &dispatcher{}
	d.SetBufferSize(20)
	d.SetMaxConcurrent(3)
	d.SetDelayBetween(100 * time.Millisecond)

	if d.buffer != 20 {
		t.Errorf("Expected buffer size 20, got %d", d.buffer)
	}
	if d.maxConcurrent != 3 {
		t.Errorf("Expected maxConcurrent 3, got %d", d.maxConcurrent)
	}
	if d.GetDelay() != 100*time.Millisecond {
		t.Errorf("Expected delay 100ms, got %v", d.GetDelay())
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
			_ = d.GetDelay()
			_ = d.IsPaused()
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
