package core

import (
	"testing"
	"time"
)

func stopWorker(w *worker, key string, delayMs int64) {

	// Delay is only to emulate that the actual test is finished!
	// Not to simulate if the unit can be stopped sanely!
	time.Sleep(time.Duration(delayMs) * time.Millisecond)

	w.mu.Lock()
	unit := w.registry[key]
	w.mu.Unlock()

	// We only stops, this is by design to test if stopping sane!
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

func TestRegisterAndUseWorkerManager(t *testing.T) {
	w1 := RegisterWorkerManager()
	w2 := UseWorker()

	if w1 != w2 {
		t.Error("Expected singleton instance from RegisterWorkerManager and UseWorker")
	}
}

func TestWorkerListenerPushTriggersFunction(t *testing.T) {
	called := false

	w := &worker{}
	w.Init()
	w.Register("listener_push", WorkerListener, 1,
		func() int64 { return 10 },
		func(payload any) bool {
			called = true
			return true
		},
		func() bool { return true },
	)

	w.Push("listener_push", "payload")

	time.Sleep(time.Duration(15) * time.Millisecond)

	if !called {
		t.Error("Expected listener to fire on pushed payload")
	}

	stopWorker(w, "scheduler_immediate", 30)
}

func TestWorkerSchedulerCallImmediate(t *testing.T) {
	called := false

	w := &worker{}
	w.Init()
	w.Register("scheduler_immediate", WorkerScheduler, 1,
		func() int64 { return 10 },
		func(payload any) bool {
			called = true
			return true
		},
		func() bool { return true },
	)

	w.Call("scheduler_immediate", CallImmediate)

	if !called {
		t.Error("Expected scheduler to call function immediately")
	}

	stopWorker(w, "scheduler_immediate", 20)
}

func TestWorkerCallDebounced(t *testing.T) {
	var callCount int

	RegisterDebouncer().Init()

	w := &worker{}
	w.Init()
	w.Register("debounced_static", WorkerScheduler, 1,
		func() int64 { return 50 },
		func(payload any) bool {
			callCount++
			return true
		},
		func() bool { return true },
	)

	w.Call("debounced_static", CallDebounced)
	w.Call("debounced_static", CallDebounced)
	w.Call("debounced_static", CallDebounced)

	time.Sleep(time.Duration(60) * time.Millisecond)
	if callCount != 1 {
		t.Errorf("Expected debounced function to fire once, but fired %d times", callCount)
	}

	stopWorker(w, "debounced_static", 120)
}

func TestWorkerCallBypassImmediateRespectsCondition(t *testing.T) {
	called := false

	w := &worker{}
	w.Init()
	w.Register("bypass_static", WorkerScheduler, 1,
		func() int64 { return 10 },
		func(payload any) bool {
			called = true
			t.Error("Function should not have been called — condition is false")
			return true
		},
		func() bool { return false },
	)

	w.Call("bypass_static", CallBypassImmediate)

	if called {
		t.Error("Function should not have been called — condition is false")
	}

	stopWorker(w, "bypass_static", 20)
}

func TestWorkerPushFlushReset(t *testing.T) {
	called := false

	w := &worker{}
	w.Init()
	w.Register("flush_reset_static", WorkerListener, 1,
		func() int64 { return 10 },
		func(payload any) bool {
			called = true
			return true
		},
		func() bool { return true },
	)

	w.Push("flush_reset_static", "payload")
	w.Flush("flush_reset_static")
	w.Reset("flush_reset_static")
	w.Push("flush_reset_static", "payload")

	time.Sleep(time.Duration(15) * time.Millisecond)

	if !called {
		t.Error("Expected function to be called after reset and push")
	}

	stopWorker(w, "flush_reset_static", 30)
}

func TestWorkerPauseResumeReload(t *testing.T) {
	called := false

	w := &worker{}
	w.Init()
	w.Register("pause_resume_static", WorkerListener, 1,
		func() int64 { return 10 },
		func(payload any) bool {
			called = true
			return true
		},
		func() bool { return true },
	)

	w.Pause()
	w.Resume()
	w.Push("pause_resume_static", "payload")

	time.Sleep(time.Duration(15) * time.Millisecond)

	if !called {
		t.Error("Expected function to be called after resume")
	}

	// Reload test
	called = false
	w.Reload()
	w.Push("pause_resume_static", "payload")

	time.Sleep(time.Duration(15) * time.Millisecond)
	if !called {
		t.Error("Expected function to be called after reload")
	}

	stopWorker(w, "pause_resume_static", 50)
}

func TestWorkerGetLastUpdate(t *testing.T) {
	w := &worker{}
	w.Init()
	w.Register("last_update_static", WorkerScheduler, 1,
		func() int64 { return 10 },
		func(payload any) bool {
			return true
		},
		func() bool { return true },
	)

	w.Call("last_update_static", CallImmediate)

	before := time.Now().Add(-1 * time.Second)
	after := w.GetLastUpdate("last_update_static")

	if after.Before(before) {
		t.Errorf("Expected last update to be after call, got %v", after)
	}

	stopWorker(w, "last_update_static", 20)
}
