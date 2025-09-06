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

type AltSeasonFetcher struct {
	Data *AltSeasonHistoricalData `json:"data"`
}

type AltSeasonHistoricalData struct {
	HistoricalValues AltSeasonHistoricalValues `json:"historicalValues"`
}

type AltSeasonHistoricalValues struct {
	Now AltSeasonSnapshot `json:"now"`
}

type AltSeasonSnapshot struct {
	AltcoinIndex string    `json:"altcoinIndex"`
	TimestampRaw string    `json:"timestamp"`
	LastUpdate   time.Time `json:"-"`
}

func (er *AltSeasonFetcher) GetRate() int64 {
	JC.PrintMemUsage("Start fetching Alt Seasons data")

	parsedURL, err := url.Parse(Config.AltSeasonEndpoint)
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
			JC.Logln("Raw Error", urlErr.Err.Error())
			if strings.Contains(urlErr.Err.Error(), "tls") {
				JC.Logln("TLS handshake error:", urlErr.Err)
				return JC.NETWORKING_BAD_CONFIG
			}
		}

		JC.Logln(fmt.Errorf("Failed to fetch Alt Seasons data from CMC: %w", err))
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
		JC.Logln(fmt.Errorf("Failed to examine Alt Seasons data: %w", err))
		return JC.NETWORKING_BAD_DATA_RECEIVED
	}

	if er.Data == nil {
		return JC.NETWORKING_BAD_DATA_RECEIVED
	}

	ts, err := strconv.ParseInt(er.Data.HistoricalValues.Now.TimestampRaw, 10, 64)

	if err == nil {
		er.Data.HistoricalValues.Now.LastUpdate = time.Unix(ts, 0)
	} else {
		er.Data.HistoricalValues.Now.LastUpdate = time.Now()
	}

	TickerCache.Insert("altcoin_index", er.Data.HistoricalValues.Now.AltcoinIndex, er.Data.HistoricalValues.Now.LastUpdate)

	JC.PrintMemUsage("End fetching Alt Seasons data")

	return JC.NETWORKING_SUCCESS
}
