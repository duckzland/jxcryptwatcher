package types

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/buger/jsonparser"

	json "github.com/goccy/go-json"

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

func (er *altSeasonFetcher) sanitizeJSON(r io.ReadCloser) (io.ReadCloser, error) {
	dec := json.NewDecoder(r)

	var raw map[string]json.RawMessage
	if err := dec.Decode(&raw); err != nil {
		return nil, err
	}

	sanitized := map[string]json.RawMessage{}

	if v, ok := raw["data"]; ok {
		sanitized["data"] = v
	}

	cleanBytes, err := json.Marshal(sanitized)
	if err != nil {
		return nil, err
	}

	return io.NopCloser(bytes.NewReader(cleanBytes)), nil
}

func (er *altSeasonFetcher) GetRate() int64 {
	return JC.GetRequest(
		UseConfig().AltSeasonEndpoint,
		func(url url.Values, req *http.Request) {
			startUnix, endUnix := JC.GetMonthBounds(time.Now())
			url.Add("start", strconv.FormatInt(startUnix, 10))
			url.Add("end", strconv.FormatInt(endUnix, 10))
		},
		func(resp *http.Response) int64 {

			sanitizedBody, err := er.sanitizeJSON(resp.Body)
			if err != nil {
				return JC.NETWORKING_BAD_DATA_RECEIVED
			}
			resp.Body.Close()
			resp.Body = sanitizedBody

			body, close, err := JC.ReadResponse(resp.Body)
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
