package types

import (
	"log"
	"os"
	"strconv"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
)

type fearGreedNullWriter struct{}

func (fearGreedNullWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func fearGreedTurnOffLogs() {
	log.SetOutput(fearGreedNullWriter{})
}

func fearGreedTurnOnLogs() {
	log.SetOutput(os.Stdout)
}

func TestFearGreedFetcherStructure(t *testing.T) {
	fearGreedTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	fetcher := NewFearGreedFetcher()
	if fetcher == nil {
		t.Error("Expected non-nil fearGreedFetcher")
	}

	fetcher.Data = &fearGreedHistoricalData{
		HistoricalValues: fearGreedHistoricalValues{
			Now: fearGreedSnapshot{
				Score:        72,
				TimestampRaw: "1695955200", // Example UNIX timestamp
			},
		},
	}

	ts, err := strconv.ParseInt(fetcher.Data.HistoricalValues.Now.TimestampRaw, 10, 64)
	if err != nil {
		t.Errorf("Failed to parse timestamp: %v", err)
	}
	fetcher.Data.HistoricalValues.Now.LastUpdate = time.Unix(ts, 0)

	if fetcher.Data.HistoricalValues.Now.Score != 72 {
		t.Error("Score not set correctly")
	}
	if fetcher.Data.HistoricalValues.Now.LastUpdate.Unix() != ts {
		t.Error("LastUpdate not set correctly")
	}
	fearGreedTurnOnLogs()
}
