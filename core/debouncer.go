package core

import (
	"sync"
	"time"
)

type Debouncer struct {
	mu      sync.Mutex
	timers  map[string]*time.Timer
	mutexes map[string]*sync.Mutex
}

func NewDebouncer() *Debouncer {
	return &Debouncer{
		timers:  make(map[string]*time.Timer),
		mutexes: make(map[string]*sync.Mutex),
	}
}

func (d *Debouncer) getMutex(key string) *sync.Mutex {
	d.mu.Lock()
	defer d.mu.Unlock()

	if m, ok := d.mutexes[key]; ok {
		return m
	}
	m := &sync.Mutex{}
	d.mutexes[key] = m
	return m
}

func (d *Debouncer) Call(key string, delay time.Duration, fn func()) {
	m := d.getMutex(key)
	m.Lock()
	defer m.Unlock()

	d.mu.Lock()
	if timer, exists := d.timers[key]; exists {
		timer.Stop()
	}
	d.timers[key] = time.AfterFunc(delay, func() {
		m.Lock()
		defer m.Unlock()
		fn()

		d.mu.Lock()
		delete(d.timers, key)
		d.mu.Unlock()
	})
	d.mu.Unlock()
}

func (d *Debouncer) Cancel(key string) {
	m := d.getMutex(key)
	m.Lock()
	defer m.Unlock()

	d.mu.Lock()
	if timer, exists := d.timers[key]; exists {
		timer.Stop()
		delete(d.timers, key)
	}
	d.mu.Unlock()
}
