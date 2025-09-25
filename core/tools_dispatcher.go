package core

import (
	"context"
	"sync"
	"time"
)

var coreDispatcher *dispatcher = nil

type dispatcher struct {
	queue         chan func()
	mu            sync.Mutex
	paused        bool
	bufferSize    int
	maxConcurrent int
	delayBetween  time.Duration
	ctx           context.Context
	cancel        context.CancelFunc
}

func (d *dispatcher) Init() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.queue != nil {
		return
	}

	if d.bufferSize == 0 {
		d.bufferSize = 1000
	}
	if d.maxConcurrent == 0 {
		d.maxConcurrent = 4
	}
	if d.delayBetween == 0 {
		d.delayBetween = 16 * time.Millisecond
	}

	d.queue = make(chan func(), d.bufferSize)
	d.paused = false

	d.ctx, d.cancel = context.WithCancel(context.Background())
}

func (d *dispatcher) Start() {
	for i := 0; i < d.maxConcurrent; i++ {
		if d.IsPaused() {
			return
		}
		go d.worker()
	}
}

func (d *dispatcher) worker() {
	ctx := d.ctx
	for {
		select {
		case <-ctx.Done():
			return
		case fn := <-d.queue:
			fn()
			time.Sleep(d.GetDelay())
		}
	}
}

func (d *dispatcher) Submit(fn func()) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.queue == nil {
		return
	}

	select {
	case d.queue <- fn:
	default:
	}
}

func (d *dispatcher) Pause() {
	d.mu.Lock()
	d.paused = true
	if d.cancel != nil {
		d.cancel()
	}
	d.mu.Unlock()
}

func (d *dispatcher) Resume() {
	d.mu.Lock()
	d.paused = false
	d.ctx, d.cancel = context.WithCancel(context.Background())
	d.mu.Unlock()

	d.Start()
}

func (d *dispatcher) SetBufferSize(size int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.bufferSize = size
}

func (d *dispatcher) SetMaxConcurrent(n int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.maxConcurrent = n
}

func (d *dispatcher) SetDelayBetween(delay time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.delayBetween = delay
}

func (d *dispatcher) GetDelay() time.Duration {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.delayBetween
}

func (d *dispatcher) IsPaused() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.paused
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
