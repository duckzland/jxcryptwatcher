package types

import (
	"encoding/json"
	"sync"

	JC "jxwatcher/core"
)

var configStorage *configType = &configType{}
var configMu sync.RWMutex

type configType struct {
	DataEndpoint      string `json:"data_endpoint"`
	ExchangeEndpoint  string `json:"exchange_endpoint"`
	AltSeasonEndpoint string `json:"altseason_endpoint"`
	FearGreedEndpoint string `json:"feargreed_endpoint"`
	CMC100Endpoint    string `json:"cmc100_endpoint"`
	MarketCapEndpoint string `json:"marketcap_endpoint"`
	Delay             int64  `json:"delay"`
	Version           string `json:"version"`
}

func (c *configType) LoadFile() *configType {
	configMu.Lock()
	defer configMu.Unlock()

	content, ok := JC.LoadFile("config.json")
	if !ok {
		JC.Logf("Failed to load config.json")
		return c
	}

	if err := json.Unmarshal([]byte(content), c); err != nil {
		JC.Logf("Failed to unmarshal config.json: %v", err)
		return c
	}

	c.updateDefault()
	JC.Logln("Configuration Loaded")
	return c
}

func (c *configType) updateDefault() *configType {
	if c.Version == "" || c.Version == "1.2.0" {
		JC.Logln("Updating old config")
		c.Version = "1.2.6"

		if c.AltSeasonEndpoint == "" {
			c.AltSeasonEndpoint = "https://api.coinmarketcap.com/data-api/v3/altcoin-season/chart"
		}
		if c.FearGreedEndpoint == "" {
			c.FearGreedEndpoint = "https://api.coinmarketcap.com/data-api/v3/fear-greed/chart"
		}
		if c.CMC100Endpoint == "" {
			c.CMC100Endpoint = "https://api.coinmarketcap.com/data-api/v3/top100/supplement"
		}
		if c.MarketCapEndpoint == "" {
			c.MarketCapEndpoint = "https://api.coinmarketcap.com/data-api/v4/global-metrics/quotes/historical"
		}

		c.SaveFile()
	}
	return c
}

func (c *configType) SaveFile() *configType {
	configMu.RLock()
	defer configMu.RUnlock()
	JC.SaveFile("config.json", configStorage)
	return c
}

func (c *configType) CheckFile() *configType {
	configMu.Lock()
	defer configMu.Unlock()

	exists, _ := JC.FileExists(JC.BuildPathRelatedToUserDirectory([]string{"config.json"}))
	if !exists {
		data := configType{
			DataEndpoint:      "https://s3.coinmarketcap.com/generated/core/crypto/cryptos.json",
			ExchangeEndpoint:  "https://api.coinmarketcap.com/data-api/v3/tools/price-conversion",
			AltSeasonEndpoint: "https://api.coinmarketcap.com/data-api/v3/altcoin-season/chart",
			FearGreedEndpoint: "https://api.coinmarketcap.com/data-api/v3/fear-greed/chart",
			CMC100Endpoint:    "https://api.coinmarketcap.com/data-api/v3/top100/supplement",
			MarketCapEndpoint: "https://api.coinmarketcap.com/data-api/v4/global-metrics/quotes/historical",
			Version:           "1.4.0",
			Delay:             60,
		}

		if !JC.SaveFile("config.json", data) {
			JC.Logln("Failed to create config.json with default values")
			configStorage = &data
			return configStorage
		} else {
			JC.Logln("Created config.json with default values")
			configStorage = &data
		}
	}

	return c
}

func (c *configType) IsValid() bool {
	configMu.RLock()
	defer configMu.RUnlock()
	return c.DataEndpoint != "" && c.ExchangeEndpoint != ""
}

func (c *configType) IsValidTickers() bool {
	configMu.RLock()
	defer configMu.RUnlock()
	return c.CanDoCMC100() || c.CanDoFearGreed() || c.CanDoMarketCap() || c.CanDoAltSeason()
}

func (c *configType) CanDoCMC100() bool {
	configMu.RLock()
	defer configMu.RUnlock()
	return c.CMC100Endpoint != ""
}

func (c *configType) CanDoMarketCap() bool {
	configMu.RLock()
	defer configMu.RUnlock()
	return c.MarketCapEndpoint != ""
}

func (c *configType) CanDoFearGreed() bool {
	configMu.RLock()
	defer configMu.RUnlock()
	return c.FearGreedEndpoint != ""
}

func (c *configType) CanDoAltSeason() bool {
	configMu.RLock()
	defer configMu.RUnlock()
	return c.AltSeasonEndpoint != ""
}

func ConfigInit() {
	configMu.Lock()
	configStorage = &configType{}
	configMu.Unlock()

	UseConfig().CheckFile().LoadFile()
}

func UseConfig() *configType {
	return configStorage
}
