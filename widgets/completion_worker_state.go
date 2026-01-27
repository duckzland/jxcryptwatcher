package widgets

import "sync/atomic"

type completionWorkerState struct {
	state atomic.Bool
}

func (s *completionWorkerState) SetCancel() {
	s.state.Store(true)
}

func (s *completionWorkerState) IsCancelled() bool {
	return s.state.Load()
}

func NewCompletionWorkerState(initial bool) *completionWorkerState {
	s := &completionWorkerState{}
	s.state.Store(initial)
	return s
}
