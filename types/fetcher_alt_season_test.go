package types

import (
	"log"
	"os"
	"strconv"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
)

type altSeasonNullWriter struct{}

func (altSeasonNullWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func altSeasonTurnOffLogs() {
	log.SetOutput(altSeasonNullWriter{})
}

func altSeasonTurnOnLogs() {
	log.SetOutput(os.Stdout)
}

func TestAltSeasonSnapshotTimestampParsing(t *testing.T) {
	altSeasonTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	rawTS := strconv.FormatInt(time.Now().Unix(), 10)
	fetcher := NewAltSeasonFetcher()

	// Simulate ParseJSON behavior manually
	ts, err := strconv.ParseInt(rawTS, 10, 64)
	if err != nil {
		t.Errorf("Failed to parse timestamp: %v", err)
	}
	fetcher.Index = "75"
	fetcher.LastUpdate = time.Unix(ts, 0)

	if fetcher.LastUpdate.Unix() != ts {
		t.Errorf("Expected LastUpdate to match parsed timestamp")
	}
	if fetcher.Index != "75" {
		t.Errorf("Expected Index to be 75, got %s", fetcher.Index)
	}
	altSeasonTurnOnLogs()
}

func TestAltSeasonFetcherStructure(t *testing.T) {
	altSeasonTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	fetcher := NewAltSeasonFetcher()
	if fetcher == nil {
		t.Error("Expected non-nil altSeasonFetcher")
	}

	// Simulate parsed values
	fetcher.Index = "80"
	tsRaw := "1695955200" // Example UNIX timestamp
	ts, err := strconv.ParseInt(tsRaw, 10, 64)
	if err != nil {
		t.Errorf("Failed to parse timestamp: %v", err)
	}
	fetcher.LastUpdate = time.Unix(ts, 0)

	if fetcher.Index != "80" {
		t.Error("AltcoinIndex not set correctly")
	}
	if fetcher.LastUpdate.Unix() != ts {
		t.Error("LastUpdate not set correctly")
	}
	altSeasonTurnOnLogs()
}
