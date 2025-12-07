package types

import (
	"log"
	"os"
	"testing"

	"fyne.io/fyne/v2/test"

	JC "jxwatcher/core"
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
	if !p.save(pm) {
		t.Error("Failed to save panels")
	}

	// Load from file
	loaded := (&panelsType{}).load()
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

	p.convert(pm)
	if pm.TotalData() != 1 {
		t.Errorf("Expected 1 panel in map, got %d", pm.TotalData())
	}
	panelsTurnOnLogs()
}

func TestPanelsTypeParseJSON(t *testing.T) {
	panelsTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	raw := []byte(`[
		{
			"source": 1,
			"target": 2,
			"value": 0.5,
			"decimals": 4,
			"source_symbol": "BTC",
			"target_symbol": "ETH"
		}
	]`)

	p := &panelsType{}
	err := p.parseJSON(raw)
	if err != nil {
		t.Errorf("Unexpected error parsing JSON: %v", err)
	}

	if len(*p) != 1 {
		t.Fatalf("Expected 1 panel, got %d", len(*p))
	}
	panel := (*p)[0]
	if panel.Source != 1 {
		t.Errorf("Expected Source=1, got %d", panel.Source)
	}
	if panel.Target != 2 {
		t.Errorf("Expected Target=2, got %d", panel.Target)
	}
	if panel.Value != 0.5 {
		t.Errorf("Expected Value=0.5, got %f", panel.Value)
	}
	if panel.Decimals != 4 {
		t.Errorf("Expected Decimals=4, got %d", panel.Decimals)
	}
	if panel.SourceSymbol != "BTC" {
		t.Errorf("Expected SourceSymbol=BTC, got %s", panel.SourceSymbol)
	}
	if panel.TargetSymbol != "ETH" {
		t.Errorf("Expected TargetSymbol=ETH, got %s", panel.TargetSymbol)
	}

	panelsTurnOnLogs()
}
