package types

import (
	"log"
	"os"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
)

type tickerNullWriter struct{}

func (tickerNullWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func tickerTurnOffLogs() {
	log.SetOutput(tickerNullWriter{})
}

func tickerTurnOnLogs() {
	log.SetOutput(os.Stdout)
}

func TestTickerCacheInsertAndGet(t *testing.T) {
	tickerTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	tc := RegisterTickerCache().Init()
	now := time.Now()

	tc.Insert("BTC", "42000", now)
	val := tc.Get("BTC")
	if val != "42000" {
		t.Errorf("Expected value '42000', got '%s'", val)
	}
	if !tc.Has("BTC") {
		t.Error("Expected key 'BTC' to exist")
	}

	updates := tc.GetRecentUpdates()
	if len(updates) != 1 {
		t.Errorf("Expected 1 recent update after first insert, got %d", len(updates))
	}
	if updates["BTC"] != "42000" {
		t.Errorf("Expected recent update value '42000', got '%s'", updates["BTC"])
	}

	tc.Insert("BTC", "42000", time.Now())
	updates = tc.GetRecentUpdates()
	if len(updates) != 0 {
		t.Errorf("Expected 0 recent updates after identical insert, got %d", len(updates))
	}

	tc.Insert("BTC", "43000", time.Now())
	updates = tc.GetRecentUpdates()
	if len(updates) != 1 {
		t.Errorf("Expected 1 recent update after changed insert, got %d", len(updates))
	}
	if updates["BTC"] != "43000" {
		t.Errorf("Expected recent update value '43000', got '%s'", updates["BTC"])
	}

	tickerTurnOnLogs()
}

func TestTickerCacheResetAndSoftReset(t *testing.T) {
	tickerTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	tc := RegisterTickerCache().Init()
	tc.Insert("ETH", "3100", time.Now())

	tc.SoftReset()
	if tc.IsUpdatedAt() != nil {
		t.Error("Expected lastUpdated to be nil after SoftReset")
	}

	tc.Insert("ETH", "3200", time.Now())
	tc.Reset()
	if tc.Has("ETH") {
		t.Error("Expected key 'ETH' to be removed after Reset")
	}
	tickerTurnOnLogs()
}

func TestTickerCacheSerializeAndHydrate(t *testing.T) {
	tickerTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	tc := RegisterTickerCache().Init()
	now := time.Now()
	tc.Insert("SOL", "19.5", now)

	snapshot := tc.Serialize()
	if len(snapshot.Data) != 1 {
		t.Errorf("Expected 1 entry in snapshot, got %d", len(snapshot.Data))
	}

	newCache := (&tickerDataCacheType{}).Init()
	newCache.Hydrate(snapshot)
	if newCache.Get("SOL") != "19.5" {
		t.Error("Hydration failed to restore value")
	}
	tickerTurnOnLogs()
}

func TestTickerCacheGetRecentUpdates(t *testing.T) {
	tickerTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	tc := RegisterTickerCache().Init()
	now := time.Now()

	tc.Insert("BTC", "42000", now)
	updates := tc.GetRecentUpdates()
	if len(updates) != 1 {
		t.Errorf("Expected 1 recent update after first insert, got %d", len(updates))
	}
	if updates["BTC"] != "42000" {
		t.Errorf("Expected recent update value '42000', got '%s'", updates["BTC"])
	}

	updates = tc.GetRecentUpdates()
	if len(updates) != 0 {
		t.Errorf("Expected 0 recent updates after retrieval, got %d", len(updates))
	}

	tc.Insert("BTC", "42000", time.Now())
	updates = tc.GetRecentUpdates()
	if len(updates) != 0 {
		t.Errorf("Expected 0 recent updates after identical insert, got %d", len(updates))
	}

	tc.Insert("BTC", "43000", time.Now())
	updates = tc.GetRecentUpdates()
	if len(updates) != 1 {
		t.Errorf("Expected 1 recent update after changed insert, got %d", len(updates))
	}
	if updates["BTC"] != "43000" {
		t.Errorf("Expected recent update value '43000', got '%s'", updates["BTC"])
	}

	tickerTurnOnLogs()
}
