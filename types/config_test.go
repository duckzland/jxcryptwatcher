package types

import (
	"log"
	"os"
	"sync"
	"testing"

	"fyne.io/fyne/v2/test"

	JC "jxwatcher/core"
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

	// Fully populated config with all fields
	configStorage = &configType{
		DataEndpoint:      "https://data",
		ExchangeEndpoint:  "https://exchange",
		AltSeasonEndpoint: "https://altseason",
		FearGreedEndpoint: "https://feargreed",
		CMC100Endpoint:    "https://cmc100",
		MarketCapEndpoint: "https://marketcap",
		RSIEndpoint:       "https://rsi",
		ETFEndpoint:       "https://etf",
		DominanceEndpoint: "https://dominance",
		Delay:             60,
		Version:           "1.3.0",
	}
	ConfigSave()

	// Load and patch
	ConfigInit()
	cfg := UseConfig()

	// Validate version bump
	if cfg.Version != "1.7.0" {
		t.Errorf("Expected version to be 1.7.0 after load, got %s", cfg.Version)
	}

	// Validate all fields are preserved
	if cfg.DataEndpoint != "https://data" {
		t.Error("DataEndpoint was overwritten")
	}
	if cfg.ExchangeEndpoint != "https://exchange" {
		t.Error("ExchangeEndpoint was overwritten")
	}
	if cfg.AltSeasonEndpoint != "https://altseason" {
		t.Error("AltSeasonEndpoint was overwritten")
	}
	if cfg.FearGreedEndpoint != "https://feargreed" {
		t.Error("FearGreedEndpoint was overwritten")
	}
	if cfg.CMC100Endpoint != "https://cmc100" {
		t.Error("CMC100Endpoint was overwritten")
	}
	if cfg.MarketCapEndpoint != "https://marketcap" {
		t.Error("MarketCapEndpoint was overwritten")
	}
	if cfg.RSIEndpoint != "https://rsi" {
		t.Error("RSIEndpoint was overwritten")
	}
	if cfg.ETFEndpoint != "https://etf" {
		t.Error("ETFEndpoint was overwritten")
	}
	if cfg.DominanceEndpoint != "https://dominance" {
		t.Error("DominanceEndpoint was overwritten")
	}
	if cfg.Delay != 60 {
		t.Errorf("Expected Delay to be 60, got %d", cfg.Delay)
	}

	configTurnOnLogs()
}

func TestConfigCheckFileCreatesDefault(t *testing.T) {
	configTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	// Clear config and force default creation
	configStorage = nil
	ConfigInit()
	cfg := UseConfig()

	if cfg.Version != "1.7.0" {
		t.Errorf("Expected default version to be 1.7.0, got %s", cfg.Version)
	}
	if !cfg.IsValid() {
		t.Error("Expected default config to be valid")
	}
	if cfg.RSIEndpoint == JC.STRING_EMPTY || cfg.ETFEndpoint == JC.STRING_EMPTY || cfg.DominanceEndpoint == JC.STRING_EMPTY {
		t.Error("Expected new endpoints to be present in default config")
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
		ETFEndpoint:       "https://custom-etf",
		DominanceEndpoint: "https://custom-dominance",
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
	if UseConfig().ETFEndpoint != "https://custom-etf" {
		t.Errorf("Expected ETFEndpoint to remain unchanged, got %s", UseConfig().ETFEndpoint)
	}
	if UseConfig().DominanceEndpoint != "https://custom-dominance" {
		t.Errorf("Expected DominanceEndpoint to remain unchanged, got %s", UseConfig().DominanceEndpoint)
	}
	configTurnOnLogs()
}

func TestUpdateDefaultVersionPatch(t *testing.T) {
	configTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	configStorage = &configType{
		Version: JC.STRING_EMPTY,
	}
	ConfigSave()

	ConfigInit()

	if UseConfig().Version != "1.7.0" {
		t.Errorf("Expected version to be updated to 1.7.0, got %s", UseConfig().Version)
	}
	if UseConfig().RSIEndpoint == JC.STRING_EMPTY || UseConfig().ETFEndpoint == JC.STRING_EMPTY || UseConfig().DominanceEndpoint == JC.STRING_EMPTY {
		t.Error("Expected endpoints to be set by updateDefault")
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
		ETFEndpoint:       "x",
		DominanceEndpoint: "x",
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
	if !UseConfig().CanDoETF() {
		t.Error("Expected CanDoETF to be true")
	}
	if !UseConfig().CanDoDominance() {
		t.Error("Expected CanDoDominance to be true")
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
