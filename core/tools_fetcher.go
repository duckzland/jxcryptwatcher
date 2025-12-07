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
}

var fetcherManager *fetcher = nil

func (m *fetcher) Init() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Can cause zombie go, don't reinit!
	if m.fetchers != nil || m.activeWorkers != nil || m.dispatcher != nil {
		return
	}

	m.fetchers = make(map[string]FetcherInterface)
	m.delay = make(map[string]time.Duration)
	m.callbacks = make(map[string]func(FetchResultInterface))
	m.conditions = make(map[string]func() bool)
	m.activeWorkers = make(map[string]context.CancelFunc)

	m.dispatcher = NewDispatcher(50, 4, 500*time.Millisecond)
	m.dispatcher.Start()
}

func (m *fetcher) Register(key string, delaySeconds int64, fetcher FetcherInterface, callback func(FetchResultInterface), conditions func() bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.fetchers[key] = fetcher
	m.delay[key] = time.Duration(delaySeconds) * time.Second
	m.callbacks[key] = callback
	m.conditions[key] = conditions
}

func (m *fetcher) Call(key string, payload any) {
	m.mu.Lock()
	if cond, ok := m.conditions[key]; ok && cond != nil && !cond() {
		m.mu.Unlock()
		return
	}

	callID := fmt.Sprintf("%s:%v", key, payload)
	if oldCancel, exists := m.activeWorkers[callID]; exists {
		oldCancel()
		delete(m.activeWorkers, callID)
	}
	ctx, cancel := context.WithCancel(context.Background())
	m.activeWorkers[callID] = cancel
	m.mu.Unlock()

	go func() {
		defer func() {
			m.mu.Lock()
			delete(m.activeWorkers, callID)
			m.mu.Unlock()
			cancel()
		}()

		select {
		case <-ctx.Done():
			return
		default:
			m.executeCall(key, payload, func(result FetchResultInterface) {
				if ctx.Err() != nil {
					return
				}

				m.mu.Lock()
				cb := m.callbacks[key]
				m.mu.Unlock()
				if cb != nil {
					cb(result)
				}
			})
		}
	}()
}

func (m *fetcher) Dispatch(payloads map[string][]string, preprocess func(totalJob int), callback func(map[string]FetchResultInterface)) {

	results := make(map[string]FetchResultInterface)
	mu := sync.Mutex{}
	total := 0

	mapKey := ""
	for key, items := range payloads {
		if cond, ok := m.conditions[key]; ok && cond != nil {
			if cond() {
				total += len(items)
			}
		}
		mapKey += key + strings.Join(items, "|")
	}

	hash := sha256.Sum256([]byte(mapKey))
	mapKey = hex.EncodeToString(hash[:])

	if preprocess != nil {
		preprocess(total)
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.mu.Lock()
	if oldCancel, exists := m.activeWorkers[mapKey]; exists {
		oldCancel()
		delete(m.activeWorkers, mapKey)
	}
	m.activeWorkers[mapKey] = cancel
	m.mu.Unlock()

	done := make(chan struct{}, total)

	for key, items := range payloads {
		if cond, ok := m.conditions[key]; ok && cond != nil {
			if !cond() {
				continue
			}
		}

		for _, item := range items {
			k := key
			payload := item

			m.dispatcher.Submit(func() {
				m.executeCall(k, payload, func(result FetchResultInterface) {
					mu.Lock()
					results[payload] = result
					mu.Unlock()
				})

				done <- struct{}{}
			})
		}
	}
	go func(ctx context.Context, mapKey string) {
		defer func() {
			cancel()
			m.mu.Lock()
			delete(m.activeWorkers, mapKey)
			m.mu.Unlock()
		}()

		count := 0
		for {
			select {
			case <-ctx.Done():
				return

			case <-done:
				count++
				if count == total {
					callback(results)
					return
				}
			}
		}
	}(ctx, mapKey)
}

func (m *fetcher) Destroy() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, cancelFunc := range m.activeWorkers {
		if cancelFunc != nil {
			cancelFunc()
		}
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
	if fetcher := m.fetchers[key]; fetcher != nil {
		fetcher.Fetch(payload, func(result FetchResultInterface) {
			result.SetSource(key)
			setResult(result)
		})
	} else {
		Logf("Fetcher for key %s is nil", key)
	}
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
