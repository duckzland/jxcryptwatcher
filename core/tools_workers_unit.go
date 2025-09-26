package core

import (
	"context"
	"sync"
	"time"
)

type workerUnit struct {
	ops    int
	delay  time.Duration
	fn     func(any) bool
	queue  chan any
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.Mutex
}

func (w *workerUnit) Call() {
	switch w.ops {
	case WORKER_SCHEDULER:
		w.mu.Lock()
		fn := w.fn
		w.mu.Unlock()
		if fn != nil {
			fn(nil)
		}

	case WORKER_LISTENER:
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

	w.ctx, w.cancel = context.WithCancel(context.Background())

	switch w.ops {

	case WORKER_SCHEDULER:
		go w.scheduler()

	case WORKER_LISTENER:
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
