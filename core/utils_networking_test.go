package core

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestFetcher_Success(t *testing.T) {
	f := NewFetcherUnit(func(ctx context.Context, payload any) (FetchResultInterface, error) {
		return NewFetchResult(123, nil), nil
	})
	f.Fetch(context.Background(), nil, func(res FetchResultInterface) {
		if res.Code() != 123 {
			t.Errorf("expected code 123, got %d", res.Code())
		}
	})
}

func TestFetcher_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	f := NewFetcherUnit(func(ctx context.Context, payload any) (FetchResultInterface, error) {
		select {
		case <-ctx.Done():
			return NewFetchResult(-1, nil), ctx.Err()
		case <-time.After(50 * time.Millisecond):
			return NewFetchResult(123, nil), nil
		}
	})
	f.Fetch(ctx, nil, func(res FetchResultInterface) {
		if !errors.Is(res.Err(), context.Canceled) {
			t.Errorf("expected context.Canceled, got %v", res.Err())
		}
		if res.Code() != -1 {
			t.Errorf("expected code -1 for cancelled, got %d", res.Code())
		}
	})
}

func TestFetcher_DeadlineExceeded(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	f := NewFetcherUnit(func(ctx context.Context, payload any) (FetchResultInterface, error) {
		select {
		case <-ctx.Done():
			return NewFetchResult(-1, nil), ctx.Err()
		case <-time.After(100 * time.Millisecond):
			return NewFetchResult(123, nil), nil
		}
	})
	f.Fetch(ctx, nil, func(res FetchResultInterface) {
		if !errors.Is(res.Err(), context.DeadlineExceeded) {
			t.Errorf("expected deadline exceeded, got %v", res.Err())
		}
		if res.Code() != -1 {
			t.Errorf("expected code -1 for deadline exceeded, got %d", res.Code())
		}
	})
}

func TestFetcher_UsesPayload(t *testing.T) {
	f := NewFetcherUnit(func(ctx context.Context, payload any) (FetchResultInterface, error) {
		if payload == "ok" {
			return NewFetchResult(200, nil), nil
		}
		return NewFetchResult(400, nil), fmt.Errorf("bad payload")
	})
	f.Fetch(context.Background(), "ok", func(res FetchResultInterface) {
		if res.Code() != 200 {
			t.Errorf("expected code 200, got %d", res.Code())
		}
	})
	f.Fetch(context.Background(), "fail", func(res FetchResultInterface) {
		if res.Code() != 400 {
			t.Errorf("expected code 400 for bad payload, got %d", res.Code())
		}
		if res.Err() == nil {
			t.Errorf("expected error for bad payload, got nil")
		}
	})
}
