package types

import (
	"sync"
	"time"

	JC "jxwatcher/core"
)

var tickerCacheStorage *tickerDataCacheType = nil
var tickerUpdateThreshold = 10 * time.Second

type tickerDataCacheSnapshot struct {
	Data        []tickerDataCacheEntry `json:"data"`
	Timestamp   time.Time              `json:"timestamp"`
	LastUpdated time.Time              `json:"last_updated"`
}

type tickerDataCacheEntry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type tickerDataCacheType struct {
	data          sync.Map
	timestamp     time.Time
	lastUpdated   *time.Time
	recentUpdates sync.Map
	mu            sync.RWMutex
}

func (tc *tickerDataCacheType) GetTimestamp() time.Time {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.timestamp
}

func (tc *tickerDataCacheType) SetTimestamp(t time.Time) {
	tc.mu.Lock()
	tc.timestamp = t
	tc.mu.Unlock()
}

func (tc *tickerDataCacheType) GetLastUpdated() *time.Time {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.lastUpdated
}

func (tc *tickerDataCacheType) SetLastUpdated(t *time.Time) {
	tc.mu.Lock()
	tc.lastUpdated = t
	tc.mu.Unlock()
}

func (tc *tickerDataCacheType) Init() *tickerDataCacheType {
	tc.data = sync.Map{}
	tc.SetTimestamp(time.Now())
	tc.SetLastUpdated(nil)
	return tc
}

func (tc *tickerDataCacheType) Get(key string) string {
	if val, ok := tc.data.Load(key); ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return JC.STRING_EMPTY
}

func (tc *tickerDataCacheType) GetRecentUpdates() map[string]string {
	updates := make(map[string]string)
	tc.recentUpdates.Range(func(k, v any) bool {
		if strVal, ok := v.(string); ok {
			updates[k.(string)] = strVal
		}
		return true
	})
	tc.recentUpdates = sync.Map{}
	return updates
}

func (tc *tickerDataCacheType) Insert(key, value string, timestamp time.Time) *tickerDataCacheType {
	if oldVal, ok := tc.data.Load(key); ok {
		old := oldVal.(string)
		if old != value {
			tc.recentUpdates.Store(key, value)
		}
	} else {
		tc.recentUpdates.Store(key, value)
	}

	// JC.Logf("Ticker received: [%s] = %s", key, value)

	tc.data.Store(key, value)
	tc.SetTimestamp(time.Now())
	tc.SetLastUpdated(&timestamp)
	return tc
}

func (tc *tickerDataCacheType) Remove(key string) *tickerDataCacheType {
	tc.data.Delete(key)
	tc.SetTimestamp(time.Now())
	return tc
}

func (tc *tickerDataCacheType) SoftReset() *tickerDataCacheType {
	tc.SetTimestamp(time.Now())
	tc.SetLastUpdated(nil)
	return tc
}

func (tc *tickerDataCacheType) Reset() *tickerDataCacheType {
	tc.data = sync.Map{}
	tc.SetTimestamp(time.Now())
	tc.SetLastUpdated(nil)
	return tc
}

func (tc *tickerDataCacheType) Has(key string) bool {
	d, ok := tc.data.Load(key)
	return ok && d != nil
}

func (tc *tickerDataCacheType) HasData() bool {
	isEmpty := true
	tc.data.Range(func(_, _ interface{}) bool {
		isEmpty = false
		return false
	})
	return !isEmpty
}

func (tc *tickerDataCacheType) IsEmpty() bool {
	return !tc.HasData()
}

func (tc *tickerDataCacheType) ShouldRefresh() bool {
	last := tc.GetLastUpdated()
	if last == nil {
		return true
	}
	return time.Now().After(last.Add(tickerUpdateThreshold)) &&
		time.Now().After(tc.GetTimestamp().Add(tickerUpdateThreshold))
}

func (tc *tickerDataCacheType) Serialize() tickerDataCacheSnapshot {
	var entries []tickerDataCacheEntry

	tc.data.Range(func(key, value any) bool {
		k, ok1 := key.(string)
		v, ok2 := value.(string)
		if ok1 && ok2 {
			entries = append(entries, tickerDataCacheEntry{
				Key:   k,
				Value: v,
			})
		}
		return true
	})

	timestamp := tc.GetTimestamp()
	last := tc.GetLastUpdated()
	var lastUpdated time.Time
	if last != nil {
		lastUpdated = *last
	}

	return tickerDataCacheSnapshot{
		Data:        entries,
		Timestamp:   timestamp,
		LastUpdated: lastUpdated,
	}
}

func (tc *tickerDataCacheType) Hydrate(snapshot tickerDataCacheSnapshot) {
	tc.data = sync.Map{}
	for _, entry := range snapshot.Data {
		tc.data.Store(entry.Key, entry.Value)
	}
	tc.SetTimestamp(snapshot.Timestamp)
	tc.SetLastUpdated(&snapshot.LastUpdated)
}

func NewTickerDataCacheSnapshot() *tickerDataCacheSnapshot {
	return &tickerDataCacheSnapshot{}
}

func RegisterTickerCache() *tickerDataCacheType {
	if tickerCacheStorage == nil {
		tickerCacheStorage = &tickerDataCacheType{}
	}

	return tickerCacheStorage
}

func UseTickerCache() *tickerDataCacheType {
	return tickerCacheStorage
}
