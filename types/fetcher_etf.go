package types

import (
	"net/http"
	"net/url"
	"strconv"
	"time"

	JC "jxwatcher/core"
)

type etfFetcher struct {
	Data   etfData   `json:"data"`
	Status etfStatus `json:"status"`
}

type etfData struct {
	Total         int64 `json:"total"`
	TotalBtcValue int64 `json:"totalBtcValue"`
	TotalEthValue int64 `json:"totalEthValue"`
}

type etfStatus struct {
	Timestamp string `json:"timestamp"`
}

func (ef *etfFetcher) GetRate() int64 {
	return JC.GetRequest(
		UseConfig().ETFEndpoint,
		ef,
		func(url url.Values, req *http.Request) {
			url.Add("category", "all")
			url.Add("range", "30d")
		},
		func(resp *http.Response, cc any) int64 {
			dec, ok := cc.(*etfFetcher)
			if !ok {
				return JC.NETWORKING_BAD_DATA_RECEIVED
			}

			parsedTime, err := time.Parse(time.RFC3339, dec.Status.Timestamp)
			if err != nil {
				parsedTime = time.Now()
			}

			tickerCacheStorage.Insert("etf",
				strconv.FormatInt(dec.Data.Total, 10),
				parsedTime)

			tickerCacheStorage.Insert("etf_btc",
				strconv.FormatInt(dec.Data.TotalBtcValue, 10),
				parsedTime)

			tickerCacheStorage.Insert("etf_eth",
				strconv.FormatInt(dec.Data.TotalEthValue, 10),
				parsedTime)

			return JC.NETWORKING_SUCCESS
		})
}

func NewETFFetcher() *etfFetcher {
	return &etfFetcher{}
}
