package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"

	"fyne.io/fyne/v2/data/binding"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/number"
)

type PanelsType []PanelType
type PanelType struct {
	Source   int64   `json:"source"`
	Target   int64   `json:"target"`
	Value    float64 `json:"value"`
	Decimals int64   `json:"decimals"`
}

var BindedData binding.UntypedList

func removeAt(index int, list binding.UntypedList) {
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
	BindedData = binding.NewUntypedList()

	for _, panel := range panels {
		str := binding.NewString()
		str.Set(generatePanelKey(panel, 0))
		BindedData.Append(str)
	}
}

func generatePanelKey(panel PanelType, rate float32) string {
	return fmt.Sprintf("%d-%d-%f-%d|%f", panel.Source, panel.Target, panel.Value, panel.Decimals, rate)
}

func savePanels() bool {

	panels := PanelsType{}
	list, _ := BindedData.Get()
	for _, x := range list {
		str, ok := x.(binding.String)
		if !ok {
			continue
		}

		val, err := str.Get()
		if err != nil {
			continue
		}

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
	if data == nil {
		return false
	}

	npk := updatePanelValue(pk, float32(data.TargetAmount))

	str := binding.NewString()
	str.Set(npk)

	BindedData.Append(str)
	Grid.Add(generatePanel(str))

	return true

}

func insertPanel(pk string, index int) bool {

	// Trying to insert to invalid index, exit
	sval, err := BindedData.GetValue(index)
	if err != nil {
		return false
	}

	// Failed to get binded string at index, exit
	str, ok := sval.(binding.String)
	if !ok {
		return false
	}

	// the new pk is invalid, throw invalid panel and exit
	if !validatePanel(pk) {
		str.Set(pk)
		Grid.Objects[index] = generateInvalidPanel(pk)
		return false
	}

	// The old pk is invalid, mutate to valid panel
	opk, _ := str.Get()
	if !validatePanel(opk) {
		Grid.Objects[index] = generatePanel(str)
	}

	// Refresh the panel value from exchange
	ex := ExchangeDataType{}
	data := ex.GetRate(pk)
	if data == nil {
		return false
	}

	npk := updatePanelValue(pk, float32(data.TargetAmount))

	// Update the panel with new value from exchange
	if npk != pk {
		str.Set(npk)
	}

	return true
}

func updatePanel(pk string) bool {

	if !validatePanelKey(pk) {
		return false
	}

	pi := getPanelIndex(pk)

	// Maybe we need better detection?
	if pi == -1 {
		return false
	}

	// Change the panel into an invalid panel
	if !validatePanel(pk) {
		Grid.Objects[pi] = generateInvalidPanel(pk)
		return true
	}

	ex := ExchangeDataType{}
	data := ex.GetRate(pk)
	if data == nil {
		return false
	}

	npk := updatePanelValue(pk, float32(data.TargetAmount))

	if npk == pk {
		return false
	}

	sval, err := BindedData.GetValue(pi)
	if err != nil {
		return false
	}

	str, ok := sval.(binding.String)
	if !ok {
		return false
	}

	str.Set(npk)

	doPanel := pi > len(Grid.Objects)

	if !doPanel {
		obj := Grid.Objects[pi]
		vobj := reflect.ValueOf(obj).Elem()
		if !vobj.FieldByName("tag").IsValid() {
			doPanel = true
		}
	}

	if !doPanel {
		obj := Grid.Objects[pi]
		xobj, ok := obj.(*DoubleClickContainer)
		if ok && xobj.getTag() != "ValidPanel" {
			doPanel = true
		}
	}

	// Build proper panel
	if doPanel {
		Grid.Objects[pi] = generatePanel(str)
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
	for i, x := range list {
		str, ok := x.(binding.String)
		if !ok {
			continue
		}

		pk, err := str.Get()
		if err != nil {
			continue
		}

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

func formatKeyAsPanelTitle(pk string) string {
	p := message.NewPrinter(language.English)

	sourceValue := getPanelSourceValue(pk)
	sourceCoin := getPanelSourceCoin(pk)
	targetCoin := getPanelTargetCoin(pk)
	sourceID := strconv.FormatInt(sourceCoin, 10)
	sourceSymbol := getTickerSymbolById(sourceID)
	targetID := strconv.FormatInt(targetCoin, 10)
	targetSymbol := getTickerSymbolById(targetID)

	frac := int(NumDecPlaces(sourceValue))
	if frac < 3 {
		frac = 2
	}

	sts := p.Sprintf("%v", number.Decimal(sourceValue, number.MaxFractionDigits(frac)))

	return fmt.Sprintf("%s %s to %s", sts, sourceSymbol, targetSymbol)
}

func formatKeyAsPanelSubtitle(pk string) string {
	p := message.NewPrinter(language.English)

	decimals := getPanelDecimals(pk)
	sourceValue := getPanelSourceValue(pk)
	targetValue := getPanelValue(pk)
	sourceCoin := getPanelSourceCoin(pk)
	targetCoin := getPanelTargetCoin(pk)
	sourceID := strconv.FormatInt(sourceCoin, 10)
	sourceSymbol := getTickerSymbolById(sourceID)
	targetID := strconv.FormatInt(targetCoin, 10)
	targetSymbol := getTickerSymbolById(targetID)

	frac := int(NumDecPlaces(sourceValue))
	if frac < 3 {
		frac = 2
	}

	evt := p.Sprintf("%v", number.Decimal(targetValue, number.MaxFractionDigits(int(decimals))))

	return fmt.Sprintf("%s %s = %s %s", "1", sourceSymbol, evt, targetSymbol)
}

func formatKeyAsPanelContent(pk string) string {
	p := message.NewPrinter(language.English)

	sourceValue := getPanelSourceValue(pk)
	targetValue := getPanelValue(pk)
	targetCoin := getPanelTargetCoin(pk)
	targetID := strconv.FormatInt(targetCoin, 10)
	targetSymbol := getTickerSymbolById(targetID)

	frac := int(NumDecPlaces(sourceValue))
	if frac < 3 {
		frac = 2
	}

	tts := p.Sprintf("%v", number.Decimal(sourceValue*float64(targetValue), number.MaxFractionDigits(frac)))

	return fmt.Sprintf("%s %s", tts, targetSymbol)
}

func isPanelValueIncrease(a, b string) int {

	if a == b {
		return 0
	}

	numA, errA := strconv.ParseFloat(a, 32)
	numB, errB := strconv.ParseFloat(b, 32)

	if errA != nil || errB != nil {
		// fmt.Printf("Error formatting")
		return 0
	}

	if numA > numB {
		// fmt.Printf("%s (%.2f) is greater than %s (%.2f)\n", a, numA, b, numB)
		return -1
	}

	if numA < numB {
		// fmt.Printf("%s (%.2f) is less than %s (%.2f)\n", a, numA, b, numB)
		return 1
	}

	return 0
}
