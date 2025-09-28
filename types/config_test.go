package types

import (
	"log"
	"os"
	"sync"
	"testing"

	"fyne.io/fyne/v2/test"
)

type configNullWriter struct{}

func (n configNullWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func configTurnOffLogs() {
	log.SetOutput(configNullWriter{})
}

func configTurnOnLogs() {
	log.SetOutput(os.Stdout)
}

func TestConfigInitLoadAndValidate(t *testing.T) {
	configTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	configStorage = &configType{
		DataEndpoint:      "https://data",
		ExchangeEndpoint:  "https://exchange",
		AltSeasonEndpoint: "https://altseason",
		FearGreedEndpoint: "https://feargreed",
		CMC100Endpoint:    "https://cmc100",
		MarketCapEndpoint: "https://marketcap",
		Version:           "1.3.0",
	}
	configStorage.SaveFile()

	configStorage = &configType{}
	configStorage.loadFile()

	if !configStorage.IsValid() {
		t.Error("Expected config to be valid")
	}
	configTurnOnLogs()
}

func TestConfigCheckFileCreatesDefault(t *testing.T) {
	configTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	configStorage = &configType{}
	configStorage.CheckFile()
	configStorage.loadFile()

	if !configStorage.IsValid() {
		t.Error("Expected default config to be valid")
	}
	configTurnOnLogs()
}

func TestSaveFileAndReload(t *testing.T) {
	configTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	configStorage = &configType{
		DataEndpoint:      "https://data",
		ExchangeEndpoint:  "https://exchange",
		AltSeasonEndpoint: "https://altseason",
		FearGreedEndpoint: "https://feargreed",
		CMC100Endpoint:    "https://cmc100",
		MarketCapEndpoint: "https://marketcap",
		Version:           "1.5.0",
		Delay:             99,
	}
	configStorage.SaveFile()

	configStorage = &configType{}
	configStorage.loadFile()

	if !configStorage.IsValid() {
		t.Error("Expected config to be valid")
	}
	if configStorage.Delay != 99 {
		t.Errorf("Expected delay 99 after reload, got %d", configStorage.Delay)
	}
	configTurnOnLogs()
}

func TestCapabilityChecks(t *testing.T) {
	configTurnOffLogs()
	configStorage = &configType{
		AltSeasonEndpoint: "x",
		FearGreedEndpoint: "x",
		CMC100Endpoint:    "x",
		MarketCapEndpoint: "x",
	}

	if !configStorage.CanDoAltSeason() {
		t.Error("Expected CanDoAltSeason to be true")
	}
	if !configStorage.CanDoFearGreed() {
		t.Error("Expected CanDoFearGreed to be true")
	}
	if !configStorage.CanDoCMC100() {
		t.Error("Expected CanDoCMC100 to be true")
	}
	if !configStorage.CanDoMarketCap() {
		t.Error("Expected CanDoMarketCap to be true")
	}
	if !configStorage.IsValidTickers() {
		t.Error("Expected IsValidTickers to be true")
	}
	configTurnOnLogs()
}

func TestConcurrentAccess(t *testing.T) {
	configTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	configStorage = &configType{
		DataEndpoint:     "x",
		ExchangeEndpoint: "x",
	}

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		configStorage.SaveFile()
	}()

	go func() {
		defer wg.Done()
		if !configStorage.IsValid() {
			t.Error("Expected config to be valid during concurrent access")
		}
	}()

	wg.Wait()
	configTurnOnLogs()
}
