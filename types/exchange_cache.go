package types

import (
	"fmt"
	"sync"
	"time"
)

var ExchangeCache exchangeDataCacheType = exchangeDataCacheType{}
var CMCUpdateThreshold = 1 * time.Minute

type exchangeDataCacheSnapshot struct {
	Data        []exchangeDataType `json:"data"`
	Timestamp   time.Time          `json:"timestamp"`
	LastUpdated time.Time          `json:"last_updated"`
}

type exchangeDataCacheType struct {
	data        sync.Map
	timestamp   time.Time
	lastUpdated *time.Time
	metaLock    sync.RWMutex
}

// Initialization
func (ec *exchangeDataCacheType) Init() *exchangeDataCacheType {
	ec.data = sync.Map{}
	ec.SetTimestamp(time.Now())
	ec.SetLastUpdated(nil)
	return ec
}

// Getters and Setters
func (ec *exchangeDataCacheType) GetTimestamp() time.Time {
	ec.metaLock.RLock()
	defer ec.metaLock.RUnlock()
	return ec.timestamp
}

func (ec *exchangeDataCacheType) SetTimestamp(t time.Time) {
	ec.metaLock.Lock()
	ec.timestamp = t
	ec.metaLock.Unlock()
}

func (ec *exchangeDataCacheType) GetLastUpdated() *time.Time {
	ec.metaLock.RLock()
	defer ec.metaLock.RUnlock()
	return ec.lastUpdated
}

func (ec *exchangeDataCacheType) SetLastUpdated(t *time.Time) {
	ec.metaLock.Lock()
	ec.lastUpdated = t
	ec.metaLock.Unlock()
}

// Core operations
func (ec *exchangeDataCacheType) Get(ck string) *exchangeDataType {
	if val, ok := ec.data.Load(ck); ok {
		ex := val.(exchangeDataType)
		return &ex
	}
	return nil
}

func (ec *exchangeDataCacheType) Insert(ex *exchangeDataType) *exchangeDataCacheType {
	ck := ec.CreateKeyFromExchangeData(ex)
	ec.data.Store(ck, *ex)
	ec.SetTimestamp(time.Now())
	ec.SetLastUpdated(&ex.Timestamp)
	return ec
}

func (ec *exchangeDataCacheType) Remove(ck string) *exchangeDataCacheType {
	ec.data.Delete(ck)
	ec.SetTimestamp(time.Now())
	return ec
}

func (ec *exchangeDataCacheType) SoftReset() *exchangeDataCacheType {
	ec.SetTimestamp(time.Now())
	ec.SetLastUpdated(nil)
	return ec
}

func (ec *exchangeDataCacheType) Reset() *exchangeDataCacheType {
	ec.data = sync.Map{}
	ec.SetTimestamp(time.Now())
	ec.SetLastUpdated(nil)
	return ec
}

// Status checks
func (ec *exchangeDataCacheType) Has(ck string) bool {
	d, ok := ec.data.Load(ck)
	return ok && d != nil
}

func (ec *exchangeDataCacheType) HasData() bool {
	isEmpty := true
	ec.data.Range(func(_, _ interface{}) bool {
		isEmpty = false
		return false
	})
	return !isEmpty
}

func (ec *exchangeDataCacheType) IsEmpty() bool {
	return !ec.HasData()
}

func (ec *exchangeDataCacheType) ShouldRefresh() bool {
	last := ec.GetLastUpdated()
	if last == nil {
		return true
	}
	return time.Now().After(last.Add(CMCUpdateThreshold))
}

// Serialization
func (ec *exchangeDataCacheType) Serialize() exchangeDataCacheSnapshot {
	var result []exchangeDataType
	cutoff := time.Now().Add(-24 * time.Hour)

	ec.data.Range(func(_, value any) bool {
		if ex, ok := value.(exchangeDataType); ok {
			if ex.Timestamp.After(cutoff) {
				result = append(result, ex)
			}
		}
		return true
	})

	timestamp := ec.GetTimestamp()
	last := ec.GetLastUpdated()
	var lastUpdated time.Time
	if last != nil {
		lastUpdated = *last
	}

	return exchangeDataCacheSnapshot{
		Data:        result,
		Timestamp:   timestamp,
		LastUpdated: lastUpdated,
	}
}

func (ec *exchangeDataCacheType) Hydrate(snapshot exchangeDataCacheSnapshot) {
	ec.data = sync.Map{}
	cutoff := time.Now().Add(-24 * time.Hour)

	for _, ex := range snapshot.Data {
		if ex.Timestamp.After(cutoff) {
			ck := ec.CreateKeyFromExchangeData(&ex)
			ec.data.Store(ck, ex)
		}
	}

	ec.SetTimestamp(snapshot.Timestamp)
	ec.SetLastUpdated(&snapshot.LastUpdated)
}

// Key generation
func (ec *exchangeDataCacheType) CreateKeyFromExchangeData(ex *exchangeDataType) string {
	return ec.CreateKeyFromInt(ex.SourceId, ex.TargetId)
}

func (ec *exchangeDataCacheType) CreateKeyFromString(sid, tid string) string {
	return fmt.Sprintf("%s-%s", sid, tid)
}

func (ec *exchangeDataCacheType) CreateKeyFromInt(sid, tid int64) string {
	return fmt.Sprintf("%d-%d", sid, tid)
}

func NewExchangeDataCacheSnapshot() *exchangeDataCacheSnapshot {
	return &exchangeDataCacheSnapshot{}
}
