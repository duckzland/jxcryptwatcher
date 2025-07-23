package types

import (
	"fmt"
)

var ExchangeCache ExchangeDataCacheType = ExchangeDataCacheType{}

type ExchangeDataCacheType struct {
	data map[string]ExchangeDataType
}

func (ec *ExchangeDataCacheType) Get(ck string) *ExchangeDataType {
	if ec.Has(ck) {
		ex := ec.data[ck]
		return &ex
	}
	return nil
}

func (ec *ExchangeDataCacheType) Insert(ex *ExchangeDataType) *ExchangeDataCacheType {

	if ec.data == nil {
		ec.Reset()
	}

	ck := ec.CreateKeyFromExchangeData(ex)
	ec.data[ck] = *ex

	return ec
}

func (ec *ExchangeDataCacheType) Remove(ck string) *ExchangeDataCacheType {
	if ec.Has(ck) {
		delete(ec.data, ck)
	}
	return ec
}

func (ec *ExchangeDataCacheType) Reset() *ExchangeDataCacheType {
	ec.data = make(map[string]ExchangeDataType)
	return ec
}

func (ec *ExchangeDataCacheType) Has(ck string) bool {
	_, ok := ec.data[ck]
	return ok
}

func (ec *ExchangeDataCacheType) CreateKeyFromExchangeData(ex *ExchangeDataType) string {
	return fmt.Sprintf("%d-%d", ex.SourceId, ex.TargetId)
}

func (ec *ExchangeDataCacheType) CreateKeyFromString(sid, tid string) string {
	return fmt.Sprintf("%s-%s", sid, tid)
}

func (ec *ExchangeDataCacheType) CreateKeyFromInt(sid, tid int64) string {
	return fmt.Sprintf("%d-%d", sid, tid)
}
