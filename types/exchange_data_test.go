package types

import (
	"log"
	"math/big"
	"os"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
)

type exchangeDataNullWriter struct{}

func (exchangeDataNullWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func exchangeDataTurnOffLogs() {
	log.SetOutput(exchangeDataNullWriter{})
}

func exchangeDataTurnOnLogs() {
	log.SetOutput(os.Stdout)
}

func TestExchangeDataBasicConstruction(t *testing.T) {
	exchangeDataTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	now := time.Now()
	ex := exchangeDataType{
		SourceSymbol: "BTC",
		SourceId:     1,
		SourceAmount: 1.0,
		TargetSymbol: "ETH",
		TargetId:     2,
		TargetAmount: big.NewFloat(15.5),
		Timestamp:    now,
	}

	if ex.SourceSymbol != "BTC" || ex.TargetSymbol != "ETH" {
		t.Error("Symbol fields not set correctly")
	}
	if ex.SourceId != 1 || ex.TargetId != 2 {
		t.Error("ID fields not set correctly")
	}
	if ex.SourceAmount != 1.0 {
		t.Error("SourceAmount not set correctly")
	}
	if ex.TargetAmount == nil || ex.TargetAmount.Cmp(big.NewFloat(15.5)) != 0 {
		t.Error("TargetAmount not set correctly")
	}
	if !ex.Timestamp.Equal(now) {
		t.Error("Timestamp not set correctly")
	}
	exchangeDataTurnOnLogs()
}

func TestExchangeDataZeroValues(t *testing.T) {
	exchangeDataTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	ex := exchangeDataType{}
	if ex.SourceSymbol != "" || ex.TargetSymbol != "" {
		t.Error("Expected empty symbols")
	}
	if ex.SourceId != 0 || ex.TargetId != 0 {
		t.Error("Expected zero IDs")
	}
	if ex.SourceAmount != 0 {
		t.Error("Expected zero SourceAmount")
	}
	if ex.TargetAmount != nil {
		t.Error("Expected nil TargetAmount")
	}
	exchangeDataTurnOnLogs()
}
