package types

import (
	"log"
	"os"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"

	JC "jxwatcher/core"
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

	cache := &exchangeDataCacheType{}
	cache.Init()
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

	updates := cache.GetRecentUpdates()
	if len(updates) != 1 {
		t.Errorf("Expected 1 recent update after first insert, got %d", len(updates))
	}
	if _, ok := updates[key]; !ok {
		t.Error("Expected recent updates to contain inserted key")
	}

	cache.Insert(ex)
	updates = cache.GetRecentUpdates()
	if len(updates) != 0 {
		t.Errorf("Expected 0 recent updates after identical insert, got %d", len(updates))
	}

	ex2 := &exchangeDataType{
		SourceSymbol: "BTC",
		SourceId:     1,
		SourceAmount: 1.0,
		TargetSymbol: "ETH",
		TargetId:     2,
		TargetAmount: JC.ToBigFloat(20.0),
		Timestamp:    now.Add(1 * time.Second),
	}
	cache.Insert(ex2)
	updates = cache.GetRecentUpdates()
	if len(updates) != 1 {
		t.Errorf("Expected 1 recent update after changed insert, got %d", len(updates))
	}
	if updates[key].Cmp(JC.ToBigFloat(20.0)) != 0 {
		t.Error("Expected recent update to contain new TargetAmount = 20.0")
	}

	exchangeCacheTurnOnLogs()
}

func TestExchangeCacheSerializeAndHydrate(t *testing.T) {
	exchangeCacheTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	cache := &exchangeDataCacheType{}
	cache.Init()
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
	newCache := &exchangeDataCacheType{}
	newCache.Init()
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

	cache := &exchangeDataCacheType{}
	cache.Init()
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
	if cache.IsUpdatedAt() != nil {
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

	cache := &exchangeDataCacheType{}
	cache.Init()
	if !cache.ShouldRefresh() {
		t.Error("Expected ShouldRefresh to be true when lastUpdated is nil")
	}

	past := time.Now().Add(-30 * time.Second)
	cache.UpdatedAt(&past)
	if !cache.ShouldRefresh() {
		t.Error("Expected ShouldRefresh to be true when lastUpdated is stale")
	}

	recent := time.Now()
	cache.UpdatedAt(&recent)
	if cache.ShouldRefresh() {
		t.Error("Expected ShouldRefresh to be false when lastUpdated is recent")
	}
	exchangeCacheTurnOnLogs()
}

func TestExchangeCacheRecentUpdates(t *testing.T) {
	exchangeCacheTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	cache := &exchangeDataCacheType{}
	cache.Init()
	now := time.Now()

	ex1 := &exchangeDataType{
		SourceSymbol: "BTC",
		SourceId:     1,
		SourceAmount: 1.0,
		TargetSymbol: "ETH",
		TargetId:     2,
		TargetAmount: JC.ToBigFloat(15.5),
		Timestamp:    now,
	}
	cache.Insert(ex1)

	updates := cache.GetRecentUpdates()
	if len(updates) != 1 {
		t.Errorf("Expected 1 recent update after first insert, got %d", len(updates))
	}

	ex2 := &exchangeDataType{
		SourceSymbol: "BTC",
		SourceId:     1,
		SourceAmount: 1.0,
		TargetSymbol: "ETH",
		TargetId:     2,
		TargetAmount: JC.ToBigFloat(15.5),
		Timestamp:    now.Add(1 * time.Second),
	}
	cache.Insert(ex2)

	updates = cache.GetRecentUpdates()
	if len(updates) != 0 {
		t.Errorf("Expected 0 recent updates after inserting identical data, got %d", len(updates))
	}

	ex3 := &exchangeDataType{
		SourceSymbol: "BTC",
		SourceId:     1,
		SourceAmount: 1.0,
		TargetSymbol: "ETH",
		TargetId:     2,
		TargetAmount: JC.ToBigFloat(20.0), // changed value
		Timestamp:    now.Add(2 * time.Second),
	}
	cache.Insert(ex3)

	updates = cache.GetRecentUpdates()
	if len(updates) != 1 {
		t.Errorf("Expected 1 recent update after changed insert, got %d", len(updates))
	}
	if updates[cache.CreateKeyFromExchangeData(ex3)].Cmp(JC.ToBigFloat(20.0)) != 0 {
		t.Error("Expected recent update to contain new TargetAmount = 20.0")
	}

	exchangeCacheTurnOnLogs()
}
