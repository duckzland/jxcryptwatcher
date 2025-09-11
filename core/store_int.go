package core

import "sync"

type IntStore struct {
	mu    sync.Mutex
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
	defer s.mu.Unlock()
	s.value = val
}

func (s *IntStore) Get() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.value
}

func (s *IntStore) IsEqual(x int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.value == x
}
