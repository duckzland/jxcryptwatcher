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

type PanelDataType struct {
	data   binding.String
	oldKey string
	parent *PanelsMap
}

func (p *PanelDataType) Init() {
	p.data = binding.NewString()
}

func (p *PanelDataType) Insert(panel PanelType, rate float32) {
	p.oldKey = p.Get()
	p.data.Set(fmt.Sprintf("%d-%d-%f-%d|%f", panel.Source, panel.Target, panel.Value, panel.Decimals, rate))
}

func (p *PanelDataType) Set(pk string) {
	// if p.parent.ValidateKey(pk) {
	p.oldKey = p.Get()
	p.data.Set(pk)
	//}
}

func (p *PanelDataType) Get() string {
	pk, err := p.data.Get()
	if err == nil {
		return pk
	}
	return ""
}

func (p *PanelDataType) GetData() binding.String {
	return p.data
}

func (p *PanelDataType) GetValueString() string {

	pk := p.Get()
	pkv := strings.Split(pk, "|")
	if len(pkv) > 0 {
		return pkv[1]
	}

	return "0"
}

func (p *PanelDataType) GetOldValueString() string {

	pk := p.oldKey
	pkv := strings.Split(pk, "|")
	if len(pkv) > 0 {
		return pkv[1]
	}

	return "0"
}

func (p *PanelDataType) GetValueFloat() float64 {

	pk := p.Get()
	pkv := strings.Split(pk, "|")
	value, err := strconv.ParseFloat(pkv[1], 64)
	if err == nil {
		return value
	}

	return 0
}

func (p *PanelDataType) GetSourceCoin() int64 {

	pk := p.Get()
	pkm := strings.Split(pk, "|")
	pkv := strings.Split(pkm[0], "-")

	source, err := strconv.ParseInt(pkv[0], 10, 64)
	if err == nil {
		return source
	}

	return 0
}

func (p *PanelDataType) GetTargetCoin() int64 {
	pk := p.Get()
	pkm := strings.Split(pk, "|")
	pkv := strings.Split(pkm[0], "-")

	target, err := strconv.ParseInt(pkv[1], 10, 64)
	if err == nil {
		return target
	}

	return 0
}

func (p *PanelDataType) GetSourceValue() float64 {
	pk := p.Get()
	pkm := strings.Split(pk, "|")
	pkv := strings.Split(pkm[0], "-")

	value, err := strconv.ParseFloat(pkv[2], 64)
	if err == nil {
		return value
	}

	return 0
}

func (p *PanelDataType) GetDecimals() int64 {
	pk := p.Get()
	pkm := strings.Split(pk, "|")
	pkv := strings.Split(pkm[0], "-")

	decimals, err := strconv.ParseInt(pkv[3], 10, 64)
	if err == nil {
		return decimals
	}

	return 0
}

func (p *PanelDataType) Update(pk string) bool {

	p.Set(pk)

	ex := ExchangeDataType{}
	data := ex.GetRate(pk)
	if data != nil {
		p.UpdateValue(float32(data.TargetAmount))
	}

	return true
}

func (p *PanelDataType) UpdateValue(rate float32) string {
	pk := p.Get()
	pkk := strings.Split(pk, "|")
	npk := fmt.Sprintf("%s|%f", pkk[0], rate)
	p.Set(npk)
	return npk
}

func (p *PanelDataType) Validate() bool {
	pk := p.Get()
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

func (p *PanelDataType) FormatTitle() string {
	pr := message.NewPrinter(language.English)

	sourceValue := p.GetSourceValue()
	sourceCoin := p.GetSourceCoin()
	targetCoin := p.GetTargetCoin()
	sourceID := strconv.FormatInt(sourceCoin, 10)
	sourceSymbol := p.parent.GetSymbolById(sourceID)
	targetID := strconv.FormatInt(targetCoin, 10)
	targetSymbol := p.parent.GetSymbolById(targetID)

	frac := int(NumDecPlaces(sourceValue))
	if frac < 3 {
		frac = 2
	}

	sts := pr.Sprintf("%v", number.Decimal(sourceValue, number.MaxFractionDigits(frac)))

	return fmt.Sprintf("%s %s to %s", sts, sourceSymbol, targetSymbol)
}

func (p *PanelDataType) FormatSubtitle() string {
	pr := message.NewPrinter(language.English)

	decimals := p.GetDecimals()
	sourceValue := p.GetSourceValue()
	targetValue := p.GetValueFloat()
	sourceCoin := p.GetSourceCoin()
	targetCoin := p.GetTargetCoin()
	sourceID := strconv.FormatInt(sourceCoin, 10)
	sourceSymbol := p.parent.GetSymbolById(sourceID)
	targetID := strconv.FormatInt(targetCoin, 10)
	targetSymbol := p.parent.GetSymbolById(targetID)

	frac := int(NumDecPlaces(sourceValue))
	if frac < 3 {
		frac = 2
	}

	evt := pr.Sprintf("%v", number.Decimal(targetValue, number.MaxFractionDigits(int(decimals))))

	return fmt.Sprintf("%s %s = %s %s", "1", sourceSymbol, evt, targetSymbol)
}

func (p *PanelDataType) FormatContent() string {
	pr := message.NewPrinter(language.English)

	sourceValue := p.GetSourceValue()
	targetValue := p.GetValueFloat()
	targetCoin := p.GetTargetCoin()
	targetID := strconv.FormatInt(targetCoin, 10)
	targetSymbol := p.parent.GetSymbolById(targetID)

	frac := int(NumDecPlaces(sourceValue))
	if frac < 3 {
		frac = 2
	}

	tts := pr.Sprintf("%v", number.Decimal(sourceValue*float64(targetValue), number.MaxFractionDigits(frac)))

	return fmt.Sprintf("%s %s", tts, targetSymbol)
}

func (p *PanelDataType) IsValueIncrease() int {

	b := p.GetValueString()
	a := p.GetOldValueString()

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

func (p *PanelDataType) IsEqualContentString(pk string) bool {
	return p.Get() == pk

}
