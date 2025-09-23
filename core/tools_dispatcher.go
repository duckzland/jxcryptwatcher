package core

import (
	"sync"
	"time"
)

var coreDispatcher *dispatcher = nil

type dispatcher struct {
	queue         chan func()
	sem           chan struct{}
	mu            sync.Mutex
	cond          *sync.Cond
	paused        bool
	started       bool
	bufferSize    int
	maxConcurrent int
	delayBetween  time.Duration
	preQueue      []func()
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

	d.paused = false
	d.queue = make(chan func(), d.bufferSize)
	d.sem = make(chan struct{}, d.maxConcurrent)
	d.cond = sync.NewCond(&d.mu)

	for _, fn := range d.preQueue {
		d.queue <- fn
	}
	d.preQueue = nil
}

func (d *dispatcher) Start() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.started || d.queue == nil {
		return
	}
	d.started = true
	go d.run()
}

func (d *dispatcher) Submit(fn func()) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.queue == nil {
		d.preQueue = append(d.preQueue, fn)
		return
	}

	d.queue <- fn
}

func (d *dispatcher) run() {
	for fn := range d.queue {
		d.sem <- struct{}{}

		go func(f func()) {
			defer func() {
				<-d.sem
				if d.delayBetween > 0 {
					time.Sleep(d.delayBetween)
				}
			}()

			d.mu.Lock()
			for d.paused {
				d.cond.Wait()
			}
			d.mu.Unlock()

			f()
		}(fn)
	}
}

func (d *dispatcher) Pause() {
	d.mu.Lock()
	d.paused = true
	d.mu.Unlock()
}

func (d *dispatcher) Resume() {
	d.mu.Lock()
	d.paused = false
	d.mu.Unlock()
	d.cond.Broadcast()
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

func RegisterDispatcher() *dispatcher {
	if coreDispatcher == nil {
		coreDispatcher = &dispatcher{}
	}
	return coreDispatcher
}

func UseDispatcher() *dispatcher {
	return coreDispatcher
}
