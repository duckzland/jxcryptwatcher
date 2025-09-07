package core

import (
	"fmt"
	"sync"
	"time"
)

var verboseWorkersDebugMessage bool = false

type CallMode int

const (
	CallImmediate CallMode = iota
	CallQueued
	CallDebounced
	CallBypassImmediate
)

type WorkerFunc func()
type BufferedWorkerFunc func(messages []string) bool

type Worker struct {
	workers         map[string]WorkerFunc
	locks           map[string]*sync.Mutex
	active          map[string]bool
	queues          map[string]chan struct{}
	conditions      map[string]func() bool
	lastRun         map[string]time.Time
	bufferedWorkers map[string]BufferedWorkerFunc
	messageQueues   map[string]chan string
	recentLogs      map[string]string
	logTimestamps   map[string]time.Time

	mu sync.Mutex
}

var WorkerManager = &Worker{
	workers:         make(map[string]WorkerFunc),
	locks:           make(map[string]*sync.Mutex),
	active:          make(map[string]bool),
	queues:          make(map[string]chan struct{}),
	conditions:      make(map[string]func() bool),
	lastRun:         make(map[string]time.Time),
	bufferedWorkers: make(map[string]BufferedWorkerFunc),
	messageQueues:   make(map[string]chan string),
	recentLogs:      make(map[string]string),
	logTimestamps:   make(map[string]time.Time),
}

func (w *Worker) logGrouped(key string, interval int64, msg string) {
	if !verboseWorkersDebugMessage {
		return
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	lastMsg := w.recentLogs[key]
	lastTime := w.logTimestamps[key]
	now := time.Now()

	if msg != lastMsg || now.Sub(lastTime).Milliseconds() >= interval {
		Logln(msg)
		w.recentLogs[key] = msg
		w.logTimestamps[key] = now
	}
}

func (w *Worker) Register(key string, fn WorkerFunc, interval int64, debounce int64, shouldRun func() bool) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.workers[key] = fn
	w.locks[key] = &sync.Mutex{}
	w.active[key] = true
	w.queues[key] = make(chan struct{}, 100)
	w.conditions[key] = shouldRun

	go w.startQueueWorker(key, false)

	if interval > 0 {
		go func() {
			ticker := time.NewTicker(time.Duration(interval) * time.Millisecond)
			defer ticker.Stop()
			for range ticker.C {
				if w.active[key] {
					w.Call(key, CallQueued)
				}
			}
		}()
	}
}

func (w *Worker) RegisterBuffered(key string, fn BufferedWorkerFunc, interval int64, debounce int64, bufferSize int64, shouldRun func() bool) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.bufferedWorkers[key] = fn
	w.locks[key] = &sync.Mutex{}
	w.active[key] = true
	w.queues[key] = make(chan struct{}, 100)
	w.conditions[key] = shouldRun
	w.messageQueues[key] = make(chan string, bufferSize)

	go w.startQueueWorker(key, true)

	if interval > 0 {
		go func() {
			ticker := time.NewTicker(time.Duration(interval) * time.Millisecond)
			defer ticker.Stop()
			for range ticker.C {
				if w.active[key] {
					w.Call(key, CallQueued)
				}
			}
		}()
	}
}

func (w *Worker) RegisterSleeper(key string, fn WorkerFunc, delayMs int64, shouldRun func() bool) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.workers[key] = fn
	w.locks[key] = &sync.Mutex{}
	w.active[key] = true
	w.queues[key] = make(chan struct{})
	w.conditions[key] = shouldRun

	go func() {
		for {
			_, ok := <-w.queues[key]
			if !ok || !w.active[key] {
				return
			}
			if cond := w.conditions[key]; cond != nil && !cond() {
				continue
			}
			if delayMs > 0 {
				time.Sleep(time.Duration(delayMs) * time.Millisecond)
			}
			w.runWorker(key)
		}
	}()
}

func (w *Worker) Call(key string, mode CallMode) {
	if mode != CallBypassImmediate {
		if cond, ok := w.conditions[key]; ok && cond != nil {
			if !cond() {
				return
			}
		}
	}

	switch mode {
	case CallImmediate, CallBypassImmediate:
		go w.runWorker(key)
	case CallQueued:
		w.queues[key] <- struct{}{}
	case CallDebounced:
		MainDebouncer.Call("worker_"+key, time.Duration(1000)*time.Millisecond, func() {
			w.queues[key] <- struct{}{}
		})
	}
}

func (w *Worker) PushMessage(key string, msg string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if ch, ok := w.messageQueues[key]; ok {
		ch <- msg
		w.queues[key] <- struct{}{}
	} else {
		w.logGrouped(key+"_pushfail", 5000, fmt.Sprintf("[PushMessage] messageQueue not found for key: %s", key))
	}
}

func (w *Worker) GetMessageChannel(key string) <-chan string {
	w.mu.Lock()
	defer w.mu.Unlock()

	if ch, ok := w.messageQueues[key]; ok {
		return ch
	}
	return nil
}

func (w *Worker) runWorker(key string) {
	w.mu.Lock()
	lock, exists := w.locks[key]
	fn, ok := w.workers[key]
	w.mu.Unlock()

	if !exists || !ok {
		w.logGrouped(key+"_missing", 5000, fmt.Sprintf("[Worker:%s] Not registered", key))
		return
	}

	lock.Lock()
	defer lock.Unlock()

	w.logGrouped(key+"_exec", 1000, fmt.Sprintf("[Worker:%s] Executing...", key))
	fn()

	w.mu.Lock()
	w.lastRun[key] = time.Now()
	w.mu.Unlock()
}

func (w *Worker) runBufferedWorker(key string) {
	w.mu.Lock()
	lock := w.locks[key]
	fn := w.bufferedWorkers[key]
	msgCh := w.messageQueues[key]
	w.mu.Unlock()

	lock.Lock()
	defer lock.Unlock()

	var messages []string
drain:
	for {
		select {
		case msg := <-msgCh:
			messages = append(messages, msg)
		default:
			break drain
		}
	}

	if len(messages) > 0 {
		w.logGrouped(key+"_exec", 1000, fmt.Sprintf("[Worker:%s] Executing with %d messages...", key, len(messages)))
		if fn(messages) {
			w.mu.Lock()
			w.lastRun[key] = time.Now()
			w.mu.Unlock()
		}
	}
}

func (w *Worker) startQueueWorker(key string, buffered bool) {
	for range w.queues[key] {
		if buffered {
			w.runBufferedWorker(key)
		} else {
			w.runWorker(key)
		}
	}
}

func (w *Worker) Stop(key string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.active[key] = false
	Logf("[Worker:%s] Marked as inactive", key)
}

func (w *Worker) GetLastUpdate(key string) time.Time {
	w.mu.Lock()
	defer w.mu.Unlock()

	if ts, ok := w.lastRun[key]; ok {
		return ts
	}
	return time.Time{}
}
