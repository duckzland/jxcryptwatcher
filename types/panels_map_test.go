package types

import (
	"log"
	"os"
	"testing"

	"fyne.io/fyne/v2/test"

	JC "jxwatcher/core"
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

	pm := &panelsMapType{}
	pm.Init()

	ref := pm.Append("1-2-0.5-BTC-ETH-4|15.5")
	if ref == nil {
		t.Fatal("Expected non-nil PanelData from Append")
	}
	if !ref.HasParent() {
		t.Error("Expected PanelData to have parent set")
	}
	if ref.GetStatus() != JC.STATE_FETCHING_NEW {
		t.Errorf("Expected status FETCHING_NEW, got %d", ref.GetStatus())
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

	pm := &panelsMapType{}
	pm.Init()

	cm := &cryptosMapType{}
	cm.Init()
	cm.Insert("1", "BTC - Bitcoin")
	cm.Insert("2", "ETH - Ethereum")
	pm.SetMaps(cm)

	original := NewPanelData()
	original.Init()
	original.Set("1-2-0.5-BTC-ETH-4|15.5")
	original.SetID("x")
	original.SetParent(pm)
	original.SetStatus(JC.STATE_ERROR)
	pm.SetData([]PanelData{original})

	if original.GetStatus() != JC.STATE_ERROR {
		t.Errorf("Expected status ERROR after SetData, got %d", original.GetStatus())
	}

	hydrated := NewPanelData()
	hydrated.Init()
	hydrated.Set("1-2-0.5-BTC-ETH-4|15.5")
	hydrated.SetID("x")
	hydrated.SetStatus(JC.STATE_LOADING)

	pm.Hydrate([]PanelData{hydrated})

	ref := pm.GetDataByID("x")
	if ref == nil {
		t.Fatal("Expected hydrated panel to exist")
	}
	if ref.GetStatus() != JC.STATE_LOADING {
		t.Errorf("Expected status LOADING after hydrate, got %d", ref.GetStatus())
	}
	if !ref.HasParent() {
		t.Error("Expected hydrated panel to have parent set")
	}

	serialized := pm.Serialize()
	if len(serialized) != 1 {
		t.Errorf("Expected 1 serialized entry, got %d", len(serialized))
	}
	panelsMapTurnOnLogs()
}
