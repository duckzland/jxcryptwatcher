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
		Logf("WorkerManager already initialized — skipping Init()")
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
	if _, exists := w.workers[key]; exists {
		Logf("Worker %s already registered", key)
		w.mu.Unlock()
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	w.workers[key] = fn
	w.locks[key] = &sync.Mutex{}
	w.active[key] = true
	w.queues[key] = make(chan struct{}, 100)
	w.conditions[key] = shouldRun
	w.cancelFuncs[key] = cancel
	w.mu.Unlock()

	go w.startQueueWorker(ctx, key, false)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
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
	if _, exists := w.bufferedWorkers[key]; exists {
		w.mu.Unlock()
		Logf("Buffered worker %s already registered — skipping", key)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	w.bufferedWorkers[key] = fn
	w.locks[key] = &sync.Mutex{}
	w.active[key] = true
	w.queues[key] = make(chan struct{}, 100)
	w.conditions[key] = shouldRun
	w.messageQueues[key] = make(chan string, bufferSize)
	w.minDelayMs[key] = minDelayMs
	if w.cancelFuncs == nil {
		w.cancelFuncs = make(map[string]context.CancelFunc)
	}
	w.cancelFuncs[key] = cancel
	w.mu.Unlock()

	go w.startQueueWorker(ctx, key, true)

	if interval > 0 {
		go func() {
			ticker := time.NewTicker(time.Duration(interval) * time.Millisecond)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					w.mu.Lock()
					active := w.active[key]
					w.mu.Unlock()
					if active {
						w.Call(key, CallQueued)
					}
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
	if _, exists := w.workers[key]; exists {
		w.mu.Unlock()
		Logf("Sleeper worker %s already registered — skipping", key)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	w.workers[key] = fn
	w.locks[key] = &sync.Mutex{}
	w.active[key] = true
	w.queues[key] = make(chan struct{}, 1)
	w.conditions[key] = shouldRun
	if w.cancelFuncs == nil {
		w.cancelFuncs = make(map[string]context.CancelFunc)
	}
	w.cancelFuncs[key] = cancel
	w.mu.Unlock()

	go func() {
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
				if delayMs > 0 {
					time.Sleep(time.Duration(delayMs) * time.Millisecond)
				}
				w.runWorker(key)
			}
		}
	}()
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

func (w *worker) PauseAll() {
	w.mu.Lock()
	defer w.mu.Unlock()
	for key := range w.active {
		w.active[key] = false
		Logf("Worker:%s Paused", key)
	}
}

func (w *worker) ResumeAll() {
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

func (w *worker) startQueueWorker(ctx context.Context, key string, buffered bool) {
	w.mu.Lock()
	queue := w.queues[key]
	w.mu.Unlock()

	for {
		select {
		case <-ctx.Done():
			return
		case <-queue:
			if buffered {
				w.runBufferedWorker(key)
			} else {
				w.runWorker(key)
			}
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
