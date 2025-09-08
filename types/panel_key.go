package types

import (
	"fmt"
	"strconv"
	"strings"

	JC "jxwatcher/core"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/number"
)

type PanelKeyType struct {
	value string
}

func (p *PanelKeyType) UpdateValue(rate float64) string {
	pkk := strings.Split(p.value, "|")
	p.value = fmt.Sprintf("%s|%.20f", pkk[0], rate)
	return p.value
}

func (p *PanelKeyType) RefreshKey() string {
	return p.GenerateKeyFromPanel(p.GetPanel(), float32(p.GetValueFloat()))
}

func (p *PanelKeyType) GenerateKey(source, target, value, sourceSymbol string, targetSymbol string, decimals string, rate float32) string {
	p.value = fmt.Sprintf("%s-%s-%s-%s-%s-%s|%.20f", source, target, value, sourceSymbol, targetSymbol, decimals, rate)
	return p.value
}

func (p *PanelKeyType) GenerateKeyFromPanel(panel PanelType, rate float32) string {
	p.value = fmt.Sprintf("%d-%d-%s-%s-%s-%d|%.20f", panel.Source, panel.Target, JC.DynamicFormatFloatToString(panel.Value), panel.SourceSymbol, panel.TargetSymbol, panel.Decimals, rate)
	return p.value
}

func (p *PanelKeyType) Validate() bool {
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

func (p *PanelKeyType) GetRawValue() string {
	return p.value
}

func (p *PanelKeyType) GetPanel() PanelType {
	return PanelType{
		Source:       p.GetSourceCoinInt(),
		Target:       p.GetTargetCoinInt(),
		Decimals:     p.GetDecimalsInt(),
		Value:        p.GetSourceValueFloat(),
		SourceSymbol: p.GetSourceSymbolString(),
		TargetSymbol: p.GetTargetSymbolString(),
	}
}

func (p *PanelKeyType) GetValueFloat() float64 {

	value, err := strconv.ParseFloat(p.GetValueString(), 64)
	if err == nil {
		return value
	}

	return 0
}

func (p *PanelKeyType) GetReverseValueFloat() float64 {

	value, err := strconv.ParseFloat(p.GetValueString(), 64)
	if err == nil {
		return 1 / value
	}

	return 0
}

func (p *PanelKeyType) GetValueString() string {

	pkv := strings.Split(p.value, "|")
	if len(pkv) > 1 {
		return pkv[1]
	}

	return ""
}

func (p *PanelKeyType) GetReverseValueString() string {

	pkv := p.GetReverseValueFloat()
	frac := JC.NumDecPlaces(pkv)
	if frac < 3 {
		frac = 2
	}

	return strconv.FormatFloat(pkv, 'f', frac, 64)
}

func (p *PanelKeyType) GetValueFormattedString() string {

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

func (p *PanelKeyType) GetReverseValueFormattedString() string {

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

func (p *PanelKeyType) GetCalculatedValueFormattedString() string {

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

func (p *PanelKeyType) GetSourceCoinInt() int64 {

	source, err := strconv.ParseInt(p.GetSourceCoinString(), 10, 64)
	if err == nil {
		return source
	}

	return 0
}

func (p *PanelKeyType) GetSourceCoinString() string {

	pkm := strings.Split(p.value, "|")
	pkv := strings.Split(pkm[0], "-")

	if len(pkv) > 0 {
		return pkv[0]
	}

	return ""
}

func (p *PanelKeyType) GetTargetCoinInt() int64 {

	target, err := strconv.ParseInt(p.GetTargetCoinString(), 10, 64)
	if err == nil {
		return target
	}

	return 0
}

func (p *PanelKeyType) GetTargetCoinString() string {

	pkm := strings.Split(p.value, "|")
	pkv := strings.Split(pkm[0], "-")

	if len(pkv) > 1 {
		return pkv[1]
	}

	return ""
}

func (p *PanelKeyType) GetSourceValueFloat() float64 {

	value, err := strconv.ParseFloat(p.GetSourceValueString(), 64)
	if err == nil {
		return value
	}

	return 0
}

func (p *PanelKeyType) GetSourceValueString() string {
	pkm := strings.Split(p.value, "|")
	pkv := strings.Split(pkm[0], "-")

	if len(pkv) > 1 {
		return pkv[2]
	}

	return ""
}

func (p *PanelKeyType) GetSourceValueFormattedString() string {
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

func (p *PanelKeyType) GetSourceSymbolString() string {

	pkm := strings.Split(p.value, "|")
	pkv := strings.Split(pkm[0], "-")

	if len(pkv) > 2 {
		return pkv[3]
	}

	return ""
}

func (p *PanelKeyType) GetTargetSymbolString() string {

	pkm := strings.Split(p.value, "|")
	pkv := strings.Split(pkm[0], "-")

	if len(pkv) > 3 {
		return pkv[4]
	}

	return ""
}

func (p *PanelKeyType) GetDecimalsInt() int64 {
	decimals, err := strconv.ParseInt(p.GetDecimalsString(), 10, 64)
	if err == nil {
		return decimals
	}

	return 0
}

func (p *PanelKeyType) GetDecimalsString() string {

	pkm := strings.Split(p.value, "|")
	pkv := strings.Split(pkm[0], "-")

	if len(pkv) > 4 {
		return pkv[5]
	}

	return ""
}
