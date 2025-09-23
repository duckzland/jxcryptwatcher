package core

import (
	"fmt"
	"sync"
	"time"
)

type CallMode int

const (
	CallImmediate CallMode = iota
	CallQueued
	CallDebounced
	CallBypassImmediate
)

type WorkerFunc func()
type BufferedWorkerFunc func(messages []string) bool

type worker struct {
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
	minDelayMs      map[string]int64
	mu              sync.Mutex
	verbose         bool
}

var workerManager *worker = nil

func (w *worker) Init() {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.workers = make(map[string]WorkerFunc)
	w.locks = make(map[string]*sync.Mutex)
	w.active = make(map[string]bool)
	w.queues = make(map[string]chan struct{})
	w.conditions = make(map[string]func() bool)
	w.lastRun = make(map[string]time.Time)
	w.bufferedWorkers = make(map[string]BufferedWorkerFunc)
	w.messageQueues = make(map[string]chan string)
	w.recentLogs = make(map[string]string)
	w.logTimestamps = make(map[string]time.Time)
	w.minDelayMs = make(map[string]int64)
	w.verbose = false
}

func (w *worker) Register(
	key string,
	debounce int64,
	fn WorkerFunc,
	getInterval func() int64,
	shouldRun func() bool,
) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.workers[key] = fn
	w.locks[key] = &sync.Mutex{}
	w.active[key] = true
	w.queues[key] = make(chan struct{}, 100)
	w.conditions[key] = shouldRun

	go w.startQueueWorker(key, false)

	go func() {
		for {
			w.mu.Lock()
			active := w.active[key]
			w.mu.Unlock()

			if !active {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			w.Call(key, CallQueued)

			interval := getInterval()
			if interval <= 0 {
				interval = 1000
			}

			time.Sleep(time.Duration(interval) * time.Millisecond)
		}
	}()
}

func (w *worker) RegisterBuffered(
	key string,
	interval int64,
	debounce int64,
	bufferSize int64,
	minDelayMs int64,
	fn BufferedWorkerFunc,
	shouldRun func() bool,
) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.bufferedWorkers[key] = fn
	w.locks[key] = &sync.Mutex{}
	w.active[key] = true
	w.queues[key] = make(chan struct{}, 100)
	w.conditions[key] = shouldRun
	w.messageQueues[key] = make(chan string, bufferSize)
	w.minDelayMs[key] = minDelayMs

	go w.startQueueWorker(key, true)

	if interval > 0 {
		go func() {
			ticker := time.NewTicker(time.Duration(interval) * time.Millisecond)
			defer ticker.Stop()
			for range ticker.C {
				w.mu.Lock()
				active := w.active[key]
				w.mu.Unlock()

				if active {
					w.Call(key, CallQueued)
				}
			}
		}()
	}
}

func (w *worker) RegisterSleeper(
	key string,
	delayMs int64,
	fn WorkerFunc,
	shouldRun func() bool,
) {

	w.mu.Lock()
	defer w.mu.Unlock()

	w.workers[key] = fn
	w.locks[key] = &sync.Mutex{}
	w.active[key] = true
	w.queues[key] = make(chan struct{})
	w.conditions[key] = shouldRun

	go func() {
		for {
			w.mu.Lock()
			queue, ok := w.queues[key]
			active := w.active[key]
			cond := w.conditions[key]
			w.mu.Unlock()

			_, ok = <-queue
			if !ok || !active {
				return
			}
			if cond != nil && !cond() {
				continue
			}
			if delayMs > 0 {
				time.Sleep(time.Duration(delayMs) * time.Millisecond)
			}
			w.runWorker(key)
		}
	}()
}

func (w *worker) Call(key string, mode CallMode) {
	w.mu.Lock()
	cond := w.conditions[key]
	w.mu.Unlock()

	if mode != CallBypassImmediate {
		if cond != nil && !cond() {
			return
		}
	}

	switch mode {
	case CallImmediate, CallBypassImmediate:
		go w.runWorker(key)

	case CallQueued:
		w.mu.Lock()
		queue := w.queues[key]
		w.mu.Unlock()
		queue <- struct{}{}

	case CallDebounced:
		UseDebouncer().Call("worker_"+key, time.Duration(1000)*time.Millisecond, func() {
			w.mu.Lock()
			queue := w.queues[key]
			w.mu.Unlock()
			queue <- struct{}{}
		})
	}
}

func (w *worker) PushMessage(key string, msg string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if ch, ok := w.messageQueues[key]; ok {
		ch <- msg
		w.queues[key] <- struct{}{}
	} else {
		w.logGrouped(key+"_pushfail", 5000, fmt.Sprintf("[PushMessage] messageQueue not found for key: %s", key))
	}
}

func (w *worker) GetMessageChannel(key string) <-chan string {
	w.mu.Lock()
	defer w.mu.Unlock()

	if ch, ok := w.messageQueues[key]; ok {
		return ch
	}
	return nil
}

func (w *worker) Stop(key string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.active[key] = false
	Logf("[Worker:%s] Marked as inactive", key)
}

func (w *worker) GetLastUpdate(key string) time.Time {
	w.mu.Lock()
	defer w.mu.Unlock()

	if ts, ok := w.lastRun[key]; ok {
		return ts
	}
	return time.Time{}
}

func (w *worker) PauseAll() {
	w.mu.Lock()
	defer w.mu.Unlock()
	for key := range w.active {
		w.active[key] = false
		Logf("[Worker:%s] Paused", key)
	}
}

func (w *worker) ResumeAll() {
	w.mu.Lock()
	defer w.mu.Unlock()
	for key := range w.active {
		w.active[key] = true
		Logf("[Worker:%s] Resumed", key)
	}
}

func (w *worker) logGrouped(key string, interval int64, msg string) {
	if !w.verbose {
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

func (w *worker) startQueueWorker(key string, buffered bool) {
	w.mu.Lock()
	queue := w.queues[key]
	w.mu.Unlock()

	for range queue {
		if buffered {
			w.runBufferedWorker(key)
		} else {
			w.runWorker(key)
		}
	}
}

func (w *worker) runWorker(key string) {
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

func (w *worker) runBufferedWorker(key string) {
	w.mu.Lock()
	lock := w.locks[key]
	fn := w.bufferedWorkers[key]
	msgCh := w.messageQueues[key]
	lastRun := w.lastRun[key]
	minDelay := w.minDelayMs[key]
	w.mu.Unlock()

	if minDelay > 0 {
		elapsed := time.Since(lastRun).Milliseconds()
		if elapsed < minDelay {
			time.Sleep(time.Duration(minDelay-elapsed) * time.Millisecond)
		}
	}

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

func RegisterWorkerManager() *worker {
	if workerManager == nil {
		InitOnce(func() {
			workerManager = &worker{}
		})
	}
	return workerManager
}

func UseWorker() *worker {
	return workerManager
}
