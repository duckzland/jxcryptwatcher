package types

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	JC "jxwatcher/core"
)

type ExchangeResults struct {
	Rates []ExchangeDataType
}

func (er *ExchangeResults) UnmarshalJSON(data []byte) error {

	var v map[string]any
	err := json.Unmarshal(data, &v)
	if err != nil {
		return err
	}

	if er.validateData(v) == false {
		return nil
	}

	sc := v["data"]
	tc := v["data"].(map[string]any)["quote"].([]any)

	for _, rate := range tc {

		if er.validateRate(rate) == false {
			continue
		}

		ex := ExchangeDataType{}

		// CMC Json data is weird the the id is in string while cryptoId is in int64 (but golang cast this as float64)
		ex.SourceSymbol = sc.(map[string]any)["symbol"].(string)
		ex.SourceId, _ = strconv.ParseInt(sc.(map[string]any)["id"].(string), 10, 64)
		ex.SourceAmount = sc.(map[string]any)["amount"].(float64)

		ex.TargetSymbol = rate.(map[string]any)["symbol"].(string)
		ex.TargetId = int64(rate.(map[string]any)["cryptoId"].(float64))
		ex.TargetAmount = rate.(map[string]any)["price"].(float64)

		er.Rates = append(er.Rates, ex)
	}

	return nil
}

func (er *ExchangeResults) validateData(v map[string]any) bool {

	if _, ok := v["data"]; !ok {
		JC.Logln("Missing 'data' field in exchange results")
		return false
	}

	if _, ok := v["data"].(map[string]any); !ok {
		JC.Logln("Invalid 'data' field format in exchange results")
		return false
	}

	if _, ok := v["data"].(map[string]any)["symbol"]; !ok {
		JC.Logln("Missing 'symbol' field in 'data'")
		return false
	}

	if _, ok := v["data"].(map[string]any)["symbol"].(string); !ok {
		JC.Logln("Invalid 'symbol' field type in 'data'")
		return false
	}

	if _, ok := v["data"].(map[string]any)["id"]; !ok {
		JC.Logln("Missing 'id' field in 'data'")
		return false
	}

	if _, ok := v["data"].(map[string]any)["id"].(string); !ok {
		JC.Logln("Invalid 'id' field type in 'data'")
		return false
	}

	if _, ok := v["data"].(map[string]any)["amount"]; !ok {
		JC.Logln("Missing 'amount' field in 'data'")
		return false
	}

	if _, ok := v["data"].(map[string]any)["amount"].(float64); !ok {
		JC.Logln("Invalid 'amount' field type in 'data'")
		return false
	}

	if _, ok := v["data"].(map[string]any)["quote"]; !ok {
		JC.Logln("Missing 'quote' field in 'data'")
		return false
	}

	if _, ok := v["data"].(map[string]any)["quote"].([]any); !ok {
		JC.Logln("Invalid 'quote' field type in 'data'")
		return false
	}

	return true
}

func (er *ExchangeResults) validateRate(rate any) bool {
	if _, ok := rate.(map[string]any); !ok {
		JC.Logln("Invalid rate format:", rate)
		return false
	}

	if _, ok := rate.(map[string]any)["symbol"]; !ok {
		JC.Logln("Missing symbol in rate:", rate)
		return false
	}

	if _, ok := rate.(map[string]any)["symbol"].(string); !ok {
		JC.Logln("Invalid symbol type in rate:", rate)
		return false
	}

	if _, ok := rate.(map[string]any)["cryptoId"]; !ok {
		JC.Logln("Missing cryptoId in rate:", rate)
		return false
	}

	if _, ok := rate.(map[string]any)["cryptoId"].(float64); !ok {
		JC.Logln("Invalid cryptoId type in rate:", rate)
		return false
	}

	if _, ok := rate.(map[string]any)["price"]; !ok {
		JC.Logln("Missing price in rate:", rate)
		return false
	}

	if _, ok := rate.(map[string]any)["price"].(float64); !ok {
		JC.Logln("Invalid price type in rate:", rate)
		return false
	}

	return true
}

func (er *ExchangeResults) GetRate(rk string) *ExchangeResults {

	JC.PrintMemUsage("Start fetching exchange rates")

	rko := strings.Split(rk, "|")

	if len(rko) != 2 {
		return nil
	}

	sid := rko[0]
	rkt := strings.Split(rko[1], ",")
	rkv := []string{}

	for _, rkt := range rkt {

		// Try to use cached data
		ck := ExchangeCache.CreateKeyFromString(sid, rkt)
		if !ExchangeCache.Has(ck) {
			rkv = append(rkv, rkt)
		}
	}

	// Seems all the query is cached don't invoke http request
	if len(rkv) == 0 {
		return nil
	}

	tid := strings.Join(rkv, ",")

	client := &http.Client{}
	req, err := http.NewRequest("GET", Config.ExchangeEndpoint, nil)

	if err != nil {
		JC.Logln("Error encountered:", err)
		return nil
	}

	q := url.Values{}
	q.Add("amount", "1")
	q.Add("id", sid)
	q.Add("convert_id", tid)

	req.URL.RawQuery = q.Encode()

	// Debug
	JC.Logf("Fetching data from %v", req.URL.RawQuery)

	resp, err := client.Do(req)
	if err != nil {
		wrappedErr := fmt.Errorf("Failed to fetch exchange data from CMC: %w", err)
		JC.Logln(wrappedErr)
		return nil
	} else {
		// JC.Log("Fetched exchange data from CMC:", req.URL.RawQuery)
	}

	defer resp.Body.Close()

	c := resp.Body

	// Debug simulating invalid json
	// payload := ""
	// payload = "{}"
	// payload = `{"data":[]}`
	// payload = `{"data":{}}`
	// payload = `{"data":{"quote":["SOL"]}}`
	// payload = `{"data":{"id": "6636", "symbol": "DOT", "amount": 0.5, "quote":["SOL"]}}`
	// payload = `{"data":{"id": "6636", "symbol": "DOT", "amount": 0.5, "quote":[{"SOL"}]}}`
	// payload = `{"data":{"id": "6636", "symbol": "DOT", "amount": 0.5, "quote":[{"SOL":[]}]}}`
	// payload = `{"data":{"id": "6636", "symbol": "DOT", "amount": 0.5, "quote":[{"SOL":{}}]}}`
	// payload = `{"data":{"id": "6636", "symbol": "DOT", "amount": 0.5, "quote":[{"SOL":{"price":""}}]}}`
	// payload = `{"data":{"id": "6636", "symbol": "DOT", "amount": 0.5, "quote":[{"SOL":{"price":"1", "symbol": "x"}}]}}`

	// c = io.NopCloser(strings.NewReader(payload))

	// Decode JSON directly from response body to save memory
	decoder := json.NewDecoder(c)
	if err := decoder.Decode(er); err != nil {
		JC.Logln(fmt.Errorf("Failed to examine exchange data: %w", err))
		return nil
	}

	// Cache the result
	for _, ex := range er.Rates {

		// Debug to force display refresh!
		ex.TargetAmount = ex.TargetAmount * (rand.Float64() * 5)

		ExchangeCache.Insert(&ex)
	}

	JC.PrintMemUsage("End fetching exchange rates")

	return er
}
