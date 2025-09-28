package types

import (
	JC "jxwatcher/core"
	"log"
	"os"
	"testing"

	"fyne.io/fyne/v2/test"
)

type panelsMapNullWriter struct{}

func (panelsMapNullWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func panelsMapTurnOffLogs() {
	log.SetOutput(panelsMapNullWriter{})
}

func panelsMapTurnOnLogs() {
	log.SetOutput(os.Stdout)
}

func TestPanelsMapInitAndAppend(t *testing.T) {
	panelsMapTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	dummy := &exchangeDataType{
		SourceId:     int64(1),
		TargetId:     int64(2),
		SourceSymbol: "BTC",
		TargetSymbol: "ETH",
		TargetAmount: JC.ToBigFloat(42.0),
	}

	RegisterExchangeCache().Init()
	UseExchangeCache().Insert(dummy)

	// Now safely call Append
	pm := &panelsMapType{}
	pm.Init()
	ref := pm.Append("1-2-0.5-BTC-ETH-4|15.5")
	if ref == nil {
		t.Error("Expected non-nil PanelData from Append")
	}
	if pm.TotalData() != 1 {
		t.Errorf("Expected 1 panel, got %d", pm.TotalData())
	}
	panelsMapTurnOnLogs()
}

func TestPanelsMapRemoveAndMove(t *testing.T) {
	panelsMapTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	pm := &panelsMapType{}
	pm.Init()
	a := pm.Append("1-2-0.5-BTC-ETH-4|15.5")
	b := pm.Append("1-3-0.5-BTC-XRP-4|25.5")
	a.SetID("a")
	b.SetID("b")

	if !pm.Move("b", 0) {
		t.Error("Expected move to succeed")
	}
	if !pm.Remove("a") {
		t.Error("Expected remove to succeed")
	}
	if pm.TotalData() != 1 {
		t.Errorf("Expected 1 panel after removal, got %d", pm.TotalData())
	}
	panelsMapTurnOnLogs()
}

func TestPanelsMapHydrateAndSerialize(t *testing.T) {
	panelsMapTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	// Initialize panelsMapType
	pm := &panelsMapType{}
	pm.Init()

	// ✅ Initialize cryptosMapType with matching coin IDs
	cm := &cryptosMapType{}
	cm.Init()
	cm.Insert("1", "BTC - Bitcoin")  // Source coin
	cm.Insert("2", "ETH - Ethereum") // Target coin
	pm.SetMaps(cm)

	// Create panel data
	p1 := NewPanelData()
	p1.Init()
	p1.Set("1-2-0.5-BTC-ETH-4|15.5")
	p1.SetID("x")

	// Hydrate and serialize
	pm.Hydrate([]PanelData{p1})
	if pm.TotalData() != 1 {
		t.Errorf("Expected 1 panel after hydrate, got %d", pm.TotalData())
	}

	serialized := pm.Serialize()
	if len(serialized) != 1 {
		t.Errorf("Expected 1 serialized entry, got %d", len(serialized))
	}
	panelsMapTurnOnLogs()
}
