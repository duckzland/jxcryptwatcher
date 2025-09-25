package core

import (
	"sync"
	"time"
)

var coreDebouncer *debouncer = nil

type debouncer struct {
	mu          sync.Mutex
	generations map[string]int
}

func (d *debouncer) Init() {
	d.mu.Lock()
	d.generations = make(map[string]int)
	d.mu.Unlock()
}

func (d *debouncer) Call(key string, delay time.Duration, fn func()) {
	d.mu.Lock()
	d.generations[key]++
	gen := d.generations[key]
	d.mu.Unlock()

	go func(gen int) {
		time.Sleep(delay)

		d.mu.Lock()
		currentGen := d.generations[key]
		d.mu.Unlock()

		if gen != currentGen {
			return
		}

		fn()
	}(gen)
}

func (d *debouncer) Cancel(key string) {
	d.mu.Lock()
	d.generations[key]++
	d.mu.Unlock()
}

func RegisterDebouncer() *debouncer {
	if coreDebouncer == nil {
		coreDebouncer = &debouncer{}
		coreDebouncer.Init()
	}
	return coreDebouncer
}

func UseDebouncer() *debouncer {
	return coreDebouncer
}
