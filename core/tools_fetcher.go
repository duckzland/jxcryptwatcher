package core

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

type FetchResultInterface interface {
	Code() int64
	Data() any
	Err() error
	Source() string
	SetSource(string)
	SetError(error)
}

type FetcherInterface interface {
	Fetch(ctx context.Context, payload any, callback func(FetchResultInterface))
}

type fetchResult struct {
	code   int64
	data   any
	err    error
	source string
	ctx    context.Context
}

func (r *fetchResult) Code() int64 {
	return r.code
}

func (r *fetchResult) Data() any {
	return r.data
}

func (r *fetchResult) Err() error {
	return r.err
}

func (r *fetchResult) Source() string {
	return r.source
}

func (r *fetchResult) SetSource(s string) {
	r.source = s
}

func (r *fetchResult) SetError(e error) {
	r.err = e
}

type fetcher struct {
	fetchers      map[string]FetcherInterface
	delay         map[string]time.Duration
	conditions    map[string]func() bool
	activeWorkers map[string]context.CancelFunc
	mu            sync.Mutex
	dispatcher    *dispatcher
	destroyed     bool
	paused        bool
}

var fetcherManager *fetcher = nil

func (m *fetcher) Init() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.fetchers != nil || m.activeWorkers != nil || m.dispatcher != nil {
		return
	}

	m.fetchers = make(map[string]FetcherInterface)
	m.delay = make(map[string]time.Duration)
	m.conditions = make(map[string]func() bool)
	m.activeWorkers = make(map[string]context.CancelFunc)
	m.destroyed = false
	m.paused = false
	m.dispatcher = NewDispatcher(50, 4, 500*time.Millisecond)
	m.dispatcher.Start()
}

func (m *fetcher) Register(key string, delaySeconds int64, fetcher FetcherInterface, conditions func() bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.destroyed || m.paused {
		return
	}

	if _, registered := m.fetchers[key]; registered {
		m.internalDestroy(key)
	}

	m.fetchers[key] = fetcher
	m.delay[key] = time.Duration(delaySeconds) * time.Second
	m.conditions[key] = conditions
}

func (m *fetcher) Deregister(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.internalDestroy(key)
}

func (m *fetcher) Call(payloads map[string][]string, preprocess func(totalJob int), onSuccess func(map[string]FetchResultInterface), onCancel func()) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.destroyed || m.paused {
		return
	}

	results := make(map[string]FetchResultInterface)
	mu := sync.Mutex{}
	total := 0
	mapKey := ""

	for key, items := range payloads {
		if cond, ok := m.conditions[key]; ok && cond != nil {
			if cond() {
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

	if oldCancel, exists := m.activeWorkers[mapKey]; exists {
		oldCancel()
		delete(m.activeWorkers, mapKey)
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.activeWorkers[mapKey] = cancel

	if total == 0 {
		if onSuccess != nil {
			onSuccess(results)
		}

		if m.activeWorkers != nil {
			delete(m.activeWorkers, mapKey)
		}

		return
	}

	cancelled := false
	done := make(chan struct{}, total)

	for key, items := range payloads {
		if cond, ok := m.conditions[key]; ok && cond != nil && !cond() {
			continue
		}

		for _, item := range items {

			k := key
			payload := item

			m.dispatcher.Submit(func() {
				defer func() {
					if !cancelled {
						done <- struct{}{}
					}
				}()

				if ctx.Err() != nil || cancelled {
					return
				}

				m.executeCall(ctx, k, payload, func(result FetchResultInterface) {
					mu.Lock()
					defer mu.Unlock()

					if ctx.Err() != nil || cancelled {
						return
					}

					results[payload] = result
				})
			})
		}
	}

	go func(ctx context.Context, mapKey string) {
		defer func() {
			m.mu.Lock()
			defer m.mu.Unlock()

			cancel()

			if m.activeWorkers != nil {
				delete(m.activeWorkers, mapKey)
			}

			if cancelled {
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
				cancelled = true
				return

			case <-ctx.Done():
				cancelled = true
				return

			case <-done:
				count++
				if count == total {
					if onSuccess != nil {
						onSuccess(results)
					}

					close(done)
					return
				}
			}
		}
	}(ctx, mapKey)
}

func (m *fetcher) Destroy() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.destroyed {
		return
	}

	m.destroyed = true

	for key := range m.activeWorkers {
		m.internalDestroy(key)
	}

	if m.dispatcher != nil {
		m.dispatcher.Destroy()
	}

	m.dispatcher = nil
	m.activeWorkers = nil
	m.fetchers = nil
	m.delay = nil
	m.conditions = nil
}

func (m *fetcher) Pause() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for key, cancel := range m.activeWorkers {
		if cancel != nil {
			cancel()
		}

		delete(m.activeWorkers, key)
	}

	m.paused = true

	m.dispatcher.Pause()
}

func (m *fetcher) Resume() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.paused = false

	m.dispatcher.Resume()
}

func (m *fetcher) internalDestroy(key string) {
	if m.activeWorkers != nil {
		if cancel, exists := m.activeWorkers[key]; exists {
			if cancel != nil {
				cancel()
			}
			delete(m.activeWorkers, key)
		}
	}

	if m.fetchers != nil {
		if _, exists := m.fetchers[key]; exists {
			delete(m.fetchers, key)
		}
	}

	if m.delay != nil {
		if _, exists := m.delay[key]; exists {
			delete(m.delay, key)
		}
	}

	if m.conditions != nil {
		if _, exists := m.conditions[key]; exists {
			delete(m.conditions, key)
		}
	}
}

func (m *fetcher) executeCall(ctx context.Context, key string, payload any, setResult func(FetchResultInterface)) {
	m.mu.Lock()
	defer m.mu.Unlock()

	f := m.fetchers[key]

	if f != nil {
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

type fetcherUnit struct {
	handler func(ctx context.Context, payload any) (FetchResultInterface, error)
}

func (df *fetcherUnit) Fetch(ctx context.Context, payload any, callback func(FetchResultInterface)) {
	result, err := df.handler(ctx, payload)

	if err != nil {
		result.SetError(err)
	}

	callback(result)
}

func NewFetchResult(code int64, data any) FetchResultInterface {
	return &fetchResult{
		code: code,
		data: data,
	}
}

func NewFetcherUnit(handler func(ctx context.Context, payload any) (FetchResultInterface, error)) *fetcherUnit {
	return &fetcherUnit{handler: handler}
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
