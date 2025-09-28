package core

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

type mockResponse struct {
	Message string `json:"message"`
}

type nullWriter struct{}

func (n nullWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func turnOffLogs() {
	log.SetOutput(nullWriter{})
}

func turnOnLogs() {
	log.SetOutput(os.Stdout)
}

func TestGetRequest_Success(t *testing.T) {
	turnOffLogs()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := mockResponse{Message: "Hello, world"}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	var result mockResponse

	code := GetRequest(ts.URL, &result, nil, func(dec any) int64 {
		r, ok := dec.(*mockResponse)
		if !ok || r.Message != "Hello, world" {
			t.Errorf("Unexpected callback data: %+v", dec)
			return NETWORKING_BAD_DATA_RECEIVED
		}
		return NETWORKING_SUCCESS
	})

	turnOnLogs()

	if code != NETWORKING_SUCCESS {
		t.Errorf("Expected NETWORKING_SUCCESS, got %d", code)
	}
}

func TestGetRequest_InvalidURL(t *testing.T) {
	turnOffLogs()

	var result mockResponse
	code := GetRequest(":://invalid-url", &result, nil, nil)

	turnOnLogs()

	if code != NETWORKING_URL_ERROR {
		t.Errorf("Expected NETWORKING_URL_ERROR, got %d", code)
	}
}

func TestGetRequest_404(t *testing.T) {
	turnOffLogs()

	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	var result mockResponse
	code := GetRequest(ts.URL, &result, nil, nil)

	turnOnLogs()

	if code != NETWORKING_BAD_CONFIG {
		t.Errorf("Expected NETWORKING_BAD_CONFIG for 404, got %d", code)
	}
}

func TestGetRequest_BadJSON(t *testing.T) {
	turnOffLogs()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not-json"))
	}))
	defer ts.Close()

	var result mockResponse
	code := GetRequest(ts.URL, &result, nil, nil)

	turnOnLogs()

	if code != NETWORKING_BAD_DATA_RECEIVED {
		t.Errorf("Expected NETWORKING_BAD_DATA_RECEIVED, got %d", code)
	}
}

func TestGetRequest_WithPrefetch(t *testing.T) {
	turnOffLogs()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("test") != "ok" {
			t.Errorf("Expected query param 'test=ok', got %v", r.URL.Query())
		}
		json.NewEncoder(w).Encode(mockResponse{Message: "Prefetch success"})
	}))
	defer ts.Close()

	var result mockResponse
	prefetch := func(q url.Values, req *http.Request) {
		q.Set("test", "ok")
	}

	code := GetRequest(ts.URL, &result, prefetch, nil)

	turnOnLogs()

	if code != NETWORKING_SUCCESS {
		t.Errorf("Expected NETWORKING_SUCCESS with prefetch, got %d", code)
	}
}
