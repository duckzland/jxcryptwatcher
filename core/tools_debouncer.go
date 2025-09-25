package core

import (
	"context"
	"sync"
	"time"
)

var coreDebouncer *debouncer = nil

type debouncer struct {
	mu          sync.Mutex
	generations map[string]int
	cancelMap   map[string]context.CancelFunc
}

func (d *debouncer) Init() {
	d.mu.Lock()
	d.generations = make(map[string]int)
	d.cancelMap = make(map[string]context.CancelFunc)
	d.mu.Unlock()
}

func (d *debouncer) Call(key string, delay time.Duration, fn func()) {
	d.mu.Lock()
	if d.generations[key] > 1000000 {
		d.generations[key] = 0
	}
	d.generations[key]++
	gen := d.generations[key]

	if cancel, exists := d.cancelMap[key]; exists {
		cancel()
	}

	ctx, cancel := context.WithCancel(context.Background())
	d.cancelMap[key] = cancel
	d.mu.Unlock()

	go func(gen int, ctx context.Context) {
		select {
		case <-time.After(delay):
			d.mu.Lock()
			currentGen := d.generations[key]
			d.mu.Unlock()

			if gen != currentGen {
				return
			}
			fn()
		case <-ctx.Done():
			return
		}
	}(gen, ctx)
}

func (d *debouncer) Cancel(key string) {
	d.mu.Lock()
	d.generations[key]++
	if cancel, exists := d.cancelMap[key]; exists {
		cancel()
		delete(d.cancelMap, key)
	}
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
