package types

import (
	"fmt"
	"strconv"
	"sync"

	"fyne.io/fyne/v2/data/binding"

	JC "jxwatcher/core"
)

type PanelData interface {
	Init()
	Set(val string)
	SetStatus(val int)
	SetID(val string)
	SetOldKey(val string)
	SetParent(val *panelsMapType)
	Get() string
	GetStatus() int
	GetID() string
	GetOldKey() string
	GetParent() *panelsMapType
	GetValueString() string
	GetOldValueString() string
	UseData() binding.String
	UseStatus() binding.Int
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
	mu     sync.RWMutex
	data   binding.String
	oldKey string
	id     string
	status binding.Int
	parent *panelsMapType
}

func (p *panelDataType) Init() {
	p.mu.Lock()
	p.data = binding.NewString()
	p.status = binding.NewInt()
	p.id = ""
	p.oldKey = ""
	p.mu.Unlock()
}

func (p *panelDataType) Set(val string) {
	p.mu.Lock()
	old, err := p.data.Get()
	if err == nil {
		p.oldKey = old
	}
	p.data.Set(val)
	p.mu.Unlock()
}

func (p *panelDataType) SetStatus(val int) {
	if p.IsStatus(val) {
		return
	}
	p.mu.Lock()
	p.status.Set(val)
	p.mu.Unlock()
}

func (p *panelDataType) SetID(val string) {
	p.mu.Lock()
	p.id = val
	p.mu.Unlock()
}

func (p *panelDataType) SetOldKey(val string) {
	p.mu.Lock()
	p.oldKey = val
	p.mu.Unlock()
}

func (p *panelDataType) SetParent(val *panelsMapType) {
	p.mu.Lock()
	p.parent = val
	p.mu.Unlock()
}

func (p *panelDataType) Get() string {
	p.mu.RLock()
	val := ""
	if p.data != nil {
		v, err := p.data.Get()
		if err == nil {
			val = v
		}
	}
	p.mu.RUnlock()
	return val
}

func (p *panelDataType) GetStatus() int {
	p.mu.RLock()
	s := p.status
	p.mu.RUnlock()
	v, err := s.Get()
	if err != nil {
		return JC.STATE_ERROR
	}
	return v
}

func (p *panelDataType) GetID() string {
	p.mu.RLock()
	v := p.id
	p.mu.RUnlock()
	return v
}

func (p *panelDataType) GetOldKey() string {
	p.mu.RLock()
	v := p.oldKey
	p.mu.RUnlock()
	return v
}

func (p *panelDataType) GetParent() *panelsMapType {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.parent
}

func (p *panelDataType) GetValueString() string {
	return p.UsePanelKey().GetValueString()
}

func (p *panelDataType) GetOldValueString() string {
	p.mu.RLock()
	old := p.oldKey
	p.mu.RUnlock()
	pko := panelKeyType{value: old}
	return pko.GetValueString()
}

func (p *panelDataType) UseData() binding.String {
	p.mu.RLock()
	d := p.data
	p.mu.RUnlock()
	return d
}

func (p *panelDataType) UseStatus() binding.Int {
	p.mu.RLock()
	d := p.status
	p.mu.RUnlock()
	return d
}

func (p *panelDataType) UsePanelKey() *panelKeyType {
	return &panelKeyType{value: p.Get()}
}

func (p *panelDataType) IsStatus(val int) bool {
	p.mu.RLock()
	s := p.status
	p.mu.RUnlock()
	v, _ := s.Get()
	return v == val
}

func (p *panelDataType) IsID(val string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.id == val
}

func (p *panelDataType) IsKey(val string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.data == nil {
		return val == ""
	}
	current, err := p.data.Get()
	if err != nil {
		return false
	}
	return current == val
}

func (p *panelDataType) IsOldKey(val string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.oldKey == val
}

func (p *panelDataType) IsEqualContentString(pk string) bool {
	return p.IsKey(pk)
}

