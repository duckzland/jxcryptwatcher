package core

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const FETCHER_MODE_DISPATCH = 0
const FETCHER_MODE_CALL = 1

type FetcherInterface interface {
	Fetch(ctx context.Context, payload any, callback func(FetchResultInterface))
}

var fetcherManager *fetcher = nil

type fetcher struct {
	fetchers      sync.Map
	conditions    sync.Map
	activeWorkers sync.Map
	dispatcher    *dispatcher
	destroyed     atomic.Bool
	paused        atomic.Bool
}

func (m *fetcher) Init() {
	if m.dispatcher != nil {
		return
	}

	m.destroyed.Store(false)
	m.paused.Store(false)

	m.dispatcher = NewDispatcher(NETWORKING_MAXIMUM_CONNECTION*2, NETWORKING_MAXIMUM_CONNECTION, 50*time.Millisecond)

	m.dispatcher.SetKey("Fetchers")
	m.dispatcher.Start()
}

func (m *fetcher) Register(key string, f FetcherInterface, cond func() bool) {
	if m.destroyed.Load() || m.paused.Load() {
		return
	}

	if _, ok := m.fetchers.Load(key); ok {
		m.Deregister(key)
	}

	m.fetchers.Store(key, f)
	m.conditions.Store(key, cond)
}

func (m *fetcher) Deregister(key string) {
	if cAny, ok := m.activeWorkers.Load(key); ok {
		if cancel, ok := cAny.(context.CancelFunc); ok && cancel != nil {
			cancel()
		}
		m.activeWorkers.Delete(key)
	}

	m.fetchers.Delete(key)
	m.conditions.Delete(key)
}

func (m *fetcher) Call(payloads map[string][]string, preprocess func(totalJob int), onSuccess func(map[string]FetchResultInterface), onCancel func()) {
	if m.destroyed.Load() || m.paused.Load() {
		return
	}

	total := 0
	mapKey := ""
	mode := FETCHER_MODE_DISPATCH

	for key, items := range payloads {
		if len(payloads) == 1 && len(items) == 1 {
			mode = FETCHER_MODE_CALL
		}

		if condAny, ok := m.conditions.Load(key); ok {
			if cond := condAny.(func() bool); cond != nil && cond() {
				total += len(items)
			}
		} else {
			total += len(items)
		}

		mapKey += key + strings.Join(items, "|")
	}

	if preprocess != nil {
		preprocess(total)
	}

	if oldCancelAny, ok := m.activeWorkers.Load(mapKey); ok {
		oldCancel := oldCancelAny.(context.CancelFunc)
		oldCancel()
		m.activeWorkers.Delete(mapKey)
	}

	if total == 0 {
		if onCancel != nil {
			onCancel()
		}

		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.activeWorkers.Store(mapKey, cancel)

	switch mode {
	case FETCHER_MODE_CALL:
		for key, items := range payloads {
			if condAny, ok := m.conditions.Load(key); ok {
				if cond := condAny.(func() bool); cond != nil && !cond() {
					continue
				}
			}

			for _, item := range items {
				k := key
				payload := item

				m.dispatcher.Submit(func() {
					defer func() {
						if ctx.Err() != nil {
							if onCancel != nil {
								onCancel()
							}
							return
						}

						cancel()
						m.activeWorkers.Delete(mapKey)
					}()

					if ctx.Err() != nil {
						return
					}

					m.execute(ctx, k, payload, func(result FetchResultInterface) {
						if ctx.Err() != nil {
							return
						}

						rs := make(map[string]FetchResultInterface, 1)
						rs[payload] = result

						if onSuccess != nil {
							onSuccess(rs)
						}
					})
				})
			}
		}

	case FETCHER_MODE_DISPATCH:
		var cancelled atomic.Bool
		cancelled.Store(false)

		results := sync.Map{}
		done := make(chan struct{}, total)

		conds := make(map[string]func() bool)
		m.conditions.Range(func(k, v any) bool {
			conds[k.(string)] = v.(func() bool)
			return true
		})

		for key, items := range payloads {
			if cond, ok := conds[key]; ok && cond != nil && !cond() {
				continue
			}
			for _, item := range items {
				k := key
				payload := item
				m.dispatcher.Submit(func() {
					defer func() {
						if !cancelled.Load() {
							done <- struct{}{}
						}
					}()
					if ctx.Err() != nil || cancelled.Load() {
						return
					}
					m.execute(ctx, k, payload, func(result FetchResultInterface) {
						if ctx.Err() != nil || cancelled.Load() {
							return
						}
						results.Store(payload, result)
					})
				})
			}
		}

		if m.paused.Load() || m.destroyed.Load() {
			cancelled.Store(true)
			cancel()
			m.activeWorkers.Delete(mapKey)
			if onCancel != nil {
				onCancel()
			}
			close(done)
			return
		}

		go func() {
			defer func() {
				cancel()
				m.activeWorkers.Delete(mapKey)
				if cancelled.Load() {
					if onCancel != nil {
						onCancel()
					}
					close(done)
				}
			}()

			count := 0
			for {
				select {
				case <-ShutdownCtx.Done():
					cancelled.Store(true)
					return

				case <-ctx.Done():
					cancelled.Store(true)
					return

				case <-done:
					count++
					if count == total {
						if onSuccess != nil {
							merged := make(map[string]FetchResultInterface)
							results.Range(func(k, v any) bool {
								merged[k.(string)] = v.(FetchResultInterface)
								return true
							})
							onSuccess(merged)
						}

						close(done)
						return
					}
				}
			}
		}()
	}
}

func (m *fetcher) Destroy() {
	if m.destroyed.Swap(true) {
		return
	}

	m.activeWorkers.Range(func(k, v any) bool {
		if cancel, ok := v.(context.CancelFunc); ok && cancel != nil {
			cancel()
		}
		m.activeWorkers.Delete(k)
		return true
	})

	if m.dispatcher != nil {
		m.dispatcher.Destroy()
	}

	m.dispatcher = nil
}

func (m *fetcher) Pause() {
	m.activeWorkers.Range(func(k, v any) bool {
		if cancel, ok := v.(context.CancelFunc); ok && cancel != nil {
			cancel()
		}
		m.activeWorkers.Delete(k)
		return true
	})

	m.paused.Store(true)

	if m.dispatcher != nil {
		m.dispatcher.Pause()
	}
}

func (m *fetcher) Resume() {
	m.paused.Store(false)

	if m.dispatcher != nil {
		m.dispatcher.Resume()
	}
}

func (m *fetcher) execute(ctx context.Context, key string, payload any, setResult func(FetchResultInterface)) {
	if ctx != nil && ctx.Err() != nil {
		return
	}

	fAny, ok := m.fetchers.Load(key)
	if ok {
		f := fAny.(FetcherInterface)
		f.Fetch(ctx, payload, func(result FetchResultInterface) {
			result.SetSource(key)
			setResult(result)
		})
		return
	}

	r := NewFetchResult(-1, nil)
	r.SetSource(key)
	r.SetError(fmt.Errorf("no fetcher for key %s", key))
	setResult(r)
}

func RegisterFetcherManager() *fetcher {
	if fetcherManager == nil {
		fetcherManager = &fetcher{}
	}
	return fetcherManager
}

func UseFetcher() *fetcher {
	return fetcherManager
}
