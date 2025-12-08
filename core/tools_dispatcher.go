package core

import (
	"context"
	"sync"
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
}

var coreDispatcher *dispatcher = nil

type dispatcher struct {
	queue         chan func()
	mu            sync.Mutex
	paused        bool
	buffer        int
	maxConcurrent int
	delay         time.Duration
	ctx           context.Context
	cancel        context.CancelFunc
	drainer       func()
	destroyed     bool
}

func (d *dispatcher) Init() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.queue != nil {
		return
	}

	if d.buffer == 0 {
		d.buffer = 100
	}
	if d.maxConcurrent == 0 {
		d.maxConcurrent = 4
	}
	if d.delay == 0 {
		d.delay = 16 * time.Millisecond
	}

	d.queue = make(chan func(), d.buffer)
	d.paused = false
	d.ctx, d.cancel = context.WithCancel(context.Background())
}

func (d *dispatcher) Start() {
	for i := 0; i < d.getMaxConcurrent(); i++ {
		if d.isPaused() || d.isDestroyed() {
			return
		}
		go d.worker(i)
	}
}

func (d *dispatcher) Drain() {
	for {
		select {
		case <-d.getQueue():
		default:
			if d.hasDrainer() {
				d.drainer()
			}
			return
		}
	}
}

func (d *dispatcher) Submit(fn func()) {

	if !d.hasQueue() {
		return
	}

	select {
	case d.getQueue() <- fn:
	default:
	}
}

func (d *dispatcher) Pause() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.paused = true

	if d.cancel != nil {
		d.cancel()
		d.ctx = nil
		d.cancel = nil
	}
}

func (d *dispatcher) Resume() {
	d.mu.Lock()
	d.paused = false
	d.ctx, d.cancel = context.WithCancel(context.Background())
	d.mu.Unlock()
	d.Start()
}

func (d *dispatcher) Destroy() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.paused = true
	d.destroyed = true

	if d.cancel != nil {
		d.cancel()
		d.ctx = nil
		d.cancel = nil
	}

	if d.queue != nil {
		close(d.queue)
		d.queue = nil
	}
}

func (d *dispatcher) SetDrainer(fn func()) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.drainer = fn
}

func (d *dispatcher) SetBufferSize(size int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.buffer = size
}

func (d *dispatcher) SetMaxConcurrent(n int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.maxConcurrent = n
}

func (d *dispatcher) SetDelayBetween(delay time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.delay = delay
}

func (d *dispatcher) getDelay() time.Duration {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.delay
}

func (d *dispatcher) isPaused() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.paused
}

func (d *dispatcher) isDestroyed() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.destroyed
}

func (d *dispatcher) hasQueue() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.queue != nil
}

func (d *dispatcher) hasDrainer() bool {
	defer d.mu.Unlock()
	d.mu.Lock()
	return d.drainer != nil
}

func (d *dispatcher) getQueue() chan func() {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.queue
}

func (d *dispatcher) getMaxConcurrent() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.maxConcurrent
}

func (d *dispatcher) call() {

	if !d.hasQueue() || d.isDestroyed() {
		return
	}

	select {
	case fn, ok := <-d.getQueue():
		if !ok {
			return
		}

		if d.isPaused() || d.isDestroyed() {
			return
		}

		if fn != nil {
			fn()
		}

	default:
	}
}

func (d *dispatcher) worker(id int) {

	defer func() {
		if !d.isDestroyed() {
			d.cancel()
			d.Drain()
		}
	}()

	if delay := d.getDelay(); delay > 0 {
		ticker := time.NewTicker(delay)
		defer ticker.Stop()

		for {

			if d.isDestroyed() {
				return
			}

			select {
			case <-ShutdownCtx.Done():
				return

			case <-d.ctx.Done():
				return

			case <-ticker.C:
				d.call()
			}
		}

	} else {
		for {
			if d.isDestroyed() {
				return
			}

			select {
			case <-ShutdownCtx.Done():
				return

			case <-d.ctx.Done():
				return

			default:
				d.call()
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
	d := &dispatcher{
		buffer:        buffer,
		maxConcurrent: maxConcurrent,
		delay:         delay,
		paused:        false,
	}
	d.Init()
	return d
}
