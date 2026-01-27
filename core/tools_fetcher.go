package core

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

const FETCHER_MODE_DISPATCH int32 = 0
const FETCHER_MODE_CALL int32 = 1

type FetcherInterface interface {
	Fetch(ctx context.Context, payload any, callback func(FetchResultInterface))
}

var fetcherManager *fetcher = nil

type fetcher struct {
	fetchers   sync.Map
	conditions sync.Map
	registry   *cancelRegistry
	dispatcher *dispatcher
	state      *stateManager
}

func (m *fetcher) Init() {
	if m.state != nil {
		return
	}

	m.state = NewStateManager(STATE_RUNNING)

	m.registry = NewCancelRegistry()

	m.dispatcher = NewDispatcher(NETWORKING_MAXIMUM_CONNECTION*2, NETWORKING_MAXIMUM_CONNECTION, 50*time.Millisecond)

	m.dispatcher.SetKey("Fetchers")
	m.dispatcher.Start()
}

func (m *fetcher) Register(key string, f FetcherInterface, cond func() bool) {
	if m.state.Is(STATE_DESTROYED) {
		return
	}

	m.Deregister(key)

	m.fetchers.Store(key, f)
	m.conditions.Store(key, cond)
}

func (m *fetcher) Deregister(key string) {
	m.registry.Cancel(key)
	m.fetchers.Delete(key)
	m.conditions.Delete(key)
}

func (m *fetcher) Call(payloads map[string][]string, preprocess func(totalJob int), onSuccess func(map[string]FetchResultInterface), onCancel func()) {
	if m.state.Is(STATE_DESTROYED) || m.state.Is(STATE_PAUSED) {
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

	m.registry.Cancel(mapKey)

	if total == 0 {
		if onCancel != nil {
			onCancel()
		}

		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.registry.Set(mapKey, cancel)

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

						m.registry.Cancel(mapKey)

						if ctx.Err() != nil {
							if onCancel != nil {
								onCancel()
							}
						}
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
						done <- struct{}{}
					}()

					if ctx.Err() != nil {
						return
					}

					m.execute(ctx, k, payload, func(result FetchResultInterface) {
						if ctx.Err() != nil {
							return
						}

						results.Store(payload, result)
					})
				})
			}
		}

		if m.state.Is(STATE_PAUSED) || m.state.Is(STATE_DESTROYED) {
			m.registry.Cancel(mapKey)

			if ctx.Err() != nil {
				if onCancel != nil {
					onCancel()
				}
			}

			return
		}

		go func() {
			defer func() {
				m.registry.Cancel(mapKey)

				if onCancel != nil {
					onCancel()
				}
			}()

			count := 0
			for {
				select {
				case <-ShutdownCtx.Done():
					return

				case <-ctx.Done():
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
	if m.state.Is(STATE_DESTROYED) {
		return
	}

	m.state.Change(STATE_DESTROYED)

	m.registry.Destroy()

	m.fetchers.Range(func(k, _ any) bool {
		m.fetchers.Delete(k)
		return true
	})

	m.conditions.Range(func(k, _ any) bool {
		m.conditions.Delete(k)
		return true
	})

	m.dispatcher.Destroy()
}

func (m *fetcher) Pause() {
	if m.state.Is(STATE_DESTROYED) {
		return
	}

	m.state.Change(STATE_PAUSED)

	m.registry.Destroy()

	m.dispatcher.Pause()
}

func (m *fetcher) Resume() {
	if m.state.Is(STATE_DESTROYED) {
		return
	}

	m.state.Change(STATE_RUNNING)

	m.dispatcher.Resume()
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
