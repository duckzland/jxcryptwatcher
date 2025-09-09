package types

import (
	"net/http"
	"net/url"
	"strconv"
	"time"

	JC "jxwatcher/core"
)

type MarketCapFetcher struct {
	Data   *MarketCapHistoricalData  `json:"data"`
	Status MarketCapHistoricalStatus `json:"status"`
}

type MarketCapHistoricalData struct {
	HistoricalValues     MarketCapHistoricalValues `json:"historicalValues"`
	ThirtyDaysPercentage float64                   `json:"thirtyDaysPercentage"`
}

type MarketCapHistoricalStatus struct {
	LastUpdate time.Time `json:"timestamp"`
}

type MarketCapHistoricalValues struct {
	Now       MarketCapSnapshot `json:"now"`
	Yesterday MarketCapSnapshot `json:"yesterday"`
}

type MarketCapSnapshot struct {
	MarketCap float64 `json:"marketCap"`
}

func (er *MarketCapFetcher) GetRate() int64 {
	return JC.GetRequest(
		Config.MarketCapEndpoint,
		er,
		func(url url.Values, req *http.Request) {
			url.Add("convertId", "2781")
			url.Add("range", "30d")
		},
		func(cc any) int64 {
			dec, ok := cc.(*MarketCapFetcher)
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

			TickerCache.Insert("market_cap", now, dec.Status.LastUpdate)
			TickerCache.Insert("market_cap_30_percentage", dix, dec.Status.LastUpdate)
			TickerCache.Insert("market_cap_24_percentage", dif, dec.Status.LastUpdate)

			return JC.NETWORKING_SUCCESS
		})

}
