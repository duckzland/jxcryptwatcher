package core

import (
	"sync"
	"time"
)

var coreDebouncer *debouncer = nil

type debouncer struct {
	mu     sync.Mutex
	timers map[string]*time.Timer
}

func (d *debouncer) Init() {
	d.mu.Lock()
	d.timers = make(map[string]*time.Timer)
	d.mu.Unlock()
}

func (d *debouncer) Call(key string, delay time.Duration, fn func()) {

	d.Cancel(key)

	d.mu.Lock()
	var timer *time.Timer
	timer = time.AfterFunc(delay, func() {
		d.mu.Lock()
		current, ok := d.timers[key]
		if !ok || current != timer {
			d.mu.Unlock()
			return
		}
		delete(d.timers, key)
		d.mu.Unlock()

		go fn()
	})

	d.timers[key] = timer
	d.mu.Unlock()
}

func (d *debouncer) Cancel(key string) {
	d.mu.Lock()
	if timer, exists := d.timers[key]; exists {
		timer.Stop()
		delete(d.timers, key)
	}
	d.mu.Unlock()
}

func RegisterDebouncer() *debouncer {
	if coreDebouncer == nil {
		coreDebouncer = &debouncer{}
	}
	return coreDebouncer
}

func UseDebouncer() *debouncer {
	return coreDebouncer
}
