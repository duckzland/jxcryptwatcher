package core

import (
	"context"
	"crypto/x509"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"sync"
	"time"
	"syscall"
)

var httpClient = &http.Client{
	Timeout: 120 * time.Second,
	Transport: &http.Transport{
		DisableKeepAlives:     false,
		MaxIdleConns:          NETWORKING_MAXIMUM_CONNECTION,
		MaxIdleConnsPerHost:   NETWORKING_MAXIMUM_CONNECTION,
		IdleConnTimeout:       30 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		TLSHandshakeTimeout:   15 * time.Second,
		ExpectContinueTimeout: 5 * time.Second,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	},
}

func GetRequest(ctx context.Context, targetUrl string, prefetch func(url url.Values, req *http.Request), callback func(ctx context.Context, resp *http.Response) int64) int64 {
	PrintPerfStats("Network Fetching Request", time.Now())

	parsedURL, err := url.Parse(targetUrl)
	if err != nil {
		Logln("Network Invalid URL:", err)
		return NETWORKING_URL_ERROR
	}

	if ctx == nil {
		ctx = context.Background()
	}

	if ctx != nil && ctx.Err() != nil {
		return NETWORKING_ERROR_CONNECTION
	}

	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", parsedURL.String(), nil)
	if err != nil {
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			Logln("Network Bad URL in request:", urlErr)
			return NETWORKING_URL_ERROR
		}

		Logln("Network Error creating request:", err)
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
	// Logf("Network Fetching data from %v", req.URL)

	resp, err := httpClient.Do(req)
	if err != nil {
		if resp != nil {
			resp.Body.Close()
		}

		if tr, ok := httpClient.Transport.(*http.Transport); ok {
			tr.CloseIdleConnections()
		}

		var dnsErr *net.DNSError
		if errors.As(err, &dnsErr) {
			if dnsErr.IsNotFound || dnsErr.IsTemporary {
				Logln("Network DNS error:", dnsErr)
				return NETWORKING_NO_INTERNET
			}
		}

		var opErr *net.OpError
		if errors.As(err, &opErr) {
			if opErr.Timeout() {
				Logln("Network Connection timeout:", opErr)
				return NETWORKING_ERROR_FIREWALL
			}
			if errors.Is(opErr.Err, syscall.ECONNREFUSED) {
				Logln("Network Connection refused:", opErr)
				return NETWORKING_ERROR_FIREWALL
			}
		}

		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			var opErr *net.OpError
			if errors.As(urlErr.Err, &opErr) {
				if opErr.Timeout() || errors.Is(opErr.Err, syscall.ECONNREFUSED) {
					Logln("Network Connection refused/timeout inside url.Error:", opErr)
					return NETWORKING_ERROR_FIREWALL
				}
			}


			var hostnameErr x509.HostnameError
			if errors.As(urlErr.Err, &hostnameErr) {
				Logln("Network TLS hostname mismatch:", hostnameErr)
				return NETWORKING_ERROR_CONNECTION
			}

			var unknownAuthErr x509.UnknownAuthorityError
			if errors.As(urlErr.Err, &unknownAuthErr) {
				Logln("Network TLS unknown authority:", unknownAuthErr)
				return NETWORKING_BAD_CONFIG
			}

			if strings.Contains(urlErr.Err.Error(), "tls") {
				Logln("Network TLS handshake/network error:", urlErr.Err)
				return NETWORKING_ERROR_CONNECTION
			}
		}

		Logf("Network Failed to fetch data: %w", err)

		return NETWORKING_ERROR_CONNECTION
	}


	defer resp.Body.Close()

	if tr, ok := httpClient.Transport.(*http.Transport); ok {
		defer tr.CloseIdleConnections()
	}

	defer runtime.GC()

	switch resp.StatusCode {
	case 200:
		if callback != nil {
			output := callback(ctx, resp)
			return output
		}
		return NETWORKING_SUCCESS

	case 400, 401, 403:
		Logf("Network Error %d: Client/config issue", resp.StatusCode)
		return NETWORKING_BAD_CONFIG

	case 404:
		Logf("Network Error 404: Not Found")
		return NETWORKING_ERROR_CONNECTION

	case 429:
		Logf("Network Error %d: Too Many Requests Rate limit exceeded", resp.StatusCode)
		return NETWORKING_RATE_LIMIT

	case 500, 502, 503, 504:
		Logf("Network Error %d: Server error", resp.StatusCode)
		return NETWORKING_ERROR_CONNECTION

	default:
		Logf("Network Error %d: Request failed", resp.StatusCode)
		return NETWORKING_ERROR_CONNECTION
	}
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
			return nil, func() {}, err
		}
	}

	cleanup := func() {
		buf = buf[:0]
		pool.Put(buf)
	}

	// Logf("Networking data buffer for [%s]: %.2fKB/%.2fKB (%+.2fKB)", key, float64(len(buf))/1024.0, float64(size)/1024.0, float64(cap(buf)-size)/1024.0)

	return buf, cleanup, nil
}
