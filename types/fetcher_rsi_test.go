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

	fetcher.Data = &rsiData{
		Overall: rsiOverall{
			AverageRSI:           55.2,
			OverboughtPercentage: 18.4,
			OversoldPercentage:   12.7,
			NeutralPercentage:    68.9,
		},
	}
	fetcher.Status = &rsiStatus{
		Timestamp: time.Unix(1695955200, 0), // Example UNIX timestamp
	}

	// Validate values
	if fetcher.Data.Overall.AverageRSI != 55.2 {
		t.Errorf("Expected AverageRSI 55.2, got %.2f", fetcher.Data.Overall.AverageRSI)
	}
	if fetcher.Data.Overall.OverboughtPercentage != 18.4 {
		t.Errorf("Expected OverboughtPercentage 18.4, got %.2f", fetcher.Data.Overall.OverboughtPercentage)
	}
	if fetcher.Data.Overall.OversoldPercentage != 12.7 {
		t.Errorf("Expected OversoldPercentage 12.7, got %.2f", fetcher.Data.Overall.OversoldPercentage)
	}
	if fetcher.Data.Overall.NeutralPercentage != 68.9 {
		t.Errorf("Expected NeutralPercentage 68.9, got %.2f", fetcher.Data.Overall.NeutralPercentage)
	}

	// Validate timestamp parsing
	ts := fetcher.Status.Timestamp.Unix()
	parsed, err := strconv.ParseInt(strconv.FormatInt(ts, 10), 10, 64)
	if err != nil {
		t.Errorf("Failed to parse timestamp: %v", err)
	}
	if parsed != 1695955200 {
		t.Errorf("Expected timestamp 1695955200, got %d", parsed)
	}

	rsiTurnOnLogs()
}
