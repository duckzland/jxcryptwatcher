package types

import (
	"bytes"
	"encoding/json"
	"io"

	"fyne.io/fyne/v2/storage"

	JC "jxwatcher/core"
)

var Config ConfigType

type ConfigType struct {
	DataEndpoint            string `json:"data_endpoint"`
	ExchangeEndpoint        string `json:"exchange_endpoint"`
	Delay                   int64  `json:"delay"`
	TickerCMC100Endpoint    string `json:"ticker_cmc100_endpoint"`
	TickerFearGreedEndpoint string `json:"ticker_feargreed_endpoint"`
	TickerMetricsEndpoint   string `json:"ticker_metrics_endpoint"`
	TickerListingsEndpoint  string `json:"ticker_listings_endpoint"`
	TickerDelay             int64  `json:"ticker_delay"`
	ProApiKey               string `json:"pro_api_key"`
	Version                 string `json:"version"`
}

func (c *ConfigType) LoadFile() *ConfigType {

	// Construct the file URI
	fileURI, err := storage.ParseURI(JC.BuildPathRelatedToUserDirectory([]string{"config.json"}))
	if err != nil {
		JC.Logln("Error getting parsing uri for file:", fileURI, err)
		return c
	}

	// Attempt to open the file with Fyne's Reader
	reader, err := storage.Reader(fileURI)
	if err != nil {
		JC.Logln("Failed to open config.json:", err)
		return c
	}
	defer reader.Close()

	// Read the file into a buffer
	buffer := bytes.NewBuffer(nil)
	if _, err := io.Copy(buffer, reader); err != nil {
		JC.Logln("Failed to read config contents:", err)
		return c
	}

	// Parse JSON into the config object
	if err := json.Unmarshal(buffer.Bytes(), c); err != nil {
		JC.Logf("Failed to unmarshal config.json: %v", err)
		return c
	}

	// New Version 1.2.0
	c.updateDefault()

	JC.Logln("Configuration Loaded")

	return c
}

func (c *ConfigType) updateDefault() *ConfigType {

	// Since version 1.2.0
	if c.Version == "" {
		c.Version = "1.2.0"

		if c.TickerCMC100Endpoint == "" {
			c.TickerCMC100Endpoint = "https://pro-api.coinmarketcap.com/v3/index/cmc100-latest"
		}

		if c.TickerFearGreedEndpoint == "" {
			c.TickerFearGreedEndpoint = "https://pro-api.coinmarketcap.com/v3/fear-and-greed/latest"
		}

		if c.TickerMetricsEndpoint == "" {
			c.TickerMetricsEndpoint = "https://pro-api.coinmarketcap.com/v1/global-metrics/quotes/latest"
		}

		if c.TickerListingsEndpoint == "" {
			c.TickerListingsEndpoint = "https://pro-api.coinmarketcap.com/v1/cryptocurrency/listings/latest"
		}

		if c.TickerDelay == 0 {
			c.TickerDelay = 1500
		}

		c.SaveFile()
	}

	return c
}

func (c *ConfigType) SaveFile() *ConfigType {

	jsonData, err := json.MarshalIndent(Config, "", "  ")
	if err != nil {
		JC.Logln("Error marshaling config:", err)
		return nil
	}

	JC.CreateFile(JC.BuildPathRelatedToUserDirectory([]string{"config.json"}), string(jsonData))

	return c
}

func (c *ConfigType) CheckFile() *ConfigType {
	exists, err := JC.FileExists(JC.BuildPathRelatedToUserDirectory([]string{"config.json"}))
	if !exists {
		data := ConfigType{
			DataEndpoint:            "https://s3.coinmarketcap.com/generated/core/crypto/cryptos.json",
			ExchangeEndpoint:        "https://api.coinmarketcap.com/data-api/v3/tools/price-conversion",
			Delay:                   60,
			TickerCMC100Endpoint:    "https://pro-api.coinmarketcap.com/v3/index/cmc100-latest",
			TickerFearGreedEndpoint: "https://pro-api.coinmarketcap.com/v3/fear-and-greed/latest",
			TickerMetricsEndpoint:   "https://pro-api.coinmarketcap.com/v1/global-metrics/quotes/latest",
			TickerListingsEndpoint:  "https://pro-api.coinmarketcap.com/v1/cryptocurrency/listings/latest",
			TickerDelay:             1500,
			ProApiKey:               "",
		}

		jsonData, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			JC.Logln(err)
			return c
		}

		if !JC.CreateFile(JC.BuildPathRelatedToUserDirectory([]string{"config.json"}), string(jsonData)) {
			JC.Logln("Failed to create config.json with default values")
			c = &data

			return c
		} else {
			JC.Logln("Created config.json with default values")
		}
	}

	if err != nil {
		JC.Logln(err)
	}

	return c
}

func (c *ConfigType) IsValid() bool {
	return c.DataEndpoint != "" && c.ExchangeEndpoint != ""
}

func (c *ConfigType) IsValidTickers() bool {
	return c.CanDoCMC100() || c.CanDoFearGreed() || c.CanDoMetrics() || c.CanDoListings()
}

func (c *ConfigType) HasProKey() bool {
	return c.ProApiKey != ""
}

func (c *ConfigType) CanDoCMC100() bool {
	return c.HasProKey() && c.TickerCMC100Endpoint != ""
}

func (c *ConfigType) CanDoFearGreed() bool {
	return c.HasProKey() && c.TickerFearGreedEndpoint != ""
}

func (c *ConfigType) CanDoMetrics() bool {
	return c.HasProKey() && c.TickerMetricsEndpoint != ""
}

func (c *ConfigType) CanDoListings() bool {
	return c.HasProKey() && c.TickerListingsEndpoint != ""
}

func ConfigInit() {
	Config = ConfigType{}
	Config.CheckFile().LoadFile()
}
