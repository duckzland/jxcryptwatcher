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

type fearGreedFetcher struct {
	Score      string
	LastUpdate time.Time
}

func (fg *fearGreedFetcher) parseJSON(data []byte) error {

	scoreBytes, _, _, err := jsonparser.Get(data, "data", "historicalValues", "now", "score")
	if err != nil {
		JC.Logln("ParseJSON error: missing score:", err)
		return err
	}
	scoreInt, _ := strconv.ParseInt(string(scoreBytes), 10, 64)
	fg.Score = strconv.FormatInt(scoreInt, 10)

	tsStr, err := jsonparser.GetString(data, "data", "historicalValues", "now", "timestamp")
	if err != nil {
		JC.Logln("ParseJSON error: missing timestamp:", err)
		fg.LastUpdate = time.Now()
		return err
	}
	tsInt, err := strconv.ParseInt(tsStr, 10, 64)
	if err == nil {
		fg.LastUpdate = time.Unix(tsInt, 0)
	} else {
		fg.LastUpdate = time.Now()
	}

	return nil
}

func (fg *fearGreedFetcher) sanitizeJSON(r io.ReadCloser) (io.ReadCloser, error) {
	dec := json.NewDecoder(r)

	var raw map[string]json.RawMessage
	if err := dec.Decode(&raw); err != nil {
		return nil, err
	}

	sanitized := map[string]json.RawMessage{}

	if data, ok := raw["data"]; ok {
		var dataObj map[string]json.RawMessage
		if err := json.Unmarshal(data, &dataObj); err != nil {
			return nil, err
		}
		if hv, ok := dataObj["historicalValues"]; ok {
			var hvObj map[string]json.RawMessage
			if err := json.Unmarshal(hv, &hvObj); err != nil {
				return nil, err
			}
			if now, ok := hvObj["now"]; ok {
				sanitized["data"] = json.RawMessage(`{"historicalValues":{"now":` + string(now) + `}}`)
			}
		}
	}

	cleanBytes, err := json.Marshal(sanitized)
	if err != nil {
		return nil, err
	}

	return io.NopCloser(bytes.NewReader(cleanBytes)), nil
}

func (fg *fearGreedFetcher) GetRate() int64 {
	return JC.GetRequest(
		UseConfig().FearGreedEndpoint,
		func(url url.Values, req *http.Request) {
			startUnix, endUnix := JC.GetMonthBounds(time.Now())
			url.Add("start", strconv.FormatInt(startUnix, 10))
			url.Add("end", strconv.FormatInt(endUnix, 10))
		},
		func(resp *http.Response) int64 {
			sanitizedBody, err := fg.sanitizeJSON(resp.Body)
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

			if err := fg.parseJSON(body); err != nil {
				return JC.NETWORKING_BAD_DATA_RECEIVED
			}

			tickerCacheStorage.Insert(TickerTypeFearGreed, fg.Score, fg.LastUpdate)

			return JC.NETWORKING_SUCCESS
		})
}

func NewFearGreedFetcher() *fearGreedFetcher {
	return &fearGreedFetcher{}
}
