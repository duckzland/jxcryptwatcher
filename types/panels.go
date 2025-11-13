package types

import (
	"bytes"
	"fmt"
	"io"
	"sync"

	"fyne.io/fyne/v2/storage"

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
		JC.Notify("Unable to load panels data from file.")
		return p
	}

	reader, err := storage.Reader(fileURI)
	if err != nil {
		JC.Logln("Failed to open panels.json:", err)
		JC.Notify("Unable to load panels data from file.")
		return p
	}
	defer reader.Close()

	buffer := bytes.NewBuffer(nil)
	if _, err := io.Copy(buffer, reader); err != nil {
		JC.Logln("Failed to read panels.json:", err)
		JC.Notify("Unable to load panels data from file.")
		return p
	}

	if err := json.Unmarshal(buffer.Bytes(), p); err != nil {
		p = &panelsType{}
		JC.Logln(fmt.Errorf("Failed to decode panels.json: %w", err))
		JC.Notify("Unable to load panels data from file.")
	} else {
		JC.Logln("Panels Loaded")
	}

	return p
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

	jsonData, err := json.MarshalIndent(np, "", "  ")
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
