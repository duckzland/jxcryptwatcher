package types

import (
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	JC "jxwatcher/core"

	"github.com/buger/jsonparser"
)

type cmc100Fetcher struct {
	Value         string
	PercentChange string
	NextUpdate    time.Time
}

func (er *cmc100Fetcher) ParseJSON(data []byte) error {
	// Extract current value as float
	valFloat, err := jsonparser.GetFloat(data, "data", "summaryData", "currentValue", "value")
	if err != nil {
		JC.Logln("ParseJSON error: missing value:", err)
		return err
	}
	er.Value = strconv.FormatFloat(valFloat, 'f', -1, 64)

	// Extract percent change as float
	changeFloat, err := jsonparser.GetFloat(data, "data", "summaryData", "currentValue", "percentChange")
	if err != nil {
		JC.Logln("ParseJSON error: missing percentChange:", err)
		return err
	}
	er.PercentChange = strconv.FormatFloat(changeFloat, 'f', -1, 64)

	// Extract next update timestamp
	timStr, err := jsonparser.GetString(data, "data", "summaryData", "nextUpdateTimestamp")
	if err != nil {
		JC.Logln("ParseJSON error: missing nextUpdateTimestamp:", err)
		er.NextUpdate = time.Now()
		return err
	}

	ts, err := strconv.ParseInt(timStr, 10, 64)
	if err == nil {
		er.NextUpdate = time.Unix(ts, 0)
	} else {
		er.NextUpdate = time.Now()
	}

	return nil
}

func (er *cmc100Fetcher) GetRate() int64 {
	return JC.GetRequest(
		UseConfig().CMC100Endpoint,
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
			tickerCacheStorage.Insert(TickerTypeCMC100, er.Value, er.NextUpdate)
			tickerCacheStorage.Insert(TickerTypeCMC10024hChange, er.PercentChange, er.NextUpdate)

			return JC.NETWORKING_SUCCESS
		})
}

func NewCMC100Fetcher() *cmc100Fetcher {
	return &cmc100Fetcher{}
}
