package core

import (
	"sync"
	"time"
)

// Debouncer manages multiple debounce timers keyed by string.
type Debouncer struct {
	mu      sync.Mutex
	timers  map[string]*time.Timer
	mutexes map[string]*sync.Mutex
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

	if timer, exists := d.timers[key]; exists {
		timer.Stop()
	}

	d.timers[key] = time.AfterFunc(delay, func() {
		m.Lock()
		defer m.Unlock()
		fn()
	})
}

func (d *Debouncer) Cancel(key string) {
	m := d.getMutex(key)
	m.Lock()
	defer m.Unlock()

	if timer, exists := d.timers[key]; exists {
		timer.Stop()
		delete(d.timers, key)
	}
}
