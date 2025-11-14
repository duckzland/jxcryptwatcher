package types

import (
	"log"
	"os"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"

	JC "jxwatcher/core"
)

type tickerDataNullWriter struct{}

func (tickerDataNullWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func tickerDataTurnOffLogs() {
	log.SetOutput(tickerDataNullWriter{})
}

func tickerDataTurnOnLogs() {
	log.SetOutput(os.Stdout)
}

func TestTickerDataInitAndSet(t *testing.T) {
	tickerDataTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	td := NewTickerData()
	td.Init()
	td.Set("123.45")

	if td.Get() != "123.45" {
		t.Errorf("Expected value '123.45', got '%s'", td.Get())
	}
	if td.GetOldKey() != JC.STRING_EMPTY {
		t.Error("Expected oldKey to be empty after first Set")
	}

	td.Set("456.78")
	if td.GetOldKey() != "123.45" {
		t.Errorf("Expected oldKey to be '123.45', got '%s'", td.GetOldKey())
	}
	tickerDataTurnOnLogs()
}

func TestTickerDataMetadataAndChecks(t *testing.T) {
	tickerDataTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	td := NewTickerData()
	td.Init()
	td.SetType("price")
	td.SetTitle("BTC Price")
	td.SetFormat("currency")
	td.SetID("btc")
	td.SetStatus(JC.STATE_LOADED)

	if !td.IsType("price") || !td.IsTitle("BTC Price") || !td.IsFormat("currency") || !td.IsID("btc") || !td.IsStatus(JC.STATE_LOADED) {
		t.Error("Metadata checks failed")
	}
	tickerDataTurnOnLogs()
}

func TestTickerDataUpdateAndDidChange(t *testing.T) {
	tickerDataTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	RegisterTickerCache().Init()
	UseTickerCache().Insert("price", "999.99", time.Now())

	td := NewTickerData()
	td.Init()
	td.SetType("price")
	td.Set("888.88")
	td.SetStatus(JC.STATE_LOADING)

	changed := td.Update()
	if !changed {
		t.Error("Expected Update to return true")
	}
	if !td.DidChange() {
		t.Error("Expected DidChange to be true")
	}
	tickerDataTurnOnLogs()
}

func TestTickerDataFormatContent(t *testing.T) {
	tickerDataTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	td := NewTickerData()
	td.Init()
	td.Set("1234.56")

	td.SetFormat("nodecimal")
	if td.FormatContent() != "1235" {
		t.Errorf("Expected '1235', got '%s'", td.FormatContent())
	}

	td.SetFormat("currency")
	if td.FormatContent() != "$1,234.56" {
		t.Errorf("Expected '$1,234.56', got '%s'", td.FormatContent())
	}

	td.SetFormat("percentage")
	if td.FormatContent() != "1234.56/100" {
		t.Errorf("Expected '1234.56/100', got '%s'", td.FormatContent())
	}
	tickerDataTurnOnLogs()
}

func TestTickerDataSerialize(t *testing.T) {
	tickerDataTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	td := NewTickerData()
	td.Init()
	td.Set("777.77")
	td.SetType("price")
	td.SetTitle("ETH Price")
	td.SetFormat("number")
	td.SetStatus(JC.STATE_LOADED)
	td.SetOldKey("666.66")

	cache := td.Serialize()
	if cache.Key != "777.77" || cache.OldKey != "666.66" || cache.Type != "price" || cache.Title != "ETH Price" || cache.Format != "number" {
		t.Error("Serialization mismatch")
	}
	tickerDataTurnOnLogs()
}
