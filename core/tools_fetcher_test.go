package core

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestFetcherInit(t *testing.T) {
	f := &fetcher{}
	f.Init()

	if f.fetchers == nil || f.delay == nil || f.callbacks == nil || f.conditions == nil || f.activeWorkers == nil {
		t.Error("Expected all internal maps to be initialized")
	}
	if f.broadcastDispatcher == nil || f.parallelDispatcher == nil {
		t.Error("Expected dispatchers to be initialized")
	}

	// Check mutex lock/unlock
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Mutex caused panic: %v", r)
			}
		}()
		f.mu.Lock()
		_ = f
		f.mu.Unlock()
	}()
}

func TestRegisterAndUseFetcherManager(t *testing.T) {
	f1 := RegisterFetcherManager()
	f2 := UseFetcher()

	if f1 != f2 {
		t.Error("Expected singleton fetcher instance")
	}
}

func TestFetcherRegisterAndCall(t *testing.T) {
	f := &fetcher{}
	f.Init()

	called := make(chan bool, 1)

	handler := func(ctx context.Context, payload any) (FetchResultInterface, error) {
		return NewFetchResult(200, "ok"), nil
	}
	fetcher := NewDynamicPayloadFetcher(handler)

	f.Register("test", 1, fetcher, func(result FetchResultInterface) {
		if result.Code() == 200 && result.Data() == "ok" {
			called <- true
		}
	}, func() bool { return true })

	f.Call("test", "payload")

	select {
	case <-called:
	case <-time.After(1 * time.Second):
		t.Error("Expected callback to be invoked")
	}
}

func TestFetcherParallelCall(t *testing.T) {
	f := &fetcher{}
	f.Init()

	keys := []string{"a", "b"}
	payloads := map[string]any{"a": "1", "b": "2"}
	results := make(chan map[string]FetchResultInterface, 1)

	for _, k := range keys {
		f.Register(k, 1, NewDynamicPayloadFetcher(func(ctx context.Context, payload any) (FetchResultInterface, error) {
			return NewFetchResult(100, payload), nil
		}), nil, func() bool { return true })
	}

	f.ParallelCall(keys, payloads, nil, func(res map[string]FetchResultInterface) {
		results <- res
	})

	select {
	case r := <-results:
		if len(r) != 2 {
			t.Errorf("Expected 2 results, got %d", len(r))
		}
	case <-time.After(2 * time.Second):
		t.Error("ParallelCall timed out")
	}
}

func TestFetcherBroadcastCall(t *testing.T) {
	f := &fetcher{}
	f.Init()

	key := "broadcast"
	payloads := []any{"x", "y", "z"}
	results := make(chan []FetchResultInterface, 1)

	f.Register(key, 1, NewDynamicPayloadFetcher(func(ctx context.Context, payload any) (FetchResultInterface, error) {
		return NewFetchResult(123, payload), nil
	}), nil, func() bool { return true })

	f.BroadcastCall(key, payloads, nil, func(res []FetchResultInterface) {
		results <- res
	})

	select {
	case r := <-results:
		if len(r) != 3 {
			t.Errorf("Expected 3 results, got %d", len(r))
		}
	case <-time.After(2 * time.Second):
		t.Error("BroadcastCall timed out")
	}
}

func TestFetcherErrorHandling(t *testing.T) {
	f := &fetcher{}
	f.Init()

	key := "error"
	errMsg := errors.New("fetch failed")
	called := make(chan error, 1)

	f.Register(key, 1, NewDynamicPayloadFetcher(func(ctx context.Context, payload any) (FetchResultInterface, error) {
		return NewFetchResult(500, nil), errMsg
	}), func(result FetchResultInterface) {
		if result.Err() != nil {
			called <- result.Err()
		}
	}, func() bool { return true })

	f.Call(key, "payload")

	select {
	case e := <-called:
		if e.Error() != "fetch failed" {
			t.Errorf("Expected error 'fetch failed', got %v", e)
		}
	case <-time.After(1 * time.Second):
		t.Error("Expected error callback to be triggered")
	}
}
