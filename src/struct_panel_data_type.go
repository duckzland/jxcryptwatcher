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
	index  int
}

func (p *PanelDataType) Init() {
	p.data = binding.NewString()
}

func (p *PanelDataType) Insert(panel PanelType, rate float32) {
	p.oldKey = p.Get()
	p.data.Set(fmt.Sprintf("%d-%d-%f-%d|%f", panel.Source, panel.Target, panel.Value, panel.Decimals, rate))
}

func (p *PanelDataType) Set(pk string) {
	p.oldKey = p.Get()
	p.data.Set(pk)
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

func (p *PanelDataType) GetSourceCoinInt() int64 {

	pk := p.GetSourceCoinString()
	source, err := strconv.ParseInt(pk, 10, 64)
	if err == nil {
		return source
	}

	return 0
}

func (p *PanelDataType) GetSourceCoinString() string {

	pk := p.Get()
	pkm := strings.Split(pk, "|")
	pkv := strings.Split(pkm[0], "-")

	if len(pkv) > 0 {
		return pkv[0]
	}

	return ""
}

func (p *PanelDataType) GetTargetCoinInt() int64 {
	pk := p.GetTargetCoinString()
	target, err := strconv.ParseInt(pk, 10, 64)
	if err == nil {
		return target
	}

	return 0
}

func (p *PanelDataType) GetTargetCoinString() string {
	pk := p.Get()
	pkm := strings.Split(pk, "|")
	pkv := strings.Split(pkm[0], "-")

	if len(pkv) > 1 {
		return pkv[1]
	}

	return ""
}

func (p *PanelDataType) GetSourceValueFloat() float64 {
	pk := p.GetSourceValueString()
	value, err := strconv.ParseFloat(pk, 64)
	if err == nil {
		return value
	}

	return 0
}

func (p *PanelDataType) GetSourceValueString() string {
	pk := p.Get()
	pkm := strings.Split(pk, "|")
	pkv := strings.Split(pkm[0], "-")

	if len(pkv) > 1 {
		return pkv[2]
	}

	return ""
}

func (p *PanelDataType) GetDecimalsInt() int64 {
	pk := p.GetDecimalsString()
	decimals, err := strconv.ParseInt(pk, 10, 64)
	if err == nil {
		return decimals
	}

	return 0
}

func (p *PanelDataType) GetDecimalsString() string {
	pk := p.Get()
	pkm := strings.Split(pk, "|")
	pkv := strings.Split(pkm[0], "-")

	if len(pkv) > 2 {
		return pkv[3]
	}

	return ""
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

	frac := int(NumDecPlaces(p.GetSourceValueFloat()))
	if frac < 3 {
		frac = 2
	}

	return fmt.Sprintf(
		"%s %s to %s",
		pr.Sprintf("%v", number.Decimal(p.GetSourceValueFloat(), number.MaxFractionDigits(frac))),
		p.parent.GetSymbolById(p.GetSourceCoinString()),
		p.parent.GetSymbolById(p.GetTargetCoinString()),
	)
}

func (p *PanelDataType) FormatSubtitle() string {
	pr := message.NewPrinter(language.English)
	frac := int(NumDecPlaces(p.GetSourceValueFloat()))
	if frac < 3 {
		frac = 2
	}

	return fmt.Sprintf(
		"%s %s = %s %s", "1",
		p.parent.GetSymbolById(p.GetSourceCoinString()),
		pr.Sprintf("%v", number.Decimal(p.GetValueFloat(), number.MaxFractionDigits(int(p.GetDecimalsInt())))),
		p.parent.GetSymbolById(p.GetTargetCoinString()),
	)
}

func (p *PanelDataType) FormatContent() string {
	pr := message.NewPrinter(language.English)
	frac := int(NumDecPlaces(p.GetSourceValueFloat()))
	if frac < 3 {
		frac = 2
	}

	return fmt.Sprintf(
		"%s %s",
		pr.Sprintf("%v", number.Decimal(
			p.GetSourceValueFloat()*float64(p.GetValueFloat()),
			number.MaxFractionDigits(frac),
		)),
		p.parent.GetSymbolById(p.GetTargetCoinString()),
	)
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
