package types

import (
	"bytes"
	"fmt"
	"io"
	"sync"

	"fyne.io/fyne/v2/storage"

	"github.com/buger/jsonparser"

	json "github.com/goccy/go-json"

	JC "jxwatcher/core"
)

var panelsMu sync.RWMutex

type panelsType []panelType

func (p *panelsType) load() *panelsType {
	panelsMu.Lock()
	defer panelsMu.Unlock()

	fileURI, err := storage.ParseURI(JC.BuildPathRelatedToUserDirectory([]string{"panels.json"}))
	if err != nil {
		JC.Logln("Error parsing URI for file:", err)
		JC.Notify(JC.NotifyUnableToLoadPanelsDataFromFile)
		return p
	}

	reader, err := storage.Reader(fileURI)
	if err != nil {
		JC.Logln("Failed to open panels.json:", err)
		JC.Notify(JC.NotifyUnableToLoadPanelsDataFromFile)
		return p
	}
	defer reader.Close()

	buffer := bytes.NewBuffer(nil)
	if _, err := io.Copy(buffer, reader); err != nil {
		JC.Logln("Failed to read panels.json:", err)
		JC.Notify(JC.NotifyUnableToLoadPanelsDataFromFile)
		return p
	}

	if err := p.parseJSON(buffer.Bytes()); err != nil {
		*p = panelsType{}
		JC.Logln(err)
		JC.Notify(JC.NotifyUnableToLoadPanelsDataFromFile)
	} else {
		JC.Logln("Panels Loaded")
	}

	return p
}

func (p *panelsType) parseJSON(data []byte) error {
	*p = nil

	_, err := jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		var panel panelType

		if src, e := jsonparser.GetInt(value, "source"); e == nil {
			panel.Source = src
		}
		if tgt, e := jsonparser.GetInt(value, "target"); e == nil {
			panel.Target = tgt
		}
		if val, e := jsonparser.GetFloat(value, "value"); e == nil {
			panel.Value = val
		}
		if dec, e := jsonparser.GetInt(value, "decimals"); e == nil {
			panel.Decimals = dec
		}
		if ss, e := jsonparser.GetString(value, "source_symbol"); e == nil {
			panel.SourceSymbol = ss
		}
		if ts, e := jsonparser.GetString(value, "target_symbol"); e == nil {
			panel.TargetSymbol = ts
		}

		*p = append(*p, panel)
	})

	if err != nil {
		return fmt.Errorf("failed to parse panels.json: %w", err)
	}
	return nil
}

func (p *panelsType) save(maps *panelsMapType) bool {
	panelsMu.RLock()
	defer panelsMu.RUnlock()

	maps.mu.RLock()
	data := make([]PanelData, len(maps.GetData()))
	copy(data, maps.GetData())
	maps.mu.RUnlock()

	np := []panelType{}
	for _, pot := range data {
		pdt := maps.GetDataByID(pot.GetID())
		if pdt == nil {
			continue
		}
		pk := pdt.UsePanelKey()

		panel := panelType{
			Source:       pk.GetSourceCoinInt(),
			Target:       pk.GetTargetCoinInt(),
			Value:        pk.GetSourceValueFloat(),
			Decimals:     pk.GetDecimalsInt(),
			SourceSymbol: pk.GetSourceSymbolString(),
			TargetSymbol: pk.GetTargetSymbolString(),
		}

		np = append(np, panel)
	}

	jsonData, err := json.MarshalIndent(np, JC.STRING_EMPTY, "  ")
	if err != nil {
		JC.Logln(err)
		return false
	}

	return JC.CreateFile(JC.BuildPathRelatedToUserDirectory([]string{"panels.json"}), string(jsonData))
}

func (p *panelsType) create() *panelsType {
	panelsMu.Lock()
	defer panelsMu.Unlock()

	JC.CreateFile(JC.BuildPathRelatedToUserDirectory([]string{"panels.json"}), "[]")
	return p
}

func (p *panelsType) check() *panelsType {
	exists, err := JC.FileExists(JC.BuildPathRelatedToUserDirectory([]string{"panels.json"}))
	if !exists {
		p.create()
	}

	if err != nil {
		JC.Logln(err)
	}

	return p
}

func (p *panelsType) convert(maps *panelsMapType) {
	panelsMu.RLock()
	defer panelsMu.RUnlock()

	for i := range *p {
		pp := &(*p)[i]

		pko := panelKeyType{}
		pko.GenerateKeyFromPanel(*pp, JC.ToBigFloat(-1))

		pp.SourceSymbol = maps.GetSymbolById(pko.GetSourceCoinString())
		pp.TargetSymbol = maps.GetSymbolById(pko.GetTargetCoinString())

		// JC.Logf("Generated key: %v", pko.GenerateKeyFromPanel(*pp, JC.ToBigFloat(-1)))

		maps.Append(pko.GenerateKeyFromPanel(*pp, JC.ToBigFloat(-1)))
	}
}

func PanelsInit() {
	maps := UsePanelMaps().GetMaps()
	UsePanelMaps().Init()
	UsePanelMaps().SetMaps(maps)

	panelsMu.Lock()
	panels := panelsType{}
	panelsMu.Unlock()

	panels.check().load().convert(UsePanelMaps())
}

func SavePanels() bool {
	panelsMu.RLock()
	panels := panelsType{}
	panelsMu.RUnlock()

	return panels.save(UsePanelMaps())
}

func RemovePanel(uuid string) bool {
	return UsePanelMaps().Remove(uuid)
}
