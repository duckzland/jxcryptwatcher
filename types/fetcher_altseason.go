package types

import (
	"net/http"
	"net/url"
	"strconv"
	"time"

	JC "jxwatcher/core"
)

type altSeasonFetcher struct {
	Data *altSeasonHistoricalData `json:"data"`
}

type altSeasonHistoricalData struct {
	HistoricalValues altSeasonHistoricalValues `json:"historicalValues"`
}

type altSeasonHistoricalValues struct {
	Now altSeasonSnapshot `json:"now"`
}

type altSeasonSnapshot struct {
	AltcoinIndex string    `json:"altcoinIndex"`
	TimestampRaw string    `json:"timestamp"`
	LastUpdate   time.Time `json:"-"`
}

func (er *altSeasonFetcher) GetRate() int64 {

	return JC.GetRequest(
		UseConfig().AltSeasonEndpoint,
		er,
		func(url url.Values, req *http.Request) {
			startUnix, endUnix := JC.GetMonthBounds(time.Now())
			url.Add("start", strconv.FormatInt(startUnix, 10))
			url.Add("end", strconv.FormatInt(endUnix, 10))
		},
		func(cc any) int64 {
			dec, ok := cc.(*altSeasonFetcher)
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

func NewAltSeasonFetcher() *altSeasonFetcher {
	return &altSeasonFetcher{}
}
