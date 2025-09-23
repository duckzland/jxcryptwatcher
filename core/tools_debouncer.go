package core

import (
	"sync"
	"time"
)

var coreDebouncer = &debouncer{
	channels: make(map[string]chan struct{}),
	stoppers: make(map[string]chan struct{}),
}

type debouncer struct {
	mu       sync.Mutex
	channels map[string]chan struct{}
	stoppers map[string]chan struct{}
}

func (d *debouncer) CallOnce(key string, delay time.Duration, fn func()) {
	d.Cancel(key)
	d.Call(key, delay, fn)
}

func (d *debouncer) Call(key string, delay time.Duration, fn func()) {
	d.mu.Lock()
	ch, exists := d.channels[key]
	stopCh := d.stoppers[key]
	if !exists {
		ch = make(chan struct{}, 1)
		stopCh = make(chan struct{})
		d.channels[key] = ch
		d.stoppers[key] = stopCh
		go func() {
			var timer *time.Timer
			for {
				select {
				case <-stopCh:
					if timer != nil {
						timer.Stop()
					}
					return

				case <-ch:
					if timer != nil {
						timer.Stop()
					}
					timer = time.NewTimer(delay)

				innerLoop:
					for {
						select {
						case <-stopCh:
							if timer != nil {
								timer.Stop()
							}
							return

						case <-ch:
							if timer != nil {
								timer.Stop()
							}
							timer = time.NewTimer(delay)

						case <-func() <-chan time.Time {
							if timer != nil {
								return timer.C
							}
							return make(chan time.Time)
						}():
							fn()
							timer = nil
							break innerLoop
						}
					}
				}
			}
		}()
	}
	d.mu.Unlock()

	select {
	case ch <- struct{}{}:
	default:
	}
}

func (d *debouncer) Cancel(key string) {
	d.mu.Lock()
	if stopCh, exists := d.stoppers[key]; exists {
		close(stopCh)
		delete(d.stoppers, key)
	}
	if ch, exists := d.channels[key]; exists {
		close(ch)
		delete(d.channels, key)
	}
	d.mu.Unlock()
}

func UseDebouncer() *debouncer {
	return coreDebouncer
}
