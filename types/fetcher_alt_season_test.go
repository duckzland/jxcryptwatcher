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
	snap := altSeasonSnapshot{
		AltcoinIndex: "75",
		TimestampRaw: rawTS,
	}

	ts, err := strconv.ParseInt(snap.TimestampRaw, 10, 64)
	if err != nil {
		t.Errorf("Failed to parse timestamp: %v", err)
	}
	snap.LastUpdate = time.Unix(ts, 0)

	if snap.LastUpdate.Unix() != ts {
		t.Errorf("Expected LastUpdate to match parsed timestamp")
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

	fetcher.Data = &altSeasonHistoricalData{
		HistoricalValues: altSeasonHistoricalValues{
			Now: altSeasonSnapshot{
				AltcoinIndex: "80",
				TimestampRaw: "1695955200", // Example UNIX timestamp
			},
		},
	}

	ts, err := strconv.ParseInt(fetcher.Data.HistoricalValues.Now.TimestampRaw, 10, 64)
	if err != nil {
		t.Errorf("Failed to parse timestamp: %v", err)
	}
	fetcher.Data.HistoricalValues.Now.LastUpdate = time.Unix(ts, 0)

	if fetcher.Data.HistoricalValues.Now.AltcoinIndex != "80" {
		t.Error("AltcoinIndex not set correctly")
	}
	if fetcher.Data.HistoricalValues.Now.LastUpdate.Unix() != ts {
		t.Error("LastUpdate not set correctly")
	}
	altSeasonTurnOnLogs()
}
