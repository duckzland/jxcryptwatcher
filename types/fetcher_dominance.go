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

type dominanceFetcher struct {
	DominanceBTC   string
	DominanceETC   string
	DominanceOther string
	LastUpdate     time.Time
}

func (df *dominanceFetcher) parseJSON(data []byte) error {

	btcBytes, _, _, err := jsonparser.Get(data, "data", "dominance", "[0]", "mcProportion")
	if err != nil {
		JC.Logln("ParseJSON error: missing BTC dominance:", err)
		return err
	}
	btcFloat, _ := strconv.ParseFloat(string(btcBytes), 64)
	df.DominanceBTC = strconv.FormatFloat(btcFloat, 'f', -1, 64)

	etcBytes, _, _, err := jsonparser.Get(data, "data", "dominance", "[1]", "mcProportion")
	if err != nil {
		JC.Logln("ParseJSON error: missing ETH dominance:", err)
		return err
	}
	etcFloat, _ := strconv.ParseFloat(string(etcBytes), 64)
	df.DominanceETC = strconv.FormatFloat(etcFloat, 'f', -1, 64)

	otherBytes, _, _, err := jsonparser.Get(data, "data", "dominance", "[2]", "mcProportion")
	if err != nil {
		JC.Logln("ParseJSON error: missing Other dominance:", err)
		return err
	}
	otherFloat, _ := strconv.ParseFloat(string(otherBytes), 64)
	df.DominanceOther = strconv.FormatFloat(otherFloat, 'f', -1, 64)

	// Timestamp
	tsStr, err := jsonparser.GetString(data, "status", "timestamp")
	if err != nil {
		JC.Logln("ParseJSON error: missing timestamp:", err)
		df.LastUpdate = time.Now()
		return err
	}
	parsedTime, err := time.Parse(time.RFC3339, tsStr)
	if err == nil {
		df.LastUpdate = parsedTime
	} else {
		df.LastUpdate = time.Now()
	}

	return nil
}

func (df *dominanceFetcher) GetRate() int64 {
	return JC.GetRequest(
		UseConfig().DominanceEndpoint,
		func(url url.Values, req *http.Request) {},
		func(resp *http.Response) int64 {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return JC.NETWORKING_BAD_DATA_RECEIVED
			}

			if err := df.parseJSON(body); err != nil {
				return JC.NETWORKING_BAD_DATA_RECEIVED
			}

			tickerCacheStorage.Insert(TickerTypeDominance, df.DominanceBTC, df.LastUpdate)
			tickerCacheStorage.Insert(TickerTypeETCDominance, df.DominanceETC, df.LastUpdate)
			tickerCacheStorage.Insert(TickerTypeOtherDominance, df.DominanceOther, df.LastUpdate)

			return JC.NETWORKING_SUCCESS
		})
}

func NewDominanceFetcher() *dominanceFetcher {
	return &dominanceFetcher{}
}
