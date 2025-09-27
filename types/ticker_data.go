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

type TickerData interface {
	Init()
	Set(rate string)
	Insert(rate string)
	Get() string
	GetData() binding.String
	HasData() bool

	GetType() string
	SetType(val string)
	GetTitle() string
	SetTitle(val string)
	GetFormat() string
	SetFormat(val string)
	GetStatus() int
	SetStatus(val int)
	GetID() string
	SetID(val string)
	GetOldKey() string
	SetOldKey(val string)

	IsType(val string) bool
	IsTitle(val string) bool
	IsFormat(val string) bool
	IsStatus(val int) bool
	IsID(val string) bool
	IsOldKey(val string) bool
	IsKey(val string) bool

	Update() bool
	FormatContent() string
	DidChange() bool
	Serialize() tickerDataCache
}

type tickerDataCache struct {
	Type   string
	Title  string
	Format string
	Status int
	Key    string
	OldKey string
}

type tickerDataType struct {
	mu       sync.RWMutex
	data     binding.String
	oldKey   string
	category string
	title    string
	format   string
	id       string
	status   int
}

func (p *tickerDataType) Init() {
	p.mu.Lock()
	p.data = binding.NewString()
	p.id = ""
	p.status = JC.STATE_LOADING
	p.oldKey = ""
	p.mu.Unlock()
}

func (p *tickerDataType) Set(rate string) {
	p.mu.Lock()
	old, err := p.data.Get()
	if err == nil {
		p.oldKey = old
	}
	p.data.Set(rate)
	p.mu.Unlock()
}

func (p *tickerDataType) Insert(rate string) {
	p.Set(rate)
}

func (p *tickerDataType) Get() string {
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

func (p *tickerDataType) GetData() binding.String {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.data
}

func (p *tickerDataType) HasData() bool {
	raw := p.Get()
	val, err := strconv.ParseFloat(raw, 64)
	return err == nil && val >= 0
}

func (p *tickerDataType) GetType() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.category
}

func (p *tickerDataType) SetType(val string) {
	p.mu.Lock()
	p.category = val
	p.mu.Unlock()
}

func (p *tickerDataType) GetTitle() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.title
}

func (p *tickerDataType) SetTitle(val string) {
	p.mu.Lock()
	p.title = val
	p.mu.Unlock()
}

func (p *tickerDataType) GetFormat() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.format
}

func (p *tickerDataType) SetFormat(val string) {
	p.mu.Lock()
	p.format = val
	p.mu.Unlock()
}

func (p *tickerDataType) GetStatus() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.status
}

func (p *tickerDataType) SetStatus(val int) {
	p.mu.Lock()
	p.status = val
	p.mu.Unlock()
}

func (p *tickerDataType) GetID() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.id
}

func (p *tickerDataType) SetID(val string) {
	p.mu.Lock()
	p.id = val
	p.mu.Unlock()
}

func (p *tickerDataType) GetOldKey() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.oldKey
}

func (p *tickerDataType) SetOldKey(val string) {
	p.mu.Lock()
	p.oldKey = val
	p.mu.Unlock()
}

func (p *tickerDataType) IsType(val string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.category == val
}

func (p *tickerDataType) IsTitle(val string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.title == val
}

func (p *tickerDataType) IsFormat(val string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.format == val
}

func (p *tickerDataType) IsStatus(val int) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.status == val
}

func (p *tickerDataType) IsID(val string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.id == val
}

func (p *tickerDataType) IsOldKey(val string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.oldKey == val
}

func (p *tickerDataType) IsKey(val string) bool {
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

func (p *tickerDataType) Update() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !tickerCacheStorage.Has(p.category) {
		return false
	}
	npk := tickerCacheStorage.Get(p.category)
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
		if nso.IsValueMatchingFloat(0, ">=") {
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

func (p *tickerDataType) FormatContent() string {
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

func (p *tickerDataType) DidChange() bool {
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

func (p *tickerDataType) Serialize() tickerDataCache {
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

func NewTickerData() *tickerDataType {
	return &tickerDataType{}
}
