package types

import (
	"bytes"
	"encoding/json"
	"io"
	"log"

	"fyne.io/fyne/v2/storage"

	JC "jxwatcher/core"
)

var Config ConfigType

type ConfigType struct {
	DataEndpoint     string `json:"data_endpoint"`
	ExchangeEndpoint string `json:"exchange_endpoint"`
	Delay            int64  `json:"delay"`
}

func (c *ConfigType) LoadFile() *ConfigType {

	// Construct the file URI
	fileURI, err := storage.ParseURI(JC.BuildPathRelatedToUserDirectory([]string{"config.json"}))
	if err != nil {
		log.Println("Error getting parsing uri for file:", fileURI, err)
		return c
	}

	// Attempt to open the file with Fyne's Reader
	reader, err := storage.Reader(fileURI)
	if err != nil {
		log.Println("Failed to open config.json:", err)
		return c
	}
	defer reader.Close()

	// Read the file into a buffer
	buffer := bytes.NewBuffer(nil)
	if _, err := io.Copy(buffer, reader); err != nil {
		log.Println("Failed to read config contents:", err)
		return c
	}

	// Parse JSON into the config object
	if err := json.Unmarshal(buffer.Bytes(), c); err != nil {
		log.Printf("Failed to unmarshal config.json: %v", err)
		return c
	}

	log.Println("Configuration Loaded")

	return c
}

func (c *ConfigType) SaveFile() *ConfigType {

	jsonData, err := json.MarshalIndent(Config, "", "  ")
	if err != nil {
		log.Println("Error marshaling config:", err)
		return nil
	}

	JC.CreateFile(JC.BuildPathRelatedToUserDirectory([]string{"config.json"}), string(jsonData))

	return c
}

func (c *ConfigType) CheckFile() *ConfigType {
	exists, err := JC.FileExists(JC.BuildPathRelatedToUserDirectory([]string{"config.json"}))
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

		if !JC.CreateFile(JC.BuildPathRelatedToUserDirectory([]string{"config.json"}), string(jsonData)) {
			log.Println("Failed to create config.json with default values")
			c = &data

			return c
		} else {
			log.Println("Created config.json with default values")
		}
	}

	if err != nil {
		log.Println(err)
	}

	return c
}

func (c *ConfigType) IsValid() bool {
	return c.DataEndpoint != "" && c.ExchangeEndpoint != ""
}

func ConfigInit() {
	Config = ConfigType{}
	Config.CheckFile().LoadFile()
}
