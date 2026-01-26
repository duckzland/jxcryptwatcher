package core

import (
	"context"
	"sync/atomic"
	"time"
)

const WORKER_UNIT_IDLE int32 = 1
const WORKER_UNIT_ACTIVE int32 = 2
const WORKER_UNIT_DESTROYED int32 = 3

type workerUnit struct {
	delay       atomic.Int64
	interval    atomic.Int64
	fn          atomic.Value
	getDelay    func() int64
	getInterval func() int64
	queue       chan any
	ctx         atomic.Value
	cancel      atomic.Value
	state       atomic.Int32
}

func (w *workerUnit) Call(payload any) {
	if w.state.Load() != WORKER_UNIT_ACTIVE {
		return
	}
	if fnAny := w.fn.Load(); fnAny != nil {
		fn := fnAny.(func(any) bool)
		fn(payload)
	}
}

func (w *workerUnit) Flush() {
	if w.state.Load() == WORKER_UNIT_DESTROYED {
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
	if w.state.Load() != WORKER_UNIT_ACTIVE {
		return
	}
	select {
	case w.queue <- payload:
	default:
	}
}

func (w *workerUnit) Start() {
	if w.state.Load() == WORKER_UNIT_DESTROYED || w.state.Load() == WORKER_UNIT_ACTIVE {
		return
	}

	if w.getDelay != nil {
		w.delay.Store(w.getDelay())
	}

	if w.getInterval != nil {
		w.interval.Store(w.getInterval())
	}

	w.state.Store(WORKER_UNIT_ACTIVE)

	ctx, cancel := context.WithCancel(context.Background())
	w.ctx.Store(&ctx)
	w.cancel.Store(&cancel)

	go w.worker()
}

func (w *workerUnit) Stop() {
	if w.state.Load() == WORKER_UNIT_DESTROYED {
		return
	}

	w.state.Store(WORKER_UNIT_IDLE)

	if cAny := w.cancel.Load(); cAny != nil {
		if cPtr := cAny.(*context.CancelFunc); cPtr != nil {
			cancel := *cPtr
			if cancel != nil {
				cancel()
			}
		}
	}
}

func (w *workerUnit) Reset() {
	w.Flush()
	if w.state.Load() == WORKER_UNIT_ACTIVE {
		w.Stop()
		w.Start()
	}
}

func (w *workerUnit) Destroy() {
	w.Flush()

	if w.state.Load() == WORKER_UNIT_DESTROYED {
		return
	}
	w.state.Store(WORKER_UNIT_DESTROYED)

	if cAny := w.cancel.Load(); cAny != nil {
		if cPtr := cAny.(*context.CancelFunc); cPtr != nil {
			cancel := *cPtr
			if cancel != nil {
				cancel()
			}
		}
	}

	w.ctx.Store((*context.Context)(nil))
	w.cancel.Store((*context.CancelFunc)(nil))

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
		if w.state.Load() != WORKER_UNIT_ACTIVE {
			if cAny := w.cancel.Load(); cAny != nil {
				if cPtr := cAny.(*context.CancelFunc); cPtr != nil {
					cancel := *cPtr
					if cancel != nil {
						cancel()
					}
				}
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

		if w.state.Load() != WORKER_UNIT_ACTIVE {
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

			if w.state.Load() != WORKER_UNIT_ACTIVE {
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
	}
	unit.fn.Store(fn)
	unit.state.Store(WORKER_UNIT_IDLE)

	unit.ctx.Store((*context.Context)(nil))
	unit.cancel.Store((*context.CancelFunc)(nil))

	return unit
}
