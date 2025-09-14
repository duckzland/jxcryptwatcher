package types

import (
	"encoding/json"
	"sync"

	JC "jxwatcher/core"
)

var Config ConfigType
var configMu sync.RWMutex

type ConfigType struct {
	DataEndpoint      string `json:"data_endpoint"`
	ExchangeEndpoint  string `json:"exchange_endpoint"`
	AltSeasonEndpoint string `json:"altseason_endpoint"`
	FearGreedEndpoint string `json:"feargreed_endpoint"`
	CMC100Endpoint    string `json:"cmc100_endpoint"`
	MarketCapEndpoint string `json:"marketcap_endpoint"`
	Delay             int64  `json:"delay"`
	Version           string `json:"version"`
}

func (c *ConfigType) LoadFile() *ConfigType {
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

func (c *ConfigType) updateDefault() *ConfigType {
	if c.Version == "" || c.Version == "1.2.0" {
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

func (c *ConfigType) SaveFile() *ConfigType {
	configMu.RLock()
	defer configMu.RUnlock()
	JC.SaveFile("config.json", Config)
	return c
}

func (c *ConfigType) CheckFile() *ConfigType {
	configMu.Lock()
	defer configMu.Unlock()

	exists, err := JC.FileExists(JC.BuildPathRelatedToUserDirectory([]string{"config.json"}))
	if !exists {
		data := ConfigType{
			DataEndpoint:      "https://s3.coinmarketcap.com/generated/core/crypto/cryptos.json",
			ExchangeEndpoint:  "https://api.coinmarketcap.com/data-api/v3/tools/price-conversion",
			AltSeasonEndpoint: "https://api.coinmarketcap.com/data-api/v3/altcoin-season/chart",
			FearGreedEndpoint: "https://api.coinmarketcap.com/data-api/v3/fear-greed/chart",
			CMC100Endpoint:    "https://api.coinmarketcap.com/data-api/v3/top100/supplement",
			MarketCapEndpoint: "https://api.coinmarketcap.com/data-api/v4/global-metrics/quotes/historical",
			Delay:             60,
		}

		if !JC.SaveFile("config.json", data) {
			JC.Logln("Failed to create config.json with default values")
			Config = data
			return &Config
		} else {
			JC.Logln("Created config.json with default values")
			Config = data
		}
	}

	if err != nil {
		JC.Logln(err)
	}

	return c
}

func (c *ConfigType) IsValid() bool {
	configMu.RLock()
	defer configMu.RUnlock()
	return c.DataEndpoint != "" && c.ExchangeEndpoint != ""
}

func (c *ConfigType) IsValidTickers() bool {
	configMu.RLock()
	defer configMu.RUnlock()
	return c.CanDoCMC100() || c.CanDoFearGreed() || c.CanDoMarketCap() || c.CanDoAltSeason()
}

func (c *ConfigType) CanDoCMC100() bool {
	configMu.RLock()
	defer configMu.RUnlock()
	return c.CMC100Endpoint != ""
}

func (c *ConfigType) CanDoMarketCap() bool {
	configMu.RLock()
	defer configMu.RUnlock()
	return c.MarketCapEndpoint != ""
}

func (c *ConfigType) CanDoFearGreed() bool {
	configMu.RLock()
	defer configMu.RUnlock()
	return c.FearGreedEndpoint != ""
}

func (c *ConfigType) CanDoAltSeason() bool {
	configMu.RLock()
	defer configMu.RUnlock()
	return c.AltSeasonEndpoint != ""
}

func ConfigInit() {
	configMu.Lock()
	Config = ConfigType{}
	configMu.Unlock()

	Config.CheckFile().LoadFile()
}
