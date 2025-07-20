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
)

/**
 * Defining Struct for panels.json
 */
type PanelsType []PanelType
type PanelType struct {
	Source   int64   `json:"source"`
	Target   int64   `json:"target"`
	Value    float64 `json:"value"`
	Decimals int64   `json:"decimals"`
}

/**
 * Global variables
 */
var Panels PanelsType

/**
 * Load Configuration Json into memory
 */
func loadPanels() {

	b := bytes.NewBuffer(nil)
	f, _ := os.Open(buildPathRelatedToUserDirectory([]string{"jxcryptwatcher", "panels.json"}))
	io.Copy(b, f)
	f.Close()

	err := json.Unmarshal(b.Bytes(), &Panels)

	if err != nil {
		wrappedErr := fmt.Errorf("Failed to load panels.json: %w", err)
		log.Fatal(wrappedErr)
	} else {
		log.Print("Panels Loaded")
	}
}

/**
 * Helper function to check fo panels.json and try to regenerate it with empty array when not found
 */
func checkPanels() {
	exists, err := fileExists(buildPathRelatedToUserDirectory([]string{"jxcryptwatcher", "panels.json"}))
	if !exists {
		data := "[]"
		createFile(buildPathRelatedToUserDirectory([]string{"jxcryptwatcher", "panels.json"}), data)
	}

	if err != nil {
		log.Fatalln(err)
	}

	loadPanels()
}

/**
 * Helper function for converting PanelType values into a string
 */
func generatePanelKey(panel PanelType, rate float32) string {
	return fmt.Sprintf("%d-%d-%f-%d|%f", panel.Source, panel.Target, panel.Value, panel.Decimals, rate)
}

/**
 * Helper function for retrieving registered panels using its panel key
 * This will return -1 if no panel found
 */
func getPanelByKey(panelKey string) int {
	pk := strings.Split(panelKey, "|")

	for i, panel := range Panels {
		pkt := generatePanelKey(panel, 0)
		pkk := strings.Split(pkt, "|")

		if pkk[0] == pk[0] {
			return i
		}
	}

	return -1
}

func savePanels() bool {

	jsonData, err := json.MarshalIndent(Panels, "", "  ")
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

func appendPanel(panel PanelType) {
	Panels = append(Panels, panel)
	pk := generatePanelKey(panel, 0)
	BindedData.Append(pk)

	data := getExchangeData(panel)
	Grid.Add(generatePanel(panel, data))

}

func insertPanel(panel PanelType, index int) {
	if len(Panels) > index {
		Panels[index] = panel
		pk := generatePanelKey(panel, 0)
		BindedData.SetValue(index, pk)

		data := getExchangeData(panel)
		Grid.Objects[index] = generatePanel(panel, data)
	}
}

func updatePanel(panel PanelType, pk string) bool {

	pi := getPanelByKey(pk)

	if pi != -1 && len(Panels) > pi {

		data := getExchangeData(panel)
		npk := generatePanelKey(panel, float32(data.TargetAmount))

		if npk != pk {
			Panels[pi] = panel
			BindedData.SetValue(pi, pk)
			data := getExchangeData(panel)
			Grid.Objects[pi] = generatePanel(panel, data)

			return true
		}
	}

	return false
}

func updatePanelByKey(pk string) bool {
	pi := getPanelByKey(pk)
	if pi != -1 && len(Panels) > pi {
		return updatePanel(Panels[pi], pk)
	}
	return false
}

func removePanel(index int) {

	if index >= 0 && index < len(Panels) {
		Panels = append(Panels[:index], Panels[index+1:]...)
	}

	if index >= 0 && index < len(Grid.Objects) {
		Grid.Objects = append(Grid.Objects[:index], Grid.Objects[index+1:]...)
	}

	removeAt(index, BindedData)
}

func createPanelObjectFromKey(panelKey string) PanelType {
	panel := PanelType{}

	pkv := strings.Split(panelKey, "|")
	if len(pkv) != 2 {
		return panel
	}

	pkt := strings.Split(pkv[0], "-")
	if len(pkt) != 4 {
		return panel
	}

	source, err := strconv.ParseInt(pkt[0], 10, 64)
	if err == nil {
		panel.Source = source
	}

	target, err := strconv.ParseInt(pkt[1], 10, 64)
	if err == nil {
		panel.Target = target
	}

	value, err := strconv.ParseFloat(pkt[2], 64)
	if err == nil {
		panel.Value = value
	}

	decimals, err := strconv.ParseInt(pkt[3], 10, 64)
	if err == nil {
		panel.Decimals = decimals
	}

	return panel
}

func validatePanel(panel PanelType) bool {
	if panel == (PanelType{}) {
		return false
	}

	if !validateCryptoId(panel.Source) {
		return false
	}

	if !validateCryptoId(panel.Target) {
		return false
	}

	return true
}
