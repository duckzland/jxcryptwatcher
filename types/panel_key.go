package types

import (
	"fmt"
	"math/big"
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

func (p *panelKeyType) UpdateValue(rate *big.Float) string {
	pkk := strings.Split(p.value, "|")

	p.value = fmt.Sprintf("%s|%s", pkk[0], rate.Text('g', -1))
	return p.value
}

func (p *panelKeyType) IsValueMatching(rate *big.Float, op string) bool {
	cmp := p.GetValueFloat().Cmp(rate)
	switch op {
	case "==", "=":
		return cmp == 0
	case "!=":
		return cmp != 0
	case "<":
		return cmp == -1
	case "<=":
		return cmp == -1 || cmp == 0
	case ">":
		return cmp == 1
	case ">=":
		return cmp == 1 || cmp == 0
	default:
		return false
	}

}

func (p *panelKeyType) IsValueMatchingFloat(val float64, op string) bool {
	return p.IsValueMatching(JC.ToBigFloat(val), op)
}

func (p *panelKeyType) IsConfigMatching(key string) bool {
	s := strings.SplitN(p.value, "|", 2)
	v := strings.SplitN(key, "|", 2)

	return len(s) > 0 && len(v) > 0 && s[0] == v[0]
}

func (p *panelKeyType) RefreshKey() string {
	return p.GenerateKeyFromPanel(p.GetPanel(), p.GetValueFloat())
}

func (p *panelKeyType) GenerateKey(source, target, value, sourceSymbol string, targetSymbol string, decimals string, rate *big.Float) string {
	p.value = fmt.Sprintf("%s-%s-%s-%s-%s-%s|%s", source, target, value, sourceSymbol, targetSymbol, decimals, rate.Text('g', -1))
	return p.value
}

func (p *panelKeyType) GenerateKeyFromPanel(panel panelType, rate *big.Float) string {
	p.value = fmt.Sprintf("%d-%d-%s-%s-%s-%d|%s",
		panel.Source,
		panel.Target,
		JC.DynamicFormatFloatToString(panel.Value),
		panel.SourceSymbol,
		panel.TargetSymbol,
		panel.Decimals,
		rate.Text('g', -1),
	)
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

func (p *panelKeyType) GetValueFloat() *big.Float {
	raw := p.GetValueString()
	f, ok := JC.ToBigString(raw)
	if ok {
		return f
	}
	return JC.ToBigFloat(0)
}

func (p *panelKeyType) GetReverseValueFloat() *big.Float {
	raw := p.GetValueString()
	val, ok := JC.ToBigString(raw)
	if !ok || val.Sign() == 0 {
		return JC.ToBigFloat(0)
	}

	return new(big.Float).Quo(JC.ToBigFloat(1), val)
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
	frac := JC.BigFloatNumDecPlaces(pkv)
	if frac < 3 {
		frac = 2
	}
	return pkv.Text('f', frac)
}

func (p *panelKeyType) GetValueFormattedString() string {
	value := p.GetValueFloat()
	source := p.GetSourceValueFloat()
	frac := JC.NumDecPlaces(source)
	dec := int(p.GetDecimalsInt())

	if frac < 3 {
		frac = 2
	}
	if frac < dec {
		frac = dec
	}

	f64, _ := value.Float64()
	pr := message.NewPrinter(language.English)

	return pr.Sprintf("%v", number.Decimal(f64, number.MaxFractionDigits(frac)))
}

func (p *panelKeyType) GetReverseValueFormattedString() string {
	value := p.GetReverseValueFloat()
	source := p.GetSourceValueFloat()
	frac := JC.NumDecPlaces(source)
	dec := int(p.GetDecimalsInt())

	if frac < 3 {
		frac = 2
	}
	if frac < dec {
		frac = dec
	}

	f64, _ := value.Float64()
	pr := message.NewPrinter(language.English)
	return pr.Sprintf("%v", number.Decimal(f64, number.MaxFractionDigits(frac)))
}

func (p *panelKeyType) GetCalculatedValueFormattedString() string {
	pr := message.NewPrinter(language.English)
	source := p.GetSourceValueFloat()
	frac := JC.NumDecPlaces(source)

	if frac < 3 {
		frac = 2
	}

	nv := new(big.Float).SetPrec(256).Mul(p.GetValueFloat(), JC.ToBigFloat(source))
	if nv.Cmp(JC.ToBigFloat(1)) < 0 {
		frac = max(int(p.GetDecimalsInt()), 4)
	}

	f64, _ := nv.Float64()
	return pr.Sprintf("%v", number.Decimal(f64, number.MaxFractionDigits(frac)))
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
