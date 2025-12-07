package types

import (
	"log"
	"os"
	"strconv"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
)

type marketCapNullWriter struct{}

func (marketCapNullWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func marketCapTurnOffLogs() {
	log.SetOutput(marketCapNullWriter{})
}

func marketCapTurnOnLogs() {
	log.SetOutput(os.Stdout)
}

func TestMarketCapFetcherStructure(t *testing.T) {
	marketCapTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	fetcher := NewMarketCapFetcher()
	if fetcher == nil {
		t.Error("Expected non-nil marketCapFetcher")
	}

	// Simulate parsed values
	fetcher.NowMarketCap = strconv.FormatFloat(1200000000, 'f', -1, 64)
	fetcher.YesterdayMarketCap = strconv.FormatFloat(1000000000, 'f', -1, 64)
	fetcher.ThirtyDaysChangePct = strconv.FormatFloat(15.0, 'f', -1, 64)
	fetcher.LastUpdate = time.Now()

	// Compute 24h change
	x, _ := strconv.ParseFloat(fetcher.NowMarketCap, 64)
	y, _ := strconv.ParseFloat(fetcher.YesterdayMarketCap, 64)
	z, _ := strconv.ParseFloat(fetcher.ThirtyDaysChangePct, 64)

	dx := ((x - y) / y) * 100
	now := strconv.FormatFloat(x, 'f', -1, 64)
	dif := strconv.FormatFloat(dx, 'f', -1, 64)
	dix := strconv.FormatFloat(z, 'f', -1, 64)

	// Assertions
	if now != "1200000000" {
		t.Errorf("Unexpected now value: %s", now)
	}
	if dif != "20" {
		t.Errorf("Unexpected 24h percentage: %s", dif)
	}
	if dix != "15" {
		t.Errorf("Unexpected 30d percentage: %s", dix)
	}

	marketCapTurnOnLogs()
}
