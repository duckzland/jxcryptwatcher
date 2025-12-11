package core

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"sync"
	"time"
)

var httpClient = &http.Client{
	Transport: &http.Transport{
		DisableKeepAlives:     false,
		MaxIdleConns:          20,
		MaxIdleConnsPerHost:   2,
		IdleConnTimeout:       30 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	},
}

func GetRequest(ctx context.Context, targetUrl string, prefetch func(url url.Values, req *http.Request), callback func(ctx context.Context, resp *http.Response) int64) int64 {
	PrintPerfStats("Fetching Request", time.Now())

	parsedURL, err := url.Parse(targetUrl)
	if err != nil {
		Logln("Invalid URL:", err)
		return NETWORKING_URL_ERROR
	}

	if ctx == nil {
		ctx = context.Background()
	}

	if ctx.Err() != nil {
		return NETWORKING_ERROR_CONNECTION
	}

	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", parsedURL.String(), nil)
	if err != nil {
		Logln("Error encountered:", err)
		return NETWORKING_ERROR_CONNECTION
	}

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

	resp, err := httpClient.Do(req)
	if err != nil {

		if resp != nil {
			resp.Body.Close()
		}

		if tr, ok := httpClient.Transport.(*http.Transport); ok {
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

	if tr, ok := httpClient.Transport.(*http.Transport); ok {
		defer tr.CloseIdleConnections()
	}

	defer runtime.GC()

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

	if callback != nil {
		output := callback(ctx, resp)
		return output
	}

	return NETWORKING_SUCCESS
}

var networkingBufPools sync.Map

func getNetworkingBufferPool(key string, size int) *sync.Pool {
	if p, ok := networkingBufPools.Load(key); ok {
		return p.(*sync.Pool)
	}
	newPool := &sync.Pool{
		New: func() any {
			return make([]byte, 0, size)
		},
	}
	networkingBufPools.Store(key, newPool)
	return newPool
}

func ReadResponse(key string, resp *http.Response, minSize int) ([]byte, func(), error) {
	size := minSize * 1024
	if size <= 0 {
		size = 4096
	}

	if resp.ContentLength > 0 {
		contentSize := int(resp.ContentLength)
		if contentSize > size {
			size = contentSize
		}
	}

	pool := getNetworkingBufferPool(key, size)
	buf := pool.Get().([]byte)
	if cap(buf) < size {
		buf = make([]byte, 0, size)
	} else {
		buf = buf[:0]
	}

	scratch := min(size, 64*1024)
	tmp := make([]byte, scratch)

	for {
		n, err := resp.Body.Read(tmp)
		if n > 0 {
			if len(buf)+n > cap(buf) {
				newCap := max(cap(buf)*2, len(buf)+n)
				newBuf := make([]byte, len(buf), newCap)
				copy(newBuf, buf)
				buf = newBuf
			}
			buf = append(buf, tmp[:n]...)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			pool.Put(buf[:0])
			return nil, nil, err
		}
	}

	cleanup := func() {
		buf = buf[:0]
		pool.Put(buf)
	}

	// Logf("Networking data buffer for [%s]: %.2fKB/%.2fKB (%+.2fKB)", key, float64(len(buf))/1024.0, float64(size)/1024.0, float64(cap(buf)-size)/1024.0)

	return buf, cleanup, nil
}
