package core

import (
	"context"
	"sync"
	"time"
)

var coreDebouncer *debouncer

type debouncer struct {
	mu          sync.Mutex
	generations map[string]int
	callbacks   map[string]func()
	cancelMap   map[string]context.CancelFunc
	destroyed   bool
}

func (d *debouncer) Init() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.generations = make(map[string]int)
	d.callbacks = make(map[string]func())
	d.cancelMap = make(map[string]context.CancelFunc)
	d.destroyed = false
}

func (d *debouncer) Call(key string, delay time.Duration, fn func()) {
	d.mu.Lock()
	if d.destroyed {
		d.mu.Unlock()
		return
	}

	if d.generations[key] > 1000000 {
		d.generations[key] = 0
	}

	d.generations[key]++
	gen := d.generations[key]

	if cancel, exists := d.cancelMap[key]; exists {
		cancel()
		delete(d.cancelMap, key)
		delete(d.callbacks, key)
	}

	ctx, cancel := context.WithCancel(context.Background())
	d.cancelMap[key] = cancel
	d.callbacks[key] = fn
	d.mu.Unlock()

	go func(gen int, ctx context.Context, cancel context.CancelFunc) {
		defer func() {
			cancel()
			d.mu.Lock()
			currentGen := d.generations[key]
			if gen == currentGen {
				delete(d.cancelMap, key)
				delete(d.callbacks, key)
			}
			d.mu.Unlock()
		}()

		select {
		case <-time.After(delay):

			time.Sleep(1 * time.Millisecond)

			if ctx.Err() != nil {
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

	if cancel, ok := d.cancelMap[key]; ok {
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
	for _, cancel := range d.cancelMap {
		if cancel != nil {
			cancel()
		}
	}
	d.generations = nil
	d.callbacks = nil
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
