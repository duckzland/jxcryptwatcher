package main

import (
	"fmt"
)

var ExchangeCache ExchangeDataCache = ExchangeDataCache{}

type ExchangeDataCache struct {
	data map[string]*ExchangeDataType
}

func (ec *ExchangeDataCache) Get(ck string) *ExchangeDataType {
	if ec.Has(ck) {
		return ec.data[ck]
	}
	return nil
}

func (ec *ExchangeDataCache) Insert(ex *ExchangeDataType) *ExchangeDataCache {

	if ec.data == nil {
		ec.Reset()
	}

	ck := ec.CreateKeyFromExchangeData(ex)
	ec.data[ck] = ex

	return ec
}

func (ec *ExchangeDataCache) Remove(ck string) *ExchangeDataCache {
	if ec.Has(ck) {
		delete(ec.data, ck)
	}
	return ec
}

func (ec *ExchangeDataCache) Reset() *ExchangeDataCache {
	ec.data = make(map[string]*ExchangeDataType)
	return ec
}

func (ec *ExchangeDataCache) Has(ck string) bool {
	_, ok := ec.data[ck]
	return ok
}

func (ec *ExchangeDataCache) CreateKeyFromExchangeData(ex *ExchangeDataType) string {
	return fmt.Sprintf("%d-%d", ex.SourceId, ex.TargetId)
}

func (ec *ExchangeDataCache) CreateKeyFromString(sid, tid string) string {
	return fmt.Sprintf("%s-%s", sid, tid)
}

func (ec *ExchangeDataCache) CreateKeyFromInt(sid, tid int64) string {
	return fmt.Sprintf("%d-%d", sid, tid)
}
