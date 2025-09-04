package types

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	JC "jxwatcher/core"
)

type TickerFearGreedFetcher struct {
	Data TickerFearGreedDataFields `json:"data"`
}

type TickerFearGreedDataFields struct {
	Value          float64   `json:"value"`
	Classification string    `json:"value_classification"`
	NextUpdate     time.Time `json:"update_time"`
}

func (er *TickerFearGreedFetcher) GetRate() int64 {

	JC.PrintMemUsage("Start fetching cmc100 data")

	if !Config.HasProKey() {
		JC.Logln("Failed to fetch Fear & Greed data due to no Pro API Key provided")
		return -1
	}

	if !Config.CanDoFearGreed() {
		JC.Logln("Failed to fetch Fear & Greed data due to no valid endpoint configured")
		return -1
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", Config.TickerFearGreedEndpoint, nil)
	if err != nil {
		JC.Logln("Error encountered:", err)
		return -1
	}

	// Add headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-CMC_PRO_API_KEY", Config.ProApiKey)
	req.Header.Set("Expires", "0")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Cache-control", "no-cache, no-store, must-revalidate")

	// Debug
	JC.Logf("Fetching data from %v", req.URL)

	resp, err := client.Do(req)
	if err != nil {
		wrappedErr := fmt.Errorf("Failed to fetch fear & greed data from CMC: %w", err)
		JC.Logln(wrappedErr)
		return -1
	} else {
		// JC.Logln("Fetched Fear & Greed Data from:", req.URL)
	}

	defer resp.Body.Close()

	// Handle HTTP status codes
	switch resp.StatusCode {
	case 401:
		JC.Logln(fmt.Sprintf("Error %d: Unauthorized Check your API key", resp.StatusCode))
		return 401
	case 429:
		JC.Logln(fmt.Sprintf("Error %d: Too Many Requests Rate limit exceeded", resp.StatusCode))
		return 429
	case 200:
		// return 200
	default:
		JC.Logln(fmt.Sprintf("Error %d: Request failed", resp.StatusCode))
		return int64(resp.StatusCode)
	}

	c := resp.Body

	// Decode JSON directly from response body to save memory
	decoder := json.NewDecoder(c)
	if err := decoder.Decode(er); err != nil {
		wrappedErr := fmt.Errorf("Failed to examine Fear & Greed data: %w", err)
		JC.Logln(wrappedErr)
		return -1
	}

	val := strconv.FormatFloat(er.Data.Value, 'f', 0, 64)

	TickerCache.Insert("feargreed", val, er.Data.NextUpdate)
	TickerCache.Insert("feargreed_classification", er.Data.Classification, er.Data.NextUpdate)

	JC.PrintMemUsage("End fetching Fear & Greed data")

	return 200
}
