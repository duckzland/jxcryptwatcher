package types

import (
	"net/http"
	"net/url"
	"strconv"
	"time"

	JC "jxwatcher/core"
)

type fearGreedFetcher struct {
	Data *fearGreedHistoricalData `json:"data"`
}

type fearGreedHistoricalData struct {
	HistoricalValues fearGreedHistoricalValues `json:"historicalValues"`
}

type fearGreedHistoricalValues struct {
	Now fearGreedSnapshot `json:"now"`
}

type fearGreedSnapshot struct {
	Score        int64     `json:"score"`
	TimestampRaw string    `json:"timestamp"`
	LastUpdate   time.Time `json:"-"`
}

func (er *fearGreedFetcher) GetRate() int64 {

	return JC.GetRequest(
		UseConfig().FearGreedEndpoint,
		er,
		func(url url.Values, req *http.Request) {
			startUnix, endUnix := JC.GetMonthBounds(time.Now())
			url.Add("start", strconv.FormatInt(startUnix, 10))
			url.Add("end", strconv.FormatInt(endUnix, 10))
		},
		func(cc any) int64 {
			dec, ok := cc.(*fearGreedFetcher)
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

			if dec.Data == nil {
				return JC.NETWORKING_BAD_DATA_RECEIVED
			}

			ms := strconv.FormatInt(dec.Data.HistoricalValues.Now.Score, 10)

			tickerCacheStorage.Insert("feargreed", ms, dec.Data.HistoricalValues.Now.LastUpdate)

			return JC.NETWORKING_SUCCESS
		})
}

func NewFearGreedFetcher() *fearGreedFetcher {
	return &fearGreedFetcher{}
}
