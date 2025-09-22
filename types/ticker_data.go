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

type tickerDataCache struct {
	Type   string
	Title  string
	Format string
	Status int
	Key    string
	OldKey string
}

type TickerDataType struct {
	mu       sync.RWMutex
	data     binding.String
	oldKey   string
	category string
	title    string
	format   string
	id       string
	status   int
}

func (p *TickerDataType) Init() {
	p.mu.Lock()
	p.data = binding.NewString()
	p.id = ""
	p.status = JC.STATE_LOADING
	p.oldKey = ""
	p.mu.Unlock()
}

func (p *TickerDataType) Set(rate string) {
	p.mu.Lock()
	old, err := p.data.Get()
	if err == nil {
		p.oldKey = old
	}
	p.data.Set(rate)
	p.mu.Unlock()
}

func (p *TickerDataType) Insert(rate string) {
	p.Set(rate)
}

func (p *TickerDataType) Get() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.data == nil {
		return ""
	}
	val, err := p.data.Get()
	if err != nil {
		return ""
	}
	return val
}

func (p *TickerDataType) GetData() binding.String {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.data
}

func (p *TickerDataType) HasData() bool {
	raw := p.Get()
	val, err := strconv.ParseFloat(raw, 64)
	return err == nil && val >= 0
}

func (p *TickerDataType) GetType() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.category
}

func (p *TickerDataType) SetType(val string) {
	p.mu.Lock()
	p.category = val
	p.mu.Unlock()
}

func (p *TickerDataType) GetTitle() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.title
}

func (p *TickerDataType) SetTitle(val string) {
	p.mu.Lock()
	p.title = val
	p.mu.Unlock()
}

func (p *TickerDataType) GetFormat() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.format
}

func (p *TickerDataType) SetFormat(val string) {
	p.mu.Lock()
	p.format = val
	p.mu.Unlock()
}

func (p *TickerDataType) GetStatus() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.status
}

func (p *TickerDataType) SetStatus(val int) {
	p.mu.Lock()
	p.status = val
	p.mu.Unlock()
}

func (p *TickerDataType) GetID() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.id
}

func (p *TickerDataType) SetID(val string) {
	p.mu.Lock()
	p.id = val
	p.mu.Unlock()
}

func (p *TickerDataType) GetOldKey() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.oldKey
}

func (p *TickerDataType) SetOldKey(val string) {
	p.mu.Lock()
	p.oldKey = val
	p.mu.Unlock()
}

func (p *TickerDataType) IsType(val string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.category == val
}

func (p *TickerDataType) IsTitle(val string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.title == val
}

func (p *TickerDataType) IsFormat(val string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.format == val
}

func (p *TickerDataType) IsStatus(val int) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.status == val
}

func (p *TickerDataType) IsID(val string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.id == val
}

func (p *TickerDataType) IsOldKey(val string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.oldKey == val
}

func (p *TickerDataType) IsKey(val string) bool {
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

func (p *TickerDataType) Update() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !TickerCache.Has(p.category) {
		return false
	}
	npk := TickerCache.Get(p.category)
	if npk == "" {
		return false
	}

	opk := ""
	if p.data != nil {
		v, err := p.data.Get()
		if err == nil {
			opk = v
		}
	}

	nso := panelKeyType{value: npk}
	nst := p.status

	switch p.status {
	case JC.STATE_LOADING, JC.STATE_FETCHING_NEW, JC.STATE_ERROR:
		if nso.GetValueFloat() >= 0 {
			nst = JC.STATE_LOADED
		}
	case JC.STATE_LOADED:
	}

	p.status = nst
	if npk != opk {
		p.oldKey = opk
		p.data.Set(npk)
	}

	return true
}

func (p *TickerDataType) FormatContent() string {
	raw := p.Get()
	val, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return raw
	}
	switch p.GetFormat() {
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
	defer p.mu.RUnlock()

	if p.status != JC.STATE_LOADED || p.oldKey == "" {
		return false
	}
	if p.data != nil {
		v, err := p.data.Get()
		if err == nil && v != p.oldKey {
			return true
		}
	}
	return false
}

func (p *TickerDataType) Serialize() tickerDataCache {
	p.mu.RLock()
	defer p.mu.RUnlock()

	key := ""
	if p.data != nil {
		v, err := p.data.Get()
		if err == nil {
			key = v
		}
	}

	return tickerDataCache{
		Type:   p.category,
		Title:  p.title,
		Format: p.format,
		Status: p.status,
		Key:    key,
		OldKey: p.oldKey,
	}
}

func NewTickerDataCache() []tickerDataCache {
	return []tickerDataCache{}
}
