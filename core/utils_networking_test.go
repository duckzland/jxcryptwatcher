package core

import (
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	json "github.com/goccy/go-json"
)

type netMockResponse struct {
	Message string `json:"message"`
}

type netNullWriter struct{}

func (n netNullWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func netTurnOffLogs() {
	log.SetOutput(netNullWriter{})
}

func netTurnOnLogs() {
	log.SetOutput(os.Stdout)
}

func TestGetRequest_Success(t *testing.T) {
	netTurnOffLogs()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := netMockResponse{Message: "Hello, world"}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	var result netMockResponse

	code := GetRequest(ts.URL, &result, nil, func(resp *http.Response, dec any) int64 {
		r, ok := dec.(*netMockResponse)
		if !ok || r.Message != "Hello, world" {
			t.Errorf("Unexpected callback data: %+v", dec)
			return NETWORKING_BAD_DATA_RECEIVED
		}
		return NETWORKING_SUCCESS
	})

	netTurnOnLogs()

	if code != NETWORKING_SUCCESS {
		t.Errorf("Expected NETWORKING_SUCCESS, got %d", code)
	}
}

func TestGetRequest_InvalidURL(t *testing.T) {
	netTurnOffLogs()

	var result netMockResponse
	code := GetRequest(":://invalid-url", &result, nil, nil)

	netTurnOnLogs()

	if code != NETWORKING_URL_ERROR {
		t.Errorf("Expected NETWORKING_URL_ERROR, got %d", code)
	}
}

func TestGetRequest_404(t *testing.T) {
	netTurnOffLogs()

	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	var result netMockResponse
	code := GetRequest(ts.URL, &result, nil, nil)

	netTurnOnLogs()

	if code != NETWORKING_BAD_CONFIG {
		t.Errorf("Expected NETWORKING_BAD_CONFIG for 404, got %d", code)
	}
}

func TestGetRequest_BadJSON(t *testing.T) {
	netTurnOffLogs()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not-json"))
	}))
	defer ts.Close()

	var result netMockResponse
	code := GetRequest(ts.URL, &result, nil, nil)

	netTurnOnLogs()

	if code != NETWORKING_BAD_DATA_RECEIVED {
		t.Errorf("Expected NETWORKING_BAD_DATA_RECEIVED, got %d", code)
	}
}

func TestGetRequest_WithPrefetch(t *testing.T) {
	netTurnOffLogs()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("test") != "ok" {
			t.Errorf("Expected query param 'test=ok', got %v", r.URL.Query())
		}
		json.NewEncoder(w).Encode(netMockResponse{Message: "Prefetch success"})
	}))
	defer ts.Close()

	var result netMockResponse
	prefetch := func(q url.Values, req *http.Request) {
		q.Set("test", "ok")
	}

	code := GetRequest(ts.URL, &result, prefetch, nil)

	netTurnOnLogs()

	if code != NETWORKING_SUCCESS {
		t.Errorf("Expected NETWORKING_SUCCESS with prefetch, got %d", code)
	}
}
