package apps

import (
	"context"
	"fmt"
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
	fetchers     map[string]AppFetcherInterface
	queues       map[string]chan FetchRequest
	delay        map[string]time.Duration
	lastActivity map[string]*time.Time
	callbacks    map[string]func(AppFetchResult)

	// Log grouping
	recentWatchdogLog string
	lastWatchdogLog   time.Time

	mu sync.Mutex
}

// --- Global Instance ---
var AppFetcherManager = &AppFetcher{
	fetchers:          make(map[string]AppFetcherInterface),
	queues:            make(map[string]chan FetchRequest),
	delay:             make(map[string]time.Duration),
	lastActivity:      make(map[string]*time.Time),
	callbacks:         make(map[string]func(AppFetchResult)),
	recentWatchdogLog: "",
	lastWatchdogLog:   time.Time{},
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
	for {
		req := <-m.queues[key]

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		m.fetchers[key].Fetch(ctx, req.Payload, func(result AppFetchResult) {
			result.Source = key
			JC.Logf("[Worker:%s] Code: %d", result.Source, result.Code)
			if result.Err != nil {
				JC.Logln("Error:", result.Err)
			} else {
				JC.Logf("Data: %+v", result.Data)
			}
			if cb, ok := m.callbacks[key]; ok {
				cb(result)
			}
		})
		cancel()

		m.mu.Lock()
		now := time.Now()
		m.lastActivity[key] = &now
		m.mu.Unlock()

		if !req.Immediate {
			time.Sleep(m.delay[key])
		}
	}
}

// --- Watchdog ---
func (m *AppFetcher) watchdog(maxIdle time.Duration) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		var stale []string

		m.mu.Lock()
		for key, last := range m.lastActivity {
			if time.Since(*last) > maxIdle {
				stale = append(stale, key)
				go m.startWorker(key)
				now := time.Now()
				m.lastActivity[key] = &now
			}
		}
		m.mu.Unlock()

		if len(stale) > 0 {
			m.logWatchdogGrouped(stale, 5*time.Second)
		}
	}
}

// --- Grouped Watchdog Log ---
func (m *AppFetcher) logWatchdogGrouped(keys []string, interval time.Duration) {
	msg := fmt.Sprintf("[WATCHDOG] Restarting stale workers: %v", keys)
	now := time.Now()

	if msg != m.recentWatchdogLog || now.Sub(m.lastWatchdogLog) >= interval {
		JC.Logln(msg)
		m.recentWatchdogLog = msg
		m.lastWatchdogLog = now
	}
}

// --- Manual Call ---
func (m *AppFetcher) Call(key string, payload any, immediate bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if immediate {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			m.fetchers[key].Fetch(ctx, payload, func(result AppFetchResult) {
				result.Source = key
				JC.Logf("[Immediate] %s → Code: %d", result.Source, result.Code)
				if result.Err != nil {
					JC.Logln("Error:", result.Err)
				} else {
					JC.Logf("Data: %+v", result.Data)
				}
				if cb, ok := m.callbacks[key]; ok {
					cb(result)
				}
			})
			cancel()
		}()
	} else {
		m.queues[key] <- FetchRequest{Payload: payload, Immediate: immediate}
	}
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
				JC.Logf("[GroupPayloadCall] Fetcher not found for key: %s", key)
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
