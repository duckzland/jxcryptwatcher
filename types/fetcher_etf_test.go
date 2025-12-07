package types

import (
	"log"
	"os"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
)

type etfNullWriter struct{}

func (etfNullWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func etfTurnOffLogs() {
	log.SetOutput(etfNullWriter{})
}

func etfTurnOnLogs() {
	log.SetOutput(os.Stdout)
}

func TestETFFetcherStructure(t *testing.T) {
	etfTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	fetcher := NewETFFetcher()
	if fetcher == nil {
		t.Error("Expected non-nil etfFetcher")
	}

	fetcher.Total = "1218600000"
	fetcher.TotalBtcValue = "985100000"
	fetcher.TotalEthValue = "233500000"
	tsStr := "2025-10-06T21:06:02.963Z"
	parsed, err := time.Parse(time.RFC3339Nano, tsStr)
	if err != nil {
		t.Errorf("Failed to parse timestamp: %v", err)
	}
	fetcher.LastUpdate = parsed

	if fetcher.Total != "1218600000" {
		t.Error("ETF total not set correctly")
	}
	if fetcher.TotalBtcValue != "985100000" {
		t.Error("ETF BTC total not set correctly")
	}
	if fetcher.TotalEthValue != "233500000" {
		t.Error("ETF ETH total not set correctly")
	}
	if fetcher.LastUpdate.Format(time.RFC3339Nano) != tsStr {
		t.Error("Parsed timestamp mismatch")
	}

	etfTurnOnLogs()
}

func TestETFFetcherParseJSON(t *testing.T) {
	etfTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	raw := []byte(`{
		"data": {
			"total": 1218600000,
			"totalBtcValue": 985100000,
			"totalEthValue": 233500000
		},
		"status": {
			"timestamp": "2025-10-06T21:06:02Z"
		}
	}`)

	fetcher := NewETFFetcher()
	err := fetcher.parseJSON(raw)
	if err != nil {
		t.Errorf("Unexpected error parsing JSON: %v", err)
	}

	if fetcher.Total != "1218600000" {
		t.Errorf("Expected Total=1218600000, got %s", fetcher.Total)
	}
	if fetcher.TotalBtcValue != "985100000" {
		t.Errorf("Expected TotalBtcValue=985100000, got %s", fetcher.TotalBtcValue)
	}
	if fetcher.TotalEthValue != "233500000" {
		t.Errorf("Expected TotalEthValue=233500000, got %s", fetcher.TotalEthValue)
	}

	expectedTS, _ := time.Parse(time.RFC3339, "2025-10-06T21:06:02Z")
	if !fetcher.LastUpdate.Equal(expectedTS) {
		t.Errorf("Expected LastUpdate=%s, got %s", expectedTS, fetcher.LastUpdate)
	}

	etfTurnOnLogs()
}
