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

	fetcher.Value = strconv.FormatFloat(100.5, 'f', -1, 64)
	fetcher.PercentChange = strconv.FormatFloat(2.3, 'f', -1, 64)
	tsRaw := "1695955200"
	ts, err := strconv.ParseInt(tsRaw, 10, 64)
	if err != nil {
		t.Errorf("Failed to parse timestamp: %v", err)
	}
	fetcher.NextUpdate = time.Unix(ts, 0)

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

func TestCMC100FetcherParseJSON(t *testing.T) {
	cmc100TurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	raw := []byte(`{
		"data": {
			"summaryData": {
				"currentValue": {
					"value": 231.79,
					"percentChange": -0.50
				},
				"nextUpdateTimestamp": "1757141100"
			}
		}
	}`)

	fetcher := NewCMC100Fetcher()
	err := fetcher.parseJSON(raw)
	if err != nil {
		t.Errorf("Unexpected error parsing JSON: %v", err)
	}

	if fetcher.Value != "231.79" {
		t.Errorf("Expected Value=231.79, got %s", fetcher.Value)
	}

	if fetcher.PercentChange != "-0.5" {
		t.Errorf("Expected PercentChange=-0.5, got %s", fetcher.PercentChange)
	}

	expectedTS, _ := strconv.ParseInt("1757141100", 10, 64)
	if fetcher.NextUpdate.Unix() != expectedTS {
		t.Errorf("Expected NextUpdate=%d, got %d", expectedTS, fetcher.NextUpdate.Unix())
	}

	cmc100TurnOnLogs()
}
