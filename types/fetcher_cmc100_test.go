package types

import (
	"log"
	"os"
	"strconv"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
)

type cmc100NullWriter struct{}

func (cmc100NullWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func cmc100TurnOffLogs() {
	log.SetOutput(cmc100NullWriter{})
}

func cmc100TurnOnLogs() {
	log.SetOutput(os.Stdout)
}

func TestCMC100FetcherStructure(t *testing.T) {
	cmc100TurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	fetcher := NewCMC100Fetcher()
	if fetcher == nil {
		t.Error("Expected non-nil cmc100Fetcher")
	}

	// Simulate parsed values
	fetcher.Value = strconv.FormatFloat(100.5, 'f', -1, 64)
	fetcher.PercentChange = strconv.FormatFloat(2.3, 'f', -1, 64)
	tsRaw := "1695955200" // Example UNIX timestamp
	ts, err := strconv.ParseInt(tsRaw, 10, 64)
	if err != nil {
		t.Errorf("Failed to parse timestamp: %v", err)
	}
	fetcher.NextUpdate = time.Unix(ts, 0)

	// Assertions
	if fetcher.Value != "100.5" {
		t.Errorf("Value field not set correctly, got %s", fetcher.Value)
	}
	if fetcher.PercentChange != "2.3" {
		t.Errorf("PercentChange field not set correctly, got %s", fetcher.PercentChange)
	}
	if fetcher.NextUpdate.Unix() != ts {
		t.Error("NextUpdate timestamp mismatch")
	}

	cmc100TurnOnLogs()
}
