package types

import (
	"fmt"
	"sync"
	"time"
)

var ExchangeCache ExchangeDataCacheType = ExchangeDataCacheType{}

type ExchangeDataCacheType struct {
	data      sync.Map
	Timestamp time.Time
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

	return ec
}

func (ec *ExchangeDataCacheType) Remove(ck string) *ExchangeDataCacheType {
	ec.data.Delete(ck)
	ec.Timestamp = time.Now()

	return ec
}

func (ec *ExchangeDataCacheType) Reset() *ExchangeDataCacheType {
	ec.data = sync.Map{}
	ec.Timestamp = time.Now()

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

func (ec *ExchangeDataCacheType) CreateKeyFromExchangeData(ex *ExchangeDataType) string {
	return ec.CreateKeyFromInt(ex.SourceId, ex.TargetId)
}

func (ec *ExchangeDataCacheType) CreateKeyFromString(sid, tid string) string {
	return fmt.Sprintf("%s-%s", sid, tid)
}

func (ec *ExchangeDataCacheType) CreateKeyFromInt(sid, tid int64) string {
	return fmt.Sprintf("%d-%d", sid, tid)
}
