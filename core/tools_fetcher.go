package core

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

var verboseFetcherDebugMessage bool = false

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

type FetchRequest struct {
	Payload any
}

type fetchResult struct {
	code   int64
	data   any
	err    error
	source string
}

func (r *fetchResult) Code() int64        { return r.code }
func (r *fetchResult) Data() any          { return r.data }
func (r *fetchResult) Err() error         { return r.err }
func (r *fetchResult) Source() string     { return r.source }
func (r *fetchResult) SetSource(s string) { r.source = s }
func (r *fetchResult) SetError(e error)   { r.err = e }

type fetcher struct {
	fetchers         map[string]FetcherInterface
	delay            map[string]time.Duration
	lastActivity     map[string]*time.Time
	callbacks        map[string]func(FetchResultInterface)
	conditions       map[string]func() bool
	recentGroupedLog string
	lastGroupedLog   time.Time
	activeWorkers    map[string]context.CancelFunc
	mu               sync.Mutex
}

var fetcherManager *fetcher = nil

func (m *fetcher) Init() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.fetchers = make(map[string]FetcherInterface)
	m.delay = make(map[string]time.Duration)
	m.lastActivity = make(map[string]*time.Time)
	m.callbacks = make(map[string]func(FetchResultInterface))
	m.conditions = make(map[string]func() bool)
	m.recentGroupedLog = ""
	m.lastGroupedLog = time.Time{}
	m.activeWorkers = make(map[string]context.CancelFunc)
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
	now := time.Now()
	m.lastActivity[key] = &now
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

		m.fetchers[key].Fetch(ctx, payload, func(result FetchResultInterface) {
			result.SetSource(key)
			lines := []string{
				fmt.Sprintf("Immediate call → %s → Code: %d", result.Source(), result.Code()),
			}
			if result.Err() != nil {
				lines = append(lines, fmt.Sprintf("Error: %v", result.Err()))
			} else {
				lines = append(lines, fmt.Sprintf("Data: %+v", result.Data()))
			}
			m.logGrouped("CALL", lines, 30*time.Second)

			if cb, ok := m.callbacks[key]; ok {
				cb(result)
			}
		})
	}()
}

func (m *fetcher) GroupCall(keys []string, payloads map[string]any, preprocess func(totalJob int), callback func(map[string]FetchResultInterface)) {
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
		go func(k string) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			payload := payloads[k]
			if fetcher := m.fetchers[k]; fetcher != nil {
				fetcher.Fetch(ctx, payload, func(result FetchResultInterface) {
					result.SetSource(k)
					mu.Lock()
					results[k] = result
					mu.Unlock()
				})
			} else {
				Logf("Fetcher for key %s is nil", k)
			}
		}(key)
	}

	go func() {
		wg.Wait()
		lines := []string{}
		for k, r := range results {
			lines = append(lines, fmt.Sprintf("%s → Code: %d", k, r.Code()))
		}
		m.logGrouped("GROUPCALL", lines, 60*time.Second)
		callback(results)
	}()
}

func (m *fetcher) GroupPayloadCall(key string, payloads []any, preprocess func(shouldProceed bool), callback func([]FetchResultInterface)) {
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
		wg.Add(1)
		go func(idx int, p any) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			fetcher, ok := m.fetchers[key]
			if !ok || fetcher == nil {
				m.logGrouped("GROUPCALL", []string{
					fmt.Sprintf("Fetcher not found for key: %s", key),
				}, 60*time.Second)
				return
			}

			fetcher.Fetch(ctx, p, func(result FetchResultInterface) {
				result.SetSource(key)
				mu.Lock()
				results[idx] = result
				mu.Unlock()
			})
		}(i, payload)
	}

	go func() {
		wg.Wait()
		lines := []string{}
		for _, r := range results {
			lines = append(lines, fmt.Sprintf("%s → Code: %d", r.Source(), r.Code()))
		}
		m.logGrouped("GROUPCALL", lines, 60*time.Second)
		callback(results)
	}()
}

func (m *fetcher) logGrouped(tag string, lines []string, interval time.Duration) {
	if !verboseFetcherDebugMessage {
		return
	}

	msg := fmt.Sprintf("[%s] %s", tag, strings.Join(lines, " | "))
	now := time.Now()

	if msg != m.recentGroupedLog || now.Sub(m.lastGroupedLog) >= interval {
		Logln(msg)
		m.recentGroupedLog = msg
		m.lastGroupedLog = now
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
