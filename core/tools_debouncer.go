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
	callbacks   map[string]func()
	cancelMap   map[string]context.CancelFunc
	destroyed   bool
}

func (d *debouncer) Init() {
	d.mu.Lock()
	d.generations = make(map[string]int)
	d.cancelMap = make(map[string]context.CancelFunc)
	d.callbacks = make(map[string]func())
	d.destroyed = false
	d.mu.Unlock()
}

func (d *debouncer) Call(key string, delay time.Duration, fn func()) {
	d.mu.Lock()
	if d.destroyed {
		d.mu.Unlock()
		return
	}

	if d.generations[key] > 1_000_000 {
		d.generations[key] = 0
	}
	d.generations[key]++
	gen := d.generations[key]

	if cancel, exists := d.cancelMap[key]; exists {
		cancel()
	}

	ctx, cancel := context.WithCancel(context.Background())
	d.cancelMap[key] = cancel
	d.callbacks[key] = fn
	d.mu.Unlock()

	go func(gen int, ctx context.Context, cancel context.CancelFunc) {
		defer func() {
			d.mu.Lock()
			delete(d.cancelMap, key)
			delete(d.callbacks, key)
			delete(d.generations, key)
			d.mu.Unlock()
			cancel()
		}()

		select {
		case <-time.After(delay):
			time.Sleep(1 * time.Millisecond)

			if err := ctx.Err(); err != nil {
				return
			}

			d.mu.Lock()
			currentGen := d.generations[key]
			fn := d.callbacks[key]
			d.mu.Unlock()

			if gen == currentGen {
				if fn != nil {
					fn()
				}
			}
			return

		case <-ctx.Done():
			return
		}
	}(gen, ctx, cancel)
}

func (d *debouncer) Cancel(key string) {
	d.mu.Lock()
	if d.destroyed {
		d.mu.Unlock()
		return
	}
	d.generations[key]++
	if cancel, exists := d.cancelMap[key]; exists {
		cancel()
		delete(d.cancelMap, key)
	}
	delete(d.callbacks, key)
	delete(d.generations, key)
	d.mu.Unlock()
}

func (d *debouncer) Destroy() {
	d.mu.Lock()
	if d.destroyed {
		d.mu.Unlock()
		return
	}
	d.destroyed = true
	d.mu.Unlock()

	d.mu.Lock()
	for key, cancel := range d.cancelMap {
		if cancel != nil {
			cancel()
		}
		delete(d.cancelMap, key)
	}
	d.callbacks = nil
	d.generations = nil
	d.cancelMap = nil
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
