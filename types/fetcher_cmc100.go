package types

import (
	"net/http"
	"net/url"
	"strconv"
	"time"

	JC "jxwatcher/core"
)

type cmc100Fetcher struct {
	Data *cmc100SummaryData `json:"data"`
}

type cmc100SummaryData struct {
	SummaryData cmc100SummaryDataFields `json:"summaryData"`
}

type cmc100SummaryDataFields struct {
	NextUpdate   string                   `json:"nextUpdateTimestamp"`
	CurrentValue cmc100CurrentValueFields `json:"currentValue"`
}

type cmc100CurrentValueFields struct {
	Value         float64 `json:"value"`
	PercentChange float64 `json:"percentChange"`
}

func (er *cmc100Fetcher) GetRate() int64 {

	return JC.GetRequest(
		Config.CMC100Endpoint,
		er,
		func(url url.Values, req *http.Request) {
			startUnix, endUnix := JC.GetMonthBounds(time.Now())
			url.Add("start", strconv.FormatInt(startUnix, 10))
			url.Add("end", strconv.FormatInt(endUnix, 10))
		},
		func(cc any) int64 {
			dec, ok := cc.(*cmc100Fetcher)
			if !ok {
				return JC.NETWORKING_BAD_DATA_RECEIVED
			}

			if dec.Data == nil {
				return JC.NETWORKING_BAD_DATA_RECEIVED
			}

			now := strconv.FormatFloat(dec.Data.SummaryData.CurrentValue.Value, 'f', -1, 64)
			dif := strconv.FormatFloat(dec.Data.SummaryData.CurrentValue.PercentChange, 'f', -1, 64)
			tim := dec.Data.SummaryData.NextUpdate

			ts, err := strconv.ParseInt(tim, 10, 64)

			var nextUpdate time.Time = time.Now()
			if err == nil {
				nextUpdate = time.Unix(ts, 0)
			}

			TickerCache.Insert("cmc100", now, nextUpdate)
			TickerCache.Insert("cmc100_24_percentage", dif, nextUpdate)

			return JC.NETWORKING_SUCCESS
		})
}

func NewCMC100Fetcher() *cmc100Fetcher {
	return &cmc100Fetcher{}
}
