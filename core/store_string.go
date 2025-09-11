package core

import "sync"

type StringStore struct {
	mu    sync.Mutex
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
	defer s.mu.Unlock()
	s.value = val
}

func (s *StringStore) Get() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.value
}

func (s *StringStore) IsEqual(x string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.value == x
}
