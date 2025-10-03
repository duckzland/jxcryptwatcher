package types

import (
	JC "jxwatcher/core"
	"log"
	"os"
	"testing"

	"fyne.io/fyne/v2/test"
)

type panelsNullWriter struct{}

func (panelsNullWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func panelsTurnOffLogs() {
	log.SetOutput(panelsNullWriter{})
}

func panelsTurnOnLogs() {
	log.SetOutput(os.Stdout)
}

func TestPanelsTypeCreateSaveLoad(t *testing.T) {
	panelsTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	// Setup cryptos map
	cm := &cryptosMapType{}
	cm.Init()
	cm.Insert("1", "BTC - Bitcoin")
	cm.Insert("2", "ETH - Ethereum")

	// Setup exchange cache
	RegisterExchangeCache().Init()

	dummy := &exchangeDataType{
		SourceId:     1,
		TargetId:     2,
		SourceSymbol: "BTC",
		TargetSymbol: "ETH",
		TargetAmount: JC.ToBigFloat(42.0),
	}
	UseExchangeCache().Insert(dummy)

	// Setup panel map
	pm := &panelsMapType{}
	pm.Init()
	pm.SetMaps(cm)
	pm.Append("1-2-0.5-BTC-ETH-4|15.5")

	// Save to file
	p := &panelsType{}
	if !p.saveFile(pm) {
		t.Error("Failed to save panels")
	}

	// Load from file
	loaded := (&panelsType{}).loadFile()
	if len(*loaded) != 1 {
		t.Errorf("Expected 1 panel, got %d", len(*loaded))
	}
	panelsTurnOnLogs()
}

func TestPanelsTypeConvertToMap(t *testing.T) {
	panelsTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	// Setup cryptos map
	cm := &cryptosMapType{}
	cm.Init()
	cm.Insert("1", "BTC - Bitcoin")
	cm.Insert("2", "ETH - Ethereum")

	// Setup exchange cache
	RegisterExchangeCache().Init()

	dummy := &exchangeDataType{
		SourceId:     1,
		TargetId:     2,
		SourceSymbol: "BTC",
		TargetSymbol: "ETH",
		TargetAmount: JC.ToBigFloat(42.0),
	}
	UseExchangeCache().Insert(dummy)

	// Setup panel map
	pm := &panelsMapType{}
	pm.Init()
	pm.SetMaps(cm)

	// Setup panelsType
	p := &panelsType{
		{
			Source:       1,
			Target:       2,
			Value:        0.5,
			Decimals:     4,
			SourceSymbol: "BTC",
			TargetSymbol: "ETH",
		},
	}

	p.convertToMap(pm)
	if pm.TotalData() != 1 {
		t.Errorf("Expected 1 panel in map, got %d", pm.TotalData())
	}
	panelsTurnOnLogs()
}
