package core

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

var verboseFetcherDebugMessage bool = false

type FetchResult struct {
	Code   int64
	Data   any
	Err    error
	Source string
}

type FetcherInterface interface {
	Fetch(ctx context.Context, payload any, callback func(FetchResult))
}

type FetchRequest struct {
	Payload any
}

type Fetcher struct {
	fetchers         map[string]FetcherInterface
	delay            map[string]time.Duration
	lastActivity     map[string]*time.Time
	callbacks        map[string]func(FetchResult)
	conditions       map[string]func() bool
	recentGroupedLog string
	lastGroupedLog   time.Time
	activeWorkers    map[string]context.CancelFunc
	mu               sync.Mutex
}

var FetcherManager = &Fetcher{
	fetchers:         make(map[string]FetcherInterface),
	delay:            make(map[string]time.Duration),
	lastActivity:     make(map[string]*time.Time),
	callbacks:        make(map[string]func(FetchResult)),
	conditions:       make(map[string]func() bool),
	recentGroupedLog: "",
	lastGroupedLog:   time.Time{},
	activeWorkers:    make(map[string]context.CancelFunc),
}

func (m *Fetcher) logGrouped(tag string, lines []string, interval time.Duration) {
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

func (m *Fetcher) Register(key string, fetcher FetcherInterface, delaySeconds int64, callback func(FetchResult), conditions func() bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.fetchers[key] = fetcher
	now := time.Now()
	m.lastActivity[key] = &now
	m.delay[key] = time.Duration(delaySeconds) * time.Second
	m.callbacks[key] = callback
	m.conditions[key] = conditions

}

func (m *Fetcher) Call(key string, payload any) {
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

		m.fetchers[key].Fetch(ctx, payload, func(result FetchResult) {
			result.Source = key
			lines := []string{
				fmt.Sprintf("Immediate call → %s → Code: %d", result.Source, result.Code),
			}
			if result.Err != nil {
				lines = append(lines, fmt.Sprintf("Error: %v", result.Err))
			} else {
				lines = append(lines, fmt.Sprintf("Data: %+v", result.Data))
			}
			m.logGrouped("CALL", lines, 30*time.Second)

			if cb, ok := m.callbacks[key]; ok {
				cb(result)
			}
		})
	}()

}

func (m *Fetcher) GroupCall(keys []string, payloads map[string]any, preprocess func(totalJob int), callback func(map[string]FetchResult)) {
	var wg sync.WaitGroup
	results := make(map[string]FetchResult)
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
				return
			}
		}

		wg.Add(1)
		go func(k string) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			payload := payloads[k]
			m.fetchers[k].Fetch(ctx, payload, func(result FetchResult) {
				result.Source = k
				mu.Lock()
				results[k] = result
				mu.Unlock()
			})
		}(key)
	}

	go func() {
		wg.Wait()
		lines := []string{}
		for k, r := range results {
			lines = append(lines, fmt.Sprintf("%s → Code: %d", k, r.Code))
		}
		m.logGrouped("GROUPCALL", lines, 60*time.Second)
		callback(results)
	}()
}

func (m *Fetcher) GroupPayloadCall(key string, payloads []any, preprocess func(shouldProceed bool), callback func([]FetchResult)) {
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
	results := make([]FetchResult, len(payloads))
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

			fetcher.Fetch(ctx, p, func(result FetchResult) {
				result.Source = key
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
			lines = append(lines, fmt.Sprintf("%s → Code: %d", r.Source, r.Code))
		}
		m.logGrouped("GROUPCALL", lines, 60*time.Second)
		callback(results)
	}()
}

type DynamicPayloadFetcher struct {
	Handler func(ctx context.Context, payload any) (FetchResult, error)
}

func (df *DynamicPayloadFetcher) Fetch(ctx context.Context, payload any, callback func(FetchResult)) {
	result, err := df.Handler(ctx, payload)
	if err != nil {
		result.Err = err
	}
	callback(result)
}

type GenericFetcher struct {
	Handler func(ctx context.Context) (FetchResult, error)
}

func (gf *GenericFetcher) Fetch(ctx context.Context, _ any, callback func(FetchResult)) {
	result, err := gf.Handler(ctx)
	if err != nil {
		result.Err = err
	}
	callback(result)
}
