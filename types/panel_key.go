package types

import (
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/number"

	JC "jxwatcher/core"
)

type panelKeyType struct {
	value string
}

func (p *panelKeyType) Set(value string) {
	p.value = value
}

func (p *panelKeyType) UpdateValue(rate float64) string {
	pkk := strings.Split(p.value, "|")
	p.value = fmt.Sprintf("%s|%.20f", pkk[0], rate)
	return p.value
}

func (p *panelKeyType) RefreshKey() string {
	return p.GenerateKeyFromPanel(p.GetPanel(), p.GetValueFloat())
}

func (p *panelKeyType) GenerateKey(source, target, value, sourceSymbol string, targetSymbol string, decimals string, rate float64) string {
	p.value = fmt.Sprintf("%s-%s-%s-%s-%s-%s|%.20f", source, target, value, sourceSymbol, targetSymbol, decimals, rate)
	return p.value
}

func (p *panelKeyType) GenerateKeyFromPanel(panel panelType, rate float64) string {
	p.value = fmt.Sprintf("%d-%d-%s-%s-%s-%d|%.20f", panel.Source, panel.Target, JC.DynamicFormatFloatToString(panel.Value), panel.SourceSymbol, panel.TargetSymbol, panel.Decimals, rate)
	return p.value
}

func (p *panelKeyType) Validate() bool {
	pkv := strings.Split(p.value, "|")
	if len(pkv) != 2 {
		return false
	}

	pkt := strings.Split(pkv[0], "-")
	if len(pkt) != 6 {
		return false
	}

	return true
}

func (p *panelKeyType) GetRawValue() string {
	return p.value
}

func (p *panelKeyType) GetPanel() panelType {
	return panelType{
		Source:       p.GetSourceCoinInt(),
		Target:       p.GetTargetCoinInt(),
		Decimals:     p.GetDecimalsInt(),
		Value:        p.GetSourceValueFloat(),
		SourceSymbol: p.GetSourceSymbolString(),
		TargetSymbol: p.GetTargetSymbolString(),
	}
}

func (p *panelKeyType) GetValueFloat() float64 {

	value, err := strconv.ParseFloat(p.GetValueString(), 64)
	if err == nil {
		return value
	}

	return 0
}

func (p *panelKeyType) GetReverseValueFloat() float64 {

	value, err := strconv.ParseFloat(p.GetValueString(), 64)
	if err == nil {
		return 1 / value
	}

	return 0
}

func (p *panelKeyType) GetValueString() string {

	pkv := strings.Split(p.value, "|")
	if len(pkv) > 1 {
		return pkv[1]
	}

	return ""
}

func (p *panelKeyType) GetReverseValueString() string {

	pkv := p.GetReverseValueFloat()
	frac := JC.NumDecPlaces(pkv)
	if frac < 3 {
		frac = 2
	}

	return strconv.FormatFloat(pkv, 'f', frac, 64)
}

func (p *panelKeyType) GetValueFormattedString() string {

	pr := message.NewPrinter(language.English)
	frac := int(JC.NumDecPlaces(p.GetSourceValueFloat()))
	dec := int(p.GetDecimalsInt())
	if frac < 3 {
		frac = 2
	}

	if frac < dec {
		frac = dec
	}

	return pr.Sprintf("%v", number.Decimal(p.GetValueFloat(), number.MaxFractionDigits(frac)))

}

func (p *panelKeyType) GetReverseValueFormattedString() string {

	pr := message.NewPrinter(language.English)
	frac := int(JC.NumDecPlaces(p.GetSourceValueFloat()))
	dec := int(p.GetDecimalsInt())
	if frac < 3 {
		frac = 2
	}

	if frac < dec {
		frac = dec
	}

	return pr.Sprintf("%v", number.Decimal(p.GetReverseValueFloat(), number.MaxFractionDigits(frac)))

}

func (p *panelKeyType) GetCalculatedValueFormattedString() string {

	pr := message.NewPrinter(language.English)
	frac := int(JC.NumDecPlaces(p.GetSourceValueFloat()))
	if frac < 3 {
		frac = 2
	}

	nv := p.GetValueFloat() * p.GetSourceValueFloat()
	if nv < 1 {
		frac = max(int(p.GetDecimalsInt()), 4)
	}

	return pr.Sprintf("%v", number.Decimal(p.GetValueFloat()*p.GetSourceValueFloat(), number.MaxFractionDigits(int(frac))))

}

func (p *panelKeyType) GetSourceCoinInt() int64 {

	source, err := strconv.ParseInt(p.GetSourceCoinString(), 10, 64)
	if err == nil {
		return source
	}

	return 0
}

func (p *panelKeyType) GetSourceCoinString() string {

	pkm := strings.Split(p.value, "|")
	pkv := strings.Split(pkm[0], "-")

	if len(pkv) > 0 {
		return pkv[0]
	}

	return ""
}

func (p *panelKeyType) GetTargetCoinInt() int64 {

	target, err := strconv.ParseInt(p.GetTargetCoinString(), 10, 64)
	if err == nil {
		return target
	}

	return 0
}

func (p *panelKeyType) GetTargetCoinString() string {

	pkm := strings.Split(p.value, "|")
	pkv := strings.Split(pkm[0], "-")

	if len(pkv) > 1 {
		return pkv[1]
	}

	return ""
}

func (p *panelKeyType) GetSourceValueFloat() float64 {

	value, err := strconv.ParseFloat(p.GetSourceValueString(), 64)
	if err == nil {
		return value
	}

	return 0
}

func (p *panelKeyType) GetSourceValueString() string {
	pkm := strings.Split(p.value, "|")
	pkv := strings.Split(pkm[0], "-")

	if len(pkv) > 1 {
		return pkv[2]
	}

	return ""
}

func (p *panelKeyType) GetSourceValueFormattedString() string {
	pr := message.NewPrinter(language.English)
	frac := int(JC.NumDecPlaces(p.GetSourceValueFloat()))
	dec := int(p.GetDecimalsInt())

	if frac < 3 {
		frac = 2
	}

	if frac < dec {
		frac = dec
	}

	return pr.Sprintf("%v", number.Decimal(p.GetSourceValueFloat(), number.MaxFractionDigits(frac)))

}

func (p *panelKeyType) GetSourceSymbolString() string {

	pkm := strings.Split(p.value, "|")
	pkv := strings.Split(pkm[0], "-")

	if len(pkv) > 2 {
		return pkv[3]
	}

	return ""
}

func (p *panelKeyType) GetTargetSymbolString() string {

	pkm := strings.Split(p.value, "|")
	pkv := strings.Split(pkm[0], "-")

	if len(pkv) > 3 {
		return pkv[4]
	}

	return ""
}

func (p *panelKeyType) GetDecimalsInt() int64 {
	decimals, err := strconv.ParseInt(p.GetDecimalsString(), 10, 64)
	if err == nil {
		return decimals
	}

	return 0
}

func (p *panelKeyType) GetDecimalsString() string {

	pkm := strings.Split(p.value, "|")
	pkv := strings.Split(pkm[0], "-")

	if len(pkv) > 4 {
		return pkv[5]
	}

	return ""
}

func NewPanelKey() *panelKeyType {
	return &panelKeyType{}
}
