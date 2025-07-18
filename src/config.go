package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

/**
 * Defining Struct for config.json
 */
type ConfigType struct {
	DataEndpoint     string `json:"data_endpoint"`
	ExchangeEndpoint string `json:"exchange_endpoint"`
	Delay            int64  `json:"delay"`
}

/**
 * Global variables
 */
var Config ConfigType

/**
 * Load Configuration Json into memory
 */
func loadConfig() {
	b := bytes.NewBuffer(nil)
	f, _ := os.Open(buildPathRelatedToUserDirectory([]string{"jxcryptwatcher", "config.json"}))
	io.Copy(b, f)
	f.Close()

	err := json.Unmarshal(b.Bytes(), &Config)

	if err != nil {
		wrappedErr := fmt.Errorf("Failed to load config.json: %w", err)
		log.Fatal(wrappedErr)
	} else {
		log.Print("Configuration Loaded")
	}
}

func saveConfig() {

	jsonData, err := json.MarshalIndent(Config, "", "  ")
	if err != nil {
		log.Fatalln(err)
	}

	// Save to file
	err = os.WriteFile(buildPathRelatedToUserDirectory([]string{"jxcryptwatcher", "config.json"}), jsonData, 0644)
	if err != nil {
		log.Fatalln(err)
	}
}

/**
 * Helper function to check fo config.json and try to regenerate it when not found
 */
func checkConfig() {
	exists, err := fileExists(buildPathRelatedToUserDirectory([]string{"jxcryptwatcher", "config.json"}))
	if !exists {
		data := ConfigType{
			DataEndpoint:     "https://s3.coinmarketcap.com/generated/core/crypto/cryptos.json",
			ExchangeEndpoint: "https://api.coinmarketcap.com/data-api/v3/tools/price-conversion",
			Delay:            60,
		}

		jsonData, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			log.Fatalln(err)
		}

		createFile(buildPathRelatedToUserDirectory([]string{"jxcryptwatcher", "config.json"}), string(jsonData))
	}

	if err != nil {
		log.Fatalln(err)
	}

	loadConfig()
}
