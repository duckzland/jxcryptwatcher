package core

import "time"

type Dispatcher struct {
	queue        chan func()
	sem          chan struct{}
	delayBetween time.Duration
}

func NewDispatcher(bufferSize int, maxConcurrent int, delayBetween time.Duration) *Dispatcher {
	d := &Dispatcher{
		queue:        make(chan func(), bufferSize),
		sem:          make(chan struct{}, maxConcurrent),
		delayBetween: delayBetween,
	}
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
			f()
		}(fn)
	}
}

func (d *Dispatcher) Submit(fn func()) {
	d.queue <- fn
}
