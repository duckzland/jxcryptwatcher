package core

import (
	"context"
	"sync"
	"time"
)

type workerUnit struct {
	ops    string
	delay  time.Duration
	fn     func(any) bool
	queue  chan any
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.Mutex
}

func (w *workerUnit) Call() {
	switch w.ops {
	case "executor":
		w.drainAndExecute()

	case "scheduler":
		w.mu.Lock()
		fn := w.fn
		w.mu.Unlock()
		if fn != nil {
			fn(nil)
		}

	case "listener":
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
	case "executor":
		go w.executor()

	case "scheduler":
		go w.scheduler()

	case "listener":
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
		case _, ok := <-w.queue:
			if !ok {
				return
			}

			w.mu.Lock()
			fn := w.fn
			w.mu.Unlock()

			fn(nil)
		}
	}
}

func (w *workerUnit) executor() {
	for {
		select {
		case <-w.ctx.Done():
			return
		case <-time.After(w.delay):
			w.drainAndExecute()
		}
	}
}

func (w *workerUnit) listener() {
	for {
		select {
		case <-w.ctx.Done():
			return
		case _, ok := <-w.queue:
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
				fn(nil)
			}
		}
	}
}

func (w *workerUnit) drainAndExecute() {
	var messages []string
drain:
	for {
		select {
		case msg := <-w.queue:
			if str, ok := msg.(string); ok {
				messages = append(messages, str)
			}
		default:
			break drain
		}
	}

	if len(messages) > 0 {
		w.mu.Lock()
		fn := w.fn
		w.mu.Unlock()
		if fn != nil {
			fn(messages)
		}
	}
}
