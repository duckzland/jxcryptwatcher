package main

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2/data/binding"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/number"
)

type PanelDataType struct {
	data   binding.String
	oldKey string
	parent *PanelsMapType
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
	if p.data == nil {
		return ""
	}
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
	return p.UsePanelKey().GetValueString()
}

func (p *PanelDataType) GetOldValueString() string {
	pko := PanelKeyType{value: p.oldKey}
	return pko.GetValueString()
}

func (p *PanelDataType) UsePanelKey() *PanelKeyType {
	pko := PanelKeyType{value: p.Get()}
	return &pko
}

func (p *PanelDataType) Update(pk string) bool {

	opk := p.Get()
	p.oldKey = opk
	npk := pk
	ck := ExchangeCache.CreateKeyFromInt(p.UsePanelKey().GetSourceCoinInt(), p.UsePanelKey().GetTargetCoinInt())

	if ExchangeCache.Has(ck) {
		data := ExchangeCache.Get(ck)
		pko := PanelKeyType{value: npk}
		npk = pko.UpdateValue(float32(data.TargetAmount))

		// Debug: Make the value always change
		// npk = pko.UpdateValue(float32(data.TargetAmount * (rand.Float64() * 5)))
	}

	if npk != opk {
		p.Set(npk)
	}

	return true
}

func (p *PanelDataType) FormatTitle() string {
	pr := message.NewPrinter(language.English)

	frac := int(NumDecPlaces(p.UsePanelKey().GetSourceValueFloat()))
	if frac < 3 {
		frac = 2
	}

	return fmt.Sprintf(
		"%s %s to %s",
		pr.Sprintf("%v", number.Decimal(p.UsePanelKey().GetSourceValueFloat(), number.MaxFractionDigits(frac))),
		p.parent.GetSymbolById(p.UsePanelKey().GetSourceCoinString()),
		p.parent.GetSymbolById(p.UsePanelKey().GetTargetCoinString()),
	)
}

func (p *PanelDataType) FormatSubtitle() string {
	pr := message.NewPrinter(language.English)
	frac := int(NumDecPlaces(p.UsePanelKey().GetSourceValueFloat()))
	if frac < 3 {
		frac = 2
	}

	return fmt.Sprintf(
		"%s %s = %s %s", "1",
		p.parent.GetSymbolById(p.UsePanelKey().GetSourceCoinString()),
		pr.Sprintf("%v", number.Decimal(p.UsePanelKey().GetValueFloat(), number.MaxFractionDigits(int(p.UsePanelKey().GetDecimalsInt())))),
		p.parent.GetSymbolById(p.UsePanelKey().GetTargetCoinString()),
	)
}

func (p *PanelDataType) FormatContent() string {
	pr := message.NewPrinter(language.English)
	frac := int(NumDecPlaces(p.UsePanelKey().GetSourceValueFloat()))
	if frac < 3 {
		frac = 2
	}

	return fmt.Sprintf(
		"%s %s",
		pr.Sprintf("%v", number.Decimal(
			p.UsePanelKey().GetSourceValueFloat()*float64(p.UsePanelKey().GetValueFloat()),
			number.MaxFractionDigits(frac),
		)),
		p.parent.GetSymbolById(p.UsePanelKey().GetTargetCoinString()),
	)
}

func (p *PanelDataType) DidChange() bool {
	return p.oldKey != p.Get()
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
