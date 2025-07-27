package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	JC "jxwatcher/core"

	"fyne.io/fyne/v2/storage"
)

type CryptosType struct {
	Values []CryptoType `json:"values"`
}

func (c *CryptosType) LoadFile() *CryptosType {

	JC.PrintMemUsage("Start loading cryptos.json")

	// Build full storage URI
	fileURI, err := storage.ParseURI(JC.BuildPathRelatedToUserDirectory([]string{"cryptos.json"}))
	if err != nil {
		log.Println("Error getting parsing uri for file:", err)
		return c
	}

	// Open the file using Fyne's storage API
	reader, err := storage.Reader(fileURI)
	if err != nil {
		log.Println("Failed to open cryptos.json:", err)
		return c
	}
	defer reader.Close()

	// Decode JSON from reader
	buffer := bytes.NewBuffer(nil)
	if _, err := io.Copy(buffer, reader); err != nil {
		log.Println("Failed to read cryptos.json:", err)
		return c
	}

	if err := json.Unmarshal(buffer.Bytes(), c); err != nil {
		wrappedErr := fmt.Errorf("Failed to decode cryptos.json: %w", err)
		log.Println(wrappedErr)
	} else {
		log.Println("Cryptos Loaded")
	}

	JC.PrintMemUsage("End loading cryptos.json")

	return c
}

func (c *CryptosType) CreateFile() *CryptosType {
	if !JC.CreateFile(JC.BuildPathRelatedToUserDirectory([]string{"cryptos.json"}), c.FetchData()) {
		return nil
	}

	return c
}

func (c *CryptosType) CheckFile() *CryptosType {
	exists, err := JC.FileExists(JC.BuildPathRelatedToUserDirectory([]string{"cryptos.json"}))

	if !exists {
		if c.CreateFile() == nil {
			log.Println("Failed to create cryptos.json with default values")
			c = &CryptosType{}
			return c
		} else {
			log.Println("Created cryptos.json with default values")
		}
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

	if c == nil || len(c.Values) == 0 {
		log.Println("No cryptos found in the data")
		return CM
	}

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
	JC.Notify("Requesting new cryptos data from exchange")

	client := &http.Client{}
	req, err := http.NewRequest("GET", Config.DataEndpoint, nil)

	if err != nil {
		log.Println("Error creating request:", err)
		JC.Notify("Failed to fetch cryptos data")
		return ""
	}

	resp, err := client.Do(req)

	if err != nil {
		log.Println("Error performing request:", err)
		JC.Notify("Failed to fetch cryptos data")
		return ""
	}

	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Println("Failed to read response body:", err)
		JC.Notify("Failed to fetch cryptos data")
		return ""
	}

	log.Println("Fetched cryptodata from CMC")
	JC.Notify("Retrieved cryptos data from exchange")

	JC.PrintMemUsage("End fetching cryptos data")

	return string(respBody)
}
