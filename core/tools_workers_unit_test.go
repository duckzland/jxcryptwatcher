package core

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestWorkerUnitListenerStartAndCall(t *testing.T) {
	called := make(chan bool, 1)

	w := &workerUnit{
		getDelay: func() int64 { return 1 },
		fn: func(payload any) bool {
			called <- true
			return true
		},
		queue: make(chan any, 1),
	}

	w.Start()
	if w.ctx == nil || w.cancel == nil {
		t.Error("Expected listener to create cancel context")
	}
	w.Push("test")

	select {
	case <-called:
		w.Stop()
	case <-time.After(200 * time.Millisecond):
		t.Error("Expected listener to call function")
	}
}

func TestWorkerUnitSchedulerTriggers(t *testing.T) {
	var mu sync.Mutex
	count := 0

	w := &workerUnit{
		getInterval: func() int64 { return 20 },
		fn: func(payload any) bool {
			mu.Lock()
			count++
			mu.Unlock()
			return true
		},
		queue: make(chan any, 10),
	}

	w.Start()
	time.Sleep(150 * time.Millisecond)
	w.Stop()

	mu.Lock()
	if count == 0 {
		t.Error("Expected scheduler to trigger function calls")
	}
	mu.Unlock()
}

func TestWorkerUnitFlush(t *testing.T) {
	w := &workerUnit{queue: make(chan any, 5)}
	for i := 0; i < 5; i++ {
		w.queue <- i
	}
	w.Flush()

	select {
	case <-w.queue:
		t.Error("Expected queue to be flushed")
	default:
	}
}

func TestWorkerUnitReset(t *testing.T) {
	called := make(chan bool, 2)

	w := &workerUnit{
		getDelay: func() int64 { return 10 },
		fn: func(payload any) bool {
			called <- true
			return true
		},
		queue: make(chan any, 2),
	}

	w.Start()
	w.Push("before reset")
	w.Reset()
	w.Push("after reset")

	select {
	case <-called:
	case <-time.After(200 * time.Millisecond):
		t.Error("Expected function to be called after reset")
	}

	w.Stop()
}

func TestWorkerUnitGetDelayOverride(t *testing.T) {
	w := &workerUnit{
		getDelay: func() int64 { return 25 },
		fn:       func(payload any) bool { return true },
		queue:    make(chan any, 1),
	}

	w.Start()
	w.mu.Lock()
	delay := w.delay
	w.mu.Unlock()

	if delay != 25*time.Millisecond {
		t.Errorf("Expected delay override to be 25ms, got %v", delay)
	}

	w.Stop()
}

func TestWorkerUnitMultipleResets(t *testing.T) {
	called := 0
	w := &workerUnit{
		getInterval: func() int64 { return 10 },
		fn: func(payload any) bool {
			called++
			return true
		},
		queue: make(chan any, 5),
	}

	for i := 0; i < 3; i++ {
		w.Start()
		w.Push(i)
		w.Reset()
	}

	time.Sleep(100 * time.Millisecond)
	w.Stop()

	if called == 0 {
		t.Error("Expected function to be called across resets")
	}
}

func TestWorkerUnitStopCleansGoroutine(t *testing.T) {
	startG := runtime.NumGoroutine()

	w := &workerUnit{
		getInterval: func() int64 { return 10 },
		fn:          func(payload any) bool { return true },
		queue:       make(chan any, 2),
	}

	w.Start()
	w.Push("payload")
	time.Sleep(50 * time.Millisecond)
	w.Stop()

	// Allow goroutines to settle
	time.Sleep(50 * time.Millisecond)
	endG := runtime.NumGoroutine()

	if endG > startG+2 {
		t.Errorf("Possible goroutine leak: start=%d end=%d", startG, endG)
	}
}

func TestWorkerUnitConcurrentPushAndCall(t *testing.T) {
	var mu sync.Mutex
	count := 0

	w := &workerUnit{
		getDelay: func() int64 { return 1 },
		fn: func(payload any) bool {
			mu.Lock()
			count++
			mu.Unlock()
			return true
		},
		queue: make(chan any, 20),
	}

	w.Start()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			w.Push(i)
		}(i)
	}
	wg.Wait()

	time.Sleep(200 * time.Millisecond)
	w.Stop()

	mu.Lock()
	if count < 10 {
		t.Errorf("Expected at least 10 calls, got %d", count)
	}
	mu.Unlock()
}
