package types

import (
	"fmt"
	"strconv"
	"sync"

	"fyne.io/fyne/v2/data/binding"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	JC "jxwatcher/core"
)

type TickerDataCache struct {
	Type   string
	Title  string
	Format string
	Status int
	Key    string
	OldKey string
}

type TickerDataType struct {
	Data   binding.String
	OldKey JC.StringStore
	Type   JC.StringStore
	Title  JC.StringStore
	Format JC.StringStore
	ID     JC.StringStore
	Status JC.IntStore
	mu     sync.Mutex
}

func (p *TickerDataType) Init() {
	p.Data = binding.NewString()
	p.ID = JC.StringStore{}
	p.Status = JC.IntStore{}
	p.OldKey = JC.StringStore{}
}

func (p *TickerDataType) Insert(rate string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.OldKey.Set(p.Get())
	p.Data.Set(rate)
}

func (p *TickerDataType) Set(rate string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.OldKey.Set(p.Get())
	p.Data.Set(rate)
}

func (p *TickerDataType) Get() string {
	if p.Data == nil {
		return ""
	}

	rate, err := p.Data.Get()
	if err == nil {
		return rate
	}

	return ""
}

func (p *TickerDataType) GetData() binding.String {
	return p.Data
}

func (p *TickerDataType) HasData() bool {
	raw := p.Get()

	val, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return false
	}

	return val >= 0
}

func (p *TickerDataType) Update() bool {

	if !TickerCache.Has(p.Type.Get()) {
		return false
	}

	npk := TickerCache.Get(p.Type.Get())
	opk := p.Get()
	nso := PanelKeyType{value: npk}
	nst := p.Status.Get()

	if npk == "" {
		return false
	}

	switch p.Status.Get() {
	case JC.STATE_LOADING, JC.STATE_FETCHING_NEW, JC.STATE_ERROR:

		// Watch out, value might be minus for things like percentage changes
		if nso.GetValueFloat() >= 0 {
			nst = JC.STATE_LOADED
		}

	case JC.STATE_LOADED:
		// Do nothing?
	}

	p.Status.Set(nst)

	// JC.Logln(fmt.Sprintf("Trying to update tickers %v with old value = %v, old status = %v to new value = %v, new status = %v", p.Type, opk, p.Status, npk, nst))

	if npk != opk {
		p.Set(npk)
	}

	return true
}

func (p *TickerDataType) FormatContent() string {
	raw := p.Get()

	val, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return raw
	}

	switch p.Format.Get() {
	case "nodecimal":
		return fmt.Sprintf("%.0f", val)
	case "number":
		return fmt.Sprintf("%.2f", val)

	case "currency":
		printer := message.NewPrinter(language.English)
		return printer.Sprintf("$%0.2f", val)

	case "shortcurrency":
		return JC.FormatShortCurrency(raw)

	case "percentage":
		return fmt.Sprintf("%s/100", raw)

	default:
		return raw
	}
}

func (p *TickerDataType) DidChange() bool {
	return !p.OldKey.IsEqual(p.Get()) && p.Status.IsEqual(JC.STATE_LOADED)
}

func (t *TickerDataType) Serialize() TickerDataCache {
	return TickerDataCache{
		Type:   t.Type.Get(),
		Title:  t.Title.Get(),
		Format: t.Format.Get(),
		Status: t.Status.Get(),
		Key:    t.Get(),
		OldKey: t.OldKey.Get(),
	}
}
