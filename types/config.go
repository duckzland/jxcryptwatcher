package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	JC "jxwatcher/core"
)

var Config ConfigType

type ConfigType struct {
	DataEndpoint     string `json:"data_endpoint"`
	ExchangeEndpoint string `json:"exchange_endpoint"`
	Delay            int64  `json:"delay"`
}

func (c *ConfigType) LoadFile() *ConfigType {
	f, err := os.Open(JC.BuildPathRelatedToUserDirectory([]string{"jxcryptwatcher", "config.json"}))
	if err != nil {
		log.Println("Failed to open config.json:", err)
		return c
	}
	defer f.Close()

	b := bytes.NewBuffer(nil)
	if _, err := io.Copy(b, f); err != nil {
		log.Println("Failed to copy file contents:", err)
		return c
	}

	if err := json.Unmarshal(b.Bytes(), c); err != nil {
		wrappedErr := fmt.Errorf("Failed to load config.json: %w", err)
		log.Println(wrappedErr)
	} else {
		log.Println("Configuration Loaded")
	}

	return c
}

func (c *ConfigType) SaveFile() *ConfigType {

	jsonData, err := json.MarshalIndent(Config, "", "  ")
	if err != nil {
		log.Println(err)
		return nil
	}

	// Save to file
	err = os.WriteFile(JC.BuildPathRelatedToUserDirectory([]string{"jxcryptwatcher", "config.json"}), jsonData, 0644)
	if err != nil {
		log.Println(err)
		return nil
	}

	log.Println("Configuration File Saved")

	return c
}

func (c *ConfigType) CheckFile() *ConfigType {
	exists, err := JC.FileExists(JC.BuildPathRelatedToUserDirectory([]string{"jxcryptwatcher", "config.json"}))
	if !exists {
		data := ConfigType{
			DataEndpoint:     "https://s3.coinmarketcap.com/generated/core/crypto/cryptos.json",
			ExchangeEndpoint: "https://api.coinmarketcap.com/data-api/v3/tools/price-conversion",
			Delay:            60,
		}

		jsonData, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			log.Println(err)
			return c
		}

		JC.CreateFile(JC.BuildPathRelatedToUserDirectory([]string{"jxcryptwatcher", "config.json"}), string(jsonData))
	}

	if err != nil {
		log.Println(err)
	}

	return c
}

func ConfigInit() {
	Config = ConfigType{}
	Config.CheckFile().LoadFile()
}
