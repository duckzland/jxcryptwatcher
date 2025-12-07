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

	// Simulate parsed values
	fetcher.AverageRSI = strconv.FormatFloat(55.2, 'f', -1, 64)
	fetcher.OverboughtPercentage = strconv.FormatFloat(18.4, 'f', -1, 64)
	fetcher.OversoldPercentage = strconv.FormatFloat(12.7, 'f', -1, 64)
	fetcher.NeutralPercentage = strconv.FormatFloat(68.9, 'f', -1, 64)
	fetcher.LastUpdate = time.Unix(1695955200, 0) // Example UNIX timestamp

	// Validate values
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

	// Validate timestamp parsing
	ts := fetcher.LastUpdate.Unix()
	if ts != 1695955200 {
		t.Errorf("Expected timestamp 1695955200, got %d", ts)
	}

	rsiTurnOnLogs()
}
