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
	SetKey(string)
}

var coreDispatcher *dispatcher = nil

type dispatcher struct {
	queue         chan func()
	mu            sync.Mutex
	buffer        int
	maxConcurrent int
	generated     int
	delay         time.Duration
	ctx           context.Context
	cancel        context.CancelFunc
	drainer       func()
	destroyed     bool
	paused        bool
	key           string
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

	d.generated = 0
	d.queue = make(chan func(), d.buffer)
	d.ctx, d.cancel = context.WithCancel(context.Background())
}

func (d *dispatcher) Start() {
	d.mu.Lock()
	generated := d.generated
	d.mu.Unlock()

	if generated == d.getMaxConcurrent() {
		return
	}

	for i := 1; i <= d.getMaxConcurrent(); i++ {
		if d.isDestroyed() {
			return
		}
		go d.worker(i)

		d.mu.Lock()
		d.generated++
		d.mu.Unlock()
	}

	Logf("Initializing Dispatcher [%s]: %d/%d", d.key, d.generated, d.getMaxConcurrent())
}

func (d *dispatcher) Pause() {
	d.mu.Lock()
	d.paused = true
	d.mu.Unlock()
	d.Drain()
}

func (d *dispatcher) Resume() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.paused = false
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
	if d.isPaused() {
		return
	}

	if !d.hasQueue() {
		return
	}

	select {
	case d.getQueue() <- fn:
	default:
	}
}

func (d *dispatcher) Destroy() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.destroyed {
		return
	}

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

func (d *dispatcher) SetKey(key string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.key = key
}

func (d *dispatcher) getDelay() time.Duration {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.delay
}

func (d *dispatcher) isDestroyed() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.destroyed
}

func (d *dispatcher) isPaused() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.paused
}

func (d *dispatcher) hasQueue() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.queue != nil
}

func (d *dispatcher) hasDrainer() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
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

func (d *dispatcher) worker(id int) {
	defer func() {
		d.mu.Lock()
		cancel := d.cancel
		d.generated--
		d.mu.Unlock()

		if !d.isDestroyed() {
			if cancel != nil {
				cancel()
			}
			d.Drain()
		}
	}()

	for {
		if d.isDestroyed() {
			return
		}

		q := d.getQueue()
		if q == nil {
			return
		}

		select {
		case <-ShutdownCtx.Done():
			return

		case <-d.ctx.Done():
			return

		case fn, ok := <-q:
			if !ok || d.isDestroyed() {
				return

			}
			if fn != nil {
				// Logf("Dispatcher firing worker: [%s] %d", d.key, id)
				fn()

				delay := max(d.getDelay(), time.Millisecond)
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
	d := &dispatcher{
		buffer:        buffer,
		maxConcurrent: maxConcurrent,
		delay:         delay,
		destroyed:     false,
	}
	d.Init()
	return d
}
