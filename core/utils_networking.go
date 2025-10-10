package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func GetRequest(targetUrl string, dec any, prefetch func(url url.Values, req *http.Request), callback func(dec any) int64) int64 {
	PrintMemUsage("Start fetching data")

	parsedURL, err := url.Parse(targetUrl)
	if err != nil {
		Logln("Invalid URL:", err)
		return NETWORKING_URL_ERROR
	}

	// Injected: context with timeout to enforce cancellation
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client := &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives:     true,
			ResponseHeaderTimeout: 10 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
		// Timeout is optional now — context handles it
	}

	// Injected: request tied to context
	req, err := http.NewRequestWithContext(ctx, "GET", parsedURL.String(), nil)
	if err != nil {
		Logln("Error encountered:", err)
		return NETWORKING_ERROR_CONNECTION
	}

	// Add headers
	req.Header.Set("User_Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:142.0) Gecko/20100101 Firefox/142.0")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Expires", "0")

	q := url.Values{}
	AlterRequests(q, req)

	if prefetch != nil {
		prefetch(q, req)
	}

	req.URL.RawQuery = q.Encode()
	// Logf("Fetching data from %v", req.URL)

	resp, err := client.Do(req)
	if err != nil {

		if tr, ok := client.Transport.(*http.Transport); ok {
			tr.CloseIdleConnections()
		}

		var opErr *net.OpError
		if errors.As(err, &opErr) {
			var dnsErr *net.DNSError
			if opErr != nil {
				if errors.As(opErr.Err, &dnsErr) && dnsErr.IsNotFound {
					Logln("DNS error: no such host")
					return NETWORKING_BAD_CONFIG
				}
			}
		}

		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			// Logln("Raw Error", urlErr.Err.Error())
			if strings.Contains(urlErr.Err.Error(), "tls") {
				Logln("TLS handshake error:", urlErr.Err)
				return NETWORKING_BAD_CONFIG
			}
		}

		Logln(fmt.Errorf("Failed to fetch data: %w", err))
		return NETWORKING_ERROR_CONNECTION
	}

	defer resp.Body.Close()

	switch resp.StatusCode {
	case 401, 404:
		Logln(fmt.Sprintf("Error %d: Unauthorized", resp.StatusCode))
		return NETWORKING_BAD_CONFIG
	case 429:
		Logln(fmt.Sprintf("Error %d: Too Many Requests Rate limit exceeded", resp.StatusCode))
		return NETWORKING_ERROR_CONNECTION
	case 200:
		// OK
	default:
		Logln(fmt.Sprintf("Error %d: Request failed", resp.StatusCode))
		return NETWORKING_ERROR_CONNECTION
	}

	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(dec); err != nil {
		Logln(fmt.Errorf("Failed to examine data: %w", err))
		return NETWORKING_BAD_DATA_RECEIVED
	}

	if callback != nil {
		return callback(dec)
	}

	PrintMemUsage("End fetching data")
	return NETWORKING_SUCCESS
}
