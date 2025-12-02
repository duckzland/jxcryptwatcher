package types

import (
	"log"
	"os"
	"testing"

	"fyne.io/fyne/v2/test"
)

type cryptosMapNullWriter struct{}

func (cryptosMapNullWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func cryptosMapTurnOffLogs() {
	log.SetOutput(cryptosMapNullWriter{})
}

func cryptosMapTurnOnLogs() {
	log.SetOutput(os.Stdout)
}

func TestCryptosMapInsertAndRetrieve(t *testing.T) {
	cryptosMapTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	cm := NewCryptosMap()
	cm.Init()
	cm.Insert("1", "1|BTC - Bitcoin")

	if cm.GetDisplayById("1") != "1|BTC - Bitcoin" {
		t.Error("Failed to retrieve inserted display value")
	}
	if cm.GetSymbolById("1") != "BTC" {
		t.Error("Failed to extract symbol from ID")
	}
	if cm.GetSymbolByDisplay("1|BTC - Bitcoin") != "BTC" {
		t.Error("Failed to extract symbol from display")
	}
	if cm.GetIdByDisplay("1|BTC - Bitcoin") != "1" {
		t.Error("Failed to extract ID from display")
	}
	if !cm.ValidateId(1) {
		t.Error("Expected ID 1 to be valid")
	}
	cryptosMapTurnOnLogs()
}
func TestCryptosMapHydrateAndOptions(t *testing.T) {
	cryptosMapTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	cm := NewCryptosMap()
	data := map[string]string{
		"1": "1|BTC - Bitcoin",
		"2": "2|ETH - Ethereum",
	}
	cm.Hydrate(cryptosMapCache{Data: data})

	opts := cm.GetOptions()
	if len(opts) != 2 {
		t.Errorf("Expected 2 options, got %d", len(opts))
	}

	search := cm.GetSearchMap()
	expected := map[string]bool{
		"1|btc - bitcoin":  true,
		"2|eth - ethereum": true,
	}
	for _, val := range search {
		if !expected[val] {
			t.Errorf("Unexpected value in search map: %s", val)
		}
	}
	cryptosMapTurnOnLogs()
}

func TestCryptosMapSerialize(t *testing.T) {
	cryptosMapTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	cm := NewCryptosMap()
	cm.Init()
	cm.Insert("1", "1|BTC - Bitcoin")
	cm.Insert("2", "2|ETH - Ethereum")
	_ = cm.GetOptions()

	cache := cm.Serialize()
	if len(cache.Data) != 2 {
		t.Errorf("Expected 2 entries in cache, got %d", len(cache.Data))
	}
	if len(cache.Maps) != 2 {
		t.Errorf("Expected 2 maps in cache, got %d", len(cache.Maps))
	}
	if len(cache.SearchMaps) != 2 {
		t.Errorf("Expected 2 search maps in cache, got %d", len(cache.SearchMaps))
	}

	// Round-trip test: hydrate from cache and verify
	cm2 := NewCryptosMap()
	cm2.Hydrate(cache)

	if len(cm2.GetOptions()) != 2 {
		t.Errorf("Expected 2 options after hydration, got %d", len(cm2.GetOptions()))
	}
	search := cm2.GetSearchMap()
	expected := map[string]bool{
		"1|btc - bitcoin":  true,
		"2|eth - ethereum": true,
	}
	for _, val := range search {
		if !expected[val] {
			t.Errorf("Unexpected value in hydrated search map: %s", val)
		}
	}
	cryptosMapTurnOnLogs()
}

func TestCryptosMapClearAndEmpty(t *testing.T) {
	cryptosMapTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	cm := NewCryptosMap()
	cm.Init()
	cm.Insert("1", "1|BTC - Bitcoin")
	if cm.IsEmpty() {
		t.Error("Expected map to be non-empty")
	}
	cm.ClearMapCache()
	if len(cm.GetOptions()) != 1 {
		t.Error("Expected options to regenerate after clear")
	}
	cm.Init()
	if !cm.IsEmpty() {
		t.Error("Expected map to be empty after Init")
	}
	cryptosMapTurnOnLogs()
}
