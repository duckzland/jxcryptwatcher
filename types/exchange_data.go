package types

import (
	"encoding/json"
	"fmt"
	JC "jxwatcher/core"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

type ExchangeDataType struct {
	SourceSymbol string
	SourceId     int64
	SourceAmount float64
	TargetSymbol string
	TargetId     int64
	TargetAmount float64
}

func (ex *ExchangeDataType) UnmarshalJSON(data []byte) error {

	var v map[string]interface{}
	err := json.Unmarshal(data, &v)
	if err != nil {
		// log.Fatal(err)
		return err
	}

	if v["data"] == nil {
		return nil
	}

	sc := v["data"]
	tc := v["data"].(map[string]interface{})["quote"].([]interface{})[0]

	// CMC Json data is weird the the id is in string while cryptoId is in int64 (but golang cast this as float64)
	ex.SourceSymbol = sc.(map[string]interface{})["symbol"].(string)
	ex.SourceId, _ = strconv.ParseInt(sc.(map[string]interface{})["id"].(string), 10, 64)
	ex.SourceAmount = sc.(map[string]interface{})["amount"].(float64)

	ex.TargetSymbol = tc.(map[string]interface{})["symbol"].(string)
	ex.TargetId = int64(tc.(map[string]interface{})["cryptoId"].(float64))
	ex.TargetAmount = tc.(map[string]interface{})["price"].(float64)

	return nil
}

func (ex *ExchangeDataType) GetRate(pk string) *ExchangeDataType {

	if !BP.ValidatePanel(pk) {
		return nil
	}

	JC.PrintMemUsage("Start fetching exchange rates")

	pko := BP.UsePanelKey(pk)
	sid := pko.GetSourceCoinInt()
	tid := pko.GetTargetCoinInt()

	// Try to use cached data
	ck := ExchangeCache.CreateKeyFromInt(sid, tid)
	if ExchangeCache.Has(ck) {
		log.Println("Using cached data for:", ck)
		JC.PrintMemUsage("End fetching exchange rates, using cached data instead")
		return ExchangeCache.Get(ck)
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", Config.ExchangeEndpoint, nil)

	if err != nil {
		log.Println("Error encountered:", err)
		return nil
	}

	q := url.Values{}
	q.Add("amount", "1")
	q.Add("id", strconv.FormatInt(sid, 10))
	q.Add("convert_id", strconv.FormatInt(tid, 10))

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
	if err := decoder.Decode(ex); err != nil {
		log.Println(fmt.Errorf("Failed to examine exchange data: %w", err))
		return nil
	}

	if err != nil {
		wrappedErr := fmt.Errorf("Failed to examine exchange data: %w", err)
		log.Println(wrappedErr)
		return nil
	}

	// Debug to force display refresh!
	// ex.TargetAmount = ex.TargetAmount * (rand.Float64() * 5)

	// Cache the result
	ExchangeCache.Insert(ex)
	JC.PrintMemUsage("End fetching exchange rates")
	return ex
}
