package core

import (
	"context"
	"sync"
	"time"
)

type WorkerMode int

const (
	WorkerScheduler WorkerMode = iota
	WorkerListener
)

type workerUnit struct {
	ops      WorkerMode
	delay    time.Duration
	fn       func(any) bool
	getDelay func() int64
	queue    chan any
	ctx      context.Context
	cancel   context.CancelFunc
	mu       sync.Mutex
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

	w.ctx, w.cancel = context.WithCancel(context.Background())

	switch w.ops {
	case WorkerScheduler:
		go w.scheduler()

	case WorkerListener:
		go w.listener()
	}
}

func (w *workerUnit) Stop() {
	w.mu.Lock()
	if w.cancel != nil {
		w.cancel()
		w.cancel = nil
	}
	w.ctx = nil
	w.mu.Unlock()
}

func (w *workerUnit) Reset() {
	w.Flush()
	w.Stop()
	w.Start()
}

func (w *workerUnit) scheduler() {
	ticker := time.NewTicker(w.delay)

	defer ticker.Stop()

	for {
		if w.ctx == nil {
			return
		}

		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			w.Push(nil)
		case x, ok := <-w.queue:
			if !ok {
				return
			}
			w.Call(x)
		}
	}
}

func (w *workerUnit) listener() {
	for {
		if w.ctx == nil {
			return
		}

		select {
		case <-w.ctx.Done():
			return
		case x, ok := <-w.queue:
			if !ok {
				return
			}

			w.mu.Lock()
			delay := w.delay
			w.mu.Unlock()

			if delay > 0 {
				time.Sleep(delay)
			}

			w.Call(x)
		}
	}
}
