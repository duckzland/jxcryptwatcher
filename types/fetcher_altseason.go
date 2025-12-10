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

type altSeasonFetcher struct {
	Index      string
	LastUpdate time.Time
}

func (er *altSeasonFetcher) parseJSON(data []byte) error {

	index, err := jsonparser.GetString(data, "data", "historicalValues", "now", "altcoinIndex")
	if err != nil {
		JC.Logln("ParseJSON error: missing altcoinIndex:", err)
		return err
	}
	er.Index = index

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

func (er *altSeasonFetcher) GetRate(ctx context.Context) int64 {
	if ctx.Err() != nil {
		return JC.NETWORKING_ERROR_CONNECTION
	}

	return JC.GetRequest(
		ctx,
		UseConfig().AltSeasonEndpoint,
		func(url url.Values, req *http.Request) {
			startUnix, endUnix := JC.GetMonthBounds(time.Now())
			url.Add("start", strconv.FormatInt(startUnix, 10))
			url.Add("end", strconv.FormatInt(endUnix, 10))
		},
		func(cctx context.Context, resp *http.Response) int64 {

			if cctx.Err() != nil {
				return JC.NETWORKING_ERROR_CONNECTION
			}

			body, close, err := JC.ReadResponse(JC.ACT_TICKER_GET_ALTSEASON, resp)
			defer close()
			if err != nil {
				return JC.NETWORKING_BAD_DATA_RECEIVED
			}

			if err := er.parseJSON(body); err != nil {
				return JC.NETWORKING_BAD_DATA_RECEIVED
			}

			tickerCacheStorage.Insert(TickerTypeAltcoinIndex, er.Index, er.LastUpdate)

			return JC.NETWORKING_SUCCESS
		})
}

func NewAltSeasonFetcher() *altSeasonFetcher {
	return &altSeasonFetcher{}
}
