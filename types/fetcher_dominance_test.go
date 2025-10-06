package types

import (
	"log"
	"os"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
)

type dominanceNullWriter struct{}

func (dominanceNullWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func dominanceTurnOffLogs() {
	log.SetOutput(dominanceNullWriter{})
}

func dominanceTurnOnLogs() {
	log.SetOutput(os.Stdout)
}

func TestDominanceFetcherStructure(t *testing.T) {
	dominanceTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	fetcher := NewDominanceFetcher()
	if fetcher == nil {
		t.Error("Expected non-nil dominanceFetcher")
	}

	fetcher.Data = dominanceData{
		Dominance: []dominanceEntry{
			{MCProportion: 58.13},
			{MCProportion: 13.20},
			{MCProportion: 28.65},
		},
	}
	fetcher.Status = dominanceStatus{
		Timestamp: "2025-10-06T21:25:08.328Z",
	}

	parsed, err := time.Parse(time.RFC3339Nano, fetcher.Status.Timestamp)
	if err != nil {
		t.Errorf("Failed to parse timestamp: %v", err)
	}

	if fetcher.Data.Dominance[0].MCProportion != 58.13 {
		t.Error("BTC dominance not set correctly")
	}
	if fetcher.Data.Dominance[1].MCProportion != 13.20 {
		t.Error("ETH dominance not set correctly")
	}
	if fetcher.Data.Dominance[2].MCProportion != 28.65 {
		t.Error("Other dominance not set correctly")
	}
	if parsed.Format(time.RFC3339Nano) != fetcher.Status.Timestamp {
		t.Error("Parsed timestamp mismatch")
	}
	dominanceTurnOnLogs()
}
