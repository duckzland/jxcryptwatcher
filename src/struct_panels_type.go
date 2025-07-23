package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"fyne.io/fyne/v2"
)

type PanelsType []PanelType

func (p *PanelsType) LoadFile() *PanelsType {
	b := bytes.NewBuffer(nil)
	f, _ := os.Open(buildPathRelatedToUserDirectory([]string{"jxcryptwatcher", "panels.json"}))
	io.Copy(b, f)
	f.Close()

	err := json.Unmarshal(b.Bytes(), p)

	if err != nil {
		wrappedErr := fmt.Errorf("Failed to load panels.json: %w", err)
		log.Fatal(wrappedErr)
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
		log.Fatalln(err)
	}

	// Save to file
	err = os.WriteFile(buildPathRelatedToUserDirectory([]string{"jxcryptwatcher", "panels.json"}), jsonData, 0644)
	if err != nil {
		log.Fatalln(err)
	}

	return true
}

func (p *PanelsType) CreateFile() *PanelsType {
	createFile(buildPathRelatedToUserDirectory([]string{"jxcryptwatcher", "panels.json"}), "[]")
	return p
}

func (p *PanelsType) CheckFile() *PanelsType {
	exists, err := fileExists(buildPathRelatedToUserDirectory([]string{"jxcryptwatcher", "panels.json"}))
	if !exists {
		p.CreateFile()
	}

	if err != nil {
		log.Fatalln(err)
	}

	return p
}

func (p *PanelsType) ConvertToMap(maps *PanelsMapType) {
	for _, panel := range *p {
		pko := PanelKeyType{}
		maps.Append(pko.GenerateKeyFromPanel(panel, 0))
	}
}

var BP PanelsMapType

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

func RemovePanel(i int) {
	BP.Remove(i)

	if i >= 0 && i < len(Grid.Objects) {
		Grid.Objects = append(Grid.Objects[:i], Grid.Objects[i+1:]...)
	}

	fyne.Do(func() {
		Grid.Refresh()
	})
}
