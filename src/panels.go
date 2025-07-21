package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"fyne.io/fyne/v2/data/binding"
)

type PanelsType []PanelType
type PanelType struct {
	Source   int64   `json:"source"`
	Target   int64   `json:"target"`
	Value    float64 `json:"value"`
	Decimals int64   `json:"decimals"`
}

var BindedData binding.UntypedList

func removeAt(index int, list binding.StringList) {
	values, _ := list.Get()
	if index < 0 || index >= len(values) {
		return // avoid out-of-bounds
	}

	// Remove item at index
	updated := append(values[:index], values[index+1:]...)
	list.Set(updated)
}

func loadPanels() PanelsType {

	b := bytes.NewBuffer(nil)
	f, _ := os.Open(buildPathRelatedToUserDirectory([]string{"jxcryptwatcher", "panels.json"}))
	io.Copy(b, f)
	f.Close()

	panels := PanelsType{}
	err := json.Unmarshal(b.Bytes(), &panels)

	if err != nil {
		wrappedErr := fmt.Errorf("Failed to load panels.json: %w", err)
		log.Fatal(wrappedErr)
	} else {
		log.Print("Panels Loaded")
	}

	return panels
}

func checkPanels() {
	exists, err := fileExists(buildPathRelatedToUserDirectory([]string{"jxcryptwatcher", "panels.json"}))
	if !exists {
		data := "[]"
		createFile(buildPathRelatedToUserDirectory([]string{"jxcryptwatcher", "panels.json"}), data)
	}

	if err != nil {
		log.Fatalln(err)
	}

	initPanels(loadPanels())
}

func initPanels(panels PanelsType) {
	// BindedData = binding.NewStringList()
	BindedData = binding.NewUntypedList()

	for _, panel := range panels {
		BindedData.Append(generatePanelKey(panel, 0))
	}
}

func generatePanelKey(panel PanelType, rate float32) string {
	return fmt.Sprintf("%d-%d-%f-%d|%f", panel.Source, panel.Target, panel.Value, panel.Decimals, rate)
}

func savePanels() bool {

	panels := PanelsType{}
	list, _ := BindedData.Get()
	for _, val := range list {
		panels = append(panels, PanelType{
			Source:   getPanelSourceCoin(val),
			Target:   getPanelTargetCoin(val),
			Value:    getPanelSourceValue(val),
			Decimals: getPanelDecimals(val),
		})
	}

	jsonData, err := json.MarshalIndent(panels, "", "  ")
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

func appendPanel(pk string) bool {

	if !validatePanel(pk) {
		return false
	}
	ex := ExchangeDataType{}
	data := ex.GetRate(pk)
	npk := updatePanelValue(pk, float32(data.TargetAmount))
	BindedData.Append(npk)
	Grid.Add(generatePanel(npk))

	return true

}

func insertPanel(pk string, index int) bool {
	if !validatePanel(pk) {
		return false
	}

	ex := ExchangeDataType{}
	data := ex.GetRate(pk)
	npk := updatePanelValue(pk, float32(data.TargetAmount))

	if npk != pk {
		BindedData.SetValue(index, npk)
		Grid.Objects[index] = generatePanel(npk)
	}

	return true
}

func updatePanel(pk string) bool {

	if !validatePanel(pk) {
		return false
	}

	pi := getPanelIndex(pk)

	if pi != -1 {

		ex := ExchangeDataType{}
		data := ex.GetRate(pk)
		npk := updatePanelValue(pk, float32(data.TargetAmount))

		if npk != pk {
			BindedData.SetValue(pi, npk)
			Grid.Objects[pi] = generatePanel(npk)

			return true
		}
	}

	return false
}

func updatePanelValue(pk string, rate float32) string {

	if validatePanelKey(pk) {
		pkk := strings.Split(pk, "|")
		return fmt.Sprintf("%s|%f", pkk[0], rate)
	}

	return pk
}

func getPanelIndex(panelKey string) int {
	list, _ := BindedData.Get()
	for i, pk := range list {
		if pk == panelKey {
			return i
		}
	}

	return -1
}

func getPanelValue(pk string) float64 {

	if validatePanelKey(pk) {
		pkv := strings.Split(pk, "|")
		value, err := strconv.ParseFloat(pkv[1], 64)
		if err == nil {
			return value
		}
	}
	return 0
}

func getPanelSourceCoin(pk string) int64 {
	if validatePanelKey(pk) {
		pkm := strings.Split(pk, "|")
		pkv := strings.Split(pkm[0], "-")

		source, err := strconv.ParseInt(pkv[0], 10, 64)
		if err == nil {
			return source
		}
	}
	return 0
}

func getPanelTargetCoin(pk string) int64 {
	if validatePanelKey(pk) {
		pkm := strings.Split(pk, "|")
		pkv := strings.Split(pkm[0], "-")

		target, err := strconv.ParseInt(pkv[1], 10, 64)
		if err == nil {
			return target
		}
	}

	return 0
}

func getPanelSourceValue(pk string) float64 {
	if validatePanelKey(pk) {
		pkm := strings.Split(pk, "|")
		pkv := strings.Split(pkm[0], "-")

		value, err := strconv.ParseFloat(pkv[2], 64)
		if err == nil {
			return value
		}
	}

	return 0
}

func getPanelDecimals(pk string) int64 {
	if validatePanelKey(pk) {
		pkm := strings.Split(pk, "|")
		pkv := strings.Split(pkm[0], "-")

		decimals, err := strconv.ParseInt(pkv[3], 10, 64)
		if err == nil {
			return decimals
		}
	}

	return 0
}

func removePanel(index int) {

	if index >= 0 && index < len(Grid.Objects) {
		Grid.Objects = append(Grid.Objects[:index], Grid.Objects[index+1:]...)
	}

	removeAt(index, BindedData)
}

func validatePanelKey(pk string) bool {
	pkv := strings.Split(pk, "|")
	if len(pkv) != 2 {
		return false
	}

	pkt := strings.Split(pkv[0], "-")
	if len(pkt) != 4 {
		return false
	}

	return true
}

func validatePanel(pk string) bool {
	if !validatePanelKey(pk) {
		return false
	}

	sid := getPanelSourceCoin(pk)
	tid := getPanelTargetCoin(pk)

	if !validateCryptoId(sid) {
		return false
	}

	if !validateCryptoId(tid) {
		return false
	}

	return true
}
