package types

import (
	"log"
	"math/big"
	"os"
	"testing"

	"fyne.io/fyne/v2/test"

	JC "jxwatcher/core"
)

type panelKeyNullWriter struct{}

func (panelKeyNullWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func panelKeyTurnOffLogs() {
	log.SetOutput(panelKeyNullWriter{})
}

func panelKeyTurnOnLogs() {
	log.SetOutput(os.Stdout)
}

func TestPanelKeyParsingAndValidation(t *testing.T) {
	panelKeyTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	raw := "1-2-0.5-BTC-ETH-4|15.5"
	pk := &panelKeyType{value: raw}

	if !pk.Validate() {
		t.Error("Expected key to be valid")
	}
	if pk.GetSourceCoinInt() != 1 || pk.GetTargetCoinInt() != 2 {
		t.Error("Source/Target coin parsing failed")
	}
	if pk.GetSourceSymbolString() != "BTC" || pk.GetTargetSymbolString() != "ETH" {
		t.Error("Symbol parsing failed")
	}
	if pk.GetDecimalsInt() != 4 {
		t.Error("Decimals parsing failed")
	}
	if pk.GetValueString() != "15.5" {
		t.Error("Value string parsing failed")
	}
	panelKeyTurnOnLogs()
}

func TestPanelKeyUpdateValue(t *testing.T) {
	panelKeyTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	pk := &panelKeyType{value: "1-2-0.5-BTC-ETH-4|15.5"}
	newVal := pk.UpdateValue(big.NewFloat(20.25))

	if newVal != "1-2-0.5-BTC-ETH-4|20.25" {
		t.Errorf("Unexpected updated value: %s", newVal)
	}
	panelKeyTurnOnLogs()
}

func TestPanelKeyComparison(t *testing.T) {
	panelKeyTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	pk := &panelKeyType{value: "1-2-0.5-BTC-ETH-4|15.5"}
	if !pk.IsValueMatchingFloat(15.5, JC.STRING_DOUBLE_EQUAL) {
		t.Error("Expected value to match")
	}
	if !pk.IsValueMatchingFloat(10.0, JC.STRING_GREATER) {
		t.Error("Expected value to be greater than 10.0")
	}
	panelKeyTurnOnLogs()
}
