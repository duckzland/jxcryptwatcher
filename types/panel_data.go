package types

import (
	"math/big"
	"strconv"
	"strings"

	JC "jxwatcher/core"
)

const fmtSpace = " "
const fmtTo = " to "
const fmtVal = "1 "
const fmtEqual = " = "

type PanelData interface {
	Init()
	Set(val string)
	SetStatus(val int)
	SetID(val string)
	SetOldKey(val string)
	SetParent(val *panelsMapType)
	SetRate(val *big.Float) bool
	Get() string
	GetStatus() int
	GetID() string
	GetOldKey() string
	GetParent() *panelsMapType
	GetValueString() string
	GetOldValueString() string
	UseData() *JC.DataBinding
	UsePanelKey() *panelKeyType
	IsStatus(val int) bool
	IsID(val string) bool
	IsKey(val string) bool
	IsOldKey(val string) bool
	IsEqualContentString(pk string) bool
	IsOnInitialValue() bool
	IsValueIncrease() int
	HasParent() bool
	RefreshData()
	RefreshKey(key string) string
	Insert(panel panelType, rate float64)
	Update(pk string) bool
	UpdateRate() bool
	UpdateStatus() bool
	Destroy()
	FormatTitle() string
	FormatSubtitle() string
	FormatBottomText() string
	FormatContent() string
	DidChange() bool
	Serialize() panelDataCache
}

type panelDataCache struct {
	Status int
	Key    string
	OldKey string
}

type panelDataType struct {
	data   *JC.DataBinding
	oldKey string
	id     string
	parent *panelsMapType
}

func (p *panelDataType) Init() {
	p.data = JC.NewDataBinding(JC.STRING_EMPTY, JC.STATE_ERROR)
	p.id = JC.STRING_EMPTY
	p.oldKey = JC.STRING_EMPTY
}

func (p *panelDataType) Set(val string) {
	if p.data != nil {
		p.oldKey = p.data.GetData()
		p.data.SetData(val)
	}
}

func (p *panelDataType) SetStatus(val int) {
	if p.IsStatus(val) || p.data == nil {
		return
	}
	p.data.SetStatus(val)
}

func (p *panelDataType) SetID(val string) {
	p.id = val
}

func (p *panelDataType) SetOldKey(val string) {
	p.oldKey = val
}

func (p *panelDataType) SetParent(val *panelsMapType) {
	p.parent = val
}

func (p *panelDataType) SetRate(val *big.Float) bool {
	if val == nil {
		return false
	}

	pk := p.UsePanelKey()
	if pk.IsValueMatching(val, JC.STRING_NOT_EQUAL) {
		p.Set(pk.UpdateValue(val))
		return true
	}

	return false
}

func (p *panelDataType) Get() string {
	if p.data == nil {
		return JC.STRING_EMPTY
	}
	return p.data.GetData()
}

func (p *panelDataType) GetStatus() int {
	if p.data == nil {
		return JC.STATE_ERROR
	}
	return p.data.GetStatus()
}

func (p *panelDataType) GetID() string {
	return p.id
}

func (p *panelDataType) GetOldKey() string {
	return p.oldKey
}

func (p *panelDataType) GetParent() *panelsMapType {
	return p.parent
}

func (p *panelDataType) GetValueString() string {
	return p.UsePanelKey().GetValueString()
}

func (p *panelDataType) GetOldValueString() string {
	pko := panelKeyType{value: p.oldKey}
	return pko.GetValueString()
}

func (p *panelDataType) UseData() *JC.DataBinding {
	return p.data
}

func (p *panelDataType) UsePanelKey() *panelKeyType {
	return &panelKeyType{value: p.Get()}
}

func (p *panelDataType) IsStatus(val int) bool {
	return p.GetStatus() == val
}

func (p *panelDataType) IsID(val string) bool {
	return p.id == val
}

func (p *panelDataType) IsKey(val string) bool {
	return p.Get() == val
}

func (p *panelDataType) IsOldKey(val string) bool {
	return p.oldKey == val
}

func (p *panelDataType) IsEqualContentString(pk string) bool {
	return p.IsKey(pk)
}

func (p *panelDataType) IsOnInitialValue() bool {
	opt := &panelKeyType{value: p.oldKey}
	return p.oldKey != JC.STRING_EMPTY &&
		opt.IsValueMatchingFloat(-1, JC.STRING_DOUBLE_EQUAL) &&
		p.IsStatus(JC.STATE_LOADED)
}

func (p *panelDataType) IsValueIncrease() int {
	b := p.GetValueString()
	a := p.GetOldValueString()

	if a == b {
		return JC.VALUE_NO_CHANGE
	}

	numA, errA := strconv.ParseFloat(a, 64)
	numB, errB := strconv.ParseFloat(b, 64)

	if errA != nil || errB != nil {
		return JC.VALUE_NO_CHANGE
	}

	if numA > numB {
		return JC.VALUE_DECREASE
	}

	if numA < numB {
		return JC.VALUE_INCREASE
	}

	return JC.VALUE_NO_CHANGE
}

func (p *panelDataType) HasParent() bool {
	return p.parent != nil
}

func (p *panelDataType) RefreshData() {
	if p.data == nil {
		return
	}
	npk := p.UsePanelKey().UpdateValue(JC.ToBigFloat(-3))
	p.data.SetData(npk)
}

