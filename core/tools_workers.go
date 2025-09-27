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
}

func (w *worker) Register(key string, ops WorkerMode, size int64, getDelay func() int64, fn func(any) bool, shouldRun func() bool) {
	w.mu.Lock()
	w.conditions[key] = shouldRun
	w.lastRun[key] = time.Now()
	w.mu.Unlock()

	if ops == WorkerScheduler {
		size = 1
	}

	w.registry[key] = &workerUnit{
		ops:      ops,
		getDelay: getDelay,
		queue:    make(chan any, size),
		fn: func(payload any) bool {
			if shouldRun != nil && !shouldRun() {
				return false
			}
			w.mu.Lock()
			w.lastRun[key] = time.Now()
			w.mu.Unlock()
			return fn(payload)
		},
	}

	w.registry[key].Start()
}

func (w *worker) Call(key string, mode CallMode) {
	w.mu.Lock()
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
	case CallImmediate, CallBypassImmediate:
		unit.Call()

	case CallQueued:
		unit.Push(nil)

	case CallDebounced:
		UseDebouncer().Call("worker_"+key, time.Second, func() {
			unit.Call()
		})
	}
}

func (w *worker) Reload() {
	w.Pause()
	w.Resume()
}

func (w *worker) Pause() {
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, unit := range w.registry {
		unit.Stop()
	}
}

func (w *worker) Resume() {
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, unit := range w.registry {
		unit.Start()
	}
}

func (w *worker) Push(key string, payload any) {
	w.mu.Lock()
	unit := w.registry[key]
	w.mu.Unlock()

	if unit != nil {
		unit.Push(payload)
	}
}

func (w *worker) Flush(key string) {
	w.mu.Lock()
	unit := w.registry[key]
	w.mu.Unlock()

	if unit != nil {
		unit.Flush()
	}
}

func (w *worker) GetLastUpdate(key string) time.Time {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.lastRun[key]
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
