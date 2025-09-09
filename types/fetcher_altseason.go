package types

import (
	"net/http"
	"net/url"
	"strconv"
	"time"

	JC "jxwatcher/core"
)

type AltSeasonFetcher struct {
	Data *AltSeasonHistoricalData `json:"data"`
}

type AltSeasonHistoricalData struct {
	HistoricalValues AltSeasonHistoricalValues `json:"historicalValues"`
}

type AltSeasonHistoricalValues struct {
	Now AltSeasonSnapshot `json:"now"`
}

type AltSeasonSnapshot struct {
	AltcoinIndex string    `json:"altcoinIndex"`
	TimestampRaw string    `json:"timestamp"`
	LastUpdate   time.Time `json:"-"`
}

func (er *AltSeasonFetcher) GetRate() int64 {

	return JC.GetRequest(
		Config.AltSeasonEndpoint,
		er,
		func(url url.Values, req *http.Request) {
			startUnix, endUnix := JC.GetMonthBounds(time.Now())
			url.Add("start", strconv.FormatInt(startUnix, 10))
			url.Add("end", strconv.FormatInt(endUnix, 10))
		},
		func(cc any) int64 {
			dec, ok := cc.(*AltSeasonFetcher)
			if !ok {
				return JC.NETWORKING_BAD_DATA_RECEIVED
			}

			if dec.Data == nil {
				return JC.NETWORKING_BAD_DATA_RECEIVED
			}

			ts, err := strconv.ParseInt(dec.Data.HistoricalValues.Now.TimestampRaw, 10, 64)

			if err == nil {
				dec.Data.HistoricalValues.Now.LastUpdate = time.Unix(ts, 0)
			} else {
				dec.Data.HistoricalValues.Now.LastUpdate = time.Now()
			}

			TickerCache.Insert("altcoin_index", dec.Data.HistoricalValues.Now.AltcoinIndex, dec.Data.HistoricalValues.Now.LastUpdate)

			return JC.NETWORKING_SUCCESS
		})
}
