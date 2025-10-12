package types

import (
	"log"
	"os"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"

	JC "jxwatcher/core"
)

type tickerMapNullWriter struct{}

func (tickerMapNullWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func tickerMapTurnOffLogs() {
	log.SetOutput(tickerMapNullWriter{})
}

func tickerMapTurnOnLogs() {
	log.SetOutput(os.Stdout)
}

func TestTickersMapUpdateAndGetData(t *testing.T) {
	tickerMapTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	// Properly register and initialize the cache
	RegisterTickerCache().Init()
	UseTickerCache().Insert("price", "999.99", time.Now())

	tm := &tickersMapType{}
	tm.Init()

	td := NewTickerData()
	td.SetID("btc")
	td.SetType("price")
	td.SetTitle("BTC Price")
	td.SetFormat("currency")

	tm.Add(td)
	if tm.IsEmpty() {
		t.Error("Expected tickers map to have data after Add")
	}
	if len(tm.GetData()) != 1 {
		t.Errorf("Expected 1 ticker, got %d", len(tm.GetData()))
	}
	tickerMapTurnOnLogs()
}

func TestTickersMapFilterAndReset(t *testing.T) {
	tickerMapTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	tm := &tickersMapType{}
	tm.Init()

	td1 := NewTickerData()
	td1.Init()
	td1.SetID("btc")
	td1.SetType("price")
	td1.Set("123.45")
	td1.SetStatus(JC.STATE_LOADED)

	td2 := NewTickerData()
	td2.Init()
	td2.SetID("eth")
	td2.SetType("volume")
	td2.Set("678.90")
	td2.SetStatus(JC.STATE_LOADED)

	tm.SetData([]TickerData{td1, td2})

	filtered := tm.GetDataByType("price")
	if len(filtered) != 1 {
		t.Errorf("Expected 1 filtered ticker, got %d", len(filtered))
	}

	tm.Reset()
	for _, td := range tm.GetData() {
		if td.Get() != "" || !td.IsStatus(JC.STATE_LOADING) {
			t.Error("Expected ticker to be reset")
		}
	}
	tickerMapTurnOnLogs()
}

func TestTickersMapSerialize(t *testing.T) {
	tickerMapTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	tm := &tickersMapType{}
	tm.Init()

	td := NewTickerData()
	tm.Add(td)

	td.SetID("btc")
	td.SetType("price")
	td.Set("123.45")
	td.SetStatus(JC.STATE_LOADED)

	serialized := tm.Serialize()

	if len(serialized) != 1 {
		t.Errorf("Expected 1 serialized entry, got %d", len(serialized))
		return
	}

	if serialized[0].Key != "123.45" {
		t.Errorf("Expected Key '123.45', got '%s'", serialized[0].Key)
	}
	if serialized[0].Type != "price" {
		t.Errorf("Expected Type 'price', got '%s'", serialized[0].Type)
	}
	tickerMapTurnOnLogs()
}
