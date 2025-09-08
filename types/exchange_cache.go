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
}

func (ec *ExchangeDataCacheType) Init() *ExchangeDataCacheType {
	ec.data = sync.Map{}
	ec.Timestamp = time.Now()
	ec.LastUpdated = nil

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
	ec.Timestamp = time.Now()
	ec.LastUpdated = &ex.Timestamp

	return ec
}

func (ec *ExchangeDataCacheType) Remove(ck string) *ExchangeDataCacheType {
	ec.data.Delete(ck)
	ec.Timestamp = time.Now()

	return ec
}

func (ec *ExchangeDataCacheType) SoftReset() *ExchangeDataCacheType {
	ec.Timestamp = time.Now()
	ec.LastUpdated = nil
	return ec
}

func (ec *ExchangeDataCacheType) Reset() *ExchangeDataCacheType {
	ec.data = sync.Map{}
	ec.Timestamp = time.Now()
	ec.LastUpdated = nil

	return ec
}

func (ec *ExchangeDataCacheType) Has(ck string) bool {
	d, ok := ec.data.Load(ck)

	return ok && d != nil
}

func (ec *ExchangeDataCacheType) HasData() bool {

	isEmpty := true
	ec.data.Range(func(key, value interface{}) bool {
		isEmpty = false
		return false
	})

	return !isEmpty
}

func (ec *ExchangeDataCacheType) ShouldRefresh() bool {
	// return true
	if ec.LastUpdated == nil {
		return true
	}

	return time.Now().After(ec.LastUpdated.Add(CMCUpdateThreshold))
}

func (ec *ExchangeDataCacheType) Serialize() ExchangeDataCacheSnapshot {
	var result []ExchangeDataType
	ec.data.Range(func(key, value any) bool {
		if ex, ok := value.(ExchangeDataType); ok {
			result = append(result, ex)
		}
		return true
	})

	var lastUpdated time.Time
	if ec.LastUpdated != nil {
		lastUpdated = *ec.LastUpdated
	}

	return ExchangeDataCacheSnapshot{
		Data:        result,
		Timestamp:   ec.Timestamp,
		LastUpdated: lastUpdated,
	}
}

func (ec *ExchangeDataCacheType) Hydrate(snapshot ExchangeDataCacheSnapshot) {
	ec.data = sync.Map{}
	for _, ex := range snapshot.Data {
		ck := ec.CreateKeyFromExchangeData(&ex)
		ec.data.Store(ck, ex)
	}
	ec.Timestamp = snapshot.Timestamp
	ec.LastUpdated = &snapshot.LastUpdated
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
