package core

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

var coreDebouncer *debouncer

type debouncer struct {
	generations sync.Map
	registry    *cancelRegistry
	destroyed   atomic.Bool
}

func (d *debouncer) Init() {
	d.destroyed.Store(false)
	d.registry = NewCancelRegistry()
}

func (d *debouncer) Call(key string, delay time.Duration, fn func()) {
	if d.destroyed.Load() {
		return
	}

	var genCounter *atomic.Int64
	if val, ok := d.generations.Load(key); ok {
		genCounter = val.(*atomic.Int64)
	} else {
		genCounter = &atomic.Int64{}
		d.generations.Store(key, genCounter)
	}

	if genCounter.Load() > 1_000_000 {
		genCounter.Store(0)
	}

	gen := genCounter.Add(1)

	if cancel, ok := d.registry.Get(key); ok {
		cancel()
		d.registry.Delete(key)
	}

	ctx, cancel := context.WithCancel(context.Background())
	d.registry.Set(key, cancel)

	timer := time.NewTimer(delay)

	go func(gen int64, ctx context.Context, cancel context.CancelFunc, timer *time.Timer, fn func()) {
		defer func() {
			cancel()

			if val, ok := d.generations.Load(key); ok {
				currentGen := val.(*atomic.Int64).Load()
				if gen == currentGen {
					d.registry.Delete(key)
				}
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
			time.Sleep(1 * time.Millisecond)

			if ctx != nil && ctx.Err() != nil {
				return
			}

			if d.destroyed.Load() {
				return
			}

			if val, ok := d.generations.Load(key); ok {
				currentGen := val.(*atomic.Int64).Load()
				if gen == currentGen && fn != nil {
					// Logf("Debouncer fired for %s/%d", key, gen)
					fn()
				}
			}

		case <-ShutdownCtx.Done():
			// Logf("Debouncer shutdown for %s/%d", key, gen)
			return

		case <-ctx.Done():
			// Logf("Debouncer cancelled for %s/%d", key, gen)
			return
		}
	}(gen, ctx, cancel, timer, fn)
}

func (d *debouncer) Cancel(key string) {
	if d.destroyed.Load() {
		return
	}

	if val, ok := d.generations.Load(key); ok {
		val.(*atomic.Int64).Add(1)
	}

	if cancel, ok := d.registry.Get(key); ok {
		cancel()
		d.registry.Delete(key)
	}

	d.generations.Delete(key)
}

func (d *debouncer) Destroy() {
	if d.destroyed.Swap(true) {
		return
	}

	d.registry.Destroy()

	d.generations = sync.Map{}
	d.registry = NewCancelRegistry()
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
