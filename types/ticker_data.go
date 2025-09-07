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

type TickerDataType struct {
	Data   binding.String
	OldKey string
	Type   string
	Title  string
	Format string
	ID     string
	Status int
	mu     sync.Mutex
}

func (p *TickerDataType) Init() {
	p.Data = binding.NewString()
}

func (p *TickerDataType) Insert(rate string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.OldKey = p.Get()
	p.Data.Set(rate)
}

func (p *TickerDataType) Set(rate string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.OldKey = p.Get()
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

func (p *TickerDataType) Update() bool {
	if TickerCache.Has(p.Type) {
		nd := TickerCache.Get(p.Type)
		if nd != "" {
			p.Set(nd)
			p.Status = JC.STATE_LOADED
		}
	}
	return true
}

func (p *TickerDataType) FormatContent() string {
	raw := p.Get()

	val, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return raw
	}

	switch p.Format {
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
	return p.OldKey != p.Get() && p.Status == JC.STATE_LOADED
}
