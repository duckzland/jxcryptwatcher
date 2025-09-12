package core

import (
	"sync"
	"time"
)

// Debouncer manages multiple debounce timers keyed by string.
type Debouncer struct {
	mu      sync.Mutex             // protects access to timers and mutexes maps
	timers  map[string]*time.Timer // shared map of timers
	mutexes map[string]*sync.Mutex // per-key mutexes for fn execution
}

// NewDebouncer initializes the debouncer.
func NewDebouncer() *Debouncer {
	return &Debouncer{
		timers:  make(map[string]*time.Timer),
		mutexes: make(map[string]*sync.Mutex),
	}
}

// getMutex returns a per-key mutex, creating one if necessary.
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

// Call debounces the function associated with the given key and delay.
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

		// Clean up the timer after execution
		d.mu.Lock()
		delete(d.timers, key)
		d.mu.Unlock()
	})
	d.mu.Unlock()
}

// Cancel stops and removes the timer for the given key.
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
