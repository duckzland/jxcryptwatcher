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
	active      bool
}

func (w *workerUnit) Call(payload any) {
	w.mu.Lock()
	if w.destroyed || !w.active {
		w.mu.Unlock()
		return
	}
	fn := w.fn
	w.mu.Unlock()

	if fn != nil {
		fn(payload)
	}
}

func (w *workerUnit) Flush() {
	w.mu.Lock()
	if w.destroyed {
		w.mu.Unlock()
		return
	}
	q := w.queue
	w.mu.Unlock()

	for {
		select {
		case <-q:
		default:
			return
		}
	}
}

func (w *workerUnit) Push(payload any) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.destroyed || !w.active {
		return
	}

	select {
	case w.queue <- payload:
	default:
	}
}

func (w *workerUnit) Start() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.destroyed {
		return
	}

	if w.active {
		return
	}

	if w.getDelay != nil {
		w.delay = time.Duration(w.getDelay()) * time.Millisecond
	}

	if w.getInterval != nil {
		w.interval = time.Duration(w.getInterval()) * time.Millisecond
	}

	w.active = true

	w.ctx, w.cancel = context.WithCancel(context.Background())

	go w.worker()
}

func (w *workerUnit) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.destroyed {
		return
	}

	w.active = false

	if w.cancel != nil {
		w.cancel()
	}
}

func (w *workerUnit) Reset() {
	w.Flush()

	w.mu.Lock()
	active := w.active
	w.mu.Unlock()

	if active {
		w.Stop()
		w.Start()
	}
}

func (w *workerUnit) newTicker() *time.Ticker {
	w.mu.Lock()
	interval := w.interval
	w.mu.Unlock()

	if interval > 0 {
		return time.NewTicker(interval)
	}
	return &time.Ticker{C: make(chan time.Time)}
}

func (w *workerUnit) worker() {

	ticker := w.newTicker()

	defer ticker.Stop()
	defer func() {
		w.mu.Lock()
		cancel := w.cancel
		active := w.active
		w.mu.Unlock()
		if cancel != nil && !active {
			cancel()
		}
	}()

	for {
		w.mu.Lock()
		ctx := w.ctx
		delay := w.delay
		active := w.active
		queue := w.queue
		w.mu.Unlock()

		if ctx == nil || !active {
			return
		}

		select {
		case <-ShutdownCtx.Done():
			return

		case <-ctx.Done():
			return

		case <-ticker.C:
			w.Push(nil)

		case x, ok := <-queue:
			if !ok {
				return
			}

			if delay > 0 {
				time.Sleep(delay)
			}

			w.mu.Lock()
			active = w.active
			w.mu.Unlock()

			if !active {
				return
			}

			w.Call(x)
		}
	}
}

func (w *workerUnit) Destroy() {
	w.Flush()

	w.mu.Lock()
	defer w.mu.Unlock()

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
		active:     false,
	}
}
