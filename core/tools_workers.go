package core

import (
	"sync"
	"sync/atomic"
	"time"
)

var workerManager *worker = nil

type CallMode int

const (
	CallImmediate CallMode = iota
	CallQueued
	CallDebounced
	CallBypassImmediate
)

type worker struct {
	registry   sync.Map
	conditions sync.Map
	lastRun    sync.Map
	destroyed  atomic.Bool
}

func (w *worker) Init() {
	w.destroyed.Store(false)
}

func (w *worker) Register(key string, size int, getDelay func() int64, getInterval func() int64, fn func(any) bool, shouldRun func() bool) {
	if w.destroyed.Load() {
		return
	}

	if existingAny, ok := w.registry.Load(key); ok {
		existing := existingAny.(*workerUnit)
		existing.Destroy()
		w.delete(key)
	}

	w.conditions.Store(key, shouldRun)
	w.lastRun.Store(key, time.Now())

	unit := NewWorkerUnit(size, getDelay, getInterval, func(payload any) bool {
		if w.destroyed.Load() {
			return false
		}

		if mode, ok := payload.(CallMode); !ok || mode != CallBypassImmediate {
			if shouldRun != nil && !shouldRun() {
				return false
			}
		}

		w.lastRun.Store(key, time.Now())
		return fn(payload)
	})

	w.registry.Store(key, unit)
	unit.Start()
}

func (w *worker) Deregister(key string) {
	if w.destroyed.Load() {
		return
	}

	if unit := w.delete(key); unit != nil {
		unit.Destroy()
	}
}

func (w *worker) Call(key string, mode CallMode) {
	if w.destroyed.Load() {
		return
	}

	unitAny, ok := w.registry.Load(key)
	if !ok {
		return
	}
	unit := unitAny.(*workerUnit)

	condAny, _ := w.conditions.Load(key)
	var cond func() bool
	if condAny != nil {
		cond = condAny.(func() bool)
	}

	if mode != CallBypassImmediate && cond != nil && !cond() {
		return
	}

	switch mode {
	case CallImmediate:
		go unit.Call(nil)

	case CallBypassImmediate:
		unit.Push(CallBypassImmediate)

	case CallQueued:
		unit.Push(nil)

	case CallDebounced:
		UseDebouncer().Call("worker_"+key, time.Second, func() {
			unit.Call(nil)
		})
	}
}

func (w *worker) Push(key string, payload any) {
	if w.destroyed.Load() {
		return
	}

	if unitAny, ok := w.registry.Load(key); ok {
		unit := unitAny.(*workerUnit)
		unit.Push(payload)
	}
}

func (w *worker) Flush(key string) {
	if w.destroyed.Load() {
		return
	}

	if unitAny, ok := w.registry.Load(key); ok {
		unit := unitAny.(*workerUnit)
		unit.Flush()
	}
}

func (w *worker) Reset(key string) {
	if w.destroyed.Load() {
		return
	}

	if unitAny, ok := w.registry.Load(key); ok {
		unit := unitAny.(*workerUnit)
		unit.Reset()
	}
}

func (w *worker) Pause() {
	if w.destroyed.Load() {
		return
	}

	w.registry.Range(func(_, v any) bool {
		if unit, ok := v.(*workerUnit); ok {
			unit.Stop()
		}
		return true
	})
}

func (w *worker) Resume() {
	if w.destroyed.Load() {
		return
	}

	w.registry.Range(func(_, v any) bool {
		if unit, ok := v.(*workerUnit); ok {
			unit.Start()
		}
		return true
	})
}

func (w *worker) Reload() {
	w.Pause()
	w.Resume()
}

func (w *worker) GetLastUpdate(key string) time.Time {
	if w.destroyed.Load() {
		return time.Time{}
	}

	if tAny, ok := w.lastRun.Load(key); ok {
		return tAny.(time.Time)
	}

	return time.Time{}
}

func (w *worker) Destroy() {
	if w.destroyed.Swap(true) {
		return
	}

	var units []*workerUnit
	w.registry.Range(func(_, v any) bool {
		if unit, ok := v.(*workerUnit); ok {
			units = append(units, unit)
		}
		return true
	})

	w.registry = sync.Map{}
	w.conditions = sync.Map{}
	w.lastRun = sync.Map{}

	for _, unit := range units {
		unit.Destroy()
	}
}

func (w *worker) delete(key string) *workerUnit {
	var unit *workerUnit

	if u, ok := w.registry.Load(key); ok {
		unit = u.(*workerUnit)
		w.registry.Delete(key)
	}

	w.conditions.Delete(key)
	w.lastRun.Delete(key)

	return unit
}

func RegisterWorkerManager() *worker {
	if workerManager == nil {
		workerManager = &worker{}
	}
	return workerManager
}

func UseWorker() *worker {
	return workerManager
}
