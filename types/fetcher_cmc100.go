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

type CMC100Fetcher struct {
	Data *CMC100SummaryData `json:"data"`
}

type CMC100SummaryData struct {
	SummaryData CMC100SummaryDataFields `json:"summaryData"`
}

type CMC100SummaryDataFields struct {
	NextUpdate   string                   `json:"nextUpdateTimestamp"`
	CurrentValue CMC100CurrentValueFields `json:"currentValue"`
}

type CMC100CurrentValueFields struct {
	Value         float64 `json:"value"`
	PercentChange float64 `json:"percentChange"`
}

func (er *CMC100Fetcher) GetRate() int64 {
	JC.PrintMemUsage("Start fetching CMC100 data")

	parsedURL, err := url.Parse(Config.CMC100Endpoint)
	if err != nil {
		JC.Logln("Invalid URL:", err)
		return JC.NETWORKING_URL_ERROR
	}

	client := &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequest("GET", parsedURL.String(), nil)
	if err != nil {
		JC.Logln("Error encountered:", err)
		return JC.NETWORKING_ERROR_CONNECTION
	}

	// Add headers
	req.Header.Set("User_Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:142.0) Gecko/20100101 Firefox/142.0")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Expires", "0")

	startUnix, endUnix := JC.GetMonthBounds(time.Now())
	q := url.Values{}
	q.Add("start", strconv.FormatInt(startUnix, 10))
	q.Add("end", strconv.FormatInt(endUnix, 10))

	req.URL.RawQuery = q.Encode()

	JC.Logf("Fetching data from %v", req.URL)

	resp, err := client.Do(req)
	if err != nil {

		var opErr *net.OpError
		if errors.As(err, &opErr) {
			var dnsErr *net.DNSError
			if opErr != nil {
				if errors.As(opErr.Err, &dnsErr) && dnsErr.IsNotFound {
					JC.Logln("DNS error: no such host")
					return JC.NETWORKING_BAD_CONFIG
				}
			}
		}

		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			if strings.Contains(urlErr.Err.Error(), "tls") {
				JC.Logln("TLS handshake error:", urlErr.Err)
				return JC.NETWORKING_BAD_CONFIG
			}
		}

		JC.Logln(fmt.Errorf("Failed to fetch CMC100 data from CMC: %w", err))
		return JC.NETWORKING_ERROR_CONNECTION
	} else {
		// JC.Logln("Fetched Fear & Greed Data from:", req.URL)
	}

	defer resp.Body.Close()

	// Handle HTTP status codes
	switch resp.StatusCode {
	case 401, 404:
		JC.Logln(fmt.Sprintf("Error %d: Unauthorized", resp.StatusCode))
		return JC.NETWORKING_BAD_CONFIG
	case 429:
		JC.Logln(fmt.Sprintf("Error %d: Too Many Requests Rate limit exceeded", resp.StatusCode))
		return JC.NETWORKING_ERROR_CONNECTION
	case 200:
		// return 200
	default:
		JC.Logln(fmt.Sprintf("Error %d: Request failed", resp.StatusCode))
		return JC.NETWORKING_ERROR_CONNECTION
	}

	c := resp.Body

	// Decode JSON directly from response body to save memory
	decoder := json.NewDecoder(c)
	if err := decoder.Decode(er); err != nil {
		JC.Logln(fmt.Errorf("Failed to examine CMC100 data: %w", err))
		return JC.NETWORKING_BAD_DATA_RECEIVED
	}

	if er.Data == nil {
		return JC.NETWORKING_BAD_DATA_RECEIVED
	}

	now := strconv.FormatFloat(er.Data.SummaryData.CurrentValue.Value, 'f', -1, 64)
	dif := strconv.FormatFloat(er.Data.SummaryData.CurrentValue.PercentChange, 'f', -1, 64)
	tim := er.Data.SummaryData.NextUpdate

	ts, err := strconv.ParseInt(tim, 10, 64)

	var nextUpdate time.Time = time.Now()
	if err == nil {
		nextUpdate = time.Unix(ts, 0)
	}

	TickerCache.Insert("cmc100", now, nextUpdate)
	TickerCache.Insert("cmc100_24_percentage", dif, nextUpdate)

	JC.PrintMemUsage("End fetching CMC100 data")

	return JC.NETWORKING_SUCCESS
}
