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

func (w *worker) Register(
	key string,
	size int64,
	getDelay func() int64,
	getInterval func() int64,
	fn func(any) bool,
	shouldRun func() bool,
) {
	unit := &workerUnit{
		getInterval: getInterval,
		getDelay:    getDelay,
		queue:       make(chan any, size),
		fn: func(payload any) bool {
			if mode, ok := payload.(CallMode); !ok || mode != CallBypassImmediate {
				if shouldRun != nil && !shouldRun() {
					return false
				}
			}
			w.mu.Lock()
			if w.destroyed {
				w.mu.Unlock()
				return false
			}
			w.lastRun[key] = time.Now()
			w.mu.Unlock()
			return fn(payload)
		},
	}

	w.mu.Lock()
	if w.destroyed {
		w.mu.Unlock()
		return
	}
	w.conditions[key] = shouldRun
	w.lastRun[key] = time.Now()
	w.registry[key] = unit
	w.mu.Unlock()

	unit.Start()
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
		unit.Push(nil)
	case CallBypassImmediate:
		unit.Push(CallBypassImmediate)
	case CallQueued:
		unit.Push(nil)
	case CallDebounced:
		UseDebouncer().Call("worker_"+key, time.Second, func() {
			unit.Push(nil)
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

	if unit == nil {
		return
	}
	unit.Push(payload)
}

func (w *worker) Flush(key string) {
	w.mu.Lock()
	if w.destroyed {
		w.mu.Unlock()
		return
	}
	unit := w.registry[key]
	w.mu.Unlock()

	if unit == nil {
		return
	}
	unit.Flush()
}

func (w *worker) Reset(key string) {
	w.mu.Lock()
	if w.destroyed {
		w.mu.Unlock()
		return
	}
	unit := w.registry[key]
	w.mu.Unlock()

	if unit == nil {
		return
	}
	unit.Reset()
}

func (w *worker) Pause() {
	w.mu.Lock()
	if w.destroyed {
		w.mu.Unlock()
		return
	}
	units := make([]*workerUnit, 0, len(w.registry))
	for _, unit := range w.registry {
		units = append(units, unit)
	}
	w.mu.Unlock()

	for _, unit := range units {
		unit.Stop()
	}
}

func (w *worker) Resume() {
	w.mu.Lock()
	if w.destroyed {
		w.mu.Unlock()
		return
	}
	units := make([]*workerUnit, 0, len(w.registry))
	for _, unit := range w.registry {
		units = append(units, unit)
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
	defer w.mu.Unlock()
	if w.destroyed {
		return time.Time{}
	}
	return w.lastRun[key]
}

func (w *worker) Destroy() {
	w.mu.Lock()
	if w.destroyed {
		w.mu.Unlock()
		return
	}
	units := make([]*workerUnit, 0, len(w.registry))
	for _, unit := range w.registry {
		units = append(units, unit)
	}
	w.destroyed = true
	w.mu.Unlock()

	for _, unit := range units {
		unit.Flush()
		unit.Stop()
	}

	w.mu.Lock()
	w.registry = nil
	w.conditions = nil
	w.lastRun = nil
	w.mu.Unlock()
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
