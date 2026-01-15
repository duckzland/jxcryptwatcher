package types

import (
	"fmt"
	"math/big"
	"time"

	JC "jxwatcher/core"
)

const exchangeCacheUpdateThreshold = 10 * time.Second

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
	JC.Database
}

func (ec *exchangeDataCacheType) Init() *exchangeDataCacheType {
	ec.Reset()

	return ec
}

func (ec *exchangeDataCacheType) GetRecentUpdates() map[string]*big.Float {
	updates := make(map[string]*big.Float)

	ec.RangeRecentUpdates(func(key, value any) bool {
		k, ok1 := key.(string)
		v, ok2 := value.(*big.Float)
		if !ok1 || !ok2 || v == nil {
			ec.DeleteRecentUpdates(key)
			return true
		}

		// copy value to avoid sharing mutable big.Float
		updates[k] = new(big.Float).Copy(v)
		ec.DeleteRecentUpdates(key)
		return true
	})

	return updates
}

func (ec *exchangeDataCacheType) Get(ck string) *exchangeDataType {
	if val, ok := ec.Load(ck); ok {
		ex := val.(exchangeDataType)
		return &ex
	}
	return nil
}

func (ec *exchangeDataCacheType) Insert(ex *exchangeDataType) *exchangeDataCacheType {
	ck := ec.CreateKeyFromExchangeData(ex)

	if oldVal, ok := ec.Load(ck); ok {
		old := oldVal.(exchangeDataType)
		if old.SourceAmount != ex.SourceAmount || old.TargetAmount.Cmp(ex.TargetAmount) != 0 {
			ec.StoreRecentUpdates(ck, ex.TargetAmount)
		}
	} else {
		ec.StoreRecentUpdates(ck, ex.TargetAmount)
	}

	ec.Store(ck, *ex)
	now := time.Now()
	ec.SetTimestamp(now)
	ec.SetLastUpdated(&ex.Timestamp)

	return ec
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

	ec.Range(func(_, value any) bool {
		if ex, ok := value.(exchangeDataType); ok {
			if ex.Timestamp.After(cutoff) && ex.TargetAmount != nil {
				raw := ex.TargetAmount.Text('g', -1)

				if raw == "NaN" ||
					raw == "0" ||
					ex.SourceAmount == 0 ||
					ex.SourceId == 0 ||
					ex.TargetId == 0 ||
					ex.SourceSymbol == JC.STRING_EMPTY ||
					ex.TargetSymbol == JC.STRING_EMPTY {
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

	ec.Reset()

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
			ec.Store(ck, ex)
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
		exchangeCacheStorage = (&exchangeDataCacheType{}).Init()
	}
	return exchangeCacheStorage
}

func UseExchangeCache() *exchangeDataCacheType {
	return exchangeCacheStorage
}
