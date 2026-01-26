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
	callbacks   sync.Map
	cancelMap   sync.Map
	destroyed   atomic.Bool
}

func (d *debouncer) Init() {
	d.destroyed.Store(false)
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

	if genCounter.Load() > 1000000 {
		genCounter.Store(0)
	}

	gen := genCounter.Add(1)

	if cancelAny, ok := d.cancelMap.Load(key); ok {
		cancelAny.(context.CancelFunc)()
		d.cancelMap.Delete(key)
		d.callbacks.Delete(key)
	}

	ctx, cancel := context.WithCancel(context.Background())
	d.cancelMap.Store(key, cancel)
	d.callbacks.Store(key, fn)

	timer := time.NewTimer(delay)

	go func(gen int64, ctx context.Context, cancel context.CancelFunc, timer *time.Timer) {
		defer func() {
			cancel()

			if val, ok := d.generations.Load(key); ok {
				currentGen := val.(*atomic.Int64).Load()
				if gen == currentGen {
					d.cancelMap.Delete(key)
					d.callbacks.Delete(key)
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
			// Allow last chance cancel
			time.Sleep(1 * time.Millisecond)

			if ctx != nil && ctx.Err() != nil {
				return
			}

			if d.destroyed.Load() {
				return
			}

			if val, ok := d.generations.Load(key); ok {
				currentGen := val.(*atomic.Int64).Load()
				if gen == currentGen {
					if fnAny, ok := d.callbacks.Load(key); ok {
						if fn := fnAny.(func()); fn != nil {
							fn()
						}
					}
				}
			}

		case <-ShutdownCtx.Done():
			return

		case <-ctx.Done():
			return
		}
	}(gen, ctx, cancel, timer)
}

func (d *debouncer) Cancel(key string) {
	if d.destroyed.Load() {
		return
	}
	d.delete(key)
}

func (d *debouncer) Destroy() {
	if d.destroyed.Swap(true) {
		return
	}

	d.cancelMap.Range(func(k, v any) bool {
		d.delete(k.(string))
		return true
	})

	// clear maps
	d.generations = sync.Map{}
	d.callbacks = sync.Map{}
	d.cancelMap = sync.Map{}
}

func (d *debouncer) delete(key string) {
	if val, ok := d.generations.Load(key); ok {
		val.(*atomic.Int64).Add(1)
	}

	if cancelAny, ok := d.cancelMap.Load(key); ok {
		cancelAny.(context.CancelFunc)()
		d.cancelMap.Delete(key)
	}

	d.callbacks.Delete(key)
	d.generations.Delete(key)
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
