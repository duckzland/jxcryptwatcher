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

	// Simulate parsed values
	fetcher.Total = "1218600000"
	fetcher.TotalBtcValue = "985100000"
	fetcher.TotalEthValue = "233500000"
	tsStr := "2025-10-06T21:06:02.963Z"
	parsed, err := time.Parse(time.RFC3339Nano, tsStr)
	if err != nil {
		t.Errorf("Failed to parse timestamp: %v", err)
	}
	fetcher.LastUpdate = parsed

	// Assertions
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
