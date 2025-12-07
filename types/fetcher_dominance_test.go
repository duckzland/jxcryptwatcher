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

	fetcher.DominanceBTC = "58.13"
	fetcher.DominanceETC = "13.20"
	fetcher.DominanceOther = "28.65"
	tsStr := "2025-10-06T21:25:08.328Z"
	parsed, err := time.Parse(time.RFC3339Nano, tsStr)
	if err != nil {
		t.Errorf("Failed to parse timestamp: %v", err)
	}
	fetcher.LastUpdate = parsed

	if fetcher.DominanceBTC != "58.13" {
		t.Error("BTC dominance not set correctly")
	}
	if fetcher.DominanceETC != "13.20" {
		t.Error("ETH dominance not set correctly")
	}
	if fetcher.DominanceOther != "28.65" {
		t.Error("Other dominance not set correctly")
	}
	if fetcher.LastUpdate.Format(time.RFC3339Nano) != tsStr {
		t.Error("Parsed timestamp mismatch")
	}

	dominanceTurnOnLogs()
}

func TestDominanceFetcherParseJSON(t *testing.T) {
	dominanceTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	raw := []byte(`{
		"data": {
			"dominance": [
				{ "mcProportion": 58.1315 },
				{ "mcProportion": 13.2099 },
				{ "mcProportion": 28.6585 }
			]
		},
		"status": {
			"timestamp": "2025-10-06T21:25:08Z"
		}
	}`)

	fetcher := NewDominanceFetcher()
	err := fetcher.parseJSON(raw)
	if err != nil {
		t.Errorf("Unexpected error parsing JSON: %v", err)
	}

	if fetcher.DominanceBTC != "58.1315" {
		t.Errorf("Expected BTC=58.1315, got %s", fetcher.DominanceBTC)
	}

	if fetcher.DominanceETC != "13.2099" {
		t.Errorf("Expected ETH=13.2099, got %s", fetcher.DominanceETC)
	}

	if fetcher.DominanceOther != "28.6585" {
		t.Errorf("Expected Other=28.6585, got %s", fetcher.DominanceOther)
	}

	expectedTS, _ := time.Parse(time.RFC3339, "2025-10-06T21:25:08Z")
	if !fetcher.LastUpdate.Equal(expectedTS) {
		t.Errorf("Expected LastUpdate=%s, got %s", expectedTS, fetcher.LastUpdate)
	}

	dominanceTurnOnLogs()
}
