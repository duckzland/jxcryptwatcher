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
	ConfigSave()

	ConfigInit()

	if !UseConfig().IsValid() {
		t.Error("Expected config to be valid")
	}
	if UseConfig().Version != "1.6.0" {
		t.Errorf("Expected version to be 1.6.0 after load, got %s", UseConfig().Version)
	}
	configTurnOnLogs()
}

func TestConfigCheckFileCreatesDefault(t *testing.T) {
	configTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	ConfigInit()

	if !UseConfig().IsValid() {
		t.Error("Expected default config to be valid")
	}
	if UseConfig().Version != "1.6.0" {
		t.Errorf("Expected default version to be 1.6.0, got %s", UseConfig().Version)
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
		RSIEndpoint:       "https://custom-rsi",
	}
	ConfigSave()

	ConfigInit()

	if !UseConfig().IsValid() {
		t.Error("Expected config to be valid")
	}
	if UseConfig().Delay != 99 {
		t.Errorf("Expected delay 99 after reload, got %d", UseConfig().Delay)
	}
	if UseConfig().RSIEndpoint != "https://custom-rsi" {
		t.Errorf("Expected RSIEndpoint to remain unchanged, got %s", UseConfig().RSIEndpoint)
	}
	configTurnOnLogs()
}

func TestUpdateDefaultVersionPatch(t *testing.T) {
	configTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	configStorage = &configType{
		Version: "",
	}
	ConfigSave()

	ConfigInit()

	if UseConfig().Version != "1.6.0" {
		t.Errorf("Expected version to be updated to 1.6.0, got %s", UseConfig().Version)
	}
	if UseConfig().RSIEndpoint == "" {
		t.Error("Expected RSIEndpoint to be set by updateDefault")
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
		RSIEndpoint:       "x",
	}
	ConfigSave()
	ConfigInit()

	if !UseConfig().CanDoAltSeason() {
		t.Error("Expected CanDoAltSeason to be true")
	}
	if !UseConfig().CanDoFearGreed() {
		t.Error("Expected CanDoFearGreed to be true")
	}
	if !UseConfig().CanDoCMC100() {
		t.Error("Expected CanDoCMC100 to be true")
	}
	if !UseConfig().CanDoMarketCap() {
		t.Error("Expected CanDoMarketCap to be true")
	}
	if !UseConfig().CanDoRSI() {
		t.Error("Expected CanDoRSI to be true")
	}
	if !UseConfig().IsValidTickers() {
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
	ConfigSave()

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		ConfigSave()
	}()

	go func() {
		defer wg.Done()
		if !UseConfig().IsValid() {
			t.Error("Expected config to be valid during concurrent access")
		}
	}()

	wg.Wait()
	configTurnOnLogs()
}
