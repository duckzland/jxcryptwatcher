package types

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	JC "jxwatcher/core"
)

type PanelsType []PanelType

func (p *PanelsType) LoadFile() *PanelsType {
	f, err := os.Open(JC.BuildPathRelatedToUserDirectory([]string{"jxcryptwatcher", "panels.json"}))
	if err != nil {
		log.Println("Failed to open panels.json:", err)
		return p
	}
	defer f.Close()

	decoder := json.NewDecoder(f)
	if err := decoder.Decode(p); err != nil {
		log.Println(fmt.Errorf("Failed to decode panels.json: %w", err))
	} else {
		log.Println("Panels Loaded")
	}

	return p
}

func (p *PanelsType) SaveFile(maps *PanelsMapType) bool {

	// It is ok to just copy the object as we are going
	// to write the data to file
	maps.mu.Lock()
	data := make([]PanelDataType, len(maps.Data))
	copy(data, maps.Data)
	maps.mu.Unlock()

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

	err = os.WriteFile(
		JC.BuildPathRelatedToUserDirectory([]string{"jxcryptwatcher", "panels.json"}),
		jsonData,
		0644,
	)
	if err != nil {
		log.Println(err)
		return false
	}

	return true
}

func (p *PanelsType) CreateFile() *PanelsType {
	JC.CreateFile(JC.BuildPathRelatedToUserDirectory([]string{"jxcryptwatcher", "panels.json"}), "[]")
	return p
}

func (p *PanelsType) CheckFile() *PanelsType {
	exists, err := JC.FileExists(JC.BuildPathRelatedToUserDirectory([]string{"jxcryptwatcher", "panels.json"}))
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
		pp := &(*p)[i] // Get pointer to the actual element

		pp.SourceSymbol = maps.GetSymbolById(strconv.FormatInt(pp.Source, 10))
		pp.TargetSymbol = maps.GetSymbolById(strconv.FormatInt(pp.Target, 10))

		pko := PanelKeyType{}
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

func RemovePanel(i int) bool {
	return BP.Remove(i)
}
