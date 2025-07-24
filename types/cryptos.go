package types

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	JC "jxwatcher/core"
)

type CryptosType struct {
	Values []CryptoType `json:"values"`
}

func (c *CryptosType) LoadFile() *CryptosType {
	JC.PrintMemUsage("Start loading cryptos.json")

	f, err := os.Open(JC.BuildPathRelatedToUserDirectory([]string{"jxcryptwatcher", "cryptos.json"}))
	if err != nil {
		log.Println("Failed to open cryptos.json:", err)
		return c
	}

	defer f.Close()

	decoder := json.NewDecoder(f)
	if err := decoder.Decode(c); err != nil {
		wrappedErr := fmt.Errorf("Failed to decode cryptos.json: %w", err)
		log.Println(wrappedErr)

	} else {
		log.Println("Cryptos Loaded")
	}

	JC.PrintMemUsage("End loading cryptos.json")

	return c
}

func (c *CryptosType) CreateFile() *CryptosType {
	JC.CreateFile(JC.BuildPathRelatedToUserDirectory([]string{"jxcryptwatcher", "cryptos.json"}), c.FetchData())

	return c
}

func (c *CryptosType) CheckFile() *CryptosType {
	exists, err := JC.FileExists(JC.BuildPathRelatedToUserDirectory([]string{"jxcryptwatcher", "cryptos.json"}))

	if !exists {
		c.CreateFile()
	}

	if err != nil {
		log.Println(err)
	}

	return c
}

func (c *CryptosType) ConvertToMap() *CryptosMapType {

	JC.PrintMemUsage("Start populating cryptos")

	CM := &CryptosMapType{}
	CM.Init()

	for _, crypto := range c.Values {
		if crypto.Status != 0 || crypto.IsActive != 0 {
			CM.Insert(strconv.FormatInt(crypto.Id, 10), crypto.CreateKey())
		}
	}

	JC.PrintMemUsage("End populating cryptos")

	return CM
}

func (c *CryptosType) FetchData() string {
	JC.PrintMemUsage("Start fetching cryptos data")

	client := &http.Client{}
	req, err := http.NewRequest("GET", Config.DataEndpoint, nil)

	if err != nil {
		log.Println("Error creating request:", err)
		return ""
	}

	resp, err := client.Do(req)

	if err != nil {
		log.Println("Error performing request:", err)
		return ""
	}

	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Println("Failed to read response body:", err)
		return ""
	}

	log.Println("Fetched cryptodata from CMC")
	JC.PrintMemUsage("End fetching cryptos data")

	return string(respBody)
}
