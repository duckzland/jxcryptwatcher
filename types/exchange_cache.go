package types

import (
	"fmt"
	"sync"
	"time"

	JC "jxwatcher/core"
)

const exchangeCacheUpdateThreshold = 10 * time.Second
const ExchangeRate = "rates"
const ExchangeRefresh = "refresh_rates"
const ExchangeUpdateRates = "update_rates"

var exchangeCacheStorage *exchangeDataCacheType = nil

type exchangeDataCacheSnapshot struct {
	Data        []exchangeDataSnapshot `json:"data"`
	Timestamp   time.Time              `json:"timestamp"`
	LastUpdated time.Time              `json:"last_updated"`
}

type exchangeDataSnapshot struct {
	SourceSymbol string    `json:"source_symbol"`
	SourceId     int64     `json:"source_id"`
	SourceAmount float64   `json:"source_amount"`
	TargetSymbol string    `json:"target_symbol"`
	TargetId     int64     `json:"target_id"`
	TargetAmount string    `json:"target_amount"`
	Timestamp    time.Time `json:"timestamp"`
}

type exchangeDataCacheType struct {
	data        sync.Map
	timestamp   time.Time
	lastUpdated *time.Time
	mu          sync.RWMutex
}

func (ec *exchangeDataCacheType) Init() *exchangeDataCacheType {
	ec.data = sync.Map{}
	ec.SetTimestamp(time.Now())
	ec.SetLastUpdated(nil)
	return ec
}

func (ec *exchangeDataCacheType) GetTimestamp() time.Time {
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	return ec.timestamp
}

func (ec *exchangeDataCacheType) SetTimestamp(t time.Time) {
	ec.mu.Lock()
	ec.timestamp = t
	ec.mu.Unlock()
}

func (ec *exchangeDataCacheType) GetLastUpdated() *time.Time {
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	return ec.lastUpdated
}

func (ec *exchangeDataCacheType) SetLastUpdated(t *time.Time) {
	ec.mu.Lock()
	ec.lastUpdated = t
	ec.mu.Unlock()
}

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
	return time.Now().After(last.Add(exchangeCacheUpdateThreshold))
}

func (ec *exchangeDataCacheType) Serialize() exchangeDataCacheSnapshot {
	var result []exchangeDataSnapshot
	cutoff := time.Now().Add(-24 * time.Hour)

	ec.data.Range(func(_, value any) bool {
		if ex, ok := value.(exchangeDataType); ok {
			if ex.Timestamp.After(cutoff) && ex.TargetAmount != nil {
				raw := ex.TargetAmount.Text('g', -1)

				if raw == "NaN" ||
					raw == "0" ||
					ex.SourceAmount == 0 ||
					ex.SourceId == 0 ||
					ex.TargetId == 0 ||
					ex.SourceSymbol == "" ||
					ex.TargetSymbol == "" {
					return true
				}

				result = append(result, exchangeDataSnapshot{
					SourceSymbol: ex.SourceSymbol,
					SourceId:     ex.SourceId,
					SourceAmount: ex.SourceAmount,
					TargetSymbol: ex.TargetSymbol,
					TargetId:     ex.TargetId,
					TargetAmount: raw,
					Timestamp:    ex.Timestamp,
				})
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

	for _, snap := range snapshot.Data {
		if snap.Timestamp.After(cutoff) {
			f, ok := JC.ToBigString(snap.TargetAmount)
			if !ok {
				f = JC.ToBigFloat(0)
			}

			ex := exchangeDataType{
				SourceSymbol: snap.SourceSymbol,
				SourceId:     snap.SourceId,
				SourceAmount: snap.SourceAmount,
				TargetSymbol: snap.TargetSymbol,
				TargetId:     snap.TargetId,
				TargetAmount: f,
				Timestamp:    snap.Timestamp,
			}

			ck := ec.CreateKeyFromExchangeData(&ex)
			ec.data.Store(ck, ex)
		}
	}

	ec.SetTimestamp(snapshot.Timestamp)
	ec.SetLastUpdated(&snapshot.LastUpdated)
}

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

func RegisterExchangeCache() *exchangeDataCacheType {
	if exchangeCacheStorage == nil {
		exchangeCacheStorage = &exchangeDataCacheType{}
	}
	return exchangeCacheStorage
}

func UseExchangeCache() *exchangeDataCacheType {
	return exchangeCacheStorage
}
