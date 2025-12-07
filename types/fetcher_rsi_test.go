package types

import (
	"log"
	"os"
	"strconv"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
)

type rsiNullWriter struct{}

func (rsiNullWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func rsiTurnOffLogs() {
	log.SetOutput(rsiNullWriter{})
}

func rsiTurnOnLogs() {
	log.SetOutput(os.Stdout)
}

func TestRSIFetcherStructure(t *testing.T) {
	rsiTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	fetcher := NewRSIFetcher()
	if fetcher == nil {
		t.Error("Expected non-nil rsiFetcher")
	}

	fetcher.AverageRSI = strconv.FormatFloat(55.2, 'f', -1, 64)
	fetcher.OverboughtPercentage = strconv.FormatFloat(18.4, 'f', -1, 64)
	fetcher.OversoldPercentage = strconv.FormatFloat(12.7, 'f', -1, 64)
	fetcher.NeutralPercentage = strconv.FormatFloat(68.9, 'f', -1, 64)
	fetcher.LastUpdate = time.Unix(1695955200, 0)

	if fetcher.AverageRSI != "55.2" {
		t.Errorf("Expected AverageRSI 55.2, got %s", fetcher.AverageRSI)
	}
	if fetcher.OverboughtPercentage != "18.4" {
		t.Errorf("Expected OverboughtPercentage 18.4, got %s", fetcher.OverboughtPercentage)
	}
	if fetcher.OversoldPercentage != "12.7" {
		t.Errorf("Expected OversoldPercentage 12.7, got %s", fetcher.OversoldPercentage)
	}
	if fetcher.NeutralPercentage != "68.9" {
		t.Errorf("Expected NeutralPercentage 68.9, got %s", fetcher.NeutralPercentage)
	}

	ts := fetcher.LastUpdate.Unix()
	if ts != 1695955200 {
		t.Errorf("Expected timestamp 1695955200, got %d", ts)
	}

	rsiTurnOnLogs()
}

func TestRSIFetcherParseJSON(t *testing.T) {
	rsiTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	// Minimal dummy JSON with the fields parseJSON expects
	raw := []byte(`{
		"data": {
			"overall": {
				"averageRsi": 61.70,
				"overboughtPercentage": 16.67,
				"oversoldPercentage": 1.00,
				"neutralPercentage": 82.33
			}
		},
		"status": {
			"timestamp": "2025-10-02T22:41:01Z"
		}
	}`)

	fetcher := NewRSIFetcher()
	err := fetcher.parseJSON(raw)
	if err != nil {
		t.Errorf("Unexpected error parsing JSON: %v", err)
	}

	if fetcher.AverageRSI != "61.7" {
		t.Errorf("Expected AverageRSI=61.7, got %s", fetcher.AverageRSI)
	}

	if fetcher.OverboughtPercentage != "16.67" {
		t.Errorf("Expected OverboughtPercentage=16.67, got %s", fetcher.OverboughtPercentage)
	}

	if fetcher.OversoldPercentage != "1" {
		t.Errorf("Expected OversoldPercentage=1, got %s", fetcher.OversoldPercentage)
	}

	if fetcher.NeutralPercentage != "82.33" {
		t.Errorf("Expected NeutralPercentage=82.33, got %s", fetcher.NeutralPercentage)
	}

	expectedTS, _ := time.Parse(time.RFC3339, "2025-10-02T22:41:01Z")
	if !fetcher.LastUpdate.Equal(expectedTS) {
		t.Errorf("Expected LastUpdate=%s, got %s", expectedTS, fetcher.LastUpdate)
	}

	rsiTurnOnLogs()
}
