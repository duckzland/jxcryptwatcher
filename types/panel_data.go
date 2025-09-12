package types

import (
	"fmt"
	"strconv"
	"sync"

	"fyne.io/fyne/v2/data/binding"

	JC "jxwatcher/core"
)

type PanelDataCache struct {
	Status int
	Key    string
	OldKey string
}

type PanelDataType struct {
	mu     sync.RWMutex
	Data   binding.String
	OldKey string
	ID     string
	Status int
	Parent *PanelsMapType
}

func (p *PanelDataType) Init() {
	p.mu.Lock()
	p.Data = binding.NewString()
	p.ID = ""
	p.OldKey = ""
	p.Status = JC.STATE_LOADING
	p.mu.Unlock()
}

func (p *PanelDataType) Set(val string) {
	p.mu.Lock()
	old, err := p.Data.Get()
	if err == nil {
		p.OldKey = old
	}
	p.Data.Set(val)
	p.mu.Unlock()
}

func (p *PanelDataType) Insert(panel PanelType, rate float32) {
	p.mu.Lock()
	old, err := p.Data.Get()
	if err == nil {
		p.OldKey = old
	}
	pkt := &PanelKeyType{}
	p.Data.Set(pkt.GenerateKeyFromPanel(panel, rate))
	p.mu.Unlock()
}

func (p *PanelDataType) Get() string {
	p.mu.RLock()
	val := ""
	if p.Data != nil {
		v, err := p.Data.Get()
		if err == nil {
			val = v
		}
	}
	p.mu.RUnlock()
	return val
}

func (p *PanelDataType) GetData() binding.String {
	p.mu.RLock()
	d := p.Data
	p.mu.RUnlock()
	return d
}

func (p *PanelDataType) GetStatus() int {
	p.mu.RLock()
	v := p.Status
	p.mu.RUnlock()
	return v
}

func (p *PanelDataType) SetStatus(val int) {
	p.mu.Lock()
	p.Status = val
	p.mu.Unlock()
}

func (p *PanelDataType) GetID() string {
	p.mu.RLock()
	v := p.ID
	p.mu.RUnlock()
	return v
}

func (p *PanelDataType) SetID(val string) {
	p.mu.Lock()
	p.ID = val
	p.mu.Unlock()
}

func (p *PanelDataType) GetOldKey() string {
	p.mu.RLock()
	v := p.OldKey
	p.mu.RUnlock()
	return v
}

func (p *PanelDataType) SetOldKey(val string) {
	p.mu.Lock()
	p.OldKey = val
	p.mu.Unlock()
}

func (p *PanelDataType) IsStatus(val int) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.Status == val
}

func (p *PanelDataType) IsID(val string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.ID == val
}

func (p *PanelDataType) IsKey(val string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.Data == nil {
		return val == ""
	}
	current, err := p.Data.Get()
	if err != nil {
		return false
	}
	return current == val
}

func (p *PanelDataType) IsOldKey(val string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.OldKey == val
}

func (p *PanelDataType) GetValueString() string {
	return p.UsePanelKey().GetValueString()
}

func (p *PanelDataType) GetOldValueString() string {
	p.mu.RLock()
	old := p.OldKey
	p.mu.RUnlock()
	pko := PanelKeyType{value: old}
	return pko.GetValueString()
}

func (p *PanelDataType) RefreshData() {
	npk := p.UsePanelKey().UpdateValue(-3)
	p.mu.Lock()
	p.Data.Set(npk)
	p.mu.Unlock()
}

func (p *PanelDataType) UsePanelKey() *PanelKeyType {
	return &PanelKeyType{value: p.Get()}
}

func (p *PanelDataType) Update(pk string) bool {
	opk := p.Get()
	npk := pk
	pks := p.UsePanelKey()

	ck := ExchangeCache.CreateKeyFromInt(pks.GetSourceCoinInt(), pks.GetTargetCoinInt())

	if ExchangeCache.Has(ck) {
		Data := ExchangeCache.Get(ck)
		if Data != nil {
			pko := PanelKeyType{value: npk}
			npk = pko.UpdateValue(Data.TargetAmount)
		}
	}

	nso := PanelKeyType{value: npk}
	nst := p.GetStatus()

	switch nst {
	case JC.STATE_LOADING, JC.STATE_FETCHING_NEW, JC.STATE_ERROR:
		if nso.GetValueFloat() >= 0 {
			nst = JC.STATE_LOADED
		}
	case JC.STATE_LOADED:
	}

	p.SetStatus(nst)

	if npk != opk {
		p.Set(npk)
		p.SetOldKey(opk)
	}

	return true
}

func (p *PanelDataType) FormatTitle() string {
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

func (p *PanelDataType) FormatSubtitle() string {
	pk := p.UsePanelKey()
	return fmt.Sprintf("1 %s = %s %s",
		pk.GetSourceSymbolString(),
		pk.GetValueFormattedString(),
		pk.GetTargetSymbolString(),
	)
}

func (p *PanelDataType) FormatBottomText() string {
	pk := p.UsePanelKey()
	return fmt.Sprintf("1 %s = %s %s",
		pk.GetTargetSymbolString(),
		pk.GetReverseValueFormattedString(),
		pk.GetSourceSymbolString(),
	)
}

func (p *PanelDataType) FormatContent() string {
	pk := p.UsePanelKey()
	return fmt.Sprintf("%s %s",
		pk.GetCalculatedValueFormattedString(),
		pk.GetTargetSymbolString(),
	)
}

func (p *PanelDataType) DidChange() bool {
	p.mu.RLock()
	old := p.OldKey
	status := p.Status
	p.mu.RUnlock()
	opt := &PanelKeyType{old}
	return old != p.Get() && opt.GetValueFloat() != -1 && status == JC.STATE_LOADED
}

func (p *PanelDataType) IsOnInitialValue() bool {
	p.mu.RLock()
	old := p.OldKey
	status := p.Status
	p.mu.RUnlock()
	opt := &PanelKeyType{old}
	return opt.GetValueFloat() == -1 && status == JC.STATE_LOADED
}

func (p *PanelDataType) IsValueIncrease() int {
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

func (p *PanelDataType) IsEqualContentString(pk string) bool {
	return p.IsKey(pk)
}

func (p *PanelDataType) Serialize() PanelDataCache {
	return PanelDataCache{
		Status: p.GetStatus(),
		Key:    p.Get(),
		OldKey: p.GetOldKey(),
	}
}
