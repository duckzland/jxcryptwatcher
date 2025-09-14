package types

import (
	"sync"
	"time"
)

var TickerCache TickerDataCacheType = TickerDataCacheType{}
var TickerUpdateThreshold = 2 * time.Minute

type TickerDataCacheSnapshot struct {
	Data        []TickerDataCacheEntry `json:"data"`
	Timestamp   time.Time              `json:"timestamp"`
	LastUpdated time.Time              `json:"last_updated"`
}

type TickerDataCacheEntry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type TickerDataCacheType struct {
	data        sync.Map
	timestamp   time.Time
	lastUpdated *time.Time
	metaLock    sync.RWMutex
}

// Getters and Setters
func (tc *TickerDataCacheType) GetTimestamp() time.Time {
	tc.metaLock.RLock()
	defer tc.metaLock.RUnlock()
	return tc.timestamp
}

func (tc *TickerDataCacheType) SetTimestamp(t time.Time) {
	tc.metaLock.Lock()
	tc.timestamp = t
	tc.metaLock.Unlock()
}

func (tc *TickerDataCacheType) GetLastUpdated() *time.Time {
	tc.metaLock.RLock()
	defer tc.metaLock.RUnlock()
	return tc.lastUpdated
}

func (tc *TickerDataCacheType) SetLastUpdated(t *time.Time) {
	tc.metaLock.Lock()
	tc.lastUpdated = t
	tc.metaLock.Unlock()
}

// Initialization
func (tc *TickerDataCacheType) Init() *TickerDataCacheType {
	tc.data = sync.Map{}
	tc.SetTimestamp(time.Now())
	tc.SetLastUpdated(nil)
	return tc
}

// Core operations
func (tc *TickerDataCacheType) Get(key string) string {
	if val, ok := tc.data.Load(key); ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return ""
}

func (tc *TickerDataCacheType) Insert(key, value string, timestamp time.Time) *TickerDataCacheType {
	tc.data.Store(key, value)
	tc.SetTimestamp(time.Now())
	tc.SetLastUpdated(&timestamp)
	return tc
}

func (tc *TickerDataCacheType) Remove(key string) *TickerDataCacheType {
	tc.data.Delete(key)
	tc.SetTimestamp(time.Now())
	return tc
}

func (tc *TickerDataCacheType) SoftReset() *TickerDataCacheType {
	tc.SetTimestamp(time.Now())
	tc.SetLastUpdated(nil)
	return tc
}

func (tc *TickerDataCacheType) Reset() *TickerDataCacheType {
	tc.data = sync.Map{}
	tc.SetTimestamp(time.Now())
	tc.SetLastUpdated(nil)
	return tc
}

// Status checks
func (tc *TickerDataCacheType) Has(key string) bool {
	d, ok := tc.data.Load(key)
	return ok && d != nil
}

func (tc *TickerDataCacheType) HasData() bool {
	isEmpty := true
	tc.data.Range(func(_, _ interface{}) bool {
		isEmpty = false
		return false
	})
	return !isEmpty
}

func (tc *TickerDataCacheType) IsEmpty() bool {
	return !tc.HasData()
}

func (tc *TickerDataCacheType) ShouldRefresh() bool {
	last := tc.GetLastUpdated()
	if last == nil {
		return true
	}
	return time.Now().After(last.Add(TickerUpdateThreshold)) &&
		time.Now().After(tc.GetTimestamp().Add(TickerUpdateThreshold))
}

// Serialization
func (tc *TickerDataCacheType) Serialize() TickerDataCacheSnapshot {
	var entries []TickerDataCacheEntry

	tc.data.Range(func(key, value any) bool {
		k, ok1 := key.(string)
		v, ok2 := value.(string)
		if ok1 && ok2 {
			entries = append(entries, TickerDataCacheEntry{
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

	return TickerDataCacheSnapshot{
		Data:        entries,
		Timestamp:   timestamp,
		LastUpdated: lastUpdated,
	}
}

func (tc *TickerDataCacheType) Hydrate(snapshot TickerDataCacheSnapshot) {
	tc.data = sync.Map{}
	for _, entry := range snapshot.Data {
		tc.data.Store(entry.Key, entry.Value)
	}
	tc.SetTimestamp(snapshot.Timestamp)
	tc.SetLastUpdated(&snapshot.LastUpdated)
}
