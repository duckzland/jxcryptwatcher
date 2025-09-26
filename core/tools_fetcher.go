package core

import (
	"context"
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

	broadcastDispatcher *dispatcher
	parallelDispatcher  *dispatcher
}

var fetcherManager *fetcher = nil

func (m *fetcher) Init() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Can cause zombie go, don't reinit!
	if m.fetchers != nil || m.activeWorkers != nil || m.broadcastDispatcher != nil || m.parallelDispatcher != nil {
		return
	}

	m.fetchers = make(map[string]FetcherInterface)
	m.delay = make(map[string]time.Duration)
	m.callbacks = make(map[string]func(FetchResultInterface))
	m.conditions = make(map[string]func() bool)
	m.activeWorkers = make(map[string]context.CancelFunc)

	m.broadcastDispatcher = NewDispatcher(1000, 4, 500*time.Millisecond)
	m.broadcastDispatcher.Start()

	m.parallelDispatcher = NewDispatcher(1000, 2, 500*time.Millisecond)
	m.parallelDispatcher.Start()
}

func (m *fetcher) Register(
	key string,
	delaySeconds int64,
	fetcher FetcherInterface,
	callback func(FetchResultInterface),
	conditions func() bool,
) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.fetchers[key] = fetcher
	m.delay[key] = time.Duration(delaySeconds) * time.Second
	m.callbacks[key] = callback
	m.conditions[key] = conditions
}

func (m *fetcher) Call(key string, payload any) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if cond, ok := m.conditions[key]; ok && cond != nil {
		if !cond() {
			return
		}
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		m.executeCall(ctx, key, payload, func(result FetchResultInterface) {
			if cb, ok := m.callbacks[key]; ok {
				cb(result)
			}
		})
	}()
}

func (m *fetcher) ParallelCall(keys []string, payloads map[string]any, preprocess func(totalJob int), callback func(map[string]FetchResultInterface)) {
	var wg sync.WaitGroup
	results := make(map[string]FetchResultInterface)
	mu := sync.Mutex{}
	total := 0

	for _, key := range keys {
		if cond, ok := m.conditions[key]; ok && cond != nil {
			if cond() {
				total++
			}
		}
	}

	if preprocess != nil {
		preprocess(total)
	}

	for _, key := range keys {
		if cond, ok := m.conditions[key]; ok && cond != nil {
			if !cond() {
				continue
			}
		}

		wg.Add(1)
		k := key

		m.parallelDispatcher.Submit(func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			m.executeCall(ctx, k, payloads[k], func(result FetchResultInterface) {
				mu.Lock()
				results[k] = result
				mu.Unlock()
			})

			wg.Done()
		})
	}

	go func() {
		wg.Wait()
		callback(results)
	}()
}

func (m *fetcher) BroadcastCall(key string, payloads []any, preprocess func(shouldProceed bool), callback func([]FetchResultInterface)) {
	if cond, ok := m.conditions[key]; ok && cond != nil {
		if !cond() {
			if preprocess != nil {
				preprocess(false)
			}
			return
		}
	}

	if preprocess != nil {
		preprocess(true)
	}

	var wg sync.WaitGroup
	results := make([]FetchResultInterface, len(payloads))
	mu := sync.Mutex{}

	for i, payload := range payloads {
		idx := i
		p := payload

		wg.Add(1)

		m.broadcastDispatcher.Submit(func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			m.executeCall(ctx, key, p, func(result FetchResultInterface) {
				mu.Lock()
				results[idx] = result
				mu.Unlock()
			})

			wg.Done()
		})
	}

	go func() {
		wg.Wait()
		callback(results)
	}()
}

func (m *fetcher) executeCall(ctx context.Context, key string, payload any, setResult func(FetchResultInterface)) {
	if fetcher := m.fetchers[key]; fetcher != nil {
		fetcher.Fetch(ctx, payload, func(result FetchResultInterface) {
			result.SetSource(key)
			setResult(result)
		})
	} else {
		Logf("Fetcher for key %s is nil", key)
	}
}

type dynamicPayloadFetcher struct {
	handler func(ctx context.Context, payload any) (FetchResultInterface, error)
}

func (df *dynamicPayloadFetcher) Fetch(ctx context.Context, payload any, callback func(FetchResultInterface)) {
	result, err := df.handler(ctx, payload)
	if err != nil {
		result.SetError(err)
	}
	callback(result)
}

type genericFetcher struct {
	handler func(ctx context.Context) (FetchResultInterface, error)
}

func (gf *genericFetcher) Fetch(ctx context.Context, _ any, callback func(FetchResultInterface)) {
	result, err := gf.handler(ctx)
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

func NewDynamicPayloadFetcher(handler func(ctx context.Context, payload any) (FetchResultInterface, error)) *dynamicPayloadFetcher {
	return &dynamicPayloadFetcher{handler: handler}
}

func NewGenericFetcher(handler func(ctx context.Context) (FetchResultInterface, error)) *genericFetcher {
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
