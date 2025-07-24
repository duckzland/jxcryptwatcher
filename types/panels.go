package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

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
		np = append(np, PanelType{
			Source:   pdt.UsePanelKey().GetSourceCoinInt(),
			Target:   pdt.UsePanelKey().GetTargetCoinInt(),
			Value:    pdt.UsePanelKey().GetSourceValueFloat(),
			Decimals: pdt.UsePanelKey().GetDecimalsInt(),
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
	for _, panel := range *p {
		pko := PanelKeyType{}
		maps.Append(pko.GenerateKeyFromPanel(panel, 0))
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
