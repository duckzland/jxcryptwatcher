package types

import (
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/buger/jsonparser"

	JC "jxwatcher/core"
)

type marketCapFetcher struct {
	NowMarketCap        string
	YesterdayMarketCap  string
	ThirtyDaysChangePct string
	LastUpdate          time.Time
}

func (mc *marketCapFetcher) parseJSON(data []byte) error {

	nowBytes, _, _, err := jsonparser.Get(data, "data", "historicalValues", "now", "marketCap")
	if err != nil {
		JC.Logln("ParseJSON error: missing now marketCap:", err)
		return err
	}
	nowFloat, _ := strconv.ParseFloat(string(nowBytes), 64)
	mc.NowMarketCap = strconv.FormatFloat(nowFloat, 'f', -1, 64)

	yesterdayBytes, _, _, err := jsonparser.Get(data, "data", "historicalValues", "yesterday", "marketCap")
	if err != nil {
		JC.Logln("ParseJSON error: missing yesterday marketCap:", err)
		return err
	}
	yesterdayFloat, _ := strconv.ParseFloat(string(yesterdayBytes), 64)
	mc.YesterdayMarketCap = strconv.FormatFloat(yesterdayFloat, 'f', -1, 64)

	thirtyBytes, _, _, err := jsonparser.Get(data, "data", "thirtyDaysPercentage")
	if err != nil {
		JC.Logln("ParseJSON error: missing thirtyDaysPercentage:", err)
		return err
	}

	thirtyFloat, _ := strconv.ParseFloat(string(thirtyBytes), 64)
	mc.ThirtyDaysChangePct = strconv.FormatFloat(thirtyFloat, 'f', -1, 64)

	tsStr, err := jsonparser.GetString(data, "status", "timestamp")
	if err != nil {
		JC.Logln("ParseJSON error: missing timestamp:", err)
		mc.LastUpdate = time.Now()
		return err
	}
	parsedTime, err := time.Parse(time.RFC3339, tsStr)
	if err == nil {
		mc.LastUpdate = parsedTime
	} else {
		mc.LastUpdate = time.Now()
	}

	return nil
}

func (mc *marketCapFetcher) GetRate() int64 {
	return JC.GetRequest(
		UseConfig().MarketCapEndpoint,
		func(url url.Values, req *http.Request) {
			url.Add("convertId", "2781")
			url.Add("range", "30d")
		},
		func(resp *http.Response) int64 {
			body, err := JC.ReadResponse(resp.Body)
			if err != nil {
				return JC.NETWORKING_BAD_DATA_RECEIVED
			}

			if err := mc.parseJSON(body); err != nil {
				return JC.NETWORKING_BAD_DATA_RECEIVED
			}

			// Compute 24h change
			x, _ := strconv.ParseFloat(mc.NowMarketCap, 64)
			y, _ := strconv.ParseFloat(mc.YesterdayMarketCap, 64)
			dx := ((x - y) / y) * 100
			dif := strconv.FormatFloat(dx, 'f', -1, 64)

			tickerCacheStorage.Insert(TickerTypeMarketCap, mc.NowMarketCap, mc.LastUpdate)
			tickerCacheStorage.Insert(TickerTypeCMC10030dChange, mc.ThirtyDaysChangePct, mc.LastUpdate)
			tickerCacheStorage.Insert(TickerTypeMarketCap24hChange, dif, mc.LastUpdate)

			return JC.NETWORKING_SUCCESS
		})
}

func NewMarketCapFetcher() *marketCapFetcher {
	return &marketCapFetcher{}
}
