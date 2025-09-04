package types

import (
	"sync"
	"time"
)

var TickerCache TickerDataCacheType = TickerDataCacheType{}
var TickerUpdateThreshold = 5 * time.Minute

type TickerDataCacheType struct {
	data        sync.Map
	Timestamp   time.Time
	LastUpdated *time.Time
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

func (ec *TickerDataCacheType) Reset() *TickerDataCacheType {
	ec.data = sync.Map{}
	ec.Timestamp = time.Now()

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

func (ec *TickerDataCacheType) ShouldRefresh() bool {
	if ec.LastUpdated == nil {
		return true
	}

	return time.Now().After(ec.LastUpdated.Add(TickerUpdateThreshold)) && time.Now().After(ec.Timestamp.Add(TickerUpdateThreshold))
}
