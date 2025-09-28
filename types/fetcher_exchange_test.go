package types

import (
	"encoding/json"
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

func TestExchangeResultsUnmarshalValid(t *testing.T) {
	exchangeResultsTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	raw := `{
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
	}`

	var result exchangeResults
	err := json.Unmarshal([]byte(raw), &result)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(result.Rates) != 1 {
		t.Errorf("Expected 1 rate, got %d", len(result.Rates))
	}
	if result.Rates[0].TargetSymbol != "ETH" {
		t.Error("TargetSymbol not parsed correctly")
	}
	exchangeResultsTurnOnLogs()
}

func TestExchangeResultsUnmarshalInvalid(t *testing.T) {
	exchangeResultsTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	raw := `{
		"data": {
			"symbol": 123,
			"id": "1",
			"amount": 1,
			"quote": []
		},
		"status": {
			"timestamp": "invalid-timestamp"
		}
	}`

	var result exchangeResults
	err := json.Unmarshal([]byte(raw), &result)
	if err != nil {
		t.Errorf("Expected graceful failure, got error: %v", err)
	}
	if len(result.Rates) != 0 {
		t.Errorf("Expected 0 rates, got %d", len(result.Rates))
	}
	exchangeResultsTurnOnLogs()
}

func TestExchangeResultsValidateRate(t *testing.T) {
	exchangeResultsTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	er := NewExchangeResults()
	valid := map[string]any{
		"symbol":   "ETH",
		"cryptoId": json.Number("1027"),
		"price":    json.Number("15.5"),
	}
	if !er.validateRate(valid) {
		t.Error("Expected valid rate to pass validation")
	}

	invalid := map[string]any{
		"symbol":   123,
		"cryptoId": "not-a-number",
		"price":    nil,
	}
	if er.validateRate(invalid) {
		t.Error("Expected invalid rate to fail validation")
	}
	exchangeResultsTurnOnLogs()
}
