package types

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2/data/binding"

	JC "jxwatcher/core"
)

type PanelDataType struct {
	Data   binding.String
	OldKey string
	Parent *PanelsMapType
	Index  int
}

func (p *PanelDataType) Init() {
	p.Data = binding.NewString()
}

func (p *PanelDataType) Insert(panel PanelType, rate float32) {
	p.OldKey = p.Get()
	p.Data.Set(fmt.Sprintf("%d-%d-%f-%d|%f", panel.Source, panel.Target, panel.Value, panel.Decimals, rate))
}

func (p *PanelDataType) Set(pk string) {
	p.OldKey = p.Get()
	p.Data.Set(pk)
}

func (p *PanelDataType) Get() string {
	if p.Data == nil {
		return ""
	}
	pk, err := p.Data.Get()
	if err == nil {
		return pk
	}
	return ""
}

func (p *PanelDataType) GetData() binding.String {
	return p.Data
}

func (p *PanelDataType) GetValueString() string {
	return p.UsePanelKey().GetValueString()
}

func (p *PanelDataType) GetOldValueString() string {
	pko := PanelKeyType{value: p.OldKey}
	return pko.GetValueString()
}

func (p *PanelDataType) UsePanelKey() *PanelKeyType {
	pko := PanelKeyType{value: p.Get()}
	return &pko
}

func (p *PanelDataType) Update(pk string) bool {

	opk := p.Get()
	p.OldKey = opk
	npk := pk
	ck := ExchangeCache.CreateKeyFromInt(p.UsePanelKey().GetSourceCoinInt(), p.UsePanelKey().GetTargetCoinInt())

	if ExchangeCache.Has(ck) {
		Data := ExchangeCache.Get(ck)
		pko := PanelKeyType{value: npk}
		npk = pko.UpdateValue(float32(Data.TargetAmount))

		// Debug: Make the value always change
		// npk = pko.UpdateValue(float32(Data.TargetAmount * (rand.Float64() * 5)))
	}

	if npk != opk {
		p.Set(npk)
	}

	return true
}

func (p *PanelDataType) FormatTitle() string {
	frac := int(JC.NumDecPlaces(p.UsePanelKey().GetSourceValueFloat()))
	if frac < 3 {
		frac = 2
	}

	pk := p.UsePanelKey()

	return fmt.Sprintf(
		"%s %s to %s",
		pk.GetSourceValueFormattedString(),
		pk.GetSourceSymbolString(),
		pk.GetTargetSymbolString(),
	)
}

func (p *PanelDataType) FormatSubtitle() string {

	pk := p.UsePanelKey()

	return fmt.Sprintf(
		"%s %s = %s %s",
		"1",
		pk.GetSourceSymbolString(),
		pk.GetValueFormattedString(),
		pk.GetTargetSymbolString(),
	)
}

func (p *PanelDataType) FormatContent() string {

	pk := p.UsePanelKey()

	return fmt.Sprintf(
		"%s %s",
		pk.GetCalculatedValueFormattedString(),
		pk.GetTargetSymbolString(),
	)
}

func (p *PanelDataType) DidChange() bool {
	return p.OldKey != p.Get()
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
