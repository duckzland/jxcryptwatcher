package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

type CryptosType struct {
	Values []CryptoType `json:"values"`
}

func (c *CryptosType) LoadFile() *CryptosType {

	PrintMemUsage("Start loading cryptos.json")

	b := bytes.NewBuffer(nil)
	f, _ := os.Open(buildPathRelatedToUserDirectory([]string{"jxcryptwatcher", "cryptos.json"}))
	io.Copy(b, f)
	f.Close()

	err := json.Unmarshal(b.Bytes(), &c)

	if err != nil {
		wrappedErr := fmt.Errorf("Failed to load cryptos.json: %w", err)
		log.Fatal(wrappedErr)
	} else {
		log.Print("Cryptos Loaded")
	}

	PrintMemUsage("End loading cryptos.json")

	return c
}

func (c *CryptosType) CreateFile() *CryptosType {
	createFile(buildPathRelatedToUserDirectory([]string{"jxcryptwatcher", "cryptos.json"}), c.FetchData())
	return c
}

func (c *CryptosType) CheckFile() *CryptosType {
	exists, err := fileExists(buildPathRelatedToUserDirectory([]string{"jxcryptwatcher", "cryptos.json"}))
	if !exists {
		c.CreateFile()
	}

	if err != nil {
		log.Fatalln(err)
	}

	return c
}

func (c *CryptosType) ConvertToMap() CryptosMap {
	PrintMemUsage("Start populating cryptos")
	CM := CryptosMap{}
	CM.Init()

	for _, crypto := range c.Values {

		// Only add crypto that is active at CMC
		if crypto.Status != 0 || crypto.IsActive != 0 {
			CM.data[strconv.FormatInt(crypto.Id, 10)] = crypto.CreateKey()
		}
	}
	PrintMemUsage("End populating cryptos")

	return CM
}

func (c *CryptosType) FetchData() string {

	client := &http.Client{}
	req, err := http.NewRequest("GET", Config.DataEndpoint, nil)

	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.Do(req)

	if err != nil {
		log.Fatal(err)
	}

	respBody, err := io.ReadAll(resp.Body)

	if err != nil {
		wrappedErr := fmt.Errorf("Failed to fetched cryptodata from CMC: %w", err)
		log.Fatal(wrappedErr)
	} else {
		log.Print("Fetched cryptodata from CMC")
	}

	return string(respBody)
}
