package core

import "sync"

type IntStore struct {
	mu    sync.RWMutex
	value int
}

func (s *IntStore) Lock() {
	s.mu.Lock()
}

func (s *IntStore) Unlock() {
	s.mu.Unlock()
}

func (s *IntStore) Set(val int) {
	s.mu.Lock()
	s.value = val
	s.mu.Unlock()
}

func (s *IntStore) Get() int {
	s.mu.RLock()
	val := s.value
	s.mu.RUnlock()
	return val
}

func (s *IntStore) IsEqual(x int) bool {
	s.mu.RLock()
	equal := s.value == x
	s.mu.RUnlock()
	return equal
}
