package types

import (
	"encoding/json"
	"log"
	"os"
	"testing"

	"fyne.io/fyne/v2/test"
)

type panelTypeNullWriter struct{}

func (panelTypeNullWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func panelTypeTurnOffLogs() {
	log.SetOutput(panelTypeNullWriter{})
}

func panelTypeTurnOnLogs() {
	log.SetOutput(os.Stdout)
}

func TestPanelTypeFieldAssignment(t *testing.T) {
	panelTypeTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	p := panelType{
		Source:       1,
		Target:       2,
		Value:        0.75,
		Decimals:     4,
		SourceSymbol: "BTC",
		TargetSymbol: "ETH",
	}

	if p.Source != 1 || p.Target != 2 {
		t.Error("Source/Target assignment failed")
	}
	if p.Value != 0.75 || p.Decimals != 4 {
		t.Error("Value/Decimals assignment failed")
	}
	if p.SourceSymbol != "BTC" || p.TargetSymbol != "ETH" {
		t.Error("Symbol assignment failed")
	}
	panelTypeTurnOnLogs()
}

func TestPanelTypeJSONMarshaling(t *testing.T) {
	panelTypeTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	original := panelType{
		Source:       1,
		Target:       2,
		Value:        0.75,
		Decimals:     4,
		SourceSymbol: "BTC",
		TargetSymbol: "ETH",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded panelType
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded != original {
		t.Error("Decoded struct does not match original")
	}
	panelTypeTurnOnLogs()
}
