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
	RSIEndpoint       string `json:"rsi_endpoint"`
	ETFEndpoint       string `json:"etf_endpoint"`
	DominanceEndpoint string `json:"dominance_endpoint"`
	Delay             int64  `json:"delay"`
	Version           string `json:"version"`
}

func (c *configType) update() bool {

	if c.Version != "1.7.0" {
		JC.Logln("Updating old config to 1.7.0")
		c.Version = "1.7.0"

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
		if c.RSIEndpoint == "" {
			c.RSIEndpoint = "https://api.coinmarketcap.com/data-api/v3/cryptocurrency/rsi/heatmap/overall"
		}
		if c.ETFEndpoint == "" {
			c.ETFEndpoint = "https://api.coinmarketcap.com/data-api/v3/etf/overview/netflow/chart"
		}
		if c.DominanceEndpoint == "" {
			c.DominanceEndpoint = "https://api.coinmarketcap.com/data-api/v3/global-metrics/dominance/overview"
		}

		return true
	}

	return false
}

func (c *configType) load() bool {
	configMu.Lock()
	defer configMu.Unlock()

	content, ok := JC.LoadFileFromStorage("config.json")
	if !ok {
		JC.Logf("Failed to load config.json")
		configStorage.check()

		return c.IsValid() && c.IsValidTickers()
	}

	if err := json.Unmarshal([]byte(content), c); err != nil {
		JC.Logf("Failed to unmarshal config.json: %v", err)
		return false
	}

	if c.update() {
		// Call directly!, using c.SaveConfig() will cause double lock!
		JC.SaveFileToStorage("config.json", configStorage)
		return false
	}

	JC.Logln("Configuration Loaded")

	return true
}

func (c *configType) save() bool {
	configMu.RLock()
	defer configMu.RUnlock()

	return JC.SaveFileToStorage("config.json", configStorage)
}

func (c *configType) check() *configType {
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
			RSIEndpoint:       "https://api.coinmarketcap.com/data-api/v3/cryptocurrency/rsi/heatmap/overall",
			ETFEndpoint:       "https://api.coinmarketcap.com/data-api/v3/etf/overview/netflow/chart",
			DominanceEndpoint: "https://api.coinmarketcap.com/data-api/v3/global-metrics/dominance/overview",
			Version:           "1.7.0",
			Delay:             60,
		}

		if !JC.SaveFileToStorage("config.json", data) {
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
	return c.CMC100Endpoint != "" || c.FearGreedEndpoint != "" || c.MarketCapEndpoint != "" || c.AltSeasonEndpoint != "" || c.RSIEndpoint != "" || c.ETFEndpoint != "" || c.DominanceEndpoint != ""
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

func (c *configType) CanDoRSI() bool {
	configMu.RLock()
	defer configMu.RUnlock()
	return c.RSIEndpoint != ""
}

func (c *configType) CanDoETF() bool {
	configMu.RLock()
	defer configMu.RUnlock()
	return c.RSIEndpoint != ""
}

func (c *configType) CanDoDominance() bool {
	configMu.RLock()
	defer configMu.RUnlock()
	return c.DominanceEndpoint != ""
}

func ConfigInit() bool {
	configMu.Lock()
	configStorage = &configType{}
	configMu.Unlock()

	return UseConfig().check().load()
}

func UseConfig() *configType {
	return configStorage
}

func ConfigSave() bool {
	return UseConfig().save()
}
