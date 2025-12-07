package core

import (
	"context"
	"sync"
	"time"
)

type workerUnit struct {
	delay       time.Duration
	interval    time.Duration
	fn          func(any) bool
	getDelay    func() int64
	getInterval func() int64
	queue       chan any
	ctx         context.Context
	cancel      context.CancelFunc
	mu          sync.Mutex
	wg          sync.WaitGroup
	bufferSize  int
}

func NewWorkerUnit(buffer int) *workerUnit {
	return &workerUnit{
		bufferSize: buffer,
		queue:      make(chan any, buffer),
	}
}

func (w *workerUnit) Call(payload any) {
	w.mu.Lock()
	fn := w.fn
	w.mu.Unlock()
	if fn != nil {
		fn(payload)
	}
}

func (w *workerUnit) Flush() {
	for {
		select {
		case <-w.queue:
		default:
			return
		}
	}
}

func (w *workerUnit) Push(payload any) {
	select {
	case w.queue <- payload:
	default:
	}
}

func (w *workerUnit) Start() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.ctx != nil {
		return
	}

	if w.getDelay != nil {
		w.delay = time.Duration(w.getDelay()) * time.Millisecond
	}
	if w.getInterval != nil {
		w.interval = time.Duration(w.getInterval()) * time.Millisecond
	}

	w.ctx, w.cancel = context.WithCancel(context.Background())
	w.wg.Add(1)
	go w.worker()
}

func (w *workerUnit) Stop() {
	w.mu.Lock()
	if w.cancel != nil {
		w.cancel()
	}
	w.mu.Unlock()

	w.wg.Wait()

	w.mu.Lock()
	w.ctx = nil
	w.cancel = nil
	w.mu.Unlock()
}

func (w *workerUnit) Reset() {
	w.Flush()
	w.Stop()
	w.Start()
}

func (w *workerUnit) newTicker() *time.Ticker {
	if w.interval > 0 {
		return time.NewTicker(w.interval)
	}
	return &time.Ticker{C: make(chan time.Time)}
}

func (w *workerUnit) worker() {
	defer w.wg.Done()
	ticker := w.newTicker()
	defer ticker.Stop()

	for {
		w.mu.Lock()
		ctx := w.ctx
		delay := w.delay
		w.mu.Unlock()

		if ctx == nil {
			return
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.Push(nil)
		case x, ok := <-w.queue:
			if !ok {
				return
			}
			if delay > 0 {
				time.Sleep(delay)
			}
			w.Call(x)
		}
	}
}
