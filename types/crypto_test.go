package types

import (
	"encoding/json"
	"log"
	"os"
	"testing"

	"fyne.io/fyne/v2/test"
)

type cryptoNullWriter struct{}

func (cryptoNullWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func cryptoTurnOffLogs() {
	log.SetOutput(cryptoNullWriter{})
}

func cryptoTurnOnLogs() {
	log.SetOutput(os.Stdout)
}

func TestCryptoTypeUnmarshalValid(t *testing.T) {
	cryptoTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	raw := []byte(`[123, "Bitcoin", "BTC", "unused", 1, 1]`)
	var c cryptoType
	err := json.Unmarshal(raw, &c)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if c.Id != 123 || c.Name != "Bitcoin" || c.Symbol != "BTC" || c.IsActive != 1 || c.Status != 1 {
		t.Errorf("Unexpected values: %+v", c)
	}
	cryptoTurnOnLogs()
}

func TestCryptoTypeUnmarshalInactive(t *testing.T) {
	cryptoTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	raw := []byte(`[123, "Bitcoin", "BTC", "unused", 0, 1]`)
	var c cryptoType
	err := json.Unmarshal(raw, &c)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if c.Id != 0 {
		t.Errorf("Expected zeroed struct for inactive crypto, got: %+v", c)
	}
	cryptoTurnOnLogs()
}

func TestCryptoTypeUnmarshalInvalidLength(t *testing.T) {
	cryptoTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	raw := []byte(`[123, "Bitcoin", "BTC"]`)
	var c cryptoType
	err := json.Unmarshal(raw, &c)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if c.Id != 0 {
		t.Errorf("Expected zeroed struct for invalid length, got: %+v", c)
	}
	cryptoTurnOnLogs()
}

func TestCryptoTypeCreateKey(t *testing.T) {
	cryptoTurnOffLogs()
	c := cryptoType{
		Id:     42,
		Name:   "Ethereum",
		Symbol: "ETH",
	}
	key := c.createKey()
	expected := "42|ETH - Ethereum"
	if key != expected {
		t.Errorf("Unexpected key. Got: %s, Want: %s", key, expected)
	}
	cryptoTurnOnLogs()
}
