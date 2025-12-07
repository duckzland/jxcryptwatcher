package types

import (
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

func TestCryptoLoaderParseValid(t *testing.T) {
	cryptoTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	raw := []byte(`[ [123, "Bitcoin", "BTC", "unused", 1, 1] ]`)
	var loader cryptosLoaderType
	err := loader.parseJSON([]byte(`{"values":` + string(raw) + `}`))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(loader.Values) != 1 {
		t.Fatalf("Expected 1 crypto, got %d", len(loader.Values))
	}
	c := loader.Values[0]
	if c.Id != 123 || c.Name != "Bitcoin" || c.Symbol != "BTC" || c.IsActive != 1 || c.Status != 1 {
		t.Errorf("Unexpected values: %+v", c)
	}
	cryptoTurnOnLogs()
}

func TestCryptoLoaderParseInactive(t *testing.T) {
	cryptoTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	raw := []byte(`[ [123, "Bitcoin", "BTC", "unused", 0, 1] ]`)
	var loader cryptosLoaderType
	err := loader.parseJSON([]byte(`{"values":` + string(raw) + `}`))
	if err == nil {
		t.Errorf("Expected error for inactive crypto, got nil")
	}
	if len(loader.Values) != 0 {
		t.Errorf("Expected no active cryptos, got: %+v", loader.Values)
	}
	cryptoTurnOnLogs()
}

func TestCryptoLoaderParseInvalidLength(t *testing.T) {
	cryptoTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	raw := []byte(`[ [123, "Bitcoin", "BTC"] ]`)
	var loader cryptosLoaderType
	err := loader.parseJSON([]byte(`{"values":` + string(raw) + `}`))
	if err == nil {
		t.Errorf("Expected non-nil error for invalid length")
	}
	if len(loader.Values) != 0 {
		t.Errorf("Expected no cryptos, got: %+v", loader.Values)
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
