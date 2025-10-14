package widgets

import (
	"strings"
	"sync"
	"time"

	JC "jxwatcher/core"
)

type runState struct {
	mu    sync.Mutex
	state bool
}

func (s *runState) SetCancel() {
	s.mu.Lock()
	s.state = true
	s.mu.Unlock()
}

func (s *runState) IsCancelled() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state
}

type completionWorker struct {
	searchable []string
	data       []string
	results    []string
	total      int
	chunk      int
	expected   int
	counter    int
	searchKey  string
	delay      time.Duration
	resultChan chan []string
	doneChan   chan struct{}
	done       func(string, []string)
	mu         sync.Mutex
}

func (c *completionWorker) Init() {
	c.expected = (c.total + c.chunk - 1) / c.chunk
	c.resultChan = make(chan []string, c.expected)
	c.doneChan = make(chan struct{}, 100000)
}

func (c *completionWorker) Search(s string, fn func(input string, results []string)) {
	s = strings.ToLower(s)
	JC.Logln("Search triggered:", s)

	c.Cancel()

	c.mu.Lock()
	c.searchKey = s
	c.done = fn
	c.mu.Unlock()

	go c.run()
}

func (c *completionWorker) Cancel() {

	c.close()

	// Allow doneChan to complete
	time.Sleep(15 * time.Millisecond)

	c.drain()

	// Allow drain to complete
	time.Sleep(15 * time.Millisecond)

	c.mu.Lock()
	c.results = []string{}
	c.counter = 0
	c.mu.Unlock()

}
func (c *completionWorker) close() {
	select {
	case c.doneChan <- struct{}{}:
	default:
	}

}

func (c *completionWorker) drain() {
	c.mu.Lock()
	c.counter = -9999
	c.mu.Unlock()

	for {
		select {
		case <-c.doneChan:
		case <-c.resultChan:
		default:
			return
		}
	}
}

func (c *completionWorker) run() {

	cancel := &runState{state: false}
	timer := time.NewTimer(c.delay)

	go func() {
		<-c.doneChan
		cancel.SetCancel()
		timer.Stop()
	}()

	<-timer.C

	if cancel.IsCancelled() {
		return
	}

	c.mu.Lock()
	c.results = []string{}
	c.counter = 0
	c.mu.Unlock()

	go c.listener(cancel)

	for i := 0; i < c.total; i += c.chunk {
		start := i
		end := min(i+c.chunk, c.total)
		go c.worker(start, end, cancel)
	}
}

func (c *completionWorker) worker(start, end int, cancel *runState) {
	local := []string{}
	for j := start; j < end; j++ {
		if cancel.IsCancelled() {
			return
		}
		if strings.Contains(c.searchable[j], c.searchKey) {
			local = append(local, c.data[j])
		}
	}

	c.mu.Lock()
	chanRef := c.resultChan
	c.mu.Unlock()

	if cancel.IsCancelled() || chanRef == nil {
		return
	}

	chanRef <- local
}

func (c *completionWorker) listener(cancel *runState) {
	for {
		if cancel.IsCancelled() {
			c.drain()
			return
		}

		select {
		case <-c.doneChan:
			c.drain()
			return

		case res, ok := <-c.resultChan:
			if !ok {
				c.drain()
				return
			}

			if cancel.IsCancelled() {
				c.drain()
				return
			}

			c.mu.Lock()
			c.results = append(c.results, res...)
			c.counter++

			shouldFire := c.counter == c.expected && c.done != nil && c.counter >= 0
			current := c.searchKey
			c.mu.Unlock()

			if shouldFire && !cancel.IsCancelled() {
				c.mu.Lock()
				c.results = JC.ReorderSearchable(c.results)
				c.mu.Unlock()

				c.done(current, c.results)

				// Important to close the run
				c.close()
			}

		}
	}
}
