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
	Timestamp   time.Time
	LastUpdated *time.Time
	metaLock    sync.RWMutex
}

func (tc *TickerDataCacheType) Init() *TickerDataCacheType {
	tc.data = sync.Map{}
	tc.metaLock.Lock()
	tc.Timestamp = time.Now()
	tc.LastUpdated = nil
	tc.metaLock.Unlock()
	return tc
}

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

	tc.metaLock.Lock()
	tc.Timestamp = time.Now()
	tc.LastUpdated = &timestamp
	tc.metaLock.Unlock()

	return tc
}

func (tc *TickerDataCacheType) Remove(key string) *TickerDataCacheType {
	tc.data.Delete(key)

	tc.metaLock.Lock()
	tc.Timestamp = time.Now()
	tc.metaLock.Unlock()

	return tc
}

func (tc *TickerDataCacheType) SoftReset() *TickerDataCacheType {
	tc.metaLock.Lock()
	tc.Timestamp = time.Now()
	tc.LastUpdated = nil
	tc.metaLock.Unlock()
	return tc
}

func (tc *TickerDataCacheType) Reset() *TickerDataCacheType {
	tc.data = sync.Map{}

	tc.metaLock.Lock()
	tc.Timestamp = time.Now()
	tc.LastUpdated = nil
	tc.metaLock.Unlock()

	return tc
}

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
	tc.metaLock.RLock()
	defer tc.metaLock.RUnlock()

	if tc.LastUpdated == nil {
		return true
	}
	return time.Now().After(tc.LastUpdated.Add(TickerUpdateThreshold)) &&
		time.Now().After(tc.Timestamp.Add(TickerUpdateThreshold))
}

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

	tc.metaLock.RLock()
	timestamp := tc.Timestamp
	var lastUpdated time.Time
	if tc.LastUpdated != nil {
		lastUpdated = *tc.LastUpdated
	}
	tc.metaLock.RUnlock()

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

	tc.metaLock.Lock()
	tc.Timestamp = snapshot.Timestamp
	tc.LastUpdated = &snapshot.LastUpdated
	tc.metaLock.Unlock()
}
