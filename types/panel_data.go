package types

import (
	"fmt"
	"strconv"
	"sync"

	"fyne.io/fyne/v2/data/binding"

	JC "jxwatcher/core"
)

type PanelDataType struct {
	Data   binding.String
	OldKey JC.StringStore
	Parent *PanelsMapType
	Status JC.IntStore
	ID     JC.StringStore
	mu     sync.Mutex
}

type PanelDataCache struct {
	Status int
	Key    string
	OldKey string
}

func (p *PanelDataType) Init() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.Data = binding.NewString()
	p.ID = JC.StringStore{}
	p.OldKey = JC.StringStore{}
	p.Status = JC.IntStore{}
}

func (p *PanelDataType) Insert(panel PanelType, rate float32) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.OldKey.Set(p.Get())
	pkt := &PanelKeyType{}
	p.Data.Set(pkt.GenerateKeyFromPanel(panel, rate))
}

func (p *PanelDataType) Set(pk string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.OldKey.Set(p.Get())
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
	pko := PanelKeyType{value: p.OldKey.Get()}
	return pko.GetValueString()
}

func (p *PanelDataType) RefreshData() {
	npk := p.UsePanelKey().UpdateValue(-3)
	p.Data.Set(npk)
}

func (p *PanelDataType) UsePanelKey() *PanelKeyType {
	pko := PanelKeyType{value: p.Get()}
	return &pko
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
	nst := p.Status.Get()

	switch p.Status.Get() {
	case JC.STATE_LOADING, JC.STATE_FETCHING_NEW, JC.STATE_ERROR:
		if nso.GetValueFloat() >= 0 {
			nst = JC.STATE_LOADED
		}
	case JC.STATE_LOADED:
		// Do nothing?
	}

	// JC.Logln(fmt.Sprintf("Trying to update panel with old value = %v, old status = %v to new value = %v, new status = %v", opk, p.Status, npk, nst))

	p.Status.Set(nst)

	if npk != opk {
		p.Set(npk)
		p.OldKey.Set(opk)
	}

	return true
}

func (p *PanelDataType) FormatTitle() string {

	pk := p.UsePanelKey()

	frac := int(JC.NumDecPlaces(pk.GetSourceValueFloat()))
	if frac < 3 {
		frac = 2
	}

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
		"1 %s = %s %s",
		pk.GetSourceSymbolString(),
		pk.GetValueFormattedString(),
		pk.GetTargetSymbolString(),
	)
}

func (p *PanelDataType) FormatBottomText() string {

	pk := p.UsePanelKey()

	return fmt.Sprintf(
		"1 %s = %s %s",
		pk.GetTargetSymbolString(),
		pk.GetReverseValueFormattedString(),
		pk.GetSourceSymbolString(),
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
	opt := &PanelKeyType{p.OldKey.Get()}
	return p.OldKey.Get() != p.Get() && opt.GetValueFloat() != -1 && p.Status.IsEqual(JC.STATE_LOADED)
}

func (p *PanelDataType) IsOnInitialValue() bool {
	opt := &PanelKeyType{p.OldKey.Get()}
	return opt.GetValueFloat() == -1 && p.Status.IsEqual(JC.STATE_LOADED)
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
		// JC.Logf("Error Formatting")
		return JC.VALUE_NO_CHANGE
	}

	if numA > numB {
		// JC.Logf("%s (%.2f) is greater than %s (%.2f)\n", a, numA, b, numB)
		return JC.VALUE_DECREASE
	}

	if numA < numB {
		// JC.Logf("%s (%.2f) is less than %s (%.2f)\n", a, numA, b, numB)
		return JC.VALUE_INCREASE
	}

	return JC.VALUE_NO_CHANGE
}

func (p *PanelDataType) IsEqualContentString(pk string) bool {
	return p.Get() == pk
}

func (p *PanelDataType) Serialize() PanelDataCache {
	return PanelDataCache{
		Status: p.Status.Get(),
		Key:    p.Get(),
		OldKey: p.OldKey.Get(),
	}
}
