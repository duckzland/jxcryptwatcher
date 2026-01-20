package core

import (
	"sync"
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
	registry   map[string]*workerUnit
	conditions map[string]func() bool
	lastRun    map[string]time.Time
	mu         sync.Mutex
	destroyed  bool
}

func (w *worker) Init() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.registry != nil {
		return
	}

	w.conditions = make(map[string]func() bool, 3)
	w.registry = make(map[string]*workerUnit, 3)
	w.lastRun = make(map[string]time.Time, 3)
	w.destroyed = false
}

func (w *worker) Register(key string, size int, getDelay func() int64, getInterval func() int64, fn func(any) bool, shouldRun func() bool) {
	var oldUnit *workerUnit

	w.mu.Lock()
	if w.destroyed {
		w.mu.Unlock()
		return
	}

	if existing, registered := w.registry[key]; registered {
		oldUnit = existing
		w.internalDeleteKey(key)
	}

	w.conditions[key] = shouldRun
	w.lastRun[key] = time.Now()

	unit := &workerUnit{
		getInterval: getInterval,
		getDelay:    getDelay,
		queue:       make(chan any, size),
		fn: func(payload any) bool {

			w.mu.Lock()
			if w.destroyed || (w.registry == nil && w.conditions == nil && w.lastRun == nil) {
				w.mu.Unlock()
				return false
			}
			w.mu.Unlock()

			if mode, ok := payload.(CallMode); !ok || mode != CallBypassImmediate {
				if shouldRun != nil && !shouldRun() {
					return false
				}
			}

			w.mu.Lock()
			if w.lastRun != nil {
				w.lastRun[key] = time.Now()
			}
			w.mu.Unlock()

			return fn(payload)
		},
	}

	w.registry[key] = unit
	w.mu.Unlock()

	if oldUnit != nil {
		oldUnit.Destroy()
	}

	unit.Start()
}

func (w *worker) Deregister(key string) {
	w.mu.Lock()
	if w.destroyed {
		w.mu.Unlock()
		return
	}
	unit := w.internalDeleteKey(key)
	w.mu.Unlock()

	if unit != nil {
		unit.Destroy()
	}
}

func (w *worker) Call(key string, mode CallMode) {
	w.mu.Lock()
	if w.destroyed {
		w.mu.Unlock()
		return
	}

	unit := w.registry[key]
	cond := w.conditions[key]
	w.mu.Unlock()

	if unit == nil {
		return
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
	w.mu.Lock()
	if w.destroyed {
		w.mu.Unlock()
		return
	}
	unit := w.registry[key]
	w.mu.Unlock()

	if unit != nil {
		unit.Push(payload)
	}
}

func (w *worker) Flush(key string) {
	w.mu.Lock()
	if w.destroyed {
		w.mu.Unlock()
		return
	}
	unit := w.registry[key]
	w.mu.Unlock()

	if unit != nil {
		unit.Flush()
	}
}

func (w *worker) Reset(key string) {
	w.mu.Lock()
	if w.destroyed {
		w.mu.Unlock()
		return
	}
	unit := w.registry[key]
	w.mu.Unlock()

	if unit != nil {
		unit.Reset()
	}
}

func (w *worker) Pause() {
	w.mu.Lock()
	if w.destroyed || w.registry == nil {
		w.mu.Unlock()
		return
	}

	units := make([]*workerUnit, 0, len(w.registry))
	for _, unit := range w.registry {
		if unit != nil {
			units = append(units, unit)
		}
	}
	w.mu.Unlock()

	for _, unit := range units {
		unit.Stop()
	}
}

func (w *worker) Resume() {
	w.mu.Lock()
	if w.destroyed || w.registry == nil {
		w.mu.Unlock()
		return
	}

	units := make([]*workerUnit, 0, len(w.registry))
	for _, unit := range w.registry {
		if unit != nil {
			units = append(units, unit)
		}
	}
	w.mu.Unlock()

	for _, unit := range units {
		unit.Start()
	}
}

func (w *worker) Reload() {
	w.Pause()
	w.Resume()
}

func (w *worker) GetLastUpdate(key string) time.Time {
	w.mu.Lock()
	if w.destroyed || w.lastRun == nil {
		w.mu.Unlock()
		return time.Time{}
	}
	t := w.lastRun[key]
	w.mu.Unlock()
	return t
}

func (w *worker) Destroy() {
	w.mu.Lock()
	if w.destroyed {
		w.mu.Unlock()
		return
	}

	w.destroyed = true

	var units []*workerUnit
	if w.registry != nil {
		for key := range w.registry {
			if unit := w.internalDeleteKey(key); unit != nil {
				units = append(units, unit)
			}
		}
	}

	w.registry = nil
	w.conditions = nil
	w.lastRun = nil
	w.mu.Unlock()

	for _, unit := range units {
		unit.Destroy()
	}
}

func (w *worker) internalDeleteKey(key string) *workerUnit {
	var unit *workerUnit

	if w.registry != nil {
		if u, exists := w.registry[key]; exists {
			unit = u
			delete(w.registry, key)
		}
	}

	if w.conditions != nil {
		delete(w.conditions, key)
	}

	if w.lastRun != nil {
		delete(w.lastRun, key)
	}

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
