package types

import (
	"net/http"
	"net/url"
	"strconv"
	"time"

	JC "jxwatcher/core"
)

type marketCapFetcher struct {
	Data   *marketCapHistoricalData  `json:"data"`
	Status marketCapHistoricalStatus `json:"status"`
}

type marketCapHistoricalData struct {
	HistoricalValues     marketCapHistoricalValues `json:"historicalValues"`
	ThirtyDaysPercentage float64                   `json:"thirtyDaysPercentage"`
}

type marketCapHistoricalStatus struct {
	LastUpdate time.Time `json:"timestamp"`
}

type marketCapHistoricalValues struct {
	Now       marketCapSnapshot `json:"now"`
	Yesterday marketCapSnapshot `json:"yesterday"`
}

type marketCapSnapshot struct {
	MarketCap float64 `json:"marketCap"`
}

func (er *marketCapFetcher) GetRate() int64 {
	return JC.GetRequest(
		UseConfig().MarketCapEndpoint,
		er,
		func(url url.Values, req *http.Request) {
			url.Add("convertId", "2781")
			url.Add("range", "30d")
		},
		func(resp *http.Response, cc any) int64 {
			dec, ok := cc.(*marketCapFetcher)
			if !ok {
				return JC.NETWORKING_BAD_DATA_RECEIVED
			}

			if dec.Data == nil {
				return JC.NETWORKING_BAD_DATA_RECEIVED
			}

			x := dec.Data.HistoricalValues.Now.MarketCap
			y := dec.Data.HistoricalValues.Yesterday.MarketCap
			z := dec.Data.ThirtyDaysPercentage

			dx := ((x - y) / y) * 100
			now := strconv.FormatFloat(x, 'f', -1, 64)
			dif := strconv.FormatFloat(dx, 'f', -1, 64)
			dix := strconv.FormatFloat(z, 'f', -1, 64)

			tickerCacheStorage.Insert(TickerTypeMarketCap, now, dec.Status.LastUpdate)
			tickerCacheStorage.Insert(TickerTypeCMC10030dChange, dix, dec.Status.LastUpdate)
			tickerCacheStorage.Insert(TickerTypeMarketCap24hChange, dif, dec.Status.LastUpdate)

			return JC.NETWORKING_SUCCESS
		})

}

func NewMarketCapFetcher() *marketCapFetcher {
	return &marketCapFetcher{}
}
