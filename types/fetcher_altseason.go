package types

import (
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/buger/jsonparser"

	JC "jxwatcher/core"
)

type altSeasonFetcher struct {
	Index      string
	LastUpdate time.Time
}

func (er *altSeasonFetcher) ParseJSON(data []byte) error {
	// Extract altcoin index
	index, err := jsonparser.GetString(data, "data", "historicalValues", "now", "altcoinIndex")
	if err != nil {
		JC.Logln("ParseJSON error: missing altcoinIndex:", err)
		return err
	}
	er.Index = index

	// Extract timestamp (raw string, numeric)
	tsRaw, err := jsonparser.GetString(data, "data", "historicalValues", "now", "timestamp")
	if err != nil {
		JC.Logln("ParseJSON error: missing timestamp:", err)
		er.LastUpdate = time.Now()
		return err
	}

	tsInt, err := strconv.ParseInt(tsRaw, 10, 64)
	if err == nil {
		er.LastUpdate = time.Unix(tsInt, 0)
	} else {
		er.LastUpdate = time.Now()
	}

	return nil
}

func (er *altSeasonFetcher) GetRate() int64 {
	return JC.GetRequest(
		UseConfig().AltSeasonEndpoint,
		nil, // don’t auto‑decode, we’ll parse manually
		func(url url.Values, req *http.Request) {
			startUnix, endUnix := JC.GetMonthBounds(time.Now())
			url.Add("start", strconv.FormatInt(startUnix, 10))
			url.Add("end", strconv.FormatInt(endUnix, 10))
		},
		func(resp *http.Response, cc any) int64 {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return JC.NETWORKING_BAD_DATA_RECEIVED
			}

			if err := er.ParseJSON(body); err != nil {
				return JC.NETWORKING_BAD_DATA_RECEIVED
			}

			// Insert into cache
			tickerCacheStorage.Insert(TickerTypeAltcoinIndex, er.Index, er.LastUpdate)

			return JC.NETWORKING_SUCCESS
		})
}

func NewAltSeasonFetcher() *altSeasonFetcher {
	return &altSeasonFetcher{}
}
