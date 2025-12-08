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
	defer d.mu.Unlock()
	d.mu.Lock()

	if d.destroyed {
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

	timer := time.NewTimer(delay)

	go func(gen int, ctx context.Context, cancel context.CancelFunc, timer *time.Timer) {
		defer func() {
			d.mu.Lock()
			defer d.mu.Unlock()

			cancel()

			currentGen := d.generations[key]
			if gen == currentGen {
				delete(d.cancelMap, key)
				delete(d.callbacks, key)
			}

			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
		}()

		select {
		case <-timer.C:

			// Allow last chance cancel
			time.Sleep(1 * time.Millisecond)

			if ctx.Err() != nil {
				return
			}

			d.mu.Lock()
			destroyed := d.destroyed
			currentGen := d.generations[key]
			fn := d.callbacks[key]
			d.mu.Unlock()

			if destroyed {
				return
			}

			if gen == currentGen && fn != nil {
				fn()
			}

		case <-ShutdownCtx.Done():
			return

		case <-ctx.Done():
			return
		}

	}(gen, ctx, cancel, timer)
}

func (d *debouncer) Cancel(key string) {
	defer d.mu.Unlock()
	d.mu.Lock()

	if d.destroyed {
		return
	}

	d.internalDestroy(key)
}

func (d *debouncer) Destroy() {
	defer d.mu.Unlock()
	d.mu.Lock()

	if d.destroyed {
		return
	}

	d.destroyed = true
	for key := range d.cancelMap {
		d.internalDestroy(key)
	}

	d.generations = nil
	d.callbacks = nil
	d.cancelMap = nil
}

func (d *debouncer) internalDestroy(key string) {

	d.generations[key]++

	if d.cancelMap != nil {
		if cancel, exists := d.cancelMap[key]; exists {
			cancel()
			delete(d.cancelMap, key)
		}
	}

	if d.callbacks != nil {
		if _, exists := d.callbacks[key]; exists {
			delete(d.callbacks, key)
		}
	}

	if d.generations != nil {
		if _, exists := d.generations[key]; exists {
			delete(d.generations, key)
		}
	}
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
