package core

import (
	"context"
	"sync"
	"time"
)

type Dispatcher interface {
	Init()
	Submit(fn func())
	GetDelay() time.Duration
	IsPaused() bool
	Pause()
	Resume()
	Start()
	Drain()
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
	ticker        *time.Ticker
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
	for i := 0; i < d.maxConcurrent; i++ {
		if d.IsPaused() {
			return
		}

		go d.worker()
	}
}

func (d *dispatcher) Drain() {
	d.mu.Lock()
	queue := d.queue
	d.mu.Unlock()
	for {
		select {
		case <-queue:
		default:
			if d.drainer != nil {
				d.drainer()
			}
			return
		}
	}
}

func (d *dispatcher) Submit(fn func()) {
	d.mu.Lock()
	queue := d.queue
	d.mu.Unlock()

	if queue == nil {
		return
	}

	select {
	case queue <- fn:
	default:
	}
}

func (d *dispatcher) Pause() {
	d.mu.Lock()
	d.paused = true
	d.mu.Unlock()
	if d.ticker != nil {
		d.ticker.Stop()
	}
	if d.cancel != nil {
		d.cancel()
		d.ctx = nil
		d.cancel = nil
	}
}

func (d *dispatcher) Resume() {
	d.mu.Lock()
	d.paused = false
	d.ctx = nil
	d.cancel = nil
	d.ctx, d.cancel = context.WithCancel(context.Background())
	d.mu.Unlock()

	d.Start()
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

func (d *dispatcher) GetDelay() time.Duration {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.delay
}

func (d *dispatcher) IsPaused() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.paused
}

func (d *dispatcher) newTicker() *time.Ticker {
	if d.GetDelay() > 0 {
		return time.NewTicker(d.GetDelay())
	}
	return &time.Ticker{C: make(chan time.Time)}
}

func (d *dispatcher) worker() {
	d.ticker = d.newTicker()
	defer d.ticker.Stop()

	for {
		if d.ctx == nil {
			return
		}

		if d.IsPaused() {
			d.ticker.Stop()
		}

		select {
		case <-d.ctx.Done():
			d.ticker.Stop()
			return

		case <-d.ticker.C:
			select {
			case fn := <-d.queue:
				if d.IsPaused() {
					continue
				}
				fn()
			default:
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
