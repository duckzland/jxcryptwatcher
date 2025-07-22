package main

import (
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2/data/binding"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/number"
)

type PanelsMap struct {
	data binding.UntypedList
	maps CryptosMap
}

func (pc *PanelsMap) Init() {
	pc.data = binding.NewUntypedList()
}

func (pc *PanelsMap) Set(data binding.UntypedList) {
	pc.data = data
}

func (pc *PanelsMap) SetMaps(maps CryptosMap) {
	pc.maps = maps
}

func (pc *PanelsMap) Remove(index int) bool {
	values, _ := pc.data.Get()
	if index < 0 || index >= len(values) {
		return false // avoid out-of-bounds
	}

	// Remove item at index
	updated := append(values[:index], values[index+1:]...)
	pc.data.Set(updated)
	return true
}

func (pc *PanelsMap) RemoveByKey(pk string) bool {
	return pc.Remove(pc.GetIndex(pk))
}

func (pc *PanelsMap) Append(pk string) *binding.String {

	if !pc.ValidatePanel(pk) {
		return nil
	}

	return pc.Create(pk)
}

func (pc *PanelsMap) AppendPanel(panel PanelType, rate float32) *binding.String {
	return pc.Append(pc.GetKey(panel, rate))
}

func (pc *PanelsMap) Create(pk string) *binding.String {

	ex := ExchangeDataType{}
	data := ex.GetRate(pk)
	npk := pk
	if data != nil {
		npk = pc.UpdateValue(pk, float32(data.TargetAmount))
	}

	str := binding.NewString()
	str.Set(npk)

	pc.data.Append(str)

	return &str
}

func (pc *PanelsMap) CreatePanel(panel PanelType, rate float32) *binding.String {
	return pc.Create(pc.GetKey(panel, rate))
}

func (pc *PanelsMap) Insert(pk string, index int) (*binding.String, int) {

	// Trying to insert to invalid index, exit
	sval, err := pc.data.GetValue(index)
	if err != nil {
		return nil, -1
	}

	// Failed to Get binded string at index, exit
	str, ok := sval.(binding.String)
	if !ok {
		return nil, -1
	}

	// Store old pk first
	opk, _ := str.Get()

	// the new pk is invalid, throw invalid panel and exit
	if !pc.ValidatePanel(pk) {
		str.Set(pk)
		return &str, 0
	}

	// Refresh the panel value from exchange
	ex := ExchangeDataType{}
	data := ex.GetRate(pk)
	npk := pk
	if data != nil {
		npk = pc.UpdateValue(pk, float32(data.TargetAmount))
	}

	// Always update!
	str.Set(npk)

	// The old pk is invalid, mutate to valid panel
	if !pc.ValidatePanel(opk) {
		return &str, 1
	}

	return &str, 2
}

func (pc *PanelsMap) InsertPanel(panel PanelType, rate float32, index int) (*binding.String, int) {
	return pc.Insert(pc.GetKey(panel, rate), index)
}

func (pc *PanelsMap) Update(pk string) (*binding.String, int) {

	if !pc.ValidateKey(pk) {
		return nil, -1
	}

	pi := pc.GetIndex(pk)

	// Maybe we need better detection?
	if pi == -1 {
		return nil, -1
	}

	// Change the panel into an invalid panel
	if !pc.ValidatePanel(pk) {
		return nil, 0
	}

	ex := ExchangeDataType{}
	data := ex.GetRate(pk)
	npk := pk
	if data != nil {
		npk = pc.UpdateValue(pk, float32(data.TargetAmount))
	}

	sval, err := pc.data.GetValue(pi)
	if err != nil {
		return nil, -1
	}

	str, ok := sval.(binding.String)
	if !ok {
		return nil, -1
	}

	if npk != pk {
		str.Set(npk)
	}

	return &str, 1
}

func (pc *PanelsMap) UpdateValue(pk string, rate float32) string {

	if pc.ValidateKey(pk) {
		pkk := strings.Split(pk, "|")
		return fmt.Sprintf("%s|%f", pkk[0], rate)
	}

	return pk
}

func (pc *PanelsMap) Get() ([]any, error) {
	return pc.data.Get()
}

func (pc *PanelsMap) GetKey(panel PanelType, rate float32) string {
	return fmt.Sprintf("%d-%d-%f-%d|%f", panel.Source, panel.Target, panel.Value, panel.Decimals, rate)
}

func (pc *PanelsMap) GetIndex(pk string) int {
	list, _ := pc.data.Get()
	for i, x := range list {
		str, ok := x.(binding.String)
		if !ok {
			continue
		}

		spk, err := str.Get()
		if err != nil {
			continue
		}

		if spk == pk {
			return i
		}
	}

	return -1
}

func (pc *PanelsMap) GetValue(pk string) float64 {

	if pc.ValidateKey(pk) {
		pkv := strings.Split(pk, "|")
		value, err := strconv.ParseFloat(pkv[1], 64)
		if err == nil {
			return value
		}
	}
	return 0
}

func (pc *PanelsMap) GetStringValue(pk string) string {

	if pc.ValidateKey(pk) {
		pkv := strings.Split(pk, "|")
		if len(pkv) > 0 {
			return pkv[1]
		}
	}
	return "0"
}

