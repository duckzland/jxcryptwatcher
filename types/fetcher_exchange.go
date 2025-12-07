package types

import (
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/buger/jsonparser"

	JC "jxwatcher/core"
)

type exchangeResults struct {
	Rates []exchangeDataType
}

func (er *exchangeResults) ParseJSON(data []byte) error {

	er.Rates = nil

	sourceSymbol, err := jsonparser.GetString(data, "data", "symbol")
	if err != nil {
		JC.Logln("ParseJSON error: missing source symbol:", err)
		return err
	}

	sourceIdStr, err := jsonparser.GetString(data, "data", "id")
	if err != nil {
		JC.Logln("ParseJSON error: missing source id:", err)
		return err
	}

	sourceId, err := strconv.ParseInt(sourceIdStr, 10, 64)
	if err != nil {
		JC.Logln("ParseJSON error: cannot parse source id:", err)
		return err
	}

	sourceAmount, err := jsonparser.GetFloat(data, "data", "amount")
	if err != nil {
		JC.Logln("ParseJSON error: missing source amount:", err)
		return err
	}

	tsStr, _ := jsonparser.GetString(data, "status", "timestamp")
	ts, _ := time.Parse(time.RFC3339Nano, tsStr)

	_, err = jsonparser.ArrayEach(data, func(value []byte, _ jsonparser.ValueType, _ int, _ error) {
		ex := exchangeDataType{}
		ex.SourceSymbol = sourceSymbol
		ex.SourceId = sourceId
		ex.SourceAmount = sourceAmount

		ex.TargetSymbol, _ = jsonparser.GetString(value, "symbol")
		ex.TargetId, _ = jsonparser.GetInt(value, "cryptoId")

		priceBytes, _, _, err := jsonparser.Get(value, "price")
		if err != nil {
			JC.Logln("ParseJSON error: missing price:", err)
			return
		}

		priceStr := string(priceBytes)
		ex.TargetAmount, _ = JC.ToBigString(priceStr)

		ex.Timestamp = ts

		er.Rates = append(er.Rates, ex)
	}, "data", "quote")

	if err != nil {
		JC.Logln("ParseJSON error: failed to iterate quotes:", err)
		return err
	}

	return nil
}

func (er *exchangeResults) GetRate(rk string) int64 {

	rko := strings.Split(rk, JC.STRING_PIPE)

	if len(rko) != 2 {
		return JC.NETWORKING_BAD_PAYLOAD
	}

	rwt := strings.Split(rko[1], ",")
	cks := make(map[string]bool)
	rkt := []string{}

	for _, id := range rwt {
		id = strings.TrimSpace(id)
		if id != JC.STRING_EMPTY && !cks[id] {
			cks[id] = true
			rkt = append(rkt, id)
		}
	}

	sid := strings.TrimSpace(rko[0])
	tid := strings.Join(rkt, ",")

	if sid == JC.STRING_EMPTY || tid == JC.STRING_EMPTY {
		return JC.NETWORKING_BAD_PAYLOAD
	}

	return JC.GetRequest(
		UseConfig().ExchangeEndpoint,
		nil,
		func(url url.Values, req *http.Request) {
			url.Add("amount", "1")
			url.Add("id", sid)
			url.Add("convert_id", tid)
		},
		func(resp *http.Response, cc any) int64 {

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return JC.NETWORKING_BAD_DATA_RECEIVED
			}

			if err := er.ParseJSON(body); err != nil {
				return JC.NETWORKING_BAD_DATA_RECEIVED
			}

			for _, ex := range er.Rates {

				// Debug to force display refresh!
				// ex.TargetAmount = ex.TargetAmount * (rand.Float64() * 5)

				// JC.Logf("Rates received: 1 %s (ID %d) = %s %s (ID %d)" ex.SourceSymbol, ex.SourceId, ex.TargetAmount.Text('f', -1), ex.TargetSymbol, ex.TargetId)

				UseExchangeCache().Insert(&ex)
			}

			return JC.NETWORKING_SUCCESS
		})
}

func NewExchangeResults() *exchangeResults {
	return &exchangeResults{}
}
