package types

import (
	"encoding/json"
	"fmt"
	JC "jxwatcher/core"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type ExchangeResults struct {
	Rates []ExchangeDataType
}

func (er *ExchangeResults) UnmarshalJSON(data []byte) error {

	var v map[string]interface{}
	err := json.Unmarshal(data, &v)
	if err != nil {
		return err
	}

	if v["data"] == nil {
		return nil
	}

	sc := v["data"]
	tc := v["data"].(map[string]interface{})["quote"].([]interface{})

	for _, rate := range tc {

		ex := ExchangeDataType{}

		// CMC Json data is weird the the id is in string while cryptoId is in int64 (but golang cast this as float64)
		ex.SourceSymbol = sc.(map[string]interface{})["symbol"].(string)
		ex.SourceId, _ = strconv.ParseInt(sc.(map[string]interface{})["id"].(string), 10, 64)
		ex.SourceAmount = sc.(map[string]interface{})["amount"].(float64)

		ex.TargetSymbol = rate.(map[string]interface{})["symbol"].(string)
		ex.TargetId = int64(rate.(map[string]interface{})["cryptoId"].(float64))
		ex.TargetAmount = rate.(map[string]interface{})["price"].(float64)

		er.Rates = append(er.Rates, ex)
	}

	return nil
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
		log.Println("Error encountered:", err)
		return nil
	}

	q := url.Values{}
	q.Add("amount", "1")
	q.Add("id", sid)
	q.Add("convert_id", tid)

	req.URL.RawQuery = q.Encode()

	// Debug
	log.Printf("Fetching data from %v", req.URL.RawQuery)

	resp, err := client.Do(req)
	if err != nil {
		wrappedErr := fmt.Errorf("Failed to fetch exchange data from CMC: %w", err)
		log.Println(wrappedErr)
		return nil
	} else {
		// log.Print("Fetched exchange data from CMC:", req.URL.RawQuery)
	}

	defer resp.Body.Close()

	// Decode JSON directly from response body to save memory
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(er); err != nil {
		log.Println(fmt.Errorf("Failed to examine exchange data: %w", err))
		return nil
	}

	// Cache the result
	for _, ex := range er.Rates {

		// Debug to force display refresh!
		// ex.TargetAmount = ex.TargetAmount * (rand.Float64() * 5)

		ExchangeCache.Insert(&ex)
	}

	JC.PrintMemUsage("End fetching exchange rates")

	return er
}
