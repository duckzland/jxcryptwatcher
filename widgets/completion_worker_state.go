package widgets

import "sync"

type completionWorkerState struct {
	mu    sync.Mutex
	state bool
}

func (s *completionWorkerState) SetCancel() {
	s.mu.Lock()
	s.state = true
	s.mu.Unlock()
}

func (s *completionWorkerState) IsCancelled() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state
}
