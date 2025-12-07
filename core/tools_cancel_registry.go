package core

import (
	"context"
	"sync"
)

type CancelRegistry struct {
	mu    sync.Mutex
	store map[string]context.CancelFunc
}

func (r *CancelRegistry) Set(tag string, cancel context.CancelFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.store[tag] = cancel
}

func (r *CancelRegistry) Get(tag string) (context.CancelFunc, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	cancel, ok := r.store[tag]
	return cancel, ok
}

func (r *CancelRegistry) Exists(tag string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.store[tag]
	return ok
}

func (r *CancelRegistry) Delete(tag string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.store, tag)
}

func (r *CancelRegistry) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.store)
}

func NewCancelRegistry(bsize int) *CancelRegistry {
	return &CancelRegistry{
		store: make(map[string]context.CancelFunc, bsize),
	}
}
