package types

import (
	"log"
	"os"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"

	JC "jxwatcher/core"
)

type panelDataNullWriter struct{}

func (panelDataNullWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func panelDataTurnOffLogs() {
	log.SetOutput(panelDataNullWriter{})
}

func panelDataTurnOnLogs() {
	log.SetOutput(os.Stdout)
}

func TestPanelDataInitAndSet(t *testing.T) {
	panelDataTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	p := NewPanelData()
	p.Init()

	p.Set("1-2-0.5-BTC-ETH-4|0.5")
	if p.Get() != "1-2-0.5-BTC-ETH-4|0.5" {
		t.Error("Set/Get mismatch")
	}
	if p.GetOldKey() != JC.STRING_EMPTY {
		t.Error("Expected oldKey to be empty after first Set")
	}

	p.Set("1-2-0.5-BTC-ETH-4|0.6")
	if p.GetOldKey() != "1-2-0.5-BTC-ETH-4|0.5" {
		t.Error("Expected oldKey to be updated after second Set")
	}
	panelDataTurnOnLogs()
}

func TestPanelDataStatusAndID(t *testing.T) {
	panelDataTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	p := NewPanelData()
	p.Init()
	p.SetStatus(JC.STATE_LOADED)
	if !p.IsStatus(JC.STATE_LOADED) {
		t.Error("Status check failed")
	}

	p.SetID("abc123")
	if !p.IsID("abc123") {
		t.Error("ID check failed")
	}
	panelDataTurnOnLogs()
}

func TestPanelDataChangeDetection(t *testing.T) {
	panelDataTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	p := NewPanelData()
	p.Init()
	p.Set("1-2-0.5-BTC-ETH-4|0.5")
	p.SetStatus(JC.STATE_LOADED)
	p.Set("1-2-0.5-BTC-ETH-4|0.6")

	if !p.DidChange() {
		t.Error("Expected DidChange to be true")
	}
	if p.IsOnInitialValue() {
		t.Error("Expected IsOnInitialValue to be false")
	}
	panelDataTurnOnLogs()
}

func TestPanelDataSetRate(t *testing.T) {
	panelDataTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	p := NewPanelData()
	p.Init()
	p.Set("1-2-0.5-BTC-ETH-4|0.5")

	if p.SetRate(nil) {
		t.Error("Expected SetRate(nil) to return false")
	}
	if p.Get() != "1-2-0.5-BTC-ETH-4|0.5" {
		t.Error("Expected value to remain unchanged when SetRate(nil)")
	}

	newVal, ok := JC.ToBigString("0.6")
	if !ok {
		t.Fatal("failed to parse test value")
	}

	if !p.SetRate(newVal) {
		t.Error("Expected SetRate with different value to return true")
	}
	if p.Get() != "1-2-0.5-BTC-ETH-4|0.6" {
		t.Errorf("Expected updated value '1-2-0.5-BTC-ETH-4|0.6', got %s", p.Get())
	}
	if p.GetOldKey() != "1-2-0.5-BTC-ETH-4|0.5" {
		t.Errorf("Expected oldKey to be '1-2-0.5-BTC-ETH-4|0.5', got %s", p.GetOldKey())
	}

	sameVal, ok := JC.ToBigString("0.6")
	if !ok {
		t.Fatal("failed to parse test value")
	}

	if p.SetRate(sameVal) {
		t.Error("Expected SetRate with same value to return false", sameVal, p.GetValueString())
	}
	if p.Get() != "1-2-0.5-BTC-ETH-4|0.6" {
		t.Error("Expected value to remain unchanged when SetRate with same value")
	}

	panelDataTurnOnLogs()
}

func TestPanelDataIsKeyAndOldKey(t *testing.T) {
	p := NewPanelData()
	p.Init()

	// First set
	p.Set("abc")
	if !p.IsKey("abc") {
		t.Error("IsKey should match current value")
	}
	if !p.IsOldKey("") {
		t.Error("OldKey should be empty initially")
	}

	// Second set
	p.Set("def")
	if !p.IsOldKey("abc") {
		t.Error("OldKey should be updated to previous value")
	}
	if !p.IsEqualContentString("def") {
		t.Error("IsEqualContentString should match current value")
	}
}

func TestPanelDataIsValueIncrease(t *testing.T) {
	p := NewPanelData()
	p.Init()

	// Old = 0.5, New = 0.6 → increase
	p.Set("1-2-0.5-BTC-ETH-4|0.5")
	p.Set("1-2-0.5-BTC-ETH-4|0.6")
	if p.IsValueIncrease() != JC.VALUE_INCREASE {
		t.Error("Expected VALUE_INCREASE")
	}

	// Old = 0.6, New = 0.5 → decrease
	p.Set("1-2-0.5-BTC-ETH-4|0.5")
	if p.IsValueIncrease() != JC.VALUE_DECREASE {
		t.Error("Expected VALUE_DECREASE")
	}

	// Old = 0.5, New = 0.5 → no change
	p.Set("1-2-0.5-BTC-ETH-4|0.5")
	if p.IsValueIncrease() != JC.VALUE_NO_CHANGE {
		t.Error("Expected VALUE_NO_CHANGE")
	}
}

func TestPanelDataDestroy(t *testing.T) {
	p := NewPanelData()
	p.Init()
	p.Set("abc")
	p.SetStatus(JC.STATE_LOADED)
	p.SetID("id123")

	p.Destroy()

	if p.Get() != "" {
		t.Error("Expected Get() to be empty after Destroy")
	}
	if p.GetID() != "" {
		t.Error("Expected ID to be empty after Destroy")
	}
	if p.GetStatus() != JC.STATE_ERROR {
		t.Error("Expected status to be STATE_ERROR after Destroy")
	}
}

func TestPanelDataSerialize(t *testing.T) {
	p := NewPanelData()
	p.Init()
	p.SetStatus(JC.STATE_LOADED)
	p.Set("1-2-0.5-BTC-ETH-4|0.5")
	p.Set("1-2-0.5-BTC-ETH-4|0.6")

	snap := p.Serialize()
	if snap.Status != JC.STATE_LOADED {
		t.Error("Serialize should capture status")
	}
	if snap.Key == "" {
		t.Error("Serialize should capture current key")
	}
	if snap.OldKey == "" {
		t.Error("Serialize should capture old key")
	}
}

func TestPanelDataFormatting(t *testing.T) {
	panelDataTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	p := NewPanelData()
	p.Init()
	p.Set("1-2-0.5-BTC-ETH-4|0.6")

	title := p.FormatTitle()
	if title != "0.5 BTC to ETH" {
		t.Errorf("FormatTitle mismatch, got %q", title)
	}

	subtitle := p.FormatSubtitle()
	if subtitle != "1 BTC = 0.6 ETH" {
		t.Errorf("FormatSubtitle mismatch, got %q", subtitle)
	}

	bottom := p.FormatBottomText()
	if bottom != "1 ETH = 1.6667 BTC" {
		t.Errorf("FormatBottomText mismatch, got %q", bottom)
	}

	content := p.FormatContent()
	if content != "0.3 ETH" {
		t.Errorf("FormatContent mismatch, got %q", content)
	}

	panelDataTurnOnLogs()
}

func TestPanelDataUpdateStatus(t *testing.T) {
	p := NewPanelData()
	p.Init()
	p.Set("1-2-0.5-BTC-ETH-4|0.5")

	p.SetStatus(JC.STATE_LOADING)
	if !p.UpdateStatus() {
		t.Error("Expected UpdateStatus to return true from STATE_LOADING")
	}
	if !p.IsStatus(JC.STATE_LOADED) {
		t.Error("Expected status to be STATE_LOADED after UpdateStatus")
	}

	p.SetStatus(JC.STATE_ERROR)
	if !p.UpdateStatus() {
		t.Error("Expected UpdateStatus to return true from STATE_ERROR")
	}
	if !p.IsStatus(JC.STATE_LOADED) {
		t.Error("Expected status to be STATE_LOADED after UpdateStatus")
	}

	p.SetStatus(JC.STATE_LOADED)
	if !p.UpdateStatus() {
		t.Error("Expected UpdateStatus to return true when already STATE_LOADED")
	}
}

func TestPanelDataUpdateAndUpdateRate(t *testing.T) {
	panelDataTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	RegisterExchangeCache().Init()

	newVal, ok := JC.ToBigString("0.75")
	if !ok {
		t.Fatal("Failed to generate big float number")
	}

	ex := exchangeDataType{
		SourceSymbol: "BTC",
		SourceId:     1,
		SourceAmount: 1.0,
		TargetSymbol: "ETH",
		TargetId:     2,
		TargetAmount: newVal,
		Timestamp:    time.Now(),
	}
	UseExchangeCache().Insert(&ex)

	p := NewPanelData()
	p.Init()
	p.Set("1-2-1-BTC-ETH-4|0.5")

	if !p.Update("1-2-1-BTC-ETH-4|0.5") {
		t.Error("Expected Update to return true when cache has new TargetAmount")
	}

	if p.GetValueString() != "0.75" {
		t.Errorf("Expected updated value string '0.75', got %s", p.GetValueString())
	}

	if p.GetOldKey() != "1-2-1-BTC-ETH-4|0.5" {
		t.Errorf("Expected OldKey to be previous key '...|0.5', got %s", p.GetOldKey())
	}

	if p.UpdateRate() {
		t.Error("Expected UpdateRate to return false when value already matches cache")
	}

	panelDataTurnOnLogs()
}
