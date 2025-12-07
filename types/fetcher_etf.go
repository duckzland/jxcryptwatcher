package types

import (
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/buger/jsonparser"

	JC "jxwatcher/core"
)

type etfFetcher struct {
	Total         string
	TotalBtcValue string
	TotalEthValue string
	LastUpdate    time.Time
}

func (ef *etfFetcher) parseJSON(data []byte) error {
	totalBytes, _, _, err := jsonparser.Get(data, "data", "total")
	if err != nil {
		JC.Logln("ParseJSON error: missing total:", err)
		return err
	}

	totalInt, _ := strconv.ParseInt(string(totalBytes), 10, 64)
	ef.Total = strconv.FormatInt(totalInt, 10)

	btcBytes, _, _, err := jsonparser.Get(data, "data", "totalBtcValue")
	if err != nil {
		JC.Logln("ParseJSON error: missing totalBtcValue:", err)
		return err
	}
	btcInt, _ := strconv.ParseInt(string(btcBytes), 10, 64)
	ef.TotalBtcValue = strconv.FormatInt(btcInt, 10)

	// Total ETH value
	ethBytes, _, _, err := jsonparser.Get(data, "data", "totalEthValue")
	if err != nil {
		JC.Logln("ParseJSON error: missing totalEthValue:", err)
		return err
	}
	ethInt, _ := strconv.ParseInt(string(ethBytes), 10, 64)
	ef.TotalEthValue = strconv.FormatInt(ethInt, 10)

	tsStr, err := jsonparser.GetString(data, "status", "timestamp")
	if err != nil {
		JC.Logln("ParseJSON error: missing timestamp:", err)
		ef.LastUpdate = time.Now()
		return err
	}

	parsedTime, err := time.Parse(time.RFC3339, tsStr)
	if err == nil {
		ef.LastUpdate = parsedTime
	} else {
		ef.LastUpdate = time.Now()
	}

	return nil
}

func (ef *etfFetcher) GetRate() int64 {
	return JC.GetRequest(
		UseConfig().ETFEndpoint,
		func(url url.Values, req *http.Request) {
			url.Add("category", "all")
			url.Add("range", "30d")
		},
		func(resp *http.Response) int64 {

			body, close, err := JC.ReadResponse(resp.Body)
			defer close()
			if err != nil {
				return JC.NETWORKING_BAD_DATA_RECEIVED
			}

			if err := ef.parseJSON(body); err != nil {
				return JC.NETWORKING_BAD_DATA_RECEIVED
			}

			tickerCacheStorage.Insert(TickerTypeETF, ef.Total, ef.LastUpdate)
			tickerCacheStorage.Insert(TickerTypeETFBTC, ef.TotalBtcValue, ef.LastUpdate)
			tickerCacheStorage.Insert(TickerTypeETF, ef.TotalEthValue, ef.LastUpdate)

			return JC.NETWORKING_SUCCESS
		})
}

func NewETFFetcher() *etfFetcher {
	return &etfFetcher{}
}