func (p *panelDataType) RefreshKey(key string) string {
	if p.parent == nil || !p.parent.ValidateKey(key) {
		return key
	}

	pkt := &panelKeyType{value: key}
	source := pkt.GetSourceCoinString()
	target := pkt.GetTargetCoinString()
	value := pkt.GetSourceValueString()
	sourceSymbol := pkt.GetSourceSymbolString()
	targetSymbol := pkt.GetTargetSymbolString()
	decimals := pkt.GetDecimalsString()
	rate := pkt.GetValueFloat()

	if sourceSymbol == JC.STRING_EMPTY {
		sourceSymbol = p.parent.GetSymbolById(source)
	}

	if targetSymbol == JC.STRING_EMPTY {
		targetSymbol = p.parent.GetSymbolById(target)
	}

	return pkt.GenerateKey(source, target, value, sourceSymbol, targetSymbol, decimals, rate)
}

func (p *panelDataType) Update(pk string) bool {
	if JC.IsShuttingDown() {
		return false
	}

	opk := p.Get()
	npk := pk
	pks := p.UsePanelKey()

	ck := UseExchangeCache().CreateKeyFromInt(pks.GetSourceCoinInt(), pks.GetTargetCoinInt())

	if UseExchangeCache().Has(ck) {
		data := UseExchangeCache().Get(ck)
		if data != nil && data.TargetAmount != nil {
			pko := panelKeyType{value: npk}
			npk = pko.UpdateValue(data.TargetAmount)
		}
	}

	nso := panelKeyType{value: npk}
	nst := p.GetStatus()
	ost := p.GetStatus()

	switch nst {
	case JC.STATE_LOADING, JC.STATE_FETCHING_NEW, JC.STATE_ERROR:
		if nso.IsValueMatchingFloat(0, JC.STRING_GREATER_EQUAL) {
			nst = JC.STATE_LOADED
		}
	case JC.STATE_LOADED:
	}

	p.SetStatus(nst)

	if npk != opk || nst != ost {
		p.Set(npk)
		p.SetOldKey(opk)
		return true
	}

	return false
}

func (p *panelDataType) UpdateRate() bool {
	if JC.IsShuttingDown() {
		return false
	}

	pk := p.UsePanelKey()
	ck := UseExchangeCache().CreateKeyFromInt(pk.GetSourceCoinInt(), pk.GetTargetCoinInt())

	if UseExchangeCache().Has(ck) {
		dt := UseExchangeCache().Get(ck)
		if dt != nil && dt.TargetAmount != nil {
			return p.SetRate(dt.TargetAmount)
		}
	}

	return false
}

func (p *panelDataType) UpdateStatus() bool {
	if JC.IsShuttingDown() {
		return false
	}

	switch p.GetStatus() {
	case JC.STATE_LOADING, JC.STATE_FETCHING_NEW, JC.STATE_ERROR:
		if p.UsePanelKey().IsValueMatchingFloat(0, JC.STRING_GREATER_EQUAL) {
			p.SetStatus(JC.STATE_LOADED)
			return true
		}
	case JC.STATE_LOADED:
		return true
	}

	return false
}

func (p *panelDataType) Destroy() {
	p.data = nil
	p.parent = nil
	p.oldKey = JC.STRING_EMPTY
	p.id = JC.STRING_EMPTY
}

func (p *panelDataType) FormatTitle() string {
	pk := p.UsePanelKey()

	var b strings.Builder
	b.WriteString(pk.GetSourceValueFormattedString())
	b.WriteString(fmtSpace)
	b.WriteString(pk.GetSourceSymbolString())
	b.WriteString(fmtTo)
	b.WriteString(pk.GetTargetSymbolString())

	return b.String()
}

func (p *panelDataType) FormatSubtitle() string {
	pk := p.UsePanelKey()

	var b strings.Builder
	b.WriteString(fmtVal)
	b.WriteString(pk.GetSourceSymbolString())
	b.WriteString(fmtEqual)
	b.WriteString(pk.GetValueFormattedString())
	b.WriteString(fmtSpace)
	b.WriteString(pk.GetTargetSymbolString())

	return b.String()
}

func (p *panelDataType) FormatBottomText() string {
	pk := p.UsePanelKey()

	var b strings.Builder
	b.WriteString(fmtVal)
	b.WriteString(pk.GetTargetSymbolString())
	b.WriteString(fmtEqual)
	b.WriteString(pk.GetReverseValueFormattedString())
	b.WriteString(fmtSpace)
	b.WriteString(pk.GetSourceSymbolString())

	return b.String()
}

func (p *panelDataType) FormatContent() string {
	pk := p.UsePanelKey()

	var b strings.Builder
	b.WriteString(pk.GetCalculatedValueFormattedString())
	b.WriteString(fmtSpace)
	b.WriteString(pk.GetTargetSymbolString())

	return b.String()
}

func (p *panelDataType) DidChange() bool {
	opt := &panelKeyType{value: p.oldKey}
	return p.oldKey != p.Get() &&
		opt.IsValueMatchingFloat(-1, JC.STRING_NOT_EQUAL) &&
		p.IsStatus(JC.STATE_LOADED)
}

func (p *panelDataType) Insert(panel panelType, rate float64) {
	if p.data != nil {
		p.oldKey = p.data.GetData()
		pkt := &panelKeyType{}
		r := JC.ToBigFloat(rate)
		p.data.SetData(pkt.GenerateKeyFromPanel(panel, r))
	}
}

func (p *panelDataType) Serialize() panelDataCache {
	return panelDataCache{
		Status: p.GetStatus(),
		Key:    p.RefreshKey(p.Get()),
		OldKey: p.RefreshKey(p.GetOldKey()),
	}
}

func NewPanelDataCache() []panelDataCache {
	return []panelDataCache{}
}

func NewPanelData() *panelDataType {
	return &panelDataType{}
}
