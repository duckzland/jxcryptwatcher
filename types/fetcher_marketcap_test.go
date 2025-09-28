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

	fetcher.Data = &marketCapHistoricalData{
		HistoricalValues: marketCapHistoricalValues{
			Now:       marketCapSnapshot{MarketCap: 1200000000},
			Yesterday: marketCapSnapshot{MarketCap: 1000000000},
		},
		ThirtyDaysPercentage: 15.0,
	}
	fetcher.Status = marketCapHistoricalStatus{
		LastUpdate: time.Now(),
	}

	x := fetcher.Data.HistoricalValues.Now.MarketCap
	y := fetcher.Data.HistoricalValues.Yesterday.MarketCap
	z := fetcher.Data.ThirtyDaysPercentage

	dx := ((x - y) / y) * 100
	now := strconv.FormatFloat(x, 'f', -1, 64)
	dif := strconv.FormatFloat(dx, 'f', -1, 64)
	dix := strconv.FormatFloat(z, 'f', -1, 64)

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
