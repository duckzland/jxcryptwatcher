package core

import (
	"context"
	"sync/atomic"
	"time"
)

type Dispatcher interface {
	Init()
	Submit(fn func())
	Pause()
	Resume()
	Start()
	Drain()
	Destroy()
	SetDrainer(fn func())
	SetBufferSize(int)
	SetDelayBetween(time.Duration)
	SetMaxConcurrent(int)
	SetKey(string)
}

var coreDispatcher *dispatcher = nil

type dispatcher struct {
	queue         chan func()
	buffer        atomic.Int32
	maxConcurrent atomic.Int32
	generated     atomic.Int32
	delay         atomic.Int64
	ctx           atomic.Value
	cancel        atomic.Value
	drainer       atomic.Value
	key           atomic.Value
	state         *stateManager
}

func (d *dispatcher) Init() {
	if d.queue != nil {
		return
	}

	if d.buffer.Load() == 0 {
		d.buffer.Store(100)
	}
	if d.maxConcurrent.Load() == 0 {
		d.maxConcurrent.Store(4)
	}
	if d.delay.Load() == 0 {
		d.delay.Store(int64(16 * time.Millisecond))
	}

	d.generated.Store(0)
	d.queue = make(chan func(), int(d.buffer.Load()))

	ctx, cancel := context.WithCancel(context.Background())
	d.ctx.Store(&ctx)
	d.cancel.Store(&cancel)

	d.state = NewStateManager(STATE_RUNNING)
}

func (d *dispatcher) Start() {
	if d.state.Is(STATE_DESTROYED) {
		return
	}

	if d.generated.Load() == d.maxConcurrent.Load() {
		return
	}

	max := d.maxConcurrent.Load()
	for i := int32(1); i <= max; i++ {
		if d.state.Is(STATE_DESTROYED) {
			return
		}
		go d.worker(i)
		d.generated.Add(1)
	}

	keyAny := d.key.Load()
	key := ""
	if keyAny != nil {
		key = keyAny.(string)
	}

	Logf("Initializing Dispatcher [%s]: %d/%d", key, d.generated.Load(), d.maxConcurrent.Load())
}

func (d *dispatcher) Pause() {
	if !d.state.Is(STATE_DESTROYED) {
		d.state.Change(STATE_PAUSED)
		d.Drain()
	}
}

func (d *dispatcher) Resume() {
	if !d.state.Is(STATE_DESTROYED) {
		d.state.Change(STATE_RUNNING)
	}
}

func (d *dispatcher) Drain() {
	for {
		if d.queue == nil {
			if fnAny := d.drainer.Load(); fnAny != nil {
				fnAny.(func())()
			}
			return
		}
		select {
		case <-d.queue:
		default:
			if fnAny := d.drainer.Load(); fnAny != nil {
				fnAny.(func())()
			}
			return
		}
	}
}

func (d *dispatcher) Submit(fn func()) {
	switch d.state.Get() {
	case STATE_PAUSED, STATE_DESTROYED:
		return
	}

	if d.queue == nil {
		return
	}

	select {
	case d.queue <- fn:
	default:
	}
}

func (d *dispatcher) Destroy() {
	if d.state.Is(STATE_DESTROYED) {
		return
	}
	d.state.Change(STATE_DESTROYED)

	if cancelPtr := d.cancel.Load(); cancelPtr != nil {
		(*cancelPtr.(*context.CancelFunc))()
		d.ctx.Store((*context.Context)(nil))
		d.cancel.Store((*context.CancelFunc)(nil))
	}

	if d.queue != nil {
		close(d.queue)
		d.queue = nil
	}
}

func (d *dispatcher) SetDrainer(fn func()) {
	d.drainer.Store(fn)
}

func (d *dispatcher) SetBufferSize(size int) {
	d.buffer.Store(int32(size))
}

func (d *dispatcher) SetMaxConcurrent(n int) {
	d.maxConcurrent.Store(int32(n))
}

func (d *dispatcher) SetDelayBetween(delay time.Duration) {
	d.delay.Store(int64(delay))
}

func (d *dispatcher) SetKey(key string) {
	d.key.Store(key)
}

func (d *dispatcher) worker(id int32) {
	defer func() {
		d.generated.Add(-1)
		if !d.state.Is(STATE_DESTROYED) {
			if cancelPtr := d.cancel.Load(); cancelPtr != nil {
				(*cancelPtr.(*context.CancelFunc))()
			}
			d.Drain()
		}
	}()

	for {
		if d.state.Is(STATE_DESTROYED) {
			return
		}
		if d.queue == nil {
			return
		}

		var ctxDone <-chan struct{}
		if ctxPtr := d.ctx.Load(); ctxPtr != nil {
			ctxDone = (*ctxPtr.(*context.Context)).Done()
		}

		select {
		case <-ShutdownCtx.Done():
			return

		case <-ctxDone:
			return

		case fn, ok := <-d.queue:
			if !ok || d.state.Is(STATE_DESTROYED) {
				return
			}
			if fn != nil {
				fn()

				delay := time.Duration(d.delay.Load())
				if delay < time.Millisecond {
					delay = time.Millisecond
				}
				time.Sleep(delay)
			}
		}
	}
}

func RegisterDispatcher() *dispatcher {
	if coreDispatcher == nil {
		coreDispatcher = &dispatcher{}
	}
	return coreDispatcher
}

func UseDispatcher() *dispatcher {
	return coreDispatcher
}

func NewDispatcher(buffer, maxConcurrent int, delay time.Duration) *dispatcher {
	d := &dispatcher{}

	d.buffer.Store(int32(buffer))
	d.maxConcurrent.Store(int32(maxConcurrent))
	d.delay.Store(int64(delay))
	d.state = NewStateManager(STATE_RUNNING)
	d.Init()

	return d
}
