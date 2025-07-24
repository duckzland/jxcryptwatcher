package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	JC "jxwatcher/core"
)

type PanelsType []PanelType

func (p *PanelsType) LoadFile() *PanelsType {
	b := bytes.NewBuffer(nil)
	f, _ := os.Open(JC.BuildPathRelatedToUserDirectory([]string{"jxcryptwatcher", "panels.json"}))
	io.Copy(b, f)
	f.Close()

	err := json.Unmarshal(b.Bytes(), p)

	if err != nil {
		wrappedErr := fmt.Errorf("Failed to load panels.json: %w", err)
		log.Println(wrappedErr)
	} else {
		log.Print("Panels Loaded")
	}

	return p
}

func (p *PanelsType) SaveFile(maps PanelsMapType) bool {

	np := []PanelType{}
	list := maps.Get()
	for _, pdt := range list {
		pk := pdt.UsePanelKey()
		np = append(np, PanelType{
			Source:   pk.GetSourceCoinInt(),
			Target:   pk.GetTargetCoinInt(),
			Value:    pk.GetSourceValueFloat(),
			Decimals: pk.GetDecimalsInt(),
		})
	}

	jsonData, err := json.MarshalIndent(np, "", "  ")
	if err != nil {
		log.Println(err)
		return false
	}

	// Save to file
	err = os.WriteFile(JC.BuildPathRelatedToUserDirectory([]string{"jxcryptwatcher", "panels.json"}), jsonData, 0644)
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
		maps.Append(pko.GenerateKeyFromPanel(*pp, 0))
	}
}

func PanelsInit() {
	BP = PanelsMapType{}
	BP.Init()

	Cryptos := CryptosType{}

	BP.SetMaps(Cryptos.CheckFile().LoadFile().ConvertToMap())
	Panels := PanelsType{}
	Panels.CheckFile().LoadFile().ConvertToMap(&BP)
}

func SavePanels() bool {
	Panels := PanelsType{}
	return Panels.SaveFile(BP)
}

func RemovePanel(i int) bool {
	return BP.Remove(i)
}
