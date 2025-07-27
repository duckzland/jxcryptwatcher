package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"fyne.io/fyne/v2/storage"

	JC "jxwatcher/core"
)

type PanelsType []PanelType

func (p *PanelsType) LoadFile() *PanelsType {

	// Build the file URI relative to Fyne's root storage
	fileURI, err := storage.ParseURI(JC.BuildPathRelatedToUserDirectory([]string{"panels.json"}))
	if err != nil {
		log.Println("Error getting parsing uri for file:", err)
		JC.Notify("Failed loading panels")
		return p
	}

	// Attempt to open the file with Fyne
	reader, err := storage.Reader(fileURI)
	if err != nil {
		log.Println("Failed to open panels.json:", err)
		JC.Notify("Failed loading panels")
		return p
	}
	defer reader.Close()

	// Read the JSON data
	buffer := bytes.NewBuffer(nil)
	if _, err := io.Copy(buffer, reader); err != nil {
		log.Println("Failed to read panels.json:", err)
		JC.Notify("Failed loading panels")
		return p
	}

	// Decode JSON into your struct
	if err := json.Unmarshal(buffer.Bytes(), p); err != nil {
		p = &PanelsType{}
		log.Println(fmt.Errorf("Failed to decode panels.json: %w", err))
		JC.Notify("Failed loading panels")
	} else {
		log.Println("Panels Loaded")
	}

	return p
}

func (p *PanelsType) SaveFile(maps *PanelsMapType) bool {

	// It is ok to just copy the object as we are going
	// to write the data to file
	maps.mu.RLock()
	data := make([]PanelDataType, len(maps.Data))
	copy(data, maps.Data)
	maps.mu.RUnlock()

	np := []PanelType{}
	for i := range data {
		pdt := maps.GetDataByIndex(i)
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
		log.Println(err)
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
		log.Println(err)
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

		log.Printf("Generated key: %v", pko.GenerateKeyFromPanel(*pp, -1))

		maps.Append(pko.GenerateKeyFromPanel(*pp, -1))
	}
}

func PanelsInit() {
	BP = PanelsMapType{}
	BP.Init()

	Cryptos := &CryptosType{}
	CM := Cryptos.CheckFile().LoadFile().ConvertToMap()

	BP.SetMaps(CM)

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