func (pc *PanelsMap) GetSourceCoin(pk string) int64 {
	if pc.ValidateKey(pk) {
		pkm := strings.Split(pk, "|")
		pkv := strings.Split(pkm[0], "-")

		source, err := strconv.ParseInt(pkv[0], 10, 64)
		if err == nil {
			return source
		}
	}
	return 0
}

func (pc *PanelsMap) GetTargetCoin(pk string) int64 {
	if pc.ValidateKey(pk) {
		pkm := strings.Split(pk, "|")
		pkv := strings.Split(pkm[0], "-")

		target, err := strconv.ParseInt(pkv[1], 10, 64)
		if err == nil {
			return target
		}
	}

	return 0
}

func (pc *PanelsMap) GetSourceValue(pk string) float64 {
	if pc.ValidateKey(pk) {
		pkm := strings.Split(pk, "|")
		pkv := strings.Split(pkm[0], "-")

		value, err := strconv.ParseFloat(pkv[2], 64)
		if err == nil {
			return value
		}
	}

	return 0
}

func (pc *PanelsMap) GetDecimals(pk string) int64 {
	if pc.ValidateKey(pk) {
		pkm := strings.Split(pk, "|")
		pkv := strings.Split(pkm[0], "-")

		decimals, err := strconv.ParseInt(pkv[3], 10, 64)
		if err == nil {
			return decimals
		}
	}

	return 0
}

func (pc *PanelsMap) ValidateKey(pk string) bool {
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

func (pc *PanelsMap) ValidatePanel(pk string) bool {
	if !pc.ValidateKey(pk) {
		return false
	}

	sid := pc.GetSourceCoin(pk)
	tid := pc.GetTargetCoin(pk)

	// @todo when cryptos got method use that!
	if !pc.maps.ValidateId(sid) {
		return false
	}

	// @todo when cryptos got method use that!
	if !pc.maps.ValidateId(tid) {
		return false
	}

	return true
}

func (pc *PanelsMap) ValidateId(id int64) bool {
	return pc.maps.ValidateId(id)
}

func (pc *PanelsMap) GetOptions() []string {
	return pc.maps.GetOptions()
}

func (pc *PanelsMap) GetDisplayById(id string) string {
	return pc.maps.GetDisplayById(id)
}

func (pc *PanelsMap) GetIdByDisplay(tk string) string {
	return pc.maps.GetIdByDisplay(tk)
}

func (pc *PanelsMap) GetSymbolById(id string) string {
	return pc.maps.GetSymbolById(id)
}

func (pc *PanelsMap) FormatPanelTitle(pk string) string {
	p := message.NewPrinter(language.English)

	sourceValue := pc.GetSourceValue(pk)
	sourceCoin := pc.GetSourceCoin(pk)
	targetCoin := pc.GetTargetCoin(pk)
	sourceID := strconv.FormatInt(sourceCoin, 10)
	sourceSymbol := pc.maps.GetSymbolById(sourceID)
	targetID := strconv.FormatInt(targetCoin, 10)
	targetSymbol := pc.maps.GetSymbolById(targetID)

	frac := int(NumDecPlaces(sourceValue))
	if frac < 3 {
		frac = 2
	}

	sts := p.Sprintf("%v", number.Decimal(sourceValue, number.MaxFractionDigits(frac)))

	return fmt.Sprintf("%s %s to %s", sts, sourceSymbol, targetSymbol)
}

func (pc *PanelsMap) FormatPanelSubtitle(pk string) string {
	p := message.NewPrinter(language.English)

	decimals := pc.GetDecimals(pk)
	sourceValue := pc.GetSourceValue(pk)
	targetValue := pc.GetValue(pk)
	sourceCoin := pc.GetSourceCoin(pk)
	targetCoin := pc.GetTargetCoin(pk)
	sourceID := strconv.FormatInt(sourceCoin, 10)
	sourceSymbol := pc.maps.GetSymbolById(sourceID)
	targetID := strconv.FormatInt(targetCoin, 10)
	targetSymbol := pc.maps.GetSymbolById(targetID)

	frac := int(NumDecPlaces(sourceValue))
	if frac < 3 {
		frac = 2
	}

	evt := p.Sprintf("%v", number.Decimal(targetValue, number.MaxFractionDigits(int(decimals))))

	return fmt.Sprintf("%s %s = %s %s", "1", sourceSymbol, evt, targetSymbol)
}

func (pc *PanelsMap) FormatPanelContent(pk string) string {
	p := message.NewPrinter(language.English)

	sourceValue := pc.GetSourceValue(pk)
	targetValue := pc.GetValue(pk)
	targetCoin := pc.GetTargetCoin(pk)
	targetID := strconv.FormatInt(targetCoin, 10)
	targetSymbol := pc.maps.GetSymbolById(targetID)

	frac := int(NumDecPlaces(sourceValue))
	if frac < 3 {
		frac = 2
	}

	tts := p.Sprintf("%v", number.Decimal(sourceValue*float64(targetValue), number.MaxFractionDigits(frac)))

	return fmt.Sprintf("%s %s", tts, targetSymbol)
}

func (pc *PanelsMap) IsValueIncrease(a, b string) int {

	if a == b {
		return 0
	}

	numA, errA := strconv.ParseFloat(a, 32)
	numB, errB := strconv.ParseFloat(b, 32)

	if errA != nil || errB != nil {
		// fmt.Printf("Error Formatting")
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
