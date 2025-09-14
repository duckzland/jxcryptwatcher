package types

import (
	"fmt"
	"sync"
	"time"
)

var ExchangeCache ExchangeDataCacheType = ExchangeDataCacheType{}
var CMCUpdateThreshold = 1 * time.Minute

type ExchangeDataCacheSnapshot struct {
	Data        []ExchangeDataType `json:"data"`
	Timestamp   time.Time          `json:"timestamp"`
	LastUpdated time.Time          `json:"last_updated"`
}

type ExchangeDataCacheType struct {
	data        sync.Map
	Timestamp   time.Time
	LastUpdated *time.Time
	metaLock    sync.RWMutex
}

func (ec *ExchangeDataCacheType) Init() *ExchangeDataCacheType {
	ec.data = sync.Map{}
	ec.metaLock.Lock()
	ec.Timestamp = time.Now()
	ec.LastUpdated = nil
	ec.metaLock.Unlock()
	return ec
}

func (ec *ExchangeDataCacheType) Get(ck string) *ExchangeDataType {
	if val, ok := ec.data.Load(ck); ok {
		ex := val.(ExchangeDataType)
		return &ex
	}
	return nil
}

func (ec *ExchangeDataCacheType) Insert(ex *ExchangeDataType) *ExchangeDataCacheType {
	ck := ec.CreateKeyFromExchangeData(ex)
	ec.data.Store(ck, *ex)

	ec.metaLock.Lock()
	ec.Timestamp = time.Now()
	ec.LastUpdated = &ex.Timestamp
	ec.metaLock.Unlock()

	return ec
}

func (ec *ExchangeDataCacheType) Remove(ck string) *ExchangeDataCacheType {
	ec.data.Delete(ck)

	ec.metaLock.Lock()
	ec.Timestamp = time.Now()
	ec.metaLock.Unlock()

	return ec
}

func (ec *ExchangeDataCacheType) SoftReset() *ExchangeDataCacheType {
	ec.metaLock.Lock()
	ec.Timestamp = time.Now()
	ec.LastUpdated = nil
	ec.metaLock.Unlock()
	return ec
}

func (ec *ExchangeDataCacheType) Reset() *ExchangeDataCacheType {
	ec.data = sync.Map{}

	ec.metaLock.Lock()
	ec.Timestamp = time.Now()
	ec.LastUpdated = nil
	ec.metaLock.Unlock()

	return ec
}

func (ec *ExchangeDataCacheType) Has(ck string) bool {
	d, ok := ec.data.Load(ck)
	return ok && d != nil
}

func (ec *ExchangeDataCacheType) HasData() bool {
	isEmpty := true
	ec.data.Range(func(_, _ interface{}) bool {
		isEmpty = false
		return false
	})
	return !isEmpty
}

func (ec *ExchangeDataCacheType) IsEmpty() bool {
	return !ec.HasData()
}

func (ec *ExchangeDataCacheType) ShouldRefresh() bool {
	ec.metaLock.RLock()
	defer ec.metaLock.RUnlock()

	if ec.LastUpdated == nil {
		return true
	}
	return time.Now().After(ec.LastUpdated.Add(CMCUpdateThreshold))
}

func (ec *ExchangeDataCacheType) Serialize() ExchangeDataCacheSnapshot {
	var result []ExchangeDataType
	cutoff := time.Now().Add(-24 * time.Hour)

	ec.data.Range(func(_, value any) bool {
		if ex, ok := value.(ExchangeDataType); ok {
			if ex.Timestamp.After(cutoff) {
				result = append(result, ex)
			}
		}
		return true
	})

	ec.metaLock.RLock()
	timestamp := ec.Timestamp
	var lastUpdated time.Time
	if ec.LastUpdated != nil {
		lastUpdated = *ec.LastUpdated
	}
	ec.metaLock.RUnlock()

	return ExchangeDataCacheSnapshot{
		Data:        result,
		Timestamp:   timestamp,
		LastUpdated: lastUpdated,
	}
}

func (ec *ExchangeDataCacheType) Hydrate(snapshot ExchangeDataCacheSnapshot) {
	ec.data = sync.Map{}
	cutoff := time.Now().Add(-24 * time.Hour)

	for _, ex := range snapshot.Data {
		if ex.Timestamp.After(cutoff) {
			ck := ec.CreateKeyFromExchangeData(&ex)
			ec.data.Store(ck, ex)
		}
	}

	ec.metaLock.Lock()
	ec.Timestamp = snapshot.Timestamp
	ec.LastUpdated = &snapshot.LastUpdated
	ec.metaLock.Unlock()
}

func (ec *ExchangeDataCacheType) CreateKeyFromExchangeData(ex *ExchangeDataType) string {
	return ec.CreateKeyFromInt(ex.SourceId, ex.TargetId)
}

func (ec *ExchangeDataCacheType) CreateKeyFromString(sid, tid string) string {
	return fmt.Sprintf("%s-%s", sid, tid)
}

func (ec *ExchangeDataCacheType) CreateKeyFromInt(sid, tid int64) string {
	return fmt.Sprintf("%d-%d", sid, tid)
}
