package types

import (
	"strconv"
	"strings"
	"sync"

	"github.com/buger/jsonparser"

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

	if c.IsVersionLessThan("1.7.0") {
		JC.Logln("Updating old config to 1.8.0")
		c.Version = "1.8.0"

		if c.DataEndpoint == JC.STRING_EMPTY {
			c.DataEndpoint = "https://s3.coinmarketcap.com/generated/core/crypto/cryptos.json"
		}
		if c.ExchangeEndpoint == JC.STRING_EMPTY {
			c.ExchangeEndpoint = "https://api.coinmarketcap.com/data-api/v3/tools/price-conversion"
		}
		if c.AltSeasonEndpoint == JC.STRING_EMPTY {
			c.AltSeasonEndpoint = "https://api.coinmarketcap.com/data-api/v3/altcoin-season/chart"
		}
		if c.FearGreedEndpoint == JC.STRING_EMPTY {
			c.FearGreedEndpoint = "https://api.coinmarketcap.com/data-api/v3/fear-greed/chart"
		}
		if c.CMC100Endpoint == JC.STRING_EMPTY {
			c.CMC100Endpoint = "https://api.coinmarketcap.com/data-api/v3/top100/supplement"
		}
		if c.MarketCapEndpoint == JC.STRING_EMPTY {
			c.MarketCapEndpoint = "https://api.coinmarketcap.com/data-api/v4/global-metrics/quotes/historical"
		}
		if c.RSIEndpoint == JC.STRING_EMPTY {
			c.RSIEndpoint = "https://api.coinmarketcap.com/data-api/v3/cryptocurrency/rsi/heatmap/overall"
		}
		if c.ETFEndpoint == JC.STRING_EMPTY {
			c.ETFEndpoint = "https://api.coinmarketcap.com/data-api/v3/etf/overview/netflow/chart"
		}
		if c.DominanceEndpoint == JC.STRING_EMPTY {
			c.DominanceEndpoint = "https://api.coinmarketcap.com/data-api/v3/global-metrics/dominance/overview"
		}

		return true
	}

	return false
}

func (c *configType) parseJSON(data []byte) error {
	if val, err := jsonparser.GetString(data, "data_endpoint"); err == nil {
		c.DataEndpoint = val
	}
	if val, err := jsonparser.GetString(data, "exchange_endpoint"); err == nil {
		c.ExchangeEndpoint = val
	}
	if val, err := jsonparser.GetString(data, "altseason_endpoint"); err == nil {
		c.AltSeasonEndpoint = val
	}
	if val, err := jsonparser.GetString(data, "feargreed_endpoint"); err == nil {
		c.FearGreedEndpoint = val
	}
	if val, err := jsonparser.GetString(data, "cmc100_endpoint"); err == nil {
		c.CMC100Endpoint = val
	}
	if val, err := jsonparser.GetString(data, "marketcap_endpoint"); err == nil {
		c.MarketCapEndpoint = val
	}
	if val, err := jsonparser.GetString(data, "rsi_endpoint"); err == nil {
		c.RSIEndpoint = val
	}
	if val, err := jsonparser.GetString(data, "etf_endpoint"); err == nil {
		c.ETFEndpoint = val
	}
	if val, err := jsonparser.GetString(data, "dominance_endpoint"); err == nil {
		c.DominanceEndpoint = val
	}
	if val, err := jsonparser.GetString(data, "version"); err == nil {
		c.Version = val
	}
	if val, err := jsonparser.GetInt(data, "delay"); err == nil {
		c.Delay = val
	}
	return nil
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

	if err := c.parseJSON([]byte(content)); err != nil {
		JC.Logf("Failed to parse config.json: %v", err)
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
			Version:           "1.8.0",
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

func (c *configType) PostInit() {
	if c.IsVersionLessThan("1.7.0") {
		JC.Logln("Updating old config to 1.8.0")
		c.Version = "1.8.0"
		c.save()
	}
}

func (c *configType) IsVersionLessThan(target string) bool {
	parse := func(s string) []int {
		parts := strings.Split(s, ".")
		nums := make([]int, len(parts))
		for i, p := range parts {
			n, _ := strconv.Atoi(p)
			nums[i] = n
		}
		return nums
	}

	a := parse(c.Version)
	b := parse(target)

	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] < b[i] {
			return true
		} else if a[i] > b[i] {
			return false
		}
	}

	return len(a) < len(b)
}

func (c *configType) IsValid() bool {
	configMu.RLock()
	defer configMu.RUnlock()
	return c.DataEndpoint != JC.STRING_EMPTY && c.ExchangeEndpoint != JC.STRING_EMPTY
}

func (c *configType) IsValidTickers() bool {
	configMu.RLock()
	defer configMu.RUnlock()
	return c.CMC100Endpoint != JC.STRING_EMPTY || c.FearGreedEndpoint != JC.STRING_EMPTY || c.MarketCapEndpoint != JC.STRING_EMPTY || c.AltSeasonEndpoint != JC.STRING_EMPTY || c.RSIEndpoint != JC.STRING_EMPTY || c.ETFEndpoint != JC.STRING_EMPTY || c.DominanceEndpoint != JC.STRING_EMPTY
}

func (c *configType) CanDoCMC100() bool {
	configMu.RLock()
	defer configMu.RUnlock()
	return c.CMC100Endpoint != JC.STRING_EMPTY
}

func (c *configType) CanDoMarketCap() bool {
	configMu.RLock()
	defer configMu.RUnlock()
	return c.MarketCapEndpoint != JC.STRING_EMPTY
}

func (c *configType) CanDoFearGreed() bool {
	configMu.RLock()
	defer configMu.RUnlock()
	return c.FearGreedEndpoint != JC.STRING_EMPTY
}

func (c *configType) CanDoAltSeason() bool {
	configMu.RLock()
	defer configMu.RUnlock()
	return c.AltSeasonEndpoint != JC.STRING_EMPTY
}

func (c *configType) CanDoRSI() bool {
	configMu.RLock()
	defer configMu.RUnlock()
	return c.RSIEndpoint != JC.STRING_EMPTY
}

func (c *configType) CanDoETF() bool {
	configMu.RLock()
	defer configMu.RUnlock()
	return c.RSIEndpoint != JC.STRING_EMPTY
}

func (c *configType) CanDoDominance() bool {
	configMu.RLock()
	defer configMu.RUnlock()
	return c.DominanceEndpoint != JC.STRING_EMPTY
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
