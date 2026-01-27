package core

import (
	"context"
	"sync/atomic"
	"time"
)

type workerUnit struct {
	delay       atomic.Int64
	interval    atomic.Int64
	fn          atomic.Value
	getDelay    func() int64
	getInterval func() int64
	queue       chan any
	ctx         atomic.Value
	registry    *cancelRegistry
	state       *stateManager
}

func (w *workerUnit) Call(payload any) {
	if !w.state.Is(STATE_RUNNING) {
		return
	}
	if fnAny := w.fn.Load(); fnAny != nil {
		fn := fnAny.(func(any) bool)
		fn(payload)
	}
}

func (w *workerUnit) Flush() {
	if w.state.Is(STATE_DESTROYED) {
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
	if !w.state.Is(STATE_RUNNING) {
		return
	}
	select {
	case w.queue <- payload:
	default:
	}
}

func (w *workerUnit) Start() {
	if w.state.Is(STATE_DESTROYED) || w.state.Is(STATE_RUNNING) {
		return
	}

	if w.getDelay != nil {
		w.delay.Store(w.getDelay())
	}

	if w.getInterval != nil {
		w.interval.Store(w.getInterval())
	}

	w.state.Change(STATE_RUNNING)

	ctx, cancel := context.WithCancel(context.Background())
	w.ctx.Store(&ctx)
	w.registry.Set("worker", cancel)

	go w.worker()
}

func (w *workerUnit) Stop() {
	if w.state.Is(STATE_DESTROYED) {
		return
	}

	w.state.Change(STATE_PAUSED)

	if cancel, ok := w.registry.Get("worker"); ok {
		cancel()
	}
}

func (w *workerUnit) Reset() {
	w.Flush()
	if w.state.Is(STATE_RUNNING) {
		w.Stop()
		w.Start()
	}
}

func (w *workerUnit) Destroy() {
	w.Flush()

	if w.state.Is(STATE_DESTROYED) {
		return
	}
	w.state.Change(STATE_DESTROYED)

	w.registry.Destroy()

	w.ctx.Store((*context.Context)(nil))

	close(w.queue)
}

func (w *workerUnit) newTicker() *time.Ticker {
	interval := w.interval.Load()
	if interval > 0 {
		return time.NewTicker(time.Duration(interval) * time.Millisecond)
	}
	return &time.Ticker{C: make(chan time.Time)}
}

func (w *workerUnit) worker() {
	ticker := w.newTicker()
	defer ticker.Stop()

	defer func() {
		if !w.state.Is(STATE_RUNNING) {
			if cancel, ok := w.registry.Get("worker"); ok {
				cancel()
			}
		}
	}()

	for {
		ctxAny := w.ctx.Load()
		if ctxAny == nil {
			return
		}
		ctxPtr := ctxAny.(*context.Context)
		if ctxPtr == nil || *ctxPtr == nil {
			return
		}
		ctx := *ctxPtr
		delay := w.delay.Load()

		if !w.state.Is(STATE_RUNNING) {
			return
		}

		select {
		case <-ShutdownCtx.Done():
			return

		case <-ctx.Done():
			return

		case <-ticker.C:
			w.Push(nil)

		case x, ok := <-w.queue:
			if !ok {
				return
			}

			if delay > 0 {
				time.Sleep(time.Duration(delay) * time.Millisecond)
			}

			if !w.state.Is(STATE_RUNNING) {
				return
			}

			w.Call(x)
		}
	}
}

func NewWorkerUnit(size int, getDelay func() int64, getInterval func() int64, fn func(any) bool) *workerUnit {
	unit := &workerUnit{
		queue:       make(chan any, size),
		getDelay:    getDelay,
		getInterval: getInterval,
		registry:    NewCancelRegistry(),
	}
	unit.fn.Store(fn)
	unit.state = NewStateManager(STATE_PAUSED)

	unit.ctx.Store((*context.Context)(nil))

	return unit
}
