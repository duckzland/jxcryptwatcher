package types

import (
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
	JC.Database
}

func (tc *tickerDataCacheType) Init() *tickerDataCacheType {
	tc.Reset()
	return tc
}

func (tc *tickerDataCacheType) Get(key string) string {
	if val, ok := tc.Load(key); ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return JC.STRING_EMPTY
}

func (tc *tickerDataCacheType) GetRecentUpdates() map[string]string {
	updates := make(map[string]string, 5)

	tc.RangeRecentUpdates(func(key, value any) bool {
		updates[key.(string)] = value.(string)
		tc.DeleteRecentUpdates(key)
		return true
	})

	return updates
}

func (tc *tickerDataCacheType) Insert(key, value string, timestamp time.Time) *tickerDataCacheType {
	if oldVal, ok := tc.Load(key); ok {
		old := oldVal.(string)
		if old != value {
			tc.StoreRecentUpdates(key, value)
		}
	} else {
		tc.StoreRecentUpdates(key, value)
	}

	// JC.Logf("Ticker received: [%s] = %s", key, value)

	tc.Store(key, value)
	tc.SetTimestamp(time.Now())
	tc.SetLastUpdated(&timestamp)
	return tc
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

	tc.Range(func(key, value any) bool {
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
	tc.Reset()

	for _, entry := range snapshot.Data {
		tc.Store(entry.Key, entry.Value)
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
