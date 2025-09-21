package core

import (
	"sync"
	"time"
)

type Dispatcher struct {
	queue        chan func()
	sem          chan struct{}
	delayBetween time.Duration

	mu     sync.Mutex
	cond   *sync.Cond
	paused bool
}

func NewDispatcher(bufferSize int, maxConcurrent int, delayBetween time.Duration) *Dispatcher {
	d := &Dispatcher{
		queue:        make(chan func(), bufferSize),
		sem:          make(chan struct{}, maxConcurrent),
		delayBetween: delayBetween,
	}
	d.cond = sync.NewCond(&d.mu)
	go d.run()
	return d
}

func (d *Dispatcher) run() {
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

func (d *Dispatcher) Submit(fn func()) {
	d.queue <- fn
}

func (d *Dispatcher) Pause() {
	d.mu.Lock()
	d.paused = true
	d.mu.Unlock()
}

func (d *Dispatcher) Resume() {
	d.mu.Lock()
	d.paused = false
	d.mu.Unlock()
	d.cond.Broadcast()
}
