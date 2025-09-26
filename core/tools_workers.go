package core

import (
	"context"
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
	bufferedWorkers map[string]BufferedWorkerFunc
	locks           map[string]*sync.Mutex
	active          map[string]bool
	queues          map[string]chan struct{}
	messageQueues   map[string]chan string
	conditions      map[string]func() bool
	lastRun         map[string]time.Time
	minDelayMs      map[string]int64
	cancelFuncs     map[string]context.CancelFunc
	mu              sync.Mutex
}

var workerManager *worker = nil

func (w *worker) Init() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.workers != nil {
		Logln("WorkerManager already initialized — skipping Init()")
		return
	}

	w.workers = make(map[string]WorkerFunc)
	w.locks = make(map[string]*sync.Mutex)
	w.active = make(map[string]bool)
	w.queues = make(map[string]chan struct{})
	w.conditions = make(map[string]func() bool)
	w.lastRun = make(map[string]time.Time)
	w.bufferedWorkers = make(map[string]BufferedWorkerFunc)
	w.messageQueues = make(map[string]chan string)
	w.minDelayMs = make(map[string]int64)
	w.cancelFuncs = make(map[string]context.CancelFunc)
}

func (w *worker) Register(key string, debounce int64, fn WorkerFunc, getInterval func() int64, shouldRun func() bool) {
	w.mu.Lock()
	if cancel, exists := w.cancelFuncs[key]; exists {
		cancel()
	}
	w.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	w.createWorker(key, 100, -1, -1, cancel, fn, nil, shouldRun)
	interval := getInterval()

	go w.processQueues(ctx, key)

	if interval > 0 {
		go w.processScheduler(ctx, key, interval, CallQueued)
	}
}

func (w *worker) RegisterBuffered(key string, bufferSize int64, minDelayMs int64, getInterval func() int64, fn BufferedWorkerFunc, shouldRun func() bool) {

	w.mu.Lock()
	if cancel, exists := w.cancelFuncs[key]; exists {
		cancel()
	}
	w.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())

	w.createWorker(key, 100, bufferSize, minDelayMs, cancel, nil, fn, shouldRun)

	interval := getInterval()

	go w.processBuffers(ctx, key)

	if interval > 0 {
		go w.processScheduler(ctx, key, interval, CallQueued)
	}
}

func (w *worker) RegisterListener(key string, delay int64, fn WorkerFunc, shouldRun func() bool) {

	w.mu.Lock()
	if cancel, exists := w.cancelFuncs[key]; exists {
		cancel()
	}
	w.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())

	w.createWorker(key, 1, -1, -1, cancel, fn, nil, shouldRun)

	go w.processListener(ctx, key, delay)
}

func (w *worker) Call(key string, mode CallMode) {
	w.mu.Lock()
	cond := w.conditions[key]
	queue := w.queues[key]
	w.mu.Unlock()

	if mode != CallBypassImmediate && cond != nil && !cond() {
		return
	}

	switch mode {
	case CallImmediate, CallBypassImmediate:
		go w.runWorker(key)
	case CallQueued:
		select {
		case queue <- struct{}{}:
		default:
			Logf("Queue full for key: %s", key)
		}
	case CallDebounced:
		UseDebouncer().Call("worker_"+key, time.Second, func() {
			select {
			case queue <- struct{}{}:
			default:
				Logf("Debounced queue full for key: %s", key)
			}
		})
	}
}

func (w *worker) GetLastUpdate(key string) time.Time {
	w.mu.Lock()
	defer w.mu.Unlock()

	if ts, ok := w.lastRun[key]; ok {
		return ts
	}
	return time.Time{}
}

func (w *worker) Pause() {
	w.mu.Lock()
	defer w.mu.Unlock()
	for key := range w.active {
		w.active[key] = false
		Logf("Worker:%s Paused", key)
	}
}

func (w *worker) Resume() {
	w.mu.Lock()
	defer w.mu.Unlock()
	for key := range w.active {
		w.active[key] = true
		Logf("Worker:%s Resumed", key)
	}
}

func (w *worker) pushMessage(key string, msg string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if ch, ok := w.messageQueues[key]; ok {
		ch <- msg
		w.queues[key] <- struct{}{}
	} else {
		Logf("MessageQueue not found for key: %s", key)
	}
}

func (w *worker) createWorker(key string, qb int64, mb int64, delay int64, cancel context.CancelFunc, fnw WorkerFunc, fnb BufferedWorkerFunc, cond func() bool) {
	w.mu.Lock()

	if fnw != nil {
		w.workers[key] = fnw
	}

	if fnb != nil {
		w.bufferedWorkers[key] = fnb
	}

	if mb > 0 {
		w.messageQueues[key] = make(chan string, mb)
	}

	if delay > 0 {
		w.minDelayMs[key] = delay
	}

	if qb > 0 {
		w.queues[key] = make(chan struct{}, qb)
	}

	w.locks[key] = &sync.Mutex{}
	w.active[key] = true
	w.conditions[key] = cond
	w.cancelFuncs[key] = cancel

	w.mu.Unlock()
}

func (w *worker) processScheduler(ctx context.Context, key string, interval int64, callType CallMode) {
	ticker := time.NewTicker(time.Duration(interval) * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.mu.Lock()
			active := w.active[key]
			cond := w.conditions[key]
			w.mu.Unlock()

			if !active {
				continue
			}

			if cond != nil && !cond() {
				continue
			}

			w.Call(key, callType)
		}
	}
}

func (w *worker) processListener(ctx context.Context, key string, delay int64) {
	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-w.queues[key]:
			if !ok {
				return
			}

			w.mu.Lock()
			active := w.active[key]
			cond := w.conditions[key]
			w.mu.Unlock()

			if !active {
				continue
			}

			if cond != nil && !cond() {
				continue
			}

			if delay > 0 {
				time.Sleep(time.Duration(delay) * time.Millisecond)
			}

			w.runWorker(key)
		}
	}
}

func (w *worker) processQueues(ctx context.Context, key string) {
	w.mu.Lock()
	queue := w.queues[key]
	w.mu.Unlock()

	for {
		select {
		case <-ctx.Done():
			return
		case <-queue:
			w.runWorker(key)
		}
	}
}

func (w *worker) processBuffers(ctx context.Context, key string) {
	w.mu.Lock()
	queue := w.queues[key]
	w.mu.Unlock()

	for {
		select {
		case <-ctx.Done():
			return
		case <-queue:
			w.runBufferedWorker(key)
		}
	}
}

func (w *worker) runWorker(key string) {
	w.mu.Lock()
	lock, exists := w.locks[key]
	fn, ok := w.workers[key]
	w.mu.Unlock()

	if !exists || !ok {
		Logf("Worker:%s Not registered", key)
		return
	}

	lock.Lock()
	defer lock.Unlock()
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
		if fn(messages) {
			w.mu.Lock()
			w.lastRun[key] = time.Now()
			w.mu.Unlock()
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
