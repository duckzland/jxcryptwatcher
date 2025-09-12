package core

import "sync"

type StringStore struct {
	mu    sync.RWMutex
	value string
}

func (s *StringStore) Lock() {
	s.mu.Lock()
}

func (s *StringStore) Unlock() {
	s.mu.Unlock()
}

func (s *StringStore) Set(val string) {
	s.mu.Lock()
	s.value = val
	s.mu.Unlock()
}

func (s *StringStore) Get() string {
	s.mu.RLock()
	val := s.value
	s.mu.RUnlock()
	return val
}

func (s *StringStore) IsEqual(x string) bool {
	s.mu.RLock()
	equal := s.value == x
	s.mu.RUnlock()
	return equal
}
