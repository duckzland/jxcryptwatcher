package core

import (
	"context"
)

type fetcherUnit struct {
	handler func(ctx context.Context, payload any) (FetchResultInterface, error)
}

func (df *fetcherUnit) Fetch(ctx context.Context, payload any, callback func(FetchResultInterface)) {
	result, err := df.handler(ctx, payload)

	if err != nil {
		result.SetError(err)
	}

	callback(result)
}

func NewFetcherUnit(handler func(ctx context.Context, payload any) (FetchResultInterface, error)) *fetcherUnit {
	return &fetcherUnit{handler: handler}
}
