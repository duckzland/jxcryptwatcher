package core

import (
	"context"
	"sync"
	"sync/atomic"
)

type cancelRegistry struct {
	data  sync.Map
	count atomic.Int64
}

func (r *cancelRegistry) Set(tag string, cancel context.CancelFunc) {
	_, loaded := r.data.Load(tag)
	r.data.Store(tag, cancel)
	if !loaded {
		r.count.Add(1)
	}
}

func (r *cancelRegistry) Get(tag string) (context.CancelFunc, bool) {
	if val, ok := r.data.Load(tag); ok {
		return val.(context.CancelFunc), true
	}
	return nil, false
}

func (r *cancelRegistry) Exists(tag string) bool {
	_, ok := r.data.Load(tag)
	return ok
}

func (r *cancelRegistry) Delete(tag string) {
	if _, ok := r.data.Load(tag); ok {
		r.data.Delete(tag)
		r.count.Add(-1)
	}
}

func (r *cancelRegistry) Cancel(tag string) {
	if v, ok := r.data.Load(tag); ok {
		cancel := v.(context.CancelFunc)
		cancel()
		r.Delete(tag)
	}
}

func (r *cancelRegistry) Len() int {
	return int(r.count.Load())
}

func (r *cancelRegistry) Range(fn func(key string, cancel context.CancelFunc) bool) {
	r.data.Range(func(k, v any) bool {
		return fn(k.(string), v.(context.CancelFunc))
	})
}

func (r *cancelRegistry) Destroy() {
	r.data.Range(func(k, v any) bool {
		r.Cancel(k.(string))
		return true
	})
}

func NewCancelRegistry() *cancelRegistry {
	return &cancelRegistry{}
}
