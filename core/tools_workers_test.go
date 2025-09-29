package core

import (
	"testing"
	"time"
)

func stopWorker(w *worker, key string, delayMs int64) {
	time.Sleep(time.Duration(delayMs) * time.Millisecond)
	w.mu.Lock()
	unit := w.registry[key]
	w.mu.Unlock()
	if unit != nil {
		unit.Stop()
	}
}

func TestWorkerInit(t *testing.T) {
	w := &worker{}
	w.Init()
	if w.registry == nil || w.conditions == nil || w.lastRun == nil {
		t.Error("Expected internal maps to be initialized")
	}
}

func TestWorkerSingleton(t *testing.T) {
	w1 := RegisterWorkerManager()
	w2 := UseWorker()
	if w1 != w2 {
		t.Error("Expected singleton instance")
	}
}

func TestWorkerCallImmediate(t *testing.T) {
	called := false
	w := &worker{}
	w.Init()
	w.Register("call_immediate", 1,
		func() int64 { return 10 },
		nil,
		func(payload any) bool {
			called = true
			return true
		},
		func() bool { return true },
	)
	w.Call("call_immediate", CallImmediate)
	if !called {
		t.Error("Expected function to be called immediately")
	}
	stopWorker(w, "call_immediate", 20)
}

func TestWorkerCallDebounced(t *testing.T) {
	var callCount int
	RegisterDebouncer().Init()
	w := &worker{}
	w.Init()
	w.Register("debounced", 1,
		nil,
		func() int64 { return 50 },
		func(payload any) bool {
			callCount++
			return true
		},
		func() bool { return true },
	)
	w.Call("debounced", CallDebounced)
	w.Call("debounced", CallDebounced)
	w.Call("debounced", CallDebounced)
	time.Sleep(60 * time.Millisecond)
	if callCount != 1 {
		t.Errorf("Expected debounced call once, got %d", callCount)
	}
	stopWorker(w, "debounced", 100)
}

func TestWorkerCallBypassImmediateIgnoresCondition(t *testing.T) {
	called := false
	w := &worker{}
	w.Init()
	w.Register("bypass", 1,
		nil,
		func() int64 { return 10 },
		func(payload any) bool {
			called = true
			return true
		},
		func() bool { return false },
	)
	w.Call("bypass", CallBypassImmediate)
	if !called {
		t.Error("Expected function to be called despite false condition")
	}
	stopWorker(w, "bypass", 20)
}

func TestWorkerPushFlushReset(t *testing.T) {
	called := false
	w := &worker{}
	w.Init()
	w.Register("reset", 1,
		func() int64 { return 10 },
		nil,
		func(payload any) bool {
			called = true
			return true
		},
		func() bool { return true },
	)
	w.Push("reset", "payload")
	w.Flush("reset")
	w.Reset("reset")
	w.Push("reset", "payload")
	time.Sleep(15 * time.Millisecond)
	if !called {
		t.Error("Expected function to be called after reset and push")
	}
	stopWorker(w, "reset", 30)
}

func TestWorkerPauseResumeReload(t *testing.T) {
	called := false
	w := &worker{}
	w.Init()
	w.Register("reload", 1,
		func() int64 { return 10 },
		nil,
		func(payload any) bool {
			called = true
			return true
		},
		func() bool { return true },
	)
	w.Pause()
	w.Resume()
	w.Push("reload", "payload")
	time.Sleep(15 * time.Millisecond)
	if !called {
		t.Error("Expected function to be called after resume")
	}
	called = false
	w.Reload()
	w.Push("reload", "payload")
	time.Sleep(15 * time.Millisecond)
	if !called {
		t.Error("Expected function to be called after reload")
	}
	stopWorker(w, "reload", 50)
}

func TestWorkerLastUpdate(t *testing.T) {
	w := &worker{}
	w.Init()
	w.Register("last_update", 1,
		nil,
		func() int64 { return 10 },
		func(payload any) bool {
			return true
		},
		func() bool { return true },
	)
	w.Call("last_update", CallImmediate)
	before := time.Now().Add(-1 * time.Second)
	after := w.GetLastUpdate("last_update")
	if after.Before(before) {
		t.Errorf("Expected last update to be recent, got %v", after)
	}
	stopWorker(w, "last_update", 20)
}
