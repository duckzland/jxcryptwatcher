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

	if f.fetchers == nil || f.conditions == nil || f.activeWorkers == nil || f.dispatcher == nil {
		t.Error("Expected internal maps and dispatcher to be initialized")
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Mutex caused panic: %v", r)
		}
	}()
	f.mu.Lock()
	_ = f
	f.mu.Unlock()
}

func TestSingletonFetcherManager(t *testing.T) {
	if RegisterFetcherManager() != UseFetcher() {
		t.Error("Expected singleton fetcher instance")
	}
}

func TestDispatch(t *testing.T) {
	f := &fetcher{}
	f.Init()

	payloads := map[string][]string{
		"a": {"1"},
		"b": {"2"},
	}
	resultChan := make(chan map[string]FetchResultInterface, 1)

	for key := range payloads {
		f.Register(
			key,
			NewFetcherUnit(func(ctx context.Context, payload any) (FetchResultInterface, error) {
				return NewFetchResult(100, payload), nil
			}),
			func() bool { return true },
		)
	}

	f.Call(payloads,
		func(totalJob int) {},
		func(res map[string]FetchResultInterface) {
			resultChan <- res
		},
		func() {})

	select {
	case res := <-resultChan:
		if len(res) != 2 {
			t.Errorf("Expected 2 results, got %d", len(res))
		}
	case <-time.After(2 * time.Second):
		t.Error("Dispatch timed out")
	}
}

func TestErrorHandling(t *testing.T) {
	f := &fetcher{}
	f.Init()

	errMsg := errors.New("fetch failed")
	done := make(chan error, 1)

	f.Register("error",
		NewFetcherUnit(func(ctx context.Context, payload any) (FetchResultInterface, error) {
			// handler sets the error on the result and returns the same error
			res := NewFetchResult(500, nil)
			res.SetError(errMsg)
			done <- errMsg
			return res, errMsg
		}),
		func() bool { return true },
	)
	payloads := make(map[string][]string, 1)
	payloads["error"] = []string{"payload"}
	f.Call(payloads,
		func(scheduledJobs int) {},
		func(results map[string]FetchResultInterface) {},
		func() {},
	)

	select {
	case err := <-done:
		if err.Error() != "fetch failed" {
			t.Errorf("Expected error 'fetch failed', got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("Expected handler to set error and signal")
	}
}

func TestFetcherDestroy(t *testing.T) {
	f := &fetcher{}
	f.Init()

	f.Register(
		"destroyTest",
		NewFetcherUnit(func(ctx context.Context, payload any) (FetchResultInterface, error) {
			return NewFetchResult(200, "ok"), nil
		}),
		func() bool { return true },
	)

	_, cancel := context.WithCancel(context.Background())
	f.mu.Lock()
	f.activeWorkers["destroyTest"] = cancel
	f.mu.Unlock()

	f.Destroy()

	f.mu.Lock()
	defer f.mu.Unlock()

	if f.fetchers != nil {
		t.Error("Expected fetchers to be nil after destroy")
	}
	if f.conditions != nil {
		t.Error("Expected conditions to be nil after destroy")
	}
	if f.activeWorkers != nil {
		t.Error("Expected activeWorkers to be nil after destroy")
	}
	if f.dispatcher != nil {
		t.Error("Expected dispatcher to be nil after destroy")
	}
}
