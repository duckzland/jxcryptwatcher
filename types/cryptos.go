package types

import (
	// "encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	JC "jxwatcher/core"

	json "github.com/goccy/go-json"
)

var cryptosLoaderStorage *cryptosLoaderType
var cryptosMu sync.RWMutex

type cryptosLoaderType struct {
	Values []cryptoType `json:"values"`
}

func (c *cryptosLoaderType) load() *cryptosLoaderType {
	JC.PrintPerfStats("Loading cryptos.json", time.Now())

	content, ok := JC.LoadFileFromStorage("cryptos.json")
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
	return c
}

func (c *cryptosLoaderType) create() *cryptosLoaderType {
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

func (c *cryptosLoaderType) check() *cryptosLoaderType {
	exists, err := JC.FileExists(JC.BuildPathRelatedToUserDirectory([]string{"cryptos.json"}))

	if !exists {
		if c.create() == nil {
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

func (c *cryptosLoaderType) convert() *cryptosMapType {
	JC.PrintPerfStats("Generating cryptos", time.Now())

	cm := &cryptosMapType{}
	cm.Init()

	if c == nil || len(c.Values) == 0 {
		JC.Logln("No cryptos found in the data")
		return cm
	}

	for _, crypto := range c.Values {
		if crypto.Status != 0 || crypto.IsActive != 0 {
			cm.Insert(strconv.FormatInt(crypto.Id, 10), crypto.createKey())
		}
	}

	return cm
}

func (c *cryptosLoaderType) GetCryptos() int64 {
	JC.PrintPerfStats("Fetching cryptos data", time.Now())
	JC.Notify("Requesting latest cryptos data from exchange...")

	return JC.GetRequest(
		UseConfig().DataEndpoint,
		nil,
		func(url url.Values, req *http.Request) {
		},
		func(resp *http.Response, cc any) int64 {
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

			return JC.NETWORKING_SUCCESS
		})
}

func CryptosLoaderInit() {
	cryptosMu.Lock()
	cryptosLoaderStorage = &cryptosLoaderType{}
	cryptosMu.Unlock()

	cm := cryptosLoaderStorage.check().load().convert()
	cm.ClearMapCache()

	UsePanelMaps().SetMaps(cm)
	UsePanelMaps().GetOptions()
}

func UseCryptosLoader() *cryptosLoaderType {
	return cryptosLoaderStorage
}
