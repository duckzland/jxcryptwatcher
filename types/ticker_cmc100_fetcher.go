package types

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	JC "jxwatcher/core"
)

type TickerCMC100Fetcher struct {
	Data TickerCMC100DataFields `json:"data"`
}

type TickerCMC100DataFields struct {
	Value                       float64   `json:"value"`
	Value24HourPercentageChange float64   `json:"value_24h_percentage_change"`
	LastUpdate                  time.Time `json:"last_update"`
	NextUpdate                  time.Time `json:"next_update"`
}

func (er *TickerCMC100Fetcher) GetRate() int64 {

	JC.PrintMemUsage("Start fetching cmc100 data")

	if !Config.HasProKey() {
		JC.Logln("Failed to fetch cmc100 data due to no Pro API Key provided")
		return -1
	}

	if !Config.CanDoCMC100() {
		JC.Logln("Failed to fetch cmc100 data due to no valid endpoint configured")
		return -1
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", Config.TickerCMC100Endpoint, nil)
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
		wrappedErr := fmt.Errorf("Failed to fetch cmc100 data from CMC: %w", err)
		JC.Logln(wrappedErr)
		return -1
	} else {
		// JC.Logln("Fetched CMC100 Data from:", req.URL)
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
		JC.Logln(fmt.Errorf("Failed to examine cmc100 data: %w", err))
		return -1
	}

	val := strconv.FormatFloat(er.Data.Value, 'f', 0, 64)
	pal := strconv.FormatFloat(er.Data.Value24HourPercentageChange, 'f', 2, 64)

	TickerCache.Insert("cmc100", val, er.Data.NextUpdate)
	TickerCache.Insert("cmc100_24_percentage", pal, er.Data.NextUpdate)

	JC.PrintMemUsage("End fetching CMC100 data")

	return 200
}
