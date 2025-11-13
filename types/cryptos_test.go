package types

import (
	"log"
	"os"
	"testing"

	json "github.com/goccy/go-json"

	"fyne.io/fyne/v2/test"

	JC "jxwatcher/core"
)

type cryptosLoaderNullWriter struct{}

func (cryptosLoaderNullWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func cryptosLoaderTurnOffLogs() {
	log.SetOutput(cryptosLoaderNullWriter{})
}

func cryptosLoaderTurnOnLogs() {
	log.SetOutput(os.Stdout)
}

func TestCryptosLoaderConvertToMap(t *testing.T) {
	cryptosLoaderTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	loader := &cryptosLoaderType{
		Values: []cryptoType{
			{Id: 1, Name: "Bitcoin", Symbol: "BTC", IsActive: 1, Status: 1},
			{Id: 2, Name: "Ethereum", Symbol: "ETH", IsActive: 1, Status: 1},
			{Id: 3, Name: "DeadCoin", Symbol: "DC", IsActive: 0, Status: 0},
		},
	}

	cm := loader.convert()
	if cm.IsEmpty() {
		t.Error("Expected map to be populated")
	}
	if cm.GetDisplayById("1") == "" || cm.GetDisplayById("2") == "" {
		t.Error("Expected valid entries for active cryptos")
	}
	if cm.GetDisplayById("3") != "" {
		t.Error("Expected inactive crypto to be excluded")
	}
	cryptosLoaderTurnOnLogs()
}

func TestCryptosLoaderLoadFileEmpty(t *testing.T) {
	cryptosLoaderTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	// Simulate empty file load
	JC.CreateFile(JC.BuildPathRelatedToUserDirectory([]string{"cryptos.json"}), `{"values":[]}`)

	loader := &cryptosLoaderType{}
	loader.load()

	if len(loader.Values) != 0 {
		t.Errorf("Expected empty values, got: %d", len(loader.Values))
	}
	cryptosLoaderTurnOnLogs()
}

func TestCryptosLoaderTransformAndUnmarshal(t *testing.T) {
	cryptosLoaderTurnOffLogs()
	t.Setenv("FYNE_STORAGE", t.TempDir())
	test.NewApp()

	raw := `{
		"fields": ["id", "name", "symbol", "slug", "is_active", "status", "rank", "address", "first_historical_data", "last_historical_data"],
		"values": [
			[1, "Bitcoin", "BTC", "bitcoin", 1, 1, 1, [], "2010-07-13T00:05:00.000Z", "2025-09-03T02:55:00.000Z"],
			[2, "Ethereum", "ETH", "ethereum", 1, 1, 2, [], "2015-08-07T00:00:00.000Z", "2025-09-03T02:55:00.000Z"]
		]
	}`

	// Simulate preprocessing: extract first 6 fields from each entry
	var rawData struct {
		Values [][]interface{} `json:"values"`
	}
	_ = json.Unmarshal([]byte(raw), &rawData)

	var trimmed struct {
		Values [][]interface{} `json:"values"`
	}
	for _, entry := range rawData.Values {
		if len(entry) >= 6 {
			trimmed.Values = append(trimmed.Values, entry[:6])
		}
	}
	processed, _ := json.Marshal(trimmed)

	var loader cryptosLoaderType
	err := json.Unmarshal(processed, &loader)
	if err != nil {
		t.Errorf("Failed to unmarshal trimmed data: %v", err)
	}
	if len(loader.Values) != 2 {
		t.Errorf("Expected 2 cryptos, got %d", len(loader.Values))
	}
	cryptosLoaderTurnOnLogs()
}
