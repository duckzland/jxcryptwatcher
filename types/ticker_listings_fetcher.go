package types

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	JC "jxwatcher/core"
)

type TickerListingsFetcher struct {
	Status       TickerListingsStatusFields `json:"status"`
	Data         []TickerListingsDataFields `json:"data"`
	AltcoinIndex float64                    `json:"-"`
}

type TickerListingsDataFields struct {
	Name   string                               `json:"name"`
	Symbol string                               `json:"symbol"`
	Quote  map[string]TickerListingsQuoteFields `json:"quote"`
}

type TickerListingsQuoteFields struct {
	Price            float64 `json:"price"`
	MarketCap        float64 `json:"market_cap"`
	PercentChange24h float64 `json:"percent_change_24h"`
	PercentChange90d float64 `json:"percent_change_90d"`
}

type TickerListingsStatusFields struct {
	LastUpdate time.Time `json:"timestamp"`
}

// Not Perfect, not exactly matching the value at cmc
func (er *TickerListingsFetcher) CalculateAltcoinIndex() {
	// Define exclusion sets
	excludeSymbols := map[string]bool{

		"BTC": true,

		// Stablecoins
		"USDT": true, "USDC": true, "DAI": true, "TUSD": true, "FDUSD": true,
		"PYUSD": true, "EURC": true, "GUSD": true,

		// Gold-backed
		"XAUT": true, "PAXG": true, "DGX": true,

		// Wrapped BTC/ETH variants
		"WBTC": true, "BTCB": true, "WETH": true, "ETHW": true,
		"STETH": true, "CBETH": true, "RETH": true, "METH": true,
		"ANKRETH": true, "WSTETH": true,

		// Other wrapped tokens
		"WBNB": true, "WAVAX": true, "WMATIC": true, "WFTM": true,
		"WONE": true, "WCELO": true, "WHT": true, "XWC": true,
	}

	// Get BTC's 90-day percent change
	var btcChange float64
	for _, coin := range er.Data {
		if coin.Symbol == "BTC" {
			if quote, ok := coin.Quote["USD"]; ok {
				btcChange = quote.PercentChange90d
			}
			break
		}
	}

	if btcChange == 0 {
		er.AltcoinIndex = 0
		return
	}

	var outperformingCount int
	var totalCount int

	for i, coin := range er.Data {
		if excludeSymbols[coin.Symbol] {
			continue
		}
		if strings.Contains(strings.ToLower(coin.Symbol), "wrapped") || strings.HasPrefix(coin.Symbol, "W") {
			continue // Catch generic wrapped tokens
		}
		quote, ok := coin.Quote["USD"]
		if !ok || quote.PercentChange90d == 0 {
			continue // Skip if no valid 90d change
		}
		totalCount++
		if quote.PercentChange90d > btcChange {
			outperformingCount++
		}

		if i == 100 {
			break
		}
	}

	if totalCount == 0 {
		er.AltcoinIndex = 0
		return
	}

	er.AltcoinIndex = float64(outperformingCount) / float64(totalCount) * 100
}
func (er *TickerListingsFetcher) GetRate() int64 {
	JC.PrintMemUsage("Start fetching Listings data")

	if !Config.HasProKey() {
		JC.Logln("Failed to fetch Listings data due to no Pro API Key provided")
		return -1
	}

	if !Config.CanDoListings() {
		JC.Logln("Failed to fetch Listings data due to no valid endpoint configured")
		return -1
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", Config.TickerListingsEndpoint, nil)
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

	JC.Logf("Fetching data from %v", req.URL)

	resp, err := client.Do(req)
	if err != nil {
		JC.Logln(fmt.Errorf("Failed to fetch Listings data from CMC: %w", err))
		return -1
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
		JC.Logln(fmt.Errorf("Failed to examine Listings data: %w", err))
		return -1
	}

	er.CalculateAltcoinIndex()

	ai := strconv.FormatFloat(er.AltcoinIndex, 'f', 0, 64)
	TickerCache.Insert("altcoin_index", ai, er.Status.LastUpdate)

	JC.PrintMemUsage("End fetching Listings data")

	return 200
}
