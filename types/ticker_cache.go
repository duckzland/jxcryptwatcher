package types

import (
	"sync"
	"time"
)

var TickerCache tickerDataCacheType = tickerDataCacheType{}
var TickerUpdateThreshold = 2 * time.Minute

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
	data        sync.Map
	timestamp   time.Time
	lastUpdated *time.Time
	metaLock    sync.RWMutex
}

// Getters and Setters
func (tc *tickerDataCacheType) GetTimestamp() time.Time {
	tc.metaLock.RLock()
	defer tc.metaLock.RUnlock()
	return tc.timestamp
}

func (tc *tickerDataCacheType) SetTimestamp(t time.Time) {
	tc.metaLock.Lock()
	tc.timestamp = t
	tc.metaLock.Unlock()
}

func (tc *tickerDataCacheType) GetLastUpdated() *time.Time {
	tc.metaLock.RLock()
	defer tc.metaLock.RUnlock()
	return tc.lastUpdated
}

func (tc *tickerDataCacheType) SetLastUpdated(t *time.Time) {
	tc.metaLock.Lock()
	tc.lastUpdated = t
	tc.metaLock.Unlock()
}

// Initialization
func (tc *tickerDataCacheType) Init() *tickerDataCacheType {
	tc.data = sync.Map{}
	tc.SetTimestamp(time.Now())
	tc.SetLastUpdated(nil)
	return tc
}

// Core operations
func (tc *tickerDataCacheType) Get(key string) string {
	if val, ok := tc.data.Load(key); ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return ""
}

func (tc *tickerDataCacheType) Insert(key, value string, timestamp time.Time) *tickerDataCacheType {
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

// Status checks
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
	return time.Now().After(last.Add(TickerUpdateThreshold)) &&
		time.Now().After(tc.GetTimestamp().Add(TickerUpdateThreshold))
}

// Serialization
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
