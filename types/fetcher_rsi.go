package types

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	JC "jxwatcher/core"
)

type rsiFetcher struct {
	Data   *rsiData   `json:"data"`
	Status *rsiStatus `json:"status"`
}

type rsiData struct {
	Overall rsiOverall `json:"overall"`
}

type rsiOverall struct {
	AverageRSI           float64 `json:"averageRsi"`
	OverboughtPercentage float64 `json:"overboughtPercentage"`
	OversoldPercentage   float64 `json:"oversoldPercentage"`
	NeutralPercentage    float64 `json:"neutralPercentage"`
}

type rsiStatus struct {
	Timestamp time.Time `json:"timestamp"`
}

func (er *rsiFetcher) GetRate() int64 {

	return JC.GetRequest(
		UseConfig().RSIEndpoint,
		er,
		func(url url.Values, req *http.Request) {
			url.Add("timeframe", "4h")
			url.Add("rsiPeriod", "14")
			url.Add("volume24Range.min", "1000000")
			url.Add("marketCapRange.min", "50000000")
		},
		func(cc any) int64 {
			dec, ok := cc.(*rsiFetcher)
			if !ok {
				return JC.NETWORKING_BAD_DATA_RECEIVED
			}

			if dec.Data == nil || dec.Status == nil {
				return JC.NETWORKING_BAD_DATA_RECEIVED
			}

			rsi := strconv.FormatFloat(dec.Data.Overall.AverageRSI, 'f', -1, 64)
			sp := strconv.FormatFloat(dec.Data.Overall.OversoldPercentage, 'f', -1, 64)
			bp := strconv.FormatFloat(dec.Data.Overall.OverboughtPercentage, 'f', -1, 64)
			np := strconv.FormatFloat(dec.Data.Overall.NeutralPercentage, 'f', -1, 64)
			ts := dec.Status.Timestamp
			ne := dec.Data.Overall.OverboughtPercentage - dec.Data.Overall.OversoldPercentage

			tickerCacheStorage.Insert("rsi", rsi, ts)
			tickerCacheStorage.Insert("pulse", fmt.Sprintf("%+.2f%%", ne), ts)
			tickerCacheStorage.Insert("rsi_oversold_percentage", sp, ts)
			tickerCacheStorage.Insert("rsi_overbought_precentage", bp, ts)
			tickerCacheStorage.Insert("rsi_neutral_percentage", np, ts)

			return JC.NETWORKING_SUCCESS
		})
}

func NewRSIFetcher() *rsiFetcher {
	return &rsiFetcher{}
}
