package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"fyne.io/fyne/v2/storage"

	JC "jxwatcher/core"
)

type PanelsType []PanelType

func (p *PanelsType) LoadFile() *PanelsType {
	fileURI, err := storage.ParseURI(JC.BuildPathRelatedToUserDirectory([]string{"panels.json"}))
	if err != nil {
		JC.Logln("Error getting parsing uri for file:", err)
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
		p = &PanelsType{}
		JC.Logln(fmt.Errorf("Failed to decode panels.json: %w", err))
		JC.Notify("Unable to load panels data from file.")
	} else {
		JC.Logln("Panels Loaded")
	}

	return p
}

func (p *PanelsType) SaveFile(maps *PanelsMapType) bool {
	maps.mu.RLock()
	data := make([]*PanelDataType, len(maps.GetData()))
	copy(data, maps.GetData())
	maps.mu.RUnlock()

	np := []PanelType{}
	for _, pot := range data {
		pdt := maps.GetDataByID(pot.GetID())
		if pdt == nil {
			continue
		}
		pk := pdt.UsePanelKey()

		panel := PanelType{
			Source:   pk.GetSourceCoinInt(),
			Target:   pk.GetTargetCoinInt(),
			Value:    pk.GetSourceValueFloat(),
			Decimals: pk.GetDecimalsInt(),
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

func (p *PanelsType) CreateFile() *PanelsType {
	JC.CreateFile(JC.BuildPathRelatedToUserDirectory([]string{"panels.json"}), "[]")
	return p
}

func (p *PanelsType) CheckFile() *PanelsType {
	exists, err := JC.FileExists(JC.BuildPathRelatedToUserDirectory([]string{"panels.json"}))
	if !exists {
		p.CreateFile()
	}

	if err != nil {
		JC.Logln(err)
	}

	return p
}

func (p *PanelsType) ConvertToMap(maps *PanelsMapType) {
	for i := range *p {
		pp := &(*p)[i]

		pko := PanelKeyType{}
		pko.GenerateKeyFromPanel(*pp, -1)

		pp.SourceSymbol = maps.GetSymbolById(pko.GetSourceCoinString())
		pp.TargetSymbol = maps.GetSymbolById(pko.GetTargetCoinString())

		JC.Logf("Generated key: %v", pko.GenerateKeyFromPanel(*pp, -1))

		maps.Append(pko.GenerateKeyFromPanel(*pp, -1))
	}
}

func PanelsInit() {
	maps := BP.GetMaps()
	BP.Init()
	BP.SetMaps(maps)

	Panels := PanelsType{}
	Panels.CheckFile().LoadFile().ConvertToMap(&BP)
}

func SavePanels() bool {
	Panels := PanelsType{}
	return Panels.SaveFile(&BP)
}

func RemovePanel(uuid string) bool {
	return BP.Remove(uuid)
}
