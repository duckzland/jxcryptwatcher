package core

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestWorkerUnitListenerStartAndCall(t *testing.T) {
	called := make(chan bool, 1)

	w := NewWorkerUnit(
		1,
		func() int64 { return 1 },
		nil,
		func(payload any) bool {
			called <- true
			return true
		},
	)

	w.Start()

	// ctx and registry should not be nil
	if w.ctx.Load() == nil || !w.registry.Exists("worker") {
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

	w := NewWorkerUnit(
		10,
		nil,
		func() int64 { return 20 },
		func(payload any) bool {
			mu.Lock()
			count++
			mu.Unlock()
			return true
		},
	)

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
	w := NewWorkerUnit(5, nil, nil, func(any) bool { return true })
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

	w := NewWorkerUnit(
		2,
		func() int64 { return 10 },
		nil,
		func(payload any) bool {
			called <- true
			return true
		},
	)

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
	w := NewWorkerUnit(
		1,
		func() int64 { return 25 },
		nil,
		func(any) bool { return true },
	)

	w.Start()
	delay := w.delay.Load()

	if delay != 25 {
		t.Errorf("Expected delay override to be 25ms, got %v", delay)
	}

	w.Stop()
}

func TestWorkerUnitMultipleResets(t *testing.T) {
	called := 0
	w := NewWorkerUnit(
		5,
		nil,
		func() int64 { return 10 },
		func(payload any) bool {
			called++
			return true
		},
	)

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

	w := NewWorkerUnit(
		2,
		nil,
		func() int64 { return 10 },
		func(any) bool { return true },
	)

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

	w := NewWorkerUnit(
		20,
		func() int64 { return 1 },
		nil,
		func(payload any) bool {
			mu.Lock()
			count++
			mu.Unlock()
			return true
		},
	)

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
