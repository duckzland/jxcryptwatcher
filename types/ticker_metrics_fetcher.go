package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	JC "jxwatcher/core"
)

type TickerMetricsFetcher struct {
	Data TickerMetricsDataFields `json:"data"`
}

type TickerMetricsDataFields struct {
	BTCDominance float64                  `json:"btc_dominance"`
	ETHDominance float64                  `json:"eth_dominance"`
	Quote        TickerMetricsQuoteFields `json:"quote"`
	NextUpdate   time.Time                `json:"last_update"`
}

type TickerMetricsQuoteFields struct {
	USD TickerMetricsQuoteUSDFields `json:"USD"`
}

type TickerMetricsQuoteUSDFields struct {
	MarketCap float64 `json:"total_market_cap"`
	Change24h float64 `json:"derivatives_24h_percentage_change"`
}

func (er *TickerMetricsFetcher) GetRate() int64 {

	JC.PrintMemUsage("Start fetching Metrics data")

	if !Config.HasProKey() {
		JC.Logln("Failed to fetch Metrics data due to no Pro API Key provided")
		return -2
	}

	if !Config.CanDoMetrics() {
		JC.Logln("Failed to fetch Metrics data due to no valid endpoint configured")
		return -3
	}

	parsedURL, err := url.Parse(Config.TickerMetricsEndpoint)
	if err != nil {
		JC.Logln("Invalid URL:", err)
		return -4
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", parsedURL.String(), nil)
	if err != nil {
		// Deep error inspection
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			// DNS error: no such host
			var dnsErr *net.DNSError
			if errors.As(urlErr.Err, &dnsErr) && dnsErr.IsNotFound {
				JC.Logln("DNS error: no such host")
				return -5
			}

			if strings.Contains(urlErr.Err.Error(), "tls") {
				JC.Logln("TLS handshake error:", urlErr.Err)
				return -6
			}
		}

		JC.Logln(fmt.Errorf("Failed to fetch Metrics data from CMC: %w", err))
		return -1
	} else {
		// JC.Logln("Fetched Fear & Greed Data from:", req.URL)
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
		wrappedErr := fmt.Errorf("Failed to fetch Metrics data from CMC: %w", err)
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
		JC.Logln(fmt.Errorf("Failed to examine Metrics data: %w", err))
		return -1
	}

	mc := strconv.FormatFloat(er.Data.Quote.USD.MarketCap, 'f', 0, 64)
	md := strconv.FormatFloat(er.Data.Quote.USD.Change24h, 'f', 0, 64)

	TickerCache.Insert("market_cap", mc, er.Data.NextUpdate)
	TickerCache.Insert("market_cap_24_percentage", md, er.Data.NextUpdate)

	JC.PrintMemUsage("End fetching Metrics data")

	return 200
}
