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
	state       *stateManager
}

func (d *debouncer) Init() {
	if d.state != nil {
		return
	}

	d.state = NewStateManager(STATE_RUNNING)
	d.registry = NewCancelRegistry()
}

func (d *debouncer) Call(key string, delay time.Duration, fn func()) {
	if d.state.Is(STATE_DESTROYED) {
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

	d.registry.Cancel(key)

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

			if d.state.Is(STATE_DESTROYED) {
				return
			}

			if val, ok := d.generations.Load(key); ok {
				currentGen := val.(*atomic.Int64).Load()
				if gen == currentGen && fn != nil {
					fn()
				}
			}

		case <-ShutdownCtx.Done():
			return

		case <-ctx.Done():
			return
		}
	}(gen, ctx, cancel, timer, fn)
}

func (d *debouncer) Cancel(key string) {
	if val, ok := d.generations.Load(key); ok {
		val.(*atomic.Int64).Add(1)
	}

	d.registry.Cancel(key)
	d.generations.Delete(key)
}

func (d *debouncer) Destroy() {
	if d.state.Is(STATE_DESTROYED) {
		return
	}

	d.state.Change(STATE_DESTROYED)

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
