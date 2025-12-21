package types

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/buger/jsonparser"

	JC "jxwatcher/core"
)

var cryptosLoaderStorage *cryptosLoaderType
var cryptosMu sync.RWMutex

type cryptosLoaderType struct {
	Values []cryptoType
}

func (c *cryptosLoaderType) load() *cryptosLoaderType {
	JC.PrintPerfStats("Loading cryptos.json", time.Now())

	data, ok := JC.LoadFileFromStorage("cryptos.json")
	if !ok {
		JC.Logln("Failed to open cryptos.json")
		JC.Notify(JC.NotifyFailedToLoadCryptosData)
		return c
	}

	if err := c.parseJSON([]byte(data)); err != nil {
		JC.Notify(JC.NotifyFailedToLoadCryptosData)
		return c
	}

	JC.Logln("Cryptos Loaded")
	return c
}

func (c *cryptosLoaderType) create() *cryptosLoaderType {
	status := c.GetCryptos(nil, nil)
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
			JC.Notify(JC.NotifyFailedToCreateCryptosDataFile)
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

func (c *cryptosLoaderType) parseJSON(data []byte) error {
	c.Values = nil
	var parsed []cryptoType

	_, err := jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		var cp cryptoType

		idFloat, err := jsonparser.GetFloat(value, "[0]")
		if err != nil {
			JC.Logln("Invalid 'id' field:", err)
			return
		}
		nameStr, err := jsonparser.GetString(value, "[1]")
		if err != nil {
			JC.Logln("Invalid 'name' field:", err)
			return
		}
		symbolStr, err := jsonparser.GetString(value, "[2]")
		if err != nil {
			JC.Logln("Invalid 'symbol' field:", err)
			return
		}
		isActiveFloat, err := jsonparser.GetFloat(value, "[4]")
		if err != nil {
			JC.Logln("Invalid 'is_active' field:", err)
			return
		}
		statusFloat, err := jsonparser.GetFloat(value, "[5]")
		if err != nil {
			JC.Logln("Invalid 'status' field:", err)
			return
		}

		cp.Id = int64(idFloat)
		cp.Name = cp.sanitizeText(nameStr, true, false, true)
		cp.Symbol = cp.sanitizeText(symbolStr, false, true, false)
		cp.IsActive = int64(isActiveFloat)
		cp.Status = int64(statusFloat)

		if cp.Id != 0 && cp.IsActive != 0 && cp.Status != 0 {
			parsed = append(parsed, cp)
		}
	}, "values")

	if err != nil {
		JC.Logln("Failed to parse cryptos:", err)
		return err
	}
	if len(parsed) == 0 {
		JC.Logln("No cryptos found in the data")
		return fmt.Errorf("empty cryptos list")
	}

	c.Values = parsed
	return nil
}

func (c *cryptosLoaderType) GetCryptos(ctx context.Context, payload any) int64 {

	JC.Notify(JC.NotifyRequestingLatestCryptosDataFromExchange)

	return JC.GetRequest(
		ctx,
		UseConfig().DataEndpoint,
		func(url url.Values, req *http.Request) {},
		func(cctx context.Context, resp *http.Response) int64 {

			if cctx != nil && cctx.Err() != nil {
				return JC.NETWORKING_ERROR_CONNECTION
			}

			body, _, err := JC.ReadResponse(JC.ACT_CRYPTO_GET_MAP, resp, 5*1024)
			if err != nil {
				return JC.NETWORKING_BAD_DATA_RECEIVED
			}

			loader := UseCryptosLoader()
			if loader == nil {
				loader = &cryptosLoaderType{}
				cryptosLoaderStorage = loader
			}

			if err := loader.parseJSON(body); err != nil {
				JC.Notify(JC.NotifyFailedToFetchCryptosData)
				return JC.NETWORKING_BAD_DATA_RECEIVED
			}

			if len(loader.Values) == 0 {
				JC.Logln("No cryptos found in the data")
				return JC.NETWORKING_BAD_DATA_RECEIVED
			}

			JC.CreateFile(JC.BuildPathRelatedToUserDirectory([]string{"cryptos.json"}), string(body))

			JC.Logln("Fetched cryptodata from CMC")
			JC.Notify(JC.NotifySuccessfullyRetrievedCryptosDataFromExch)
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

	if JC.IsMobile {
		UsePanelMaps().GetOptions()
	}
}

func UseCryptosLoader() *cryptosLoaderType {
	return cryptosLoaderStorage
}
