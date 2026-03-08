package types

import (
	JC "jxwatcher/core"
	"log"
	"os"
	"testing"

	json "github.com/goccy/go-json"

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

		// Watcher fields
		Rate:      123.45,
		Sent:      10,
		Operator:  1,
		Limit:     100,
		Duration:  60,
		Timestamp: 1678900000,
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
	if p.Rate != 123.45 || p.Sent != 10 || p.Operator != 1 ||
		p.Limit != 100 || p.Duration != 60 || p.Timestamp != 1678900000 {
		t.Error("Watcher field assignment failed")
	}
	panelTypeTurnOnLogs()
}

func TestPanelTypeJSONMarshaling(t *testing.T) {
	panelTypeTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())

	JC.App = test.NewApp()

	original := panelType{
		Source:       1,
		Target:       2,
		Value:        0.75,
		Decimals:     4,
		SourceSymbol: "BTC",
		TargetSymbol: "ETH",

		// Watcher fields
		Rate:      123.45,
		Sent:      10,
		Operator:  1,
		Limit:     100,
		Duration:  60,
		Timestamp: 1678900000,
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
