package types

import (
	"math/big"
	"strconv"
	"strings"

	JC "jxwatcher/core"
)

type panelKeyType struct {
	value string
}

func (p *panelKeyType) Set(value string) {
	p.value = value
}

func (p *panelKeyType) UpdateValue(rate *big.Float) string {
	pkk := strings.Split(p.value, JC.STRING_PIPE)

	p.value = pkk[0] + JC.STRING_PIPE + rate.Text('g', -1)
	return p.value
}

func (p *panelKeyType) IsValueMatching(rate *big.Float, op string) bool {
	cmp := p.GetValueFloat().Cmp(rate)
	switch op {
	case JC.STRING_DOUBLE_EQUAL, JC.STRING_EQUAL:
		return cmp == 0
	case JC.STRING_NOT_EQUAL:
		return cmp != 0
	case JC.STRING_LESS:
		return cmp == -1
	case JC.STRING_LESS_EQUAL:
		return cmp == -1 || cmp == 0
	case JC.STRING_GREATER:
		return cmp == 1
	case JC.STRING_GREATER_EQUAL:
		return cmp == 1 || cmp == 0
	default:
		return false
	}

}

func (p *panelKeyType) IsValueMatchingFloat(val float64, op string) bool {
	return p.IsValueMatching(JC.ToBigFloat(val), op)
}

func (p *panelKeyType) IsConfigMatching(key string) bool {
	s := strings.SplitN(p.value, JC.STRING_PIPE, 2)
	v := strings.SplitN(key, JC.STRING_PIPE, 2)

	return len(s) > 0 && len(v) > 0 && s[0] == v[0]
}

func (p *panelKeyType) RefreshKey() string {
	return p.GenerateKeyFromPanel(p.GetPanel(), p.GetValueFloat())
}

func (p *panelKeyType) GenerateKey(source, target, value, sourceSymbol string, targetSymbol string, decimals string, rate *big.Float) string {
	var b strings.Builder
	b.WriteString(source)
	b.WriteString(JC.STRING_MINUS)
	b.WriteString(target)
	b.WriteString(JC.STRING_MINUS)
	b.WriteString(value)
	b.WriteString(JC.STRING_MINUS)
	b.WriteString(sourceSymbol)
	b.WriteString(JC.STRING_MINUS)
	b.WriteString(targetSymbol)
	b.WriteString(JC.STRING_MINUS)
	b.WriteString(decimals)
	b.WriteString(JC.STRING_PIPE)
	b.WriteString(rate.Text('g', -1))
	p.value = b.String()
	return p.value
}

func (p *panelKeyType) GenerateKeyFromPanel(panel panelType, rate *big.Float) string {
	var b strings.Builder
	b.WriteString(strconv.FormatInt(panel.Source, 10))
	b.WriteString(JC.STRING_MINUS)
	b.WriteString(strconv.FormatInt(panel.Target, 10))
	b.WriteString(JC.STRING_MINUS)
	b.WriteString(JC.DynamicFormatFloatToString(panel.Value))
	b.WriteString(JC.STRING_MINUS)
	b.WriteString(panel.SourceSymbol)
	b.WriteString(JC.STRING_MINUS)
	b.WriteString(panel.TargetSymbol)
	b.WriteString(JC.STRING_MINUS)
	b.WriteString(strconv.FormatInt(panel.Decimals, 10))
	b.WriteString(JC.STRING_PIPE)
	b.WriteString(rate.Text('g', -1))
	p.value = b.String()
	return p.value
}

func (p *panelKeyType) Validate() bool {
	pkv := strings.Split(p.value, JC.STRING_PIPE)
	if len(pkv) != 2 {
		return false
	}

	pkt := strings.Split(pkv[0], JC.STRING_MINUS)
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

	pkv := strings.Split(p.value, JC.STRING_PIPE)
	if len(pkv) > 1 {
		return pkv[1]
	}

	return JC.STRING_EMPTY
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
	return JC.FormatNumberWithCommas(f64, frac)
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
	return JC.FormatNumberWithCommas(f64, frac)
}

func (p *panelKeyType) GetCalculatedValueFormattedString() string {
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
	return JC.FormatNumberWithCommas(f64, frac)
}

func (p *panelKeyType) GetSourceCoinInt() int64 {

	source, err := strconv.ParseInt(p.GetSourceCoinString(), 10, 64)
	if err == nil {
		return source
	}

	return 0
}

func (p *panelKeyType) GetSourceCoinString() string {

	pkm := strings.Split(p.value, JC.STRING_PIPE)
	pkv := strings.Split(pkm[0], JC.STRING_MINUS)

	if len(pkv) > 0 {
		return pkv[0]
	}

	return JC.STRING_EMPTY
}

func (p *panelKeyType) GetTargetCoinInt() int64 {

	target, err := strconv.ParseInt(p.GetTargetCoinString(), 10, 64)
	if err == nil {
		return target
	}

	return 0
}

func (p *panelKeyType) GetTargetCoinString() string {

	pkm := strings.Split(p.value, JC.STRING_PIPE)
	pkv := strings.Split(pkm[0], JC.STRING_MINUS)

	if len(pkv) > 1 {
		return pkv[1]
	}

	return JC.STRING_EMPTY
}

func (p *panelKeyType) GetSourceValueFloat() float64 {

	value, err := strconv.ParseFloat(p.GetSourceValueString(), 64)
	if err == nil {
		return value
	}

	return 0
}

func (p *panelKeyType) GetSourceValueString() string {
	pkm := strings.Split(p.value, JC.STRING_PIPE)
	pkv := strings.Split(pkm[0], JC.STRING_MINUS)

	if len(pkv) > 1 {
		return pkv[2]
	}

	return JC.STRING_EMPTY
}

func (p *panelKeyType) GetSourceValueFormattedString() string {
	frac := int(JC.NumDecPlaces(p.GetSourceValueFloat()))
	dec := int(p.GetDecimalsInt())

	if frac < 3 {
		frac = 2
	}
	if frac < dec {
		frac = dec
	}

	f64 := p.GetSourceValueFloat()
	return JC.FormatNumberWithCommas(f64, frac)
}

func (p *panelKeyType) GetSourceSymbolString() string {

	pkm := strings.Split(p.value, JC.STRING_PIPE)
	pkv := strings.Split(pkm[0], JC.STRING_MINUS)

	if len(pkv) > 2 {
		return pkv[3]
	}

	return JC.STRING_EMPTY
}

func (p *panelKeyType) GetTargetSymbolString() string {

	pkm := strings.Split(p.value, JC.STRING_PIPE)
	pkv := strings.Split(pkm[0], JC.STRING_MINUS)

	if len(pkv) > 3 {
		return pkv[4]
	}

	return JC.STRING_EMPTY
}

func (p *panelKeyType) GetDecimalsInt() int64 {
	decimals, err := strconv.ParseInt(p.GetDecimalsString(), 10, 64)
	if err == nil {
		return decimals
	}

	return 0
}

func (p *panelKeyType) GetDecimalsString() string {

	pkm := strings.Split(p.value, JC.STRING_PIPE)
	pkv := strings.Split(pkm[0], JC.STRING_MINUS)

	if len(pkv) > 4 {
		return pkv[5]
	}

	return JC.STRING_EMPTY
}

func NewPanelKey() *panelKeyType {
	return &panelKeyType{}
}
