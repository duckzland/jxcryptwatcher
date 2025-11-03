package core

import (
	"errors"
	"testing"
	"time"
)

func TestFetcherInit(t *testing.T) {
	f := &fetcher{}
	f.Init()

	if f.fetchers == nil || f.delay == nil || f.callbacks == nil || f.conditions == nil || f.activeWorkers == nil || f.dispatcher == nil {
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

func TestRegisterAndCall(t *testing.T) {
	f := &fetcher{}
	f.Init()

	done := make(chan bool, 1)
	f.Register("test", 1, NewDynamicPayloadFetcher(func(payload any) (FetchResultInterface, error) {
		return NewFetchResult(200, "ok"), nil
	}), func(result FetchResultInterface) {
		if result.Code() == 200 && result.Data() == "ok" {
			done <- true
		}
	}, func() bool { return true })

	f.Call("test", "payload")

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Error("Expected callback to be invoked")
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
		f.Register(key, 1, NewDynamicPayloadFetcher(func(payload any) (FetchResultInterface, error) {
			return NewFetchResult(100, payload), nil
		}), nil, func() bool { return true })
	}

	f.Dispatch(payloads, nil, func(res map[string]FetchResultInterface) {
		resultChan <- res
	})

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

	f.Register("error", 1, NewDynamicPayloadFetcher(func(payload any) (FetchResultInterface, error) {
		return NewFetchResult(500, nil), errMsg
	}), func(result FetchResultInterface) {
		if result.Err() != nil {
			done <- result.Err()
		}
	}, func() bool { return true })

	f.Call("error", "payload")

	select {
	case err := <-done:
		if err.Error() != "fetch failed" {
			t.Errorf("Expected error 'fetch failed', got %v", err)
		}
	case <-time.After(time.Second):
		t.Error("Expected error callback to be triggered")
	}
}
