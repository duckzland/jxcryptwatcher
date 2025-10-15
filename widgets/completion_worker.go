package widgets

import (
	"context"
	"strings"
	"sync"
	"time"

	JC "jxwatcher/core"
)

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
	closeChan  chan struct{}
	done       func(string, []string)
	mu         sync.Mutex
}

func (c *completionWorker) Init() {
	c.expected = (c.total + c.chunk - 1) / c.chunk
	c.resultChan = make(chan []string, c.expected)
	c.closeChan = make(chan struct{}, 100000)
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

	// Allow closeChan to complete
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
	case c.closeChan <- struct{}{}:
	default:
	}

}

func (c *completionWorker) drain() {
	c.mu.Lock()
	c.counter = -9999
	c.mu.Unlock()

	for {
		select {
		case <-c.closeChan:
		case <-c.resultChan:
		default:
			return
		}
	}
}

func (c *completionWorker) run() {

	// If stale after the delay and additional expected operational duration
	// Force close to prevent ghost go routine
	to := 500 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), c.delay+to)

	state := &completionWorkerState{state: false}
	timer := time.NewTimer(c.delay)

	go func() {
		select {
		case <-ctx.Done():
			state.SetCancel()
			return

		case <-c.closeChan:
			state.SetCancel()
			cancel()
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			return
		}
	}()

	select {
	case <-ctx.Done():
		return

	case <-timer.C:
		if state.IsCancelled() {
			return
		}

		c.mu.Lock()
		c.results = []string{}
		c.counter = 0
		c.mu.Unlock()

		go c.listener(state)

		for i := 0; i < c.total; i += c.chunk {
			start := i
			end := min(i+c.chunk, c.total)
			go c.worker(start, end, state)
		}
	}

	JC.TraceGoroutines()
}

func (c *completionWorker) worker(start, end int, state *completionWorkerState) {
	if state.IsCancelled() {
		return
	}

	local := []string{}
	for j := start; j < end; j++ {
		if state.IsCancelled() {
			return
		}
		if strings.Contains(c.searchable[j], c.searchKey) {
			local = append(local, c.data[j])
		}
	}

	if state.IsCancelled() {
		return
	}

	c.mu.Lock()
	if c.resultChan != nil {
		c.resultChan <- local
	}
	c.mu.Unlock()
}

func (c *completionWorker) listener(state *completionWorkerState) {

	ctx, _ := context.WithTimeout(context.Background(), 500*time.Millisecond)

	for {
		if state.IsCancelled() {
			c.drain()
			return
		}

		select {
		case <-ctx.Done():
			return

		case <-c.closeChan:
			c.drain()
			return

		case res, ok := <-c.resultChan:
			if !ok {
				c.drain()
				return
			}

			if state.IsCancelled() {
				c.drain()
				return
			}

			c.mu.Lock()
			c.results = append(c.results, res...)
			c.counter++

			shouldFire := c.counter == c.expected && c.done != nil && c.counter >= 0
			current := c.searchKey
			c.mu.Unlock()

			if shouldFire && !state.IsCancelled() {
				c.mu.Lock()
				c.results = JC.ReorderSearchable(c.results)
				c.mu.Unlock()

				c.done(current, c.results)

				// Important to close the run and go routine!
				c.close()
				return
			}
		}
	}
}
