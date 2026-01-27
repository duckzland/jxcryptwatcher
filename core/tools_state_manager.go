package core

import (
	"sync/atomic"
)

type stateManager struct {
	v atomic.Int64
}

func (m *stateManager) Is(state int64) bool {
	return m.v.Load() == state
}

func (m *stateManager) Change(state int64) {
	m.v.Store(state)
}

func (m *stateManager) Get() int64 {
	return m.v.Load()
}

func (m *stateManager) CompareAndChange(oldState, newState int64) bool {
	return m.v.CompareAndSwap(oldState, newState)
}

func NewStateManager(initial int64) *stateManager {
	m := &stateManager{}
	m.v.Store(initial)
	return m
}
