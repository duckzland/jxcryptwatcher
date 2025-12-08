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
	bufferSize  int
	destroyed   bool
}

func (w *workerUnit) Call(payload any) {
	defer w.mu.Unlock()
	w.mu.Lock()

	fn := w.fn

	if w.destroyed {
		return
	}

	if fn != nil {
		fn(payload)
	}
}

func (w *workerUnit) Flush() {
	defer w.mu.Unlock()
	w.mu.Lock()

	if w.destroyed {
		return
	}

	for {
		select {
		case <-w.queue:
		default:
			return
		}
	}
}

func (w *workerUnit) Push(payload any) {
	defer w.mu.Unlock()
	w.mu.Lock()

	if w.destroyed {
		return
	}

	select {
	case w.queue <- payload:
	default:
	}
}

func (w *workerUnit) Start() {
	defer w.mu.Unlock()
	w.mu.Lock()

	if w.destroyed {
		return
	}

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

	go w.worker()
}

func (w *workerUnit) Stop() {
	defer w.mu.Unlock()
	w.mu.Lock()

	if w.destroyed {
		return
	}

	if w.cancel != nil {
		w.cancel()
	}

	w.ctx = nil
	w.cancel = nil
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

	ticker := w.newTicker()

	defer ticker.Stop()
	defer w.cancel()

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

func (w *workerUnit) Destroy() {
	w.Flush()

	defer w.mu.Unlock()
	w.mu.Lock()

	if w.destroyed {
		return
	}

	w.destroyed = true

	if w.cancel != nil {
		w.cancel()
	}

	w.ctx = nil
	w.cancel = nil

	close(w.queue)
}

func NewWorkerUnit(buffer int) *workerUnit {
	return &workerUnit{
		bufferSize: buffer,
		queue:      make(chan any, buffer),
		destroyed:  false,
	}
}
