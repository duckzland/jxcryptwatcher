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

	if f.dispatcher == nil {
		t.Error("Expected dispatcher to be initialized")
	}

	if f.destroyed.Load() {
		t.Error("Fetcher should not be marked destroyed after Init")
	}

	if f.paused.Load() {
		t.Error("Fetcher should not be marked paused after Init")
	}
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
			res := NewFetchResult(500, nil)
			res.SetError(errMsg)
			done <- errMsg
			return res, errMsg
		}),
		func() bool { return true },
	)

	payloads := map[string][]string{"error": {"payload"}}

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
	f.activeWorkers.Store("destroyTest", cancel)

	f.Destroy()

	if !f.destroyed.Load() {
		t.Error("Expected fetcher to be marked destroyed after Destroy")
	}

	var found bool
	f.fetchers.Range(func(_, _ any) bool { found = true; return false })
	if found {
		t.Error("Expected fetchers to be empty after destroy")
	}

	f.conditions.Range(func(_, _ any) bool { found = true; return false })
	if found {
		t.Error("Expected conditions to be empty after destroy")
	}

	f.activeWorkers.Range(func(_, _ any) bool { found = true; return false })
	if found {
		t.Error("Expected activeWorkers to be empty after destroy")
	}

	if f.dispatcher != nil {
		t.Error("Expected dispatcher to be nil after destroy")
	}
}
