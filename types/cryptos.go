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

	JC "jxwatcher/core"
)

type CryptosType struct {
	Values []CryptoType `json:"values"`
}

func (c *CryptosType) LoadFile() *CryptosType {

	JC.PrintMemUsage("Start loading cryptos.json")

	b := bytes.NewBuffer(nil)
	f, _ := os.Open(JC.BuildPathRelatedToUserDirectory([]string{"jxcryptwatcher", "cryptos.json"}))
	io.Copy(b, f)
	f.Close()

	err := json.Unmarshal(b.Bytes(), &c)

	if err != nil {
		wrappedErr := fmt.Errorf("Failed to load cryptos.json: %w", err)
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

func (c *CryptosType) ConvertToMap() CryptosMapType {
	JC.PrintMemUsage("Start populating cryptos")
	CM := CryptosMapType{}
	CM.Init()

	for _, crypto := range c.Values {

		// Only add crypto that is active at CMC
		if crypto.Status != 0 || crypto.IsActive != 0 {
			CM.data[strconv.FormatInt(crypto.Id, 10)] = crypto.CreateKey()
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
		log.Println("Fetched cryptodata from CMC")
	}

	JC.PrintMemUsage("End fetching cryptos data")

	return string(respBody)
}
