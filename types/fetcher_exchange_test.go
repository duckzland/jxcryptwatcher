package types

import (
	"log"
	"os"
	"testing"

	"fyne.io/fyne/v2/test"
)

type exchangeResultsNullWriter struct{}

func (exchangeResultsNullWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func exchangeResultsTurnOffLogs() {
	log.SetOutput(exchangeResultsNullWriter{})
}

func exchangeResultsTurnOnLogs() {
	log.SetOutput(os.Stdout)
}

func TestExchangeResultsParseJSONValid(t *testing.T) {
	exchangeResultsTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	raw := []byte(`{
		"data": {
			"symbol": "BTC",
			"id": "1",
			"amount": 1,
			"quote": [
				{
					"symbol": "ETH",
					"cryptoId": 1027,
					"price": 15.5
				}
			]
		},
		"status": {
			"timestamp": "2025-09-29T03:00:00.000Z"
		}
	}`)

	er := NewExchangeResults()
	err := er.parseJSON(raw)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(er.Rates) != 1 {
		t.Errorf("Expected 1 rate, got %d", len(er.Rates))
	}
	if er.Rates[0].TargetSymbol != "ETH" {
		t.Error("TargetSymbol not parsed correctly")
	}
	exchangeResultsTurnOnLogs()
}

func TestExchangeResultsParseJSONInvalid(t *testing.T) {
	exchangeResultsTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	raw := []byte(`{
		"data": {
			"symbol": 123,
			"id": "1",
			"amount": 1,
			"quote": []
		},
		"status": {
			"timestamp": "invalid-timestamp"
		}
	}`)

	er := NewExchangeResults()
	err := er.parseJSON(raw)
	if err == nil {
		t.Errorf("Expected error for invalid payload, got nil")
	}
	if len(er.Rates) != 0 {
		t.Errorf("Expected 0 rates, got %d", len(er.Rates))
	}
	exchangeResultsTurnOnLogs()
}
