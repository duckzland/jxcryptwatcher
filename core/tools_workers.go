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

	w.conditions = make(map[string]func() bool)
	w.registry = make(map[string]*workerUnit)
	w.lastRun = make(map[string]time.Time)
	w.destroyed = false
}

func (w *worker) Register(key string, size int64, getDelay func() int64, getInterval func() int64, fn func(any) bool, shouldRun func() bool) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.destroyed {
		return
	}

	if _, registered := w.registry[key]; registered {
		w.internalDestroy(key)
	}

	w.conditions[key] = shouldRun
	w.lastRun[key] = time.Now()

	w.registry[key] = &workerUnit{
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
			w.lastRun[key] = time.Now()
			w.mu.Unlock()

			return fn(payload)
		},
	}

	w.registry[key].Start()
}

func (w *worker) Deregister(key string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.internalDestroy(key)
}

func (w *worker) Call(key string, mode CallMode) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.destroyed {
		return
	}

	unit := w.registry[key]
	cond := w.conditions[key]

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
	defer w.mu.Unlock()

	if w.destroyed {
		return
	}

	if unit, exists := w.registry[key]; exists && unit != nil {
		unit.Push(payload)
	}
}

func (w *worker) Flush(key string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.destroyed {
		return
	}

	if unit, exists := w.registry[key]; exists && unit != nil {
		unit.Flush()
	}
}

func (w *worker) Reset(key string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.destroyed {
		return
	}

	if unit, exists := w.registry[key]; exists && unit != nil {
		unit.Reset()
	}
}

func (w *worker) Pause() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.destroyed || w.registry == nil {
		return
	}

	for _, unit := range w.registry {
		unit.Stop()
	}

}

func (w *worker) Resume() {
	defer w.mu.Unlock()
	w.mu.Lock()

	if w.destroyed || w.registry == nil {
		return
	}

	for _, unit := range w.registry {
		unit.Start()
	}
}

func (w *worker) Reload() {
	w.Pause()
	w.Resume()
}

func (w *worker) GetLastUpdate(key string) time.Time {
	defer w.mu.Unlock()
	w.mu.Lock()

	if w.destroyed {
		return time.Time{}
	}

	return w.lastRun[key]
}

func (w *worker) Destroy() {
	defer w.mu.Unlock()
	w.mu.Lock()

	if w.destroyed {
		return
	}

	w.destroyed = true

	if w.registry != nil {
		for key := range w.registry {
			w.internalDestroy(key)
		}
	}

	w.registry = nil
	w.conditions = nil
	w.lastRun = nil
}

func (w *worker) internalDestroy(key string) {
	if w.registry != nil {
		if unit, exists := w.registry[key]; exists {
			if unit != nil {
				unit.Destroy()
			}

			delete(w.registry, key)
		}
	}

	if w.conditions != nil {
		if _, exists := w.conditions[key]; exists {
			delete(w.conditions, key)
		}
	}

	if w.lastRun != nil {
		if _, exists := w.lastRun[key]; exists {
			delete(w.lastRun, key)
		}
	}
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
