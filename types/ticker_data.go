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
	mu     sync.RWMutex
	Data   binding.String
	OldKey string
	Type   string
	Title  string
	Format string
	ID     string
	Status int
}

func (p *TickerDataType) Init() {
	p.mu.Lock()
	p.Data = binding.NewString()
	p.ID = ""
	p.Status = JC.STATE_LOADING
	p.OldKey = ""
	p.mu.Unlock()
}

func (p *TickerDataType) Set(rate string) {
	p.mu.Lock()
	old, err := p.Data.Get()
	if err == nil {
		p.OldKey = old
	}
	p.Data.Set(rate)
	p.mu.Unlock()
}

func (p *TickerDataType) Insert(rate string) {
	p.Set(rate)
}

func (p *TickerDataType) Get() string {
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

func (p *TickerDataType) GetData() binding.String {
	p.mu.RLock()
	d := p.Data
	p.mu.RUnlock()
	return d
}

func (p *TickerDataType) HasData() bool {
	raw := p.Get()
	val, err := strconv.ParseFloat(raw, 64)
	return err == nil && val >= 0
}

func (p *TickerDataType) GetType() string {
	p.mu.RLock()
	v := p.Type
	p.mu.RUnlock()
	return v
}

func (p *TickerDataType) GetTitle() string {
	p.mu.RLock()
	v := p.Title
	p.mu.RUnlock()
	return v
}

func (p *TickerDataType) GetFormat() string {
	p.mu.RLock()
	v := p.Format
	p.mu.RUnlock()
	return v
}

func (p *TickerDataType) GetStatus() int {
	p.mu.RLock()
	v := p.Status
	p.mu.RUnlock()
	return v
}

func (p *TickerDataType) GetOldKey() string {
	p.mu.RLock()
	v := p.OldKey
	p.mu.RUnlock()
	return v
}

func (p *TickerDataType) GetID() string {
	p.mu.RLock()
	v := p.ID
	p.mu.RUnlock()
	return v
}

func (p *TickerDataType) SetType(val string) {
	p.mu.Lock()
	p.Type = val
	p.mu.Unlock()
}

func (p *TickerDataType) SetTitle(val string) {
	p.mu.Lock()
	p.Title = val
	p.mu.Unlock()
}

func (p *TickerDataType) SetFormat(val string) {
	p.mu.Lock()
	p.Format = val
	p.mu.Unlock()
}

func (p *TickerDataType) SetStatus(val int) {
	p.mu.Lock()
	p.Status = val
	p.mu.Unlock()
}

func (p *TickerDataType) SetID(val string) {
	p.mu.Lock()
	p.ID = val
	p.mu.Unlock()
}

func (p *TickerDataType) SetOldKey(val string) {
	p.mu.Lock()
	p.OldKey = val
	p.mu.Unlock()
}

func (p *TickerDataType) IsType(val string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.Type == val
}

func (p *TickerDataType) IsTitle(val string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.Title == val
}

func (p *TickerDataType) IsFormat(val string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.Format == val
}

func (p *TickerDataType) IsStatus(val int) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.Status == val
}

func (p *TickerDataType) IsID(val string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.ID == val
}

func (p *TickerDataType) IsOldKey(val string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.OldKey == val
}

func (p *TickerDataType) IsKey(val string) bool {
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

func (p *TickerDataType) Update() bool {
	p.mu.Lock()
	if !TickerCache.Has(p.Type) {
		p.mu.Unlock()
		return false
	}
	npk := TickerCache.Get(p.Type)
	opk := ""
	if p.Data != nil {
		v, err := p.Data.Get()
		if err == nil {
			opk = v
		}
	}
	nso := PanelKeyType{value: npk}
	nst := p.Status
	if npk == "" {
		p.mu.Unlock()
		return false
	}
	switch p.Status {
	case JC.STATE_LOADING, JC.STATE_FETCHING_NEW, JC.STATE_ERROR:
		if nso.GetValueFloat() >= 0 {
			nst = JC.STATE_LOADED
		}
	case JC.STATE_LOADED:
	}
	p.Status = nst
	if npk != opk {
		p.OldKey = opk
		p.Data.Set(npk)
	}
	p.mu.Unlock()
	return true
}

func (p *TickerDataType) FormatContent() string {
	raw := p.Get()
	val, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return raw
	}
	format := p.GetFormat()
	switch format {
	case "nodecimal":
		return fmt.Sprintf("%.0f", val)
	case "number":
		return fmt.Sprintf("%.2f", val)
	case "currency":
		printer := message.NewPrinter(language.English)
		return printer.Sprintf("$%.2f", val)
	case "shortcurrency":
		return JC.FormatShortCurrency(raw)
	case "percentage":
		return fmt.Sprintf("%s/100", raw)
	default:
		return raw
	}
}

func (p *TickerDataType) DidChange() bool {
	p.mu.RLock()
	changed := p.Status == JC.STATE_LOADED && p.OldKey != ""
	if p.Data != nil {
		v, err := p.Data.Get()
		if err == nil && v != p.OldKey {
			changed = changed && true
		} else {
			changed = false
		}
	}
	p.mu.RUnlock()
	return changed
}

func (p *TickerDataType) Serialize() TickerDataCache {
	p.mu.RLock()
	key := ""
	if p.Data != nil {
		v, err := p.Data.Get()
		if err == nil {
			key = v
		}
	}
	cache := TickerDataCache{
		Type:   p.Type,
		Title:  p.Title,
		Format: p.Format,
		Status: p.Status,
		Key:    key,
		OldKey: p.OldKey,
	}
	p.mu.RUnlock()
	return cache
}
