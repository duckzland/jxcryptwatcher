package types

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/buger/jsonparser"

	JC "jxwatcher/core"
)

type cmc100Fetcher struct {
	Value         string
	PercentChange string
	NextUpdate    time.Time
}

func (er *cmc100Fetcher) parseJSON(data []byte) error {

	valFloat, err := jsonparser.GetFloat(data, "data", "summaryData", "currentValue", "value")
	if err != nil {
		JC.Logln("ParseJSON error: missing value:", err)
		return err
	}
	er.Value = strconv.FormatFloat(valFloat, 'f', -1, 64)

	changeFloat, err := jsonparser.GetFloat(data, "data", "summaryData", "currentValue", "percentChange")
	if err != nil {
		JC.Logln("ParseJSON error: missing percentChange:", err)
		return err
	}
	er.PercentChange = strconv.FormatFloat(changeFloat, 'f', -1, 64)

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

func (er *cmc100Fetcher) GetRate(ctx context.Context, payload any) int64 {

	if ctx.Err() != nil {
		return JC.NETWORKING_ERROR_CONNECTION
	}

	return JC.GetRequest(
		ctx,
		UseConfig().CMC100Endpoint,
		func(url url.Values, req *http.Request) {
			startUnix, endUnix := JC.GetMonthBounds(time.Now())
			url.Add("start", strconv.FormatInt(startUnix, 10))
			url.Add("end", strconv.FormatInt(endUnix, 10))
		},
		func(cctx context.Context, resp *http.Response) int64 {

			if cctx.Err() != nil {
				return JC.NETWORKING_ERROR_CONNECTION
			}

			body, close, err := JC.ReadResponse(JC.ACT_TICKER_GET_CMC100, resp, 2)
			defer close()
			if err != nil {
				return JC.NETWORKING_BAD_DATA_RECEIVED
			}

			if err := er.parseJSON(body); err != nil {
				return JC.NETWORKING_BAD_DATA_RECEIVED
			}

			tickerCacheStorage.Insert(TickerTypeCMC100, er.Value, er.NextUpdate)
			tickerCacheStorage.Insert(TickerTypeCMC10024hChange, er.PercentChange, er.NextUpdate)

			return JC.NETWORKING_SUCCESS
		})
}

func NewCMC100Fetcher() *cmc100Fetcher {
	return &cmc100Fetcher{}
}
