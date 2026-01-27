package widgets

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	JC "jxwatcher/core"
)

type completionWorker struct {
	searchable []string
	data       []string
	results    []string
	total      atomic.Int64
	chunk      atomic.Int64
	expected   atomic.Int64
	counter    atomic.Int64
	searchKey  atomic.Value
	delay      atomic.Int64
	resultChan chan []string
	closeChan  chan struct{}
	done       atomic.Value
	mu         sync.Mutex
}

func (c *completionWorker) Init() {
	c.expected.Store((c.total.Load() + c.chunk.Load() - 1) / c.chunk.Load())
	c.resultChan = make(chan []string, int(c.expected.Load()))
	c.closeChan = make(chan struct{}, 100)
}

func (c *completionWorker) Search(s string, fn func(input string, results []string)) {
	c.Cancel()

	c.mu.Lock()
	c.searchKey.Store(strings.ToLower(s))
	c.done.Store(fn)
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
	c.counter.Store(0)
	c.mu.Unlock()

}

func (c *completionWorker) Destroy() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.resultChan != nil {
		close(c.resultChan)
		c.resultChan = nil
	}
	if c.closeChan != nil {
		close(c.closeChan)
		c.closeChan = nil
	}

	c.searchable = nil
	c.data = nil
	c.results = nil

	c.total.Store(0)
	c.chunk.Store(0)
	c.expected.Store(0)
	c.counter.Store(0)
	c.searchKey.Store("")
	c.delay.Store(0)

	var f func(string, []string)
	c.done.Store(f)

}

func (c *completionWorker) close() {
	select {
	case c.closeChan <- struct{}{}:
	default:
	}

}

func (c *completionWorker) drain() {
	c.mu.Lock()
	c.counter.Store(-9999)
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.delay.Load())+to)

	state := NewCompletionWorkerState(false)
	timer := time.NewTimer(time.Duration(c.delay.Load()))

	defer timer.Stop()

	go func() {
		select {
		case <-JC.ShutdownCtx.Done():
			state.SetCancel()
			cancel()
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			return

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
	case <-JC.ShutdownCtx.Done():
		return

	case <-ctx.Done():
		return

	case <-timer.C:
		if state.IsCancelled() {
			return
		}

		c.mu.Lock()
		c.results = []string{}
		c.counter.Store(0)
		c.mu.Unlock()

		go c.listener(state)

		for i := 0; i < int(c.total.Load()); i += int(c.chunk.Load()) {
			start := i
			end := min(i+int(c.chunk.Load()), int(c.total.Load()))
			go c.worker(start, end, state)
		}
	}
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
		if strings.Contains(c.searchable[j], c.searchKey.Load().(string)) {
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

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	for {
		if state.IsCancelled() {
			c.drain()
			return
		}

		select {
		case <-JC.ShutdownCtx.Done():
			return

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
			c.counter.Add(1)

			shouldFire := c.counter.Load() == c.expected.Load() && c.done.Load() != nil && c.counter.Load() >= 0
			current := c.searchKey.Load().(string)
			c.mu.Unlock()

			if shouldFire && !state.IsCancelled() {
				c.mu.Lock()
				c.results = JC.ReorderSearchable(c.results)
				c.mu.Unlock()

				c.done.Load().(func(string, []string))(current, c.results)

				// Important to close the run and go routine!
				c.close()
				return
			}
		}
	}
}

func NewCompletionWorker(searchable []string, data []string, total int, chunk int, delay time.Duration) *completionWorker {
	c := &completionWorker{
		searchable: searchable,
		data:       data,
	}

	c.total.Store(int64(total))
	c.chunk.Store(int64(chunk))
	c.delay.Store(int64(delay))

	return c
}
