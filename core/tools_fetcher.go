package core

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
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
	Fetch(payload any, callback func(FetchResultInterface))
}

type fetchResult struct {
	code   int64
	data   any
	err    error
	source string
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
	callbacks     map[string]func(FetchResultInterface)
	conditions    map[string]func() bool
	activeWorkers map[string]context.CancelFunc
	mu            sync.Mutex
	dispatcher    *dispatcher
	destroyed     bool
}

var fetcherManager *fetcher = nil

func (m *fetcher) Init() {
	defer m.mu.Unlock()
	m.mu.Lock()

	if m.fetchers != nil || m.activeWorkers != nil || m.dispatcher != nil {
		return
	}

	m.fetchers = make(map[string]FetcherInterface)
	m.delay = make(map[string]time.Duration)
	m.callbacks = make(map[string]func(FetchResultInterface))
	m.conditions = make(map[string]func() bool)
	m.activeWorkers = make(map[string]context.CancelFunc)
	m.destroyed = false
	m.dispatcher = NewDispatcher(50, 4, 500*time.Millisecond)
	m.dispatcher.Start()
}

func (m *fetcher) Register(key string, delaySeconds int64, fetcher FetcherInterface, callback func(FetchResultInterface), conditions func() bool) {
	defer m.mu.Unlock()
	m.mu.Lock()

	if m.destroyed {
		return
	}

	m.fetchers[key] = fetcher
	m.delay[key] = time.Duration(delaySeconds) * time.Second
	m.callbacks[key] = callback
	m.conditions[key] = conditions
}

func (m *fetcher) Call(key string, payload any) {
	defer m.mu.Unlock()
	m.mu.Lock()

	if m.destroyed {
		return
	}

	if cond, ok := m.conditions[key]; ok && cond != nil && !cond() {
		return
	}

	callID := fmt.Sprintf("%s:%v", key, payload)
	if oldCancel, exists := m.activeWorkers[callID]; exists {
		oldCancel()
		delete(m.activeWorkers, callID)
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.activeWorkers[callID] = cancel

	go func(callID string, ctx context.Context, cancel context.CancelFunc) {
		defer func() {
			defer m.mu.Unlock()
			m.mu.Lock()

			cancel()

			if m.activeWorkers != nil {
				delete(m.activeWorkers, callID)
			}
		}()

		select {
		case <-ctx.Done():
			return

		default:
			m.executeCall(key, payload, func(result FetchResultInterface) {
				defer m.mu.Unlock()
				m.mu.Lock()

				if ctx.Err() != nil {
					return
				}

				if m.callbacks == nil || m.destroyed {
					return
				}

				cb := m.callbacks[key]

				if cb != nil {
					cb(result)
				}
			})
		}

	}(callID, ctx, cancel)
}

func (m *fetcher) Dispatch(payloads map[string][]string, preprocess func(totalJob int), callback func(map[string]FetchResultInterface)) {
	defer m.mu.Unlock()
	m.mu.Lock()

	if m.destroyed {
		return
	}

	results := make(map[string]FetchResultInterface)
	mu := sync.Mutex{}
	total := 0
	mapKeySeed := ""

	for key, items := range payloads {
		if cond, ok := m.conditions[key]; ok && cond != nil {
			if cond() {
				total += len(items)
			}
		} else {
			total += len(items)
		}
		mapKeySeed += key + strings.Join(items, "|")
	}

	hash := sha256.Sum256([]byte(mapKeySeed))
	mapKey := hex.EncodeToString(hash[:])

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
		if callback != nil {
			callback(results)
		}

		if m.activeWorkers != nil {
			delete(m.activeWorkers, mapKey)
		}

		return
	}

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
					done <- struct{}{}
				}()

				if ctx.Err() != nil {
					return
				}

				m.executeCall(k, payload, func(result FetchResultInterface) {
					mu.Lock()
					defer mu.Unlock()

					if ctx.Err() != nil {
						return
					}

					results[payload] = result
				})
			})
		}
	}
	go func(ctx context.Context, mapKey string) {
		defer func() {
			defer m.mu.Unlock()
			m.mu.Lock()

			cancel()

			if m.activeWorkers != nil {
				delete(m.activeWorkers, mapKey)
			}
		}()

		count := 0

		for {
			select {
			case <-ctx.Done():
				return

			case <-done:
				count++
				if count == total {
					if callback != nil {
						callback(results)
					}

					close(done)
					return
				}
			}
		}
	}(ctx, mapKey)
}

func (m *fetcher) Destroy() {
	defer m.mu.Unlock()
	m.mu.Lock()

	if m.destroyed {
		return
	}

	m.destroyed = true

	for key, cancelFunc := range m.activeWorkers {
		if cancelFunc != nil {
			cancelFunc()
		}
		delete(m.activeWorkers, key)
	}

	m.activeWorkers = nil

	if m.dispatcher != nil {
		m.dispatcher.Destroy()
		m.dispatcher = nil
	}

	m.fetchers = nil
	m.delay = nil
	m.callbacks = nil
	m.conditions = nil
}

func (m *fetcher) executeCall(key string, payload any, setResult func(FetchResultInterface)) {
	defer m.mu.Unlock()
	m.mu.Lock()

	f := m.fetchers[key]

	if f != nil {
		f.Fetch(payload, func(result FetchResultInterface) {
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

type dynamicPayloadFetcher struct {
	handler func(payload any) (FetchResultInterface, error)
}

func (df *dynamicPayloadFetcher) Fetch(payload any, callback func(FetchResultInterface)) {
	result, err := df.handler(payload)

	if err != nil {
		result.SetError(err)
	}

	callback(result)
}

type genericFetcher struct {
	handler func() (FetchResultInterface, error)
}

func (gf *genericFetcher) Fetch(_ any, callback func(FetchResultInterface)) {
	result, err := gf.handler()

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

func NewDynamicPayloadFetcher(handler func(payload any) (FetchResultInterface, error)) *dynamicPayloadFetcher {
	return &dynamicPayloadFetcher{handler: handler}
}

func NewGenericFetcher(handler func() (FetchResultInterface, error)) *genericFetcher {
	return &genericFetcher{handler: handler}
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
