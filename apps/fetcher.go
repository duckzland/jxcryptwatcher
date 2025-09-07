package apps

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	JC "jxwatcher/core"
)

// --- Result Struct ---
type AppFetchResult struct {
	Code   int64
	Data   any
	Err    error
	Source string
}

// --- Fetcher Interface ---
type AppFetcherInterface interface {
	Fetch(ctx context.Context, payload any, callback func(AppFetchResult))
}

// --- Request Struct ---
type FetchRequest struct {
	Payload   any
	Immediate bool
}

// --- Manager ---
type AppFetcher struct {
	fetchers         map[string]AppFetcherInterface
	queues           map[string]chan FetchRequest
	delay            map[string]time.Duration
	lastActivity     map[string]*time.Time
	callbacks        map[string]func(AppFetchResult)
	recentGroupedLog string
	lastGroupedLog   time.Time
	activeWorkers    map[string]context.CancelFunc
	mu               sync.Mutex
}

// --- Global Instance ---
var AppFetcherManager = &AppFetcher{
	fetchers:         make(map[string]AppFetcherInterface),
	queues:           make(map[string]chan FetchRequest),
	delay:            make(map[string]time.Duration),
	lastActivity:     make(map[string]*time.Time),
	callbacks:        make(map[string]func(AppFetchResult)),
	recentGroupedLog: "",
	lastGroupedLog:   time.Time{},
	activeWorkers:    make(map[string]context.CancelFunc),
}

// --- Centralized Grouped Logger ---
func (m *AppFetcher) logGrouped(tag string, lines []string, interval time.Duration) {
	msg := fmt.Sprintf("[%s] %s", tag, strings.Join(lines, " | "))
	now := time.Now()

	if msg != m.recentGroupedLog || now.Sub(m.lastGroupedLog) >= interval {
		JC.Logln(msg)
		m.recentGroupedLog = msg
		m.lastGroupedLog = now
	}
}

// --- Registration ---
func (m *AppFetcher) Register(key string, fetcher AppFetcherInterface, delaySeconds int64, callback func(AppFetchResult)) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.fetchers[key] = fetcher
	m.queues[key] = make(chan FetchRequest, 100)
	now := time.Now()
	m.lastActivity[key] = &now
	m.delay[key] = time.Duration(delaySeconds) * time.Second
	m.callbacks[key] = callback

	go m.startWorker(key)
	go m.watchdog(30 * time.Second)
}

// --- Worker Loop ---
func (m *AppFetcher) startWorker(key string) {
	ctx, cancel := context.WithCancel(context.Background())

	m.mu.Lock()
	m.activeWorkers[key] = cancel
	m.mu.Unlock()

	go func() {
		for {
			select {
			case req := <-m.queues[key]:
				select {
				case <-ctx.Done():
					m.logGrouped("WORKER", []string{
						fmt.Sprintf("Terminated before processing: %s", key),
					}, 60*time.Second)
					return
				default:
					// proceed with fetch
				}

				m.fetchers[key].Fetch(ctx, req.Payload, func(result AppFetchResult) {
					if cb := m.callbacks[key]; cb != nil {
						cb(result)
					}
				})

				m.mu.Lock()
				now := time.Now()
				m.lastActivity[key] = &now
				m.mu.Unlock()

				if !req.Immediate {
					time.Sleep(m.delay[key])
				}
			case <-ctx.Done():
				// m.logGrouped("WORKER", []string{
				// 	fmt.Sprintf("Terminated by watchdog: %s", key),
				// }, 60*time.Second)
				return
			}
		}
	}()
}

// --- Watchdog ---
func (m *AppFetcher) watchdog(maxIdle time.Duration) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {

		m.mu.Lock()
		for key, last := range m.lastActivity {
			if time.Since(*last) > maxIdle {
				if cancel, ok := m.activeWorkers[key]; ok {
					cancel()
					delete(m.activeWorkers, key)
					m.logGrouped("WATCHDOG", []string{
						fmt.Sprintf("Killed ghost worker: %s", key),
					}, 60*time.Second)
				}
			}
		}
		m.mu.Unlock()
	}
}

// --- Manual Call ---
func (m *AppFetcher) Call(key string, payload any, immediate bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if immediate {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			m.fetchers[key].Fetch(ctx, payload, func(result AppFetchResult) {
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
	} else {
		m.queues[key] <- FetchRequest{Payload: payload, Immediate: immediate}
	}
}

func (m *AppFetcher) CallWithCallback(key string, payload any, immediate bool, callback func(AppFetchResult)) {
	m.mu.Lock()
	fetcher, exists := m.fetchers[key]
	m.mu.Unlock()

	if !exists || fetcher == nil {
		m.logGrouped("CALLBACK", []string{
			fmt.Sprintf("Fetcher not found for key: %s", key),
		}, 30*time.Second)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	go fetcher.Fetch(ctx, payload, func(result AppFetchResult) {
		result.Source = key
		m.logGrouped("CALLBACK", []string{
			fmt.Sprintf("Callback → %s → Code: %d", result.Source, result.Code),
		}, 30*time.Second)
		callback(result)
	})
}

// --- Scheduled Payload Logic ---
func (m *AppFetcher) SchedulePayload(key string, interval time.Duration, logic func() any) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			payload := logic()
			m.Call(key, payload, true)
		}
	}()
}

// --- Grouped Fetch ---
func (m *AppFetcher) GroupCall(keys []string, payloads map[string]any, callback func(map[string]AppFetchResult)) {
	var wg sync.WaitGroup
	results := make(map[string]AppFetchResult)
	mu := sync.Mutex{}

	for _, key := range keys {
		wg.Add(1)
		go func(k string) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			payload := payloads[k]
			m.fetchers[k].Fetch(ctx, payload, func(result AppFetchResult) {
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

func (m *AppFetcher) GroupPayloadCall(key string, payloads []any, callback func([]AppFetchResult)) {
	var wg sync.WaitGroup
	results := make([]AppFetchResult, len(payloads))
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

			fetcher.Fetch(ctx, p, func(result AppFetchResult) {
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

// --- Fetcher Types ---
type DynamicPayloadFetcher struct {
	Handler func(ctx context.Context, payload any) (AppFetchResult, error)
}

func (df *DynamicPayloadFetcher) Fetch(ctx context.Context, payload any, callback func(AppFetchResult)) {
	result, err := df.Handler(ctx, payload)
	if err != nil {
		result.Err = err
	}
	callback(result)
}

type GenericFetcher struct {
	Handler func(ctx context.Context) (AppFetchResult, error)
}

func (gf *GenericFetcher) Fetch(ctx context.Context, _ any, callback func(AppFetchResult)) {
	result, err := gf.Handler(ctx)
	if err != nil {
		result.Err = err
	}
	callback(result)
}
