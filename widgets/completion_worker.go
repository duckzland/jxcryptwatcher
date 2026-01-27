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
	searchable sync.Map
	data       sync.Map
	results    sync.Map
	total      atomic.Int64
	chunk      atomic.Int64
	expected   atomic.Int64
	counter    atomic.Int64
	searchKey  atomic.Value
	delay      atomic.Int64
	resultChan chan []string
	closeChan  chan struct{}
	done       atomic.Value
}

func (c *completionWorker) Init() {
	c.expected.Store((c.total.Load() + c.chunk.Load() - 1) / c.chunk.Load())
	c.resultChan = make(chan []string, int(c.expected.Load()))
	c.closeChan = make(chan struct{}, 100)
}

func (c *completionWorker) Search(s string, fn func(input string, results []string)) {
	c.Cancel()
	c.searchKey.Store(strings.ToLower(s))
	c.done.Store(fn)

	go c.run()
}

func (c *completionWorker) Cancel() {
	if c.closeChan != nil {
		c.close()
		time.Sleep(15 * time.Millisecond)
	}

	if c.resultChan != nil && c.closeChan != nil {
		c.drain()
		time.Sleep(15 * time.Millisecond)
	}

	c.results = sync.Map{}
	c.counter.Store(0)
}

func (c *completionWorker) Destroy() {
	if c.resultChan != nil {
		close(c.resultChan)
		c.resultChan = nil
	}
	if c.closeChan != nil {
		close(c.closeChan)
		c.closeChan = nil
	}

	c.searchable = sync.Map{}
	c.data = sync.Map{}
	c.results = sync.Map{}

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
	if c.closeChan == nil {
		return
	}
	select {
	case c.closeChan <- struct{}{}:
	default:
	}
}

func (c *completionWorker) drain() {
	if c.closeChan == nil || c.resultChan == nil {
		return
	}

	c.counter.Store(-9999)

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

		c.results = sync.Map{}
		c.counter.Store(0)

		go c.listener(state)

		total := int(c.total.Load())
		chunk := int(c.chunk.Load())

		for i := 0; i < total; i += chunk {
			start := i
			end := min(i+chunk, total)
			go c.worker(start, end, state)
		}
	}
}

func (c *completionWorker) worker(start, end int, state *completionWorkerState) {
	if state.IsCancelled() {
		return
	}

	local := []string{}
	key := c.searchKey.Load().(string)

	for j := start; j < end; j++ {
		if state.IsCancelled() {
			return
		}

		sv, ok1 := c.searchable.Load(j)
		dv, ok2 := c.data.Load(j)
		if !ok1 || !ok2 {
			continue
		}

		if strings.Contains(sv.(string), key) {
			local = append(local, dv.(string))
		}
	}

	if state.IsCancelled() {
		return
	}

	if c.resultChan != nil {
		c.resultChan <- local
	}
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

			for _, v := range res {
				c.results.Store(v, v)
			}

			c.counter.Add(1)

			shouldFire := c.counter.Load() == c.expected.Load() &&
				c.done.Load() != nil &&
				c.counter.Load() >= 0

			current := c.searchKey.Load().(string)

			if shouldFire && !state.IsCancelled() {

				ttl := 0
				c.results.Range(func(_, v any) bool {
					ttl++
					return true
				})

				tmp := make([]string, 0, ttl)
				c.results.Range(func(_, v any) bool {
					tmp = append(tmp, v.(string))
					return true
				})

				c.done.Load().(func(string, []string))(current, JC.ReorderSearchable(tmp))

				// Important to close the run and go routine!
				c.close()
				return
			}
		}
	}
}

func NewCompletionWorker(searchable []string, data []string, total int, chunk int, delay time.Duration) *completionWorker {
	c := &completionWorker{}

	for i, v := range searchable {
		c.searchable.Store(i, v)
	}
	for i, v := range data {
		c.data.Store(i, v)
	}

	c.total.Store(int64(total))
	c.chunk.Store(int64(chunk))
	c.delay.Store(int64(delay))

	return c
}