func (p *panelDataType) IsOnInitialValue() bool {
	p.mu.RLock()
	old := p.oldKey
	p.mu.RUnlock()
	opt := &panelKeyType{value: old}
	return opt.IsValueMatchingFloat(-1, "==") && p.IsStatus(JC.STATE_LOADED)
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
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.parent != nil
}

func (p *panelDataType) RefreshData() {
	npk := p.UsePanelKey().UpdateValue(JC.ToBigFloat(-3))
	p.mu.Lock()
	p.data.Set(npk)
	p.mu.Unlock()
}

func (p *panelDataType) RefreshKey(key string) string {
	if !p.parent.ValidateKey(key) {
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

	parent := p.GetParent()

	if parent != nil && sourceSymbol == "" {
		sourceSymbol = parent.GetSymbolById(source)
	}

	if parent != nil && targetSymbol == "" {
		targetSymbol = parent.GetSymbolById(target)
	}

	return pkt.GenerateKey(source, target, value, sourceSymbol, targetSymbol, decimals, rate)
}

func (p *panelDataType) Update(pk string) bool {
	opk := p.Get()
	npk := pk
	pks := p.UsePanelKey()

	ck := UseExchangeCache().CreateKeyFromInt(pks.GetSourceCoinInt(), pks.GetTargetCoinInt())

	if UseExchangeCache().Has(ck) {
		Data := UseExchangeCache().Get(ck)
		if Data != nil && Data.TargetAmount != nil {
			pko := panelKeyType{value: npk}
			npk = pko.UpdateValue(Data.TargetAmount)
		}
	}

	nso := panelKeyType{value: npk}
	nst := p.GetStatus()
	ost := p.GetStatus()

	switch nst {
	case JC.STATE_LOADING, JC.STATE_FETCHING_NEW, JC.STATE_ERROR:
		if nso.IsValueMatchingFloat(0, ">=") {
			nst = JC.STATE_LOADED
		}
	case JC.STATE_LOADED:
	}

	p.SetStatus(nst)

	if npk != opk || nst != ost {
		// JC.Logln("Updating panel:", npk, opk, p.status)
		p.Set(npk)
		p.SetOldKey(opk)
		return true
	}

	return false
}

func (p *panelDataType) FormatTitle() string {
	pk := p.UsePanelKey()
	frac := int(JC.NumDecPlaces(pk.GetSourceValueFloat()))
	if frac < 3 {
		frac = 2
	}
	return fmt.Sprintf("%s %s to %s",
		pk.GetSourceValueFormattedString(),
		pk.GetSourceSymbolString(),
		pk.GetTargetSymbolString(),
	)
}

func (p *panelDataType) FormatSubtitle() string {
	pk := p.UsePanelKey()
	return fmt.Sprintf("1 %s = %s %s",
		pk.GetSourceSymbolString(),
		pk.GetValueFormattedString(),
		pk.GetTargetSymbolString(),
	)
}

func (p *panelDataType) FormatBottomText() string {
	pk := p.UsePanelKey()
	return fmt.Sprintf("1 %s = %s %s",
		pk.GetTargetSymbolString(),
		pk.GetReverseValueFormattedString(),
		pk.GetSourceSymbolString(),
	)
}

func (p *panelDataType) FormatContent() string {
	pk := p.UsePanelKey()
	return fmt.Sprintf("%s %s",
		pk.GetCalculatedValueFormattedString(),
		pk.GetTargetSymbolString(),
	)
}

func (p *panelDataType) DidChange() bool {
	p.mu.RLock()
	old := p.oldKey
	p.mu.RUnlock()
	opt := &panelKeyType{value: old}
	return old != p.Get() && opt.IsValueMatchingFloat(-1, "!=") && p.IsStatus(JC.STATE_LOADED)
}

func (p *panelDataType) Insert(panel panelType, rate float64) {
	p.mu.Lock()
	old, err := p.data.Get()
	if err == nil {
		p.oldKey = old
	}
	pkt := &panelKeyType{}
	r := JC.ToBigFloat(rate)
	p.data.Set(pkt.GenerateKeyFromPanel(panel, r))
	p.mu.Unlock()
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
