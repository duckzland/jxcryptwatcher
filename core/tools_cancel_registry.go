package core

import (
	"context"
	"sync"
	"sync/atomic"
)

type CancelRegistry struct {
	store    sync.Map
	lenCount atomic.Int64
}

func (r *CancelRegistry) Set(tag string, cancel context.CancelFunc) {
	if _, loaded := r.store.LoadOrStore(tag, cancel); loaded {
		r.store.Store(tag, cancel)
	} else {
		r.lenCount.Add(1)
	}
}

func (r *CancelRegistry) Get(tag string) (context.CancelFunc, bool) {
	if val, ok := r.store.Load(tag); ok {
		return val.(context.CancelFunc), true
	}
	return nil, false
}

func (r *CancelRegistry) Exists(tag string) bool {
	_, ok := r.store.Load(tag)
	return ok
}

func (r *CancelRegistry) Delete(tag string) {
	if _, ok := r.store.Load(tag); ok {
		r.store.Delete(tag)
		r.lenCount.Add(-1)
	}
}

func (r *CancelRegistry) Len() int {
	return int(r.lenCount.Load())
}

func NewCancelRegistry(_ int) *CancelRegistry {
	return &CancelRegistry{}
}
