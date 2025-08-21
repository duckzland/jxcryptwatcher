package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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
		JC.Logln("Error getting parsing uri for file:", err)
		JC.Notify("Failed to load cryptos data")

		return c
	}

	// Open the file using Fyne's storage API
	reader, err := storage.Reader(fileURI)
	if err != nil {
		JC.Logln("Failed to open cryptos.json:", err)
		JC.Notify("Failed to load cryptos data")

		return c
	}
	defer reader.Close()

	// Decode JSON from reader
	buffer := bytes.NewBuffer(nil)
	if _, err := io.Copy(buffer, reader); err != nil {
		JC.Logln("Failed to read cryptos.json:", err)
		JC.Notify("Failed to load cryptos data")

		return c
	}

	if err := json.Unmarshal(buffer.Bytes(), c); err != nil {

		wrappedErr := fmt.Errorf("Failed to decode cryptos.json: %w", err)
		JC.Notify("Failed to load cryptos data")
		JC.Logln(wrappedErr)

	} else {

		JC.Logln("Cryptos Loaded")

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
			JC.Logln("Failed to create cryptos.json with default values")
			JC.Notify("Failed to create cryptos data file")
			c = &CryptosType{}

			return c
		} else {
			JC.Logln("Created cryptos.json with default values")
		}
	}

	if err != nil {
		JC.Logln(err)
	}

	return c
}

func (c *CryptosType) ConvertToMap() *CryptosMapType {

	JC.PrintMemUsage("Start populating cryptos")

	CM := &CryptosMapType{}
	CM.Init()

	if c == nil || len(c.Values) == 0 {
		JC.Logln("No cryptos found in the data")
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
		JC.Logln("Error creating request:", err)
		JC.Notify("Failed to fetch cryptos data")
		return ""
	}

	resp, err := client.Do(req)

	if err != nil {
		JC.Logln("Error performing request:", err)
		JC.Notify("Failed to fetch cryptos data")
		return ""
	}

	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)

	if err != nil {
		JC.Logln("Failed to read response body:", err)
		JC.Notify("Failed to fetch cryptos data")
		return ""
	}

	JC.Logln("Fetched cryptodata from CMC")
	JC.Notify("Retrieved cryptos data from exchange")

	JC.PrintMemUsage("End fetching cryptos data")

	payload := string(respBody)

	// Debug simulating invalid json
	// payload = "{///}"
	// payload = "{}"
	// payload = `{"values":{}}`
	// payload = `{"values":[{}]}`
	// payload = `{"values":[[]]}`

	return payload
}
