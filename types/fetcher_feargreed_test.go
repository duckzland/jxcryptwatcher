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

	fetcher.Score = strconv.FormatInt(72, 10)
	tsRaw := "1695955200"
	ts, err := strconv.ParseInt(tsRaw, 10, 64)
	if err != nil {
		t.Errorf("Failed to parse timestamp: %v", err)
	}
	fetcher.LastUpdate = time.Unix(ts, 0)

	// Assertions
	if fetcher.Score != "72" {
		t.Error("Score not set correctly")
	}
	if fetcher.LastUpdate.Unix() != ts {
		t.Error("LastUpdate not set correctly")
	}

	fearGreedTurnOnLogs()
}

func TestFearGreedFetcherParseJSON(t *testing.T) {
	fearGreedTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	raw := []byte(`{
		"data": {
			"historicalValues": {
				"now": {
					"score": 41,
					"timestamp": "1757136190"
				}
			}
		}
	}`)

	fetcher := NewFearGreedFetcher()
	err := fetcher.parseJSON(raw)
	if err != nil {
		t.Errorf("Unexpected error parsing JSON: %v", err)
	}

	if fetcher.Score != "41" {
		t.Errorf("Expected Score=41, got %s", fetcher.Score)
	}

	expectedTS, _ := strconv.ParseInt("1757136190", 10, 64)
	if fetcher.LastUpdate.Unix() != expectedTS {
		t.Errorf("Expected LastUpdate=%d, got %d", expectedTS, fetcher.LastUpdate.Unix())
	}

	fearGreedTurnOnLogs()
}
