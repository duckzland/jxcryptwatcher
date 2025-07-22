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

func (p *PanelsType) SaveFile(maps PanelsMap) bool {

	np := []PanelType{}
	list, _ := maps.Get()
	for _, x := range list {
		pdt, ok := x.(PanelDataType)
		if !ok {
			continue
		}

		np = append(np, PanelType{
			Source:   pdt.GetSourceCoin(),
			Target:   pdt.GetTargetCoin(),
			Value:    pdt.GetSourceValue(),
			Decimals: pdt.GetDecimals(),
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

func (p *PanelsType) ConvertToMap(maps PanelsMap) {
	for _, panel := range *p {
		pk := maps.GenerateKeyFromPanel(panel, 0)
		maps.Append(pk)
	}
}

var BP PanelsMap

func PanelsInit() {
	BP = PanelsMap{}
	BP.Init()

	Cryptos := CryptosType{}

	BP.SetMaps(Cryptos.CheckFile().LoadFile().ConvertToMap())
	Panels := PanelsType{}
	Panels.CheckFile().LoadFile().ConvertToMap(BP)
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
