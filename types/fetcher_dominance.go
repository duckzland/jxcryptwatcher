package types

import (
	"net/http"
	"net/url"
	"strconv"
	"time"

	JC "jxwatcher/core"
)

type dominanceFetcher struct {
	Data   dominanceData   `json:"data"`
	Status dominanceStatus `json:"status"`
}

type dominanceData struct {
	Dominance []dominanceEntry `json:"dominance"`
}

type dominanceEntry struct {
	MCProportion float64 `json:"mcProportion"`
}

type dominanceStatus struct {
	Timestamp string `json:"timestamp"`
}

func (df *dominanceFetcher) GetRate() int64 {
	return JC.GetRequest(
		UseConfig().DominanceEndpoint,
		df,
		func(url url.Values, req *http.Request) {},
		func(resp *http.Response, cc any) int64 {
			dec, ok := cc.(*dominanceFetcher)
			if !ok || len(dec.Data.Dominance) < 3 {
				return JC.NETWORKING_BAD_DATA_RECEIVED
			}

			parsedTime, err := time.Parse(time.RFC3339, dec.Status.Timestamp)
			if err != nil {
				parsedTime = time.Now()
			}

			tickerCacheStorage.Insert("dominance",
				strconv.FormatFloat(dec.Data.Dominance[0].MCProportion, 'f', -1, 64),
				parsedTime)

			tickerCacheStorage.Insert("etc_dominance",
				strconv.FormatFloat(dec.Data.Dominance[1].MCProportion, 'f', -1, 64),
				parsedTime)

			tickerCacheStorage.Insert("other_dominance",
				strconv.FormatFloat(dec.Data.Dominance[2].MCProportion, 'f', -1, 64),
				parsedTime)

			return JC.NETWORKING_SUCCESS
		})
}

func NewDominanceFetcher() *dominanceFetcher {
	return &dominanceFetcher{}
}
