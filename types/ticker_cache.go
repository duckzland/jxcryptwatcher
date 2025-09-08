package types

import (
	"sync"
	"time"
)

var TickerCache TickerDataCacheType = TickerDataCacheType{}
var TickerUpdateThreshold = 5 * time.Minute

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
}

func (ec *TickerDataCacheType) Init() *TickerDataCacheType {
	ec.data = sync.Map{}
	ec.Timestamp = time.Now()
	ec.LastUpdated = nil

	return ec
}

func (ec *TickerDataCacheType) Get(key string) string {
	if val, ok := ec.data.Load(key); ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return ""
}

func (ec *TickerDataCacheType) Insert(key, value string, timestamp time.Time) *TickerDataCacheType {
	ec.data.Store(key, value)
	ec.Timestamp = time.Now()
	ec.LastUpdated = &timestamp

	return ec
}

func (ec *TickerDataCacheType) Remove(key string) *TickerDataCacheType {
	ec.data.Delete(key)
	ec.Timestamp = time.Now()

	return ec
}

func (ec *TickerDataCacheType) SoftReset() *TickerDataCacheType {
	ec.Timestamp = time.Now()
	ec.LastUpdated = nil

	return ec
}

func (ec *TickerDataCacheType) Reset() *TickerDataCacheType {
	ec.data = sync.Map{}
	ec.Timestamp = time.Now()
	ec.LastUpdated = nil

	return ec
}

func (ec *TickerDataCacheType) Has(key string) bool {
	d, ok := ec.data.Load(key)

	return ok && d != nil
}

func (ec *TickerDataCacheType) HasData() bool {

	isEmpty := true
	ec.data.Range(func(key, value interface{}) bool {
		isEmpty = false
		return false
	})

	return !isEmpty
}

func (ec *TickerDataCacheType) IsEmpty() bool {
	return ec.HasData() == false
}

func (ec *TickerDataCacheType) ShouldRefresh() bool {
	if ec.LastUpdated == nil {
		return true
	}

	return time.Now().After(ec.LastUpdated.Add(TickerUpdateThreshold)) && time.Now().After(ec.Timestamp.Add(TickerUpdateThreshold))
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

	var lastUpdated time.Time
	if tc.LastUpdated != nil {
		lastUpdated = *tc.LastUpdated
	}

	return TickerDataCacheSnapshot{
		Data:        entries,
		Timestamp:   tc.Timestamp,
		LastUpdated: lastUpdated,
	}
}

func (tc *TickerDataCacheType) Hydrate(snapshot TickerDataCacheSnapshot) {
	tc.data = sync.Map{}
	for _, entry := range snapshot.Data {
		tc.data.Store(entry.Key, entry.Value)
	}
	tc.Timestamp = snapshot.Timestamp
	tc.LastUpdated = &snapshot.LastUpdated
}
