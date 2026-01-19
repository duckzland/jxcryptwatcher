package types

import (
	"time"

	JC "jxwatcher/core"
)

var tickerCacheStorage *tickerDataCacheType = nil

type tickerDataCacheSnapshot struct {
	Data      []tickerDataCacheEntry `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
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
	tc.SetUpdateTreshold(10 * time.Second)
	return tc
}

func (tc *tickerDataCacheType) Get(key string) string {
	if val, ok := tc.UseData().Load(key); ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return JC.STRING_EMPTY
}

func (tc *tickerDataCacheType) GetRecentUpdates() map[string]string {
	updates := make(map[string]string, 5)

	tc.UseUpdates().Range(func(key, value any) bool {
		updates[key.(string)] = value.(string)
		tc.UseUpdates().Delete(key)
		return true
	})

	return updates
}

func (tc *tickerDataCacheType) Insert(key, value string, timestamp time.Time) *tickerDataCacheType {
	if oldVal, ok := tc.UseData().Load(key); ok {
		old := oldVal.(string)
		if old != value {
			tc.UseUpdates().Store(key, value)
		}
	} else {
		tc.UseUpdates().Store(key, value)
	}

	// JC.Logf("Ticker received: [%s] = %s", key, value)

	tc.UseData().Store(key, value)
	tc.UpdatedAt(&timestamp)
	return tc
}

func (tc *tickerDataCacheType) Serialize() tickerDataCacheSnapshot {
	var entries []tickerDataCacheEntry

	tc.UseData().Range(func(key, value any) bool {
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

	last := tc.IsUpdatedAt()
	var lastUpdated time.Time
	if last != nil {
		lastUpdated = *last
	}

	return tickerDataCacheSnapshot{
		Data:      entries,
		Timestamp: lastUpdated,
	}
}

func (tc *tickerDataCacheType) Hydrate(snapshot tickerDataCacheSnapshot) {
	tc.Reset()

	for _, entry := range snapshot.Data {
		tc.UseData().Store(entry.Key, entry.Value)
	}

	tc.UpdatedAt(&snapshot.Timestamp)
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
