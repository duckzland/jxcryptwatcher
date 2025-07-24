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
	b := bytes.NewBuffer(nil)
	f, _ := os.Open(JC.BuildPathRelatedToUserDirectory([]string{"jxcryptwatcher", "config.json"}))
	io.Copy(b, f)
	f.Close()

	err := json.Unmarshal(b.Bytes(), c)

	if err != nil {
		wrappedErr := fmt.Errorf("Failed to load config.json: %w", err)
		log.Println(wrappedErr)
	} else {
		log.Print("Configuration Loaded")
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
