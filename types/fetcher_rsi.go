package types

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/buger/jsonparser"

	JC "jxwatcher/core"
)

type rsiFetcher struct {
	AverageRSI           string
	OverboughtPercentage string
	OversoldPercentage   string
	NeutralPercentage    string
	LastUpdate           time.Time
}

func (rf *rsiFetcher) parseJSON(data []byte) error {

	rsiBytes, _, _, err := jsonparser.Get(data, "data", "overall", "averageRsi")
	if err != nil {
		JC.Logln("ParseJSON error: missing averageRsi:", err)
		return err
	}
	rsiFloat, _ := strconv.ParseFloat(string(rsiBytes), 64)
	rf.AverageRSI = strconv.FormatFloat(rsiFloat, 'f', -1, 64)

	bpBytes, _, _, err := jsonparser.Get(data, "data", "overall", "overboughtPercentage")
	if err != nil {
		JC.Logln("ParseJSON error: missing overboughtPercentage:", err)
		return err
	}
	bpFloat, _ := strconv.ParseFloat(string(bpBytes), 64)
	rf.OverboughtPercentage = strconv.FormatFloat(bpFloat, 'f', -1, 64)

	spBytes, _, _, err := jsonparser.Get(data, "data", "overall", "oversoldPercentage")
	if err != nil {
		JC.Logln("ParseJSON error: missing oversoldPercentage:", err)
		return err
	}
	spFloat, _ := strconv.ParseFloat(string(spBytes), 64)
	rf.OversoldPercentage = strconv.FormatFloat(spFloat, 'f', -1, 64)

	npBytes, _, _, err := jsonparser.Get(data, "data", "overall", "neutralPercentage")
	if err != nil {
		JC.Logln("ParseJSON error: missing neutralPercentage:", err)
		return err
	}
	npFloat, _ := strconv.ParseFloat(string(npBytes), 64)
	rf.NeutralPercentage = strconv.FormatFloat(npFloat, 'f', -1, 64)

	tsStr, err := jsonparser.GetString(data, "status", "timestamp")
	if err != nil {
		JC.Logln("ParseJSON error: missing timestamp:", err)
		rf.LastUpdate = time.Now()
		return err
	}
	parsedTime, err := time.Parse(time.RFC3339, tsStr)
	if err == nil {
		rf.LastUpdate = parsedTime
	} else {
		rf.LastUpdate = time.Now()
	}

	return nil
}

func (rf *rsiFetcher) GetRate() int64 {
	return JC.GetRequest(
		UseConfig().RSIEndpoint,
		func(url url.Values, req *http.Request) {
			url.Add("timeframe", "4h")
			url.Add("rsiPeriod", "14")
			url.Add("volume24Range.min", "1000000")
			url.Add("marketCapRange.min", "50000000")
		},
		func(resp *http.Response) int64 {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return JC.NETWORKING_BAD_DATA_RECEIVED
			}

			if err := rf.parseJSON(body); err != nil {
				return JC.NETWORKING_BAD_DATA_RECEIVED
			}

			// Compute pulse (overbought - oversold)
			bp, _ := strconv.ParseFloat(rf.OverboughtPercentage, 64)
			sp, _ := strconv.ParseFloat(rf.OversoldPercentage, 64)
			ne := bp - sp

			// ✅ Same insertion logic as old code
			tickerCacheStorage.Insert(TickerTypeRSI, rf.AverageRSI, rf.LastUpdate)
			tickerCacheStorage.Insert(TickerTypePulse, fmt.Sprintf("%+.2f%%", ne), rf.LastUpdate)
			tickerCacheStorage.Insert(TickerTypeRSIOversold, rf.OversoldPercentage, rf.LastUpdate)
			tickerCacheStorage.Insert(TickerTypeRSIOverbought, rf.OverboughtPercentage, rf.LastUpdate)
			tickerCacheStorage.Insert(TickerTypeRSINeutral, rf.NeutralPercentage, rf.LastUpdate)

			return JC.NETWORKING_SUCCESS
		})
}

func NewRSIFetcher() *rsiFetcher {
	return &rsiFetcher{}
}
