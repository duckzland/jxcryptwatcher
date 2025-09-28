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

	fetcher.Data = &cmc100SummaryData{
		SummaryData: cmc100SummaryDataFields{
			NextUpdate: "1695955200", // Example UNIX timestamp
			CurrentValue: cmc100CurrentValueFields{
				Value:         100.5,
				PercentChange: 2.3,
			},
		},
	}

	ts, err := strconv.ParseInt(fetcher.Data.SummaryData.NextUpdate, 10, 64)
	if err != nil {
		t.Errorf("Failed to parse timestamp: %v", err)
	}
	parsed := time.Unix(ts, 0)

	if fetcher.Data.SummaryData.CurrentValue.Value != 100.5 {
		t.Error("Value field not set correctly")
	}
	if fetcher.Data.SummaryData.CurrentValue.PercentChange != 2.3 {
		t.Error("PercentChange field not set correctly")
	}
	if parsed.Unix() != ts {
		t.Error("Parsed timestamp mismatch")
	}
	cmc100TurnOnLogs()
}
