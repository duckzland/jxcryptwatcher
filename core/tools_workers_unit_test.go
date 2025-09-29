package core

import (
	"sync"
	"testing"
	"time"
)

func TestWorkerUnitListenerStartAndCall(t *testing.T) {
	called := make(chan bool, 1)

	w := &workerUnit{
		getInterval: nil,
		getDelay:    func() int64 { return 1 },
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
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected listener to call function")
	}
}

func TestWorkerUnitSchedulerTriggers(t *testing.T) {
	var mu sync.Mutex
	count := 0

	w := &workerUnit{
		getDelay:    nil,
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
	time.Sleep(100 * time.Millisecond)
	w.Stop()

	mu.Lock()
	if count == 0 {
		t.Error("Expected scheduler to trigger function calls")
	}
	mu.Unlock()
}

func TestWorkerUnitFlush(t *testing.T) {
	w := &workerUnit{
		queue: make(chan any, 5),
	}
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
	called := make(chan bool, 1)

	w := &workerUnit{
		getInterval: nil,
		getDelay:    func() int64 { return 10 },
		fn: func(payload any) bool {
			called <- true
			return true
		},
		queue: make(chan any, 1),
	}

	w.Start()
	w.Push("before reset")
	w.Reset()
	w.Push("after reset")

	select {
	case <-called:
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected function to be called after reset")
	}

	w.Stop()
}

func TestWorkerUnitGetDelayOverride(t *testing.T) {
	w := &workerUnit{
		getInterval: nil,
		getDelay:    func() int64 { return 25 },
		fn:          func(payload any) bool { return true },
		queue:       make(chan any, 1),
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
