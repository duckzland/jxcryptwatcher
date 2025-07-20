package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

/**
 * Defining struct for endpoint exhange data
 */
type ExchangeDataType struct {
	SourceSymbol string
	SourceId     int64
	SourceAmount float64
	TargetSymbol string
	TargetId     int64
	TargetAmount float64
}

/**
 * Custom UnmarshalJSON for CMC Exchange sjon
 */
func (ex *ExchangeDataType) UnmarshalJSON(data []byte) error {

	var v map[string]interface{}
	err := json.Unmarshal(data, &v)
	if err != nil {
		log.Fatal(err)
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

/**
 * Get the exchange data from CMC
 */
func getExchangeData(Panel PanelType) ExchangeDataType {

	client := &http.Client{}
	req, err := http.NewRequest("GET", Config.ExchangeEndpoint, nil)

	if err != nil {
		log.Fatal(err)
	}

	q := url.Values{}
	q.Add("amount", "1") // always get 1 exchange value then we can multiply later.
	// q.Add("amount", strconv.FormatFloat(Panel.Value, 'f', 4, 64))
	q.Add("id", strconv.FormatInt(Panel.Source, 10))
	q.Add("convert_id", strconv.FormatInt(Panel.Target, 10))

	req.URL.RawQuery = q.Encode()

	// Debug
	// fmt.Println(req.URL.RawQuery)

	resp, err := client.Do(req)
	if err != nil {
		wrappedErr := fmt.Errorf("Failed to fetch exchange data from CMC: %w", err)
		log.Fatal(wrappedErr)
	} else {
		// log.Print("Fetched exchange data from CMC:", req.URL.RawQuery)
	}

	respBody, _ := io.ReadAll(resp.Body)

	// @todo better error reporting, check the response http status
	// and CMC error status as well
	// fmt.Println(resp.Status)
	// fmt.Println(string(respBody))

	var Exchange ExchangeDataType
	err = json.Unmarshal([]byte(string(respBody)), &Exchange)

	if err != nil {
		wrappedErr := fmt.Errorf("Failed to examine exchange data: %w", err)
		log.Fatal(wrappedErr)
	}

	// Debug to force display refresh!
	// Exchange.TargetAmount = Exchange.TargetAmount * (rand.Float64() * 5)

	return Exchange
}
