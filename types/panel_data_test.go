package types

import (
	"log"
	"os"
	"testing"

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

	p.Set("123|BTC - ETH|0.5")
	if p.Get() != "123|BTC - ETH|0.5" {
		t.Error("Set/Get mismatch")
	}
	if p.GetOldKey() != "" {
		t.Error("Expected oldKey to be empty after first Set")
	}

	p.Set("123|BTC - ETH|0.6")
	if p.GetOldKey() != "123|BTC - ETH|0.5" {
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
	p.Set("123|BTC - ETH|0.5")
	p.SetStatus(JC.STATE_LOADED)
	p.Set("123|BTC - ETH|0.6")

	if !p.DidChange() {
		t.Error("Expected DidChange to be true")
	}
	if p.IsOnInitialValue() {
		t.Error("Expected IsOnInitialValue to be false")
	}
	panelDataTurnOnLogs()
}
