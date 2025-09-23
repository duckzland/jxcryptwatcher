package types

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	JC "jxwatcher/core"
)

var CryptosLoader cryptosLoaderType
var cryptosMu sync.RWMutex

type cryptosLoaderType struct {
	Values []cryptoType `json:"values"`
}

func (c *cryptosLoaderType) LoadFile() *cryptosLoaderType {
	JC.PrintMemUsage("Start loading cryptos.json")

	content, ok := JC.LoadFile("cryptos.json")
	if !ok {
		JC.Logln("Failed to open cryptos.json")
		JC.Notify("Failed to load cryptos data")
	}

	if err := json.Unmarshal([]byte(content), c); err != nil {
		wrappedErr := fmt.Errorf("Failed to decode cryptos.json: %w", err)
		JC.Notify("Failed to load cryptos data")
		JC.Logln(wrappedErr)
		return c
	}

	JC.Logln("Cryptos Loaded")
	JC.PrintMemUsage("End loading cryptos.json")
	return c
}

func (c *cryptosLoaderType) CreateFile() *cryptosLoaderType {
	status := c.GetCryptos()
	switch status {
	case JC.NETWORKING_FAILED_CREATE_FILE:
		JC.Logln("Failed to create cryptos.json with new values")
		return nil
	case JC.NETWORKING_BAD_DATA_RECEIVED, JC.NETWORKING_ERROR_CONNECTION, JC.NETWORKING_URL_ERROR:
		return nil
	}
	return c
}

func (c *cryptosLoaderType) CheckFile() *cryptosLoaderType {
	exists, err := JC.FileExists(JC.BuildPathRelatedToUserDirectory([]string{"cryptos.json"}))

	if !exists {
		if c.CreateFile() == nil {
			JC.Logln("Failed to create cryptos.json with default values")
			JC.Notify("Failed to create cryptos data file")
			c = &cryptosLoaderType{}
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

func (c *cryptosLoaderType) ConvertToMap() *cryptosMapType {
	JC.PrintMemUsage("Start populating cryptos")

	CM := &cryptosMapType{}
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

func (c *cryptosLoaderType) GetCryptos() int64 {
	JC.PrintMemUsage("Start fetching cryptos data")
	JC.Notify("Requesting latest cryptos data from exchange...")

	parsedURL, err := url.Parse(Config.DataEndpoint)
	if err != nil {
		JC.Logln("Invalid URL:", err)
		return JC.NETWORKING_URL_ERROR
	}

	client := &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequest("GET", parsedURL.String(), nil)
	if err != nil {
		JC.Logln("Error creating request:", err)
		JC.Notify("Failed to fetch cryptos data")
		return JC.NETWORKING_ERROR_CONNECTION
	}

	req.Header.Set("User_Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:142.0) Gecko/20100101 Firefox/142.0")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Expires", "0")

	JC.Logf("Fetching data from %v", req.URL)

	resp, err := client.Do(req)
	if err != nil {
		JC.Logln("Error performing request:", err)
		JC.Notify("Failed to fetch cryptos data from exchange.")
		return JC.NETWORKING_ERROR_CONNECTION
	}

	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		JC.Logln("Failed to read response body:", err)
		JC.Notify("Failed to fetch cryptos data")
		return JC.NETWORKING_BAD_DATA_RECEIVED
	}

	payload := string(respBody)
	if payload == "" {
		return JC.NETWORKING_FAILED_CREATE_FILE
	}

	if err := json.Unmarshal(respBody, &c); err != nil {
		JC.Logln(fmt.Errorf("Failed to examine cryptos data: %w", err))
		return JC.NETWORKING_BAD_DATA_RECEIVED
	}

	if c == nil || len(c.Values) == 0 {
		JC.Logln("No cryptos found in the data")
		return JC.NETWORKING_BAD_DATA_RECEIVED
	}

	if !JC.CreateFile(JC.BuildPathRelatedToUserDirectory([]string{"cryptos.json"}), payload) {
		return JC.NETWORKING_FAILED_CREATE_FILE
	}

	JC.Logln("Fetched cryptodata from CMC")
	JC.Notify("Successfully retrieved cryptos data from exchange.")
	JC.PrintMemUsage("End fetching cryptos data")

	return JC.NETWORKING_SUCCESS
}

func CryptosLoaderInit() {
	cryptosMu.Lock()
	CryptosLoader = cryptosLoaderType{}
	cryptosMu.Unlock()

	CM := CryptosLoader.CheckFile().LoadFile().ConvertToMap()
	CM.ClearMapCache()

	UsePanelMaps().SetMaps(CM)
	UsePanelMaps().GetOptions()
}
