package types

import (
	"log"
	"os"
	"testing"
	"time"

	JC "jxwatcher/core"

	"fyne.io/fyne/v2/test"
)

type exchangeCacheNullWriter struct{}

func (exchangeCacheNullWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func exchangeCacheTurnOffLogs() {
	log.SetOutput(exchangeCacheNullWriter{})
}

func exchangeCacheTurnOnLogs() {
	log.SetOutput(os.Stdout)
}

func TestExchangeCacheInsertAndGet(t *testing.T) {
	exchangeCacheTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	cache := (&exchangeDataCacheType{}).Init()
	now := time.Now()

	ex := &exchangeDataType{
		SourceSymbol: "BTC",
		SourceId:     1,
		SourceAmount: 1.0,
		TargetSymbol: "ETH",
		TargetId:     2,
		TargetAmount: JC.ToBigFloat(15.5),
		Timestamp:    now,
	}

	cache.Insert(ex)
	key := cache.CreateKeyFromExchangeData(ex)
	got := cache.Get(key)

	if got == nil || got.TargetSymbol != "ETH" {
		t.Error("Failed to retrieve inserted exchange data")
	}
	exchangeCacheTurnOnLogs()
}

func TestExchangeCacheSerializeAndHydrate(t *testing.T) {
	exchangeCacheTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	cache := (&exchangeDataCacheType{}).Init()
	now := time.Now()

	ex := &exchangeDataType{
		SourceSymbol: "BTC",
		SourceId:     1,
		SourceAmount: 1.0,
		TargetSymbol: "ETH",
		TargetId:     2,
		TargetAmount: JC.ToBigFloat(15.5),
		Timestamp:    now,
	}
	cache.Insert(ex)

	snapshot := cache.Serialize()
	newCache := (&exchangeDataCacheType{}).Init()
	newCache.Hydrate(snapshot)

	key := newCache.CreateKeyFromExchangeData(ex)
	got := newCache.Get(key)

	if got == nil || got.TargetSymbol != "ETH" {
		t.Error("Failed to hydrate exchange data from snapshot")
	}
	exchangeCacheTurnOnLogs()
}

func TestExchangeCacheSoftResetAndReset(t *testing.T) {
	exchangeCacheTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	cache := (&exchangeDataCacheType{}).Init()
	cache.Insert(&exchangeDataType{
		SourceSymbol: "BTC",
		SourceId:     1,
		SourceAmount: 1.0,
		TargetSymbol: "ETH",
		TargetId:     2,
		TargetAmount: JC.ToBigFloat(15.5),
		Timestamp:    time.Now(),
	})

	cache.SoftReset()
	if cache.GetLastUpdated() != nil {
		t.Error("Expected lastUpdated to be nil after SoftReset")
	}

	cache.Reset()
	if cache.HasData() {
		t.Error("Expected cache to be empty after Reset")
	}
	exchangeCacheTurnOnLogs()
}

func TestExchangeCacheShouldRefresh(t *testing.T) {
	exchangeCacheTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	cache := (&exchangeDataCacheType{}).Init()
	if !cache.ShouldRefresh() {
		t.Error("Expected ShouldRefresh to be true when lastUpdated is nil")
	}

	past := time.Now().Add(-30 * time.Second)
	cache.SetLastUpdated(&past)
	if !cache.ShouldRefresh() {
		t.Error("Expected ShouldRefresh to be true when lastUpdated is stale")
	}

	recent := time.Now()
	cache.SetLastUpdated(&recent)
	if cache.ShouldRefresh() {
		t.Error("Expected ShouldRefresh to be false when lastUpdated is recent")
	}
	exchangeCacheTurnOnLogs()
}
