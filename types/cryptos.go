package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"jxwatcher/core"
)

type CryptosType struct {
	Values []CryptoType `json:"values"`
}

func (c *CryptosType) LoadFile() *CryptosType {

	core.PrintMemUsage("Start loading cryptos.json")

	b := bytes.NewBuffer(nil)
	f, _ := os.Open(core.BuildPathRelatedToUserDirectory([]string{"jxcryptwatcher", "cryptos.json"}))
	io.Copy(b, f)
	f.Close()

	err := json.Unmarshal(b.Bytes(), &c)

	if err != nil {
		wrappedErr := fmt.Errorf("Failed to load cryptos.json: %w", err)
		log.Println(wrappedErr)
	} else {
		log.Print("Cryptos Loaded")
	}

	core.PrintMemUsage("End loading cryptos.json")

	return c
}

func (c *CryptosType) CreateFile() *CryptosType {
	core.CreateFile(core.BuildPathRelatedToUserDirectory([]string{"jxcryptwatcher", "cryptos.json"}), c.FetchData())
	return c
}

func (c *CryptosType) CheckFile() *CryptosType {
	exists, err := core.FileExists(core.BuildPathRelatedToUserDirectory([]string{"jxcryptwatcher", "cryptos.json"}))
	if !exists {
		c.CreateFile()
	}

	if err != nil {
		log.Println(err)
	}

	return c
}

func (c *CryptosType) ConvertToMap() CryptosMapType {
	core.PrintMemUsage("Start populating cryptos")
	CM := CryptosMapType{}
	CM.Init()

	for _, crypto := range c.Values {

		// Only add crypto that is active at CMC
		if crypto.Status != 0 || crypto.IsActive != 0 {
			CM.data[strconv.FormatInt(crypto.Id, 10)] = crypto.CreateKey()
		}
	}
	core.PrintMemUsage("End populating cryptos")

	return CM
}

func (c *CryptosType) FetchData() string {

	client := &http.Client{}
	req, err := http.NewRequest("GET", Config.DataEndpoint, nil)

	if err != nil {
		log.Println(err)
		return ""
	}

	resp, err := client.Do(req)

	if err != nil {
		log.Println(err)
		return ""
	}

	respBody, err := io.ReadAll(resp.Body)

	if err != nil {
		wrappedErr := fmt.Errorf("Failed to fetched cryptodata from CMC: %w", err)
		log.Println(wrappedErr)
		return ""
	} else {
		log.Print("Fetched cryptodata from CMC")
	}

	return string(respBody)
}
