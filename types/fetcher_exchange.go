package types

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	JC "jxwatcher/core"
)

type exchangeResults struct {
	Rates []exchangeDataType
}

func (er *exchangeResults) UnmarshalJSON(data []byte) error {

	var v map[string]interface{}

	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.UseNumber()

	err := decoder.Decode(&v)
	if err != nil {
		return err
	}

	if er.validateData(v) == false {
		return nil
	}

	sc := v["data"]
	tc := v["data"].(map[string]any)["quote"].([]any)

	st := v["status"]
	tm := st.(map[string]any)["timestamp"].(string)
	tx, _ := time.Parse(time.RFC3339Nano, tm)

	for _, rate := range tc {

		if er.validateRate(rate) == false {
			continue
		}

		ex := exchangeDataType{}

		// CMC Json data is weird the the id is in string while cryptoId is in int64 (but golang cast this as float64)
		ex.SourceSymbol = sc.(map[string]any)["symbol"].(string)
		ex.SourceId, _ = strconv.ParseInt(sc.(map[string]any)["id"].(string), 10, 64)
		ex.SourceAmount, _ = sc.(map[string]any)["amount"].(json.Number).Float64()

		ex.TargetSymbol = rate.(map[string]any)["symbol"].(string)
		ex.TargetId, _ = rate.(map[string]any)["cryptoId"].(json.Number).Int64()

		price := rate.(map[string]any)["price"].(json.Number).String()
		ex.TargetAmount, _ = JC.ToBigString(price)

		ex.Timestamp = tx

		er.Rates = append(er.Rates, ex)
	}

	return nil
}

func (er *exchangeResults) validateData(v map[string]any) bool {

	if _, ok := v["data"]; !ok {
		JC.Logln("Missing 'data' field in exchange results")
		return false
	}

	if _, ok := v["status"]; !ok {
		JC.Logln("Missing 'status' field in exchange results")
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

	if _, ok := v["data"].(map[string]any)["amount"].(json.Number); !ok {
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

	if _, ok := v["status"].(map[string]any)["timestamp"]; !ok {
		JC.Logln("Missing 'timestamp' field in 'status'")
		return false
	}

	if tm, ok := v["status"].(map[string]any)["timestamp"].(string); !ok {
		JC.Logln("Invalid 'timestamp' field type in 'status'")

		_, err := time.Parse(time.RFC3339Nano, tm)
		if err != nil {
			JC.Logln("Invalid 'timestamp' value in 'status'")
		}
		return false
	}

	return true
}

func (er *exchangeResults) validateRate(rate any) bool {
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

	if _, ok := rate.(map[string]any)["cryptoId"].(json.Number); !ok {
		JC.Logln("Invalid cryptoId type in rate:", rate)
		return false
	}

	if _, ok := rate.(map[string]any)["price"]; !ok {
		JC.Logln("Missing price in rate:", rate)
		return false
	}

	if _, ok := rate.(map[string]any)["price"].(json.Number); !ok {
		JC.Logln("Invalid price type in rate:", rate)
		return false
	}

	return true
}

func (er *exchangeResults) GetRate(rk string) int64 {

	rko := strings.Split(rk, "|")

	if len(rko) != 2 {
		return JC.NETWORKING_BAD_PAYLOAD
	}

	rwt := strings.Split(rko[1], ",")
	cks := make(map[string]bool)
	rkt := []string{}

	for _, id := range rwt {
		id = strings.TrimSpace(id)
		if id != "" && !cks[id] {
			cks[id] = true
			rkt = append(rkt, id)
		}
	}

	sid := strings.TrimSpace(rko[0])
	tid := strings.Join(rkt, ",")

	if sid == "" || tid == "" {
		return JC.NETWORKING_BAD_PAYLOAD
	}

	return JC.GetRequest(
		UseConfig().ExchangeEndpoint,
		er,
		func(url url.Values, req *http.Request) {
			url.Add("amount", "1")
			url.Add("id", sid)
			url.Add("convert_id", tid)
		},
		func(cc any) int64 {
			dec, ok := cc.(*exchangeResults)
			if !ok {
				return JC.NETWORKING_BAD_DATA_RECEIVED
			}

			for _, ex := range dec.Rates {

				// Debug to force display refresh!
				// ex.TargetAmount = ex.TargetAmount * (rand.Float64() * 5)

				// JC.Logf("Rates received: 1 %s (ID %d) = %s %s (ID %d)",
				// 	ex.SourceSymbol,
				// 	ex.SourceId,
				// 	ex.TargetAmount.Text('f', -1),
				// 	ex.TargetSymbol,
				// 	ex.TargetId,
				// )

				UseExchangeCache().Insert(&ex)
			}

			return JC.NETWORKING_SUCCESS
		})
}

func NewExchangeResults() *exchangeResults {
	return &exchangeResults{}
}
