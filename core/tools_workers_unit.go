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

func (w *workerUnit) Flush() {
	for {
		select {
		case <-w.queue:
		default:
			return
		}
	}
}

func (w *workerUnit) Call() {
	switch w.ops {
	case WorkerScheduler:
		w.mu.Lock()
		fn := w.fn
		w.mu.Unlock()
		if fn != nil {
			fn(nil)
		}

	case WorkerListener:
		w.Push(struct{}{})
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

func (w *workerUnit) scheduler() {
	t := time.NewTicker(w.delay)
	defer t.Stop()
	for {
		select {
		case <-w.ctx.Done():
			return
		case <-t.C:
			w.Call()
		case x, ok := <-w.queue:
			if !ok {
				return
			}

			w.mu.Lock()
			fn := w.fn
			w.mu.Unlock()

			if fn != nil {
				fn(x)
			}
		}
	}
}

func (w *workerUnit) listener() {
	for {
		select {
		case <-w.ctx.Done():
			return
		case x, ok := <-w.queue:
			if !ok {
				return
			}
			w.mu.Lock()
			fn := w.fn
			delay := w.delay
			w.mu.Unlock()

			if delay > 0 {
				time.Sleep(delay)
			}
			if fn != nil {
				fn(x)
			}
		}
	}
}
