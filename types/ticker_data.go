package types

import (
	"math"
	"strconv"
	"sync"

	"fyne.io/fyne/v2/data/binding"

	JC "jxwatcher/core"
)

const TickerFormatNodecimal = "nodecimal"
const TickerFormatNumber = "number"
const TickerFormatCurrency = "currency"
const TickerFormatShortCurrency = "shortcurrency"
const TickerFormatShortCurrencyWithSign = "shortcurrency_withsign"
const TickerFormatPercentage = "percentage"
const TickerFormatShortPercentage = "shortpercentage"
const TickerFormatPulse = "pulse"
const TickerTypeMarketCap = "market_cap"
const TickerTypePulse = "pulse"
const TickerTypeCMC100 = "cmc100"
const TickerTypeAltcoinIndex = "altcoin_index"
const TickerTypeFearGreed = "feargreed"
const TickerTypeRSI = "rsi"
const TickerTypeRSIOverbought = "rsi_overbought_precentage"
const TickerTypeRSIOversold = "rsi_oversold_percentage"
const TickerTypeRSINeutral = "rsi_neutral_percentage"
const TickerTypeETF = "etf"
const TickerTypeETFBTC = "etf_btc"
const TickerTypeETFETH = "etf_eth"
const TickerTypeDominance = "dominance"
const TickerTypeETCDominance = "etc_dominance"
const TickerTypeOtherDominance = "other_dominance"
const TickerTypeMarketCap24hChange = "market_cap_24_percentage"
const TickerTypeCMC10024hChange = "cmc100_24_percentage"
const TickerTypeCMC10030dChange = "market_cap_30_percentage"

type TickerData interface {
	Init()
	Set(rate string)
	SetType(val string)
	SetTitle(val string)
	SetFormat(val string)
	SetStatus(val int)
	SetID(val string)
	SetOldKey(val string)
	Get() string
	GetType() string
	GetTitle() string
	GetFormat() string
	GetStatus() int
	GetID() string
	GetOldKey() string
	UseData() binding.String
	UseStatus() binding.Int
	HasData() bool
	IsType(val string) bool
	IsTitle(val string) bool
	IsFormat(val string) bool
	IsStatus(val int) bool
	IsID(val string) bool
	IsOldKey(val string) bool
	IsKey(val string) bool
	Insert(rate string)
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
	status   binding.Int
}

func (p *tickerDataType) Init() {
	p.mu.Lock()
	p.data = binding.NewString()
	p.status = binding.NewInt()
	p.id = ""
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

func (p *tickerDataType) SetType(val string) {
	p.mu.Lock()
	p.category = val
	p.mu.Unlock()
}

func (p *tickerDataType) SetTitle(val string) {
	p.mu.Lock()
	p.title = val
	p.mu.Unlock()
}

func (p *tickerDataType) SetFormat(val string) {
	p.mu.Lock()
	p.format = val
	p.mu.Unlock()
}

func (p *tickerDataType) SetStatus(val int) {
	if p.IsStatus(val) {
		return
	}
	p.mu.Lock()
	p.status.Set(val)
	p.mu.Unlock()
}

func (p *tickerDataType) SetID(val string) {
	p.mu.Lock()
	p.id = val
	p.mu.Unlock()
}

func (p *tickerDataType) SetOldKey(val string) {
	p.mu.Lock()
	p.oldKey = val
	p.mu.Unlock()
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

func (p *tickerDataType) GetType() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.category
}

func (p *tickerDataType) GetTitle() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.title
}

func (p *tickerDataType) GetFormat() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.format
}

func (p *tickerDataType) GetStatus() int {
	p.mu.RLock()
	s := p.status
	p.mu.RUnlock()
	v, err := s.Get()
	if err != nil {
		return JC.STATE_ERROR
	}
	return v
}

func (p *tickerDataType) GetID() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.id
}

func (p *tickerDataType) GetOldKey() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.oldKey
}

func (p *tickerDataType) UseData() binding.String {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.data
}

func (p *tickerDataType) UseStatus() binding.Int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.status
}

func (p *tickerDataType) HasData() bool {
	return p.Get() != ""
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
	s := p.status
	p.mu.RUnlock()
	v, _ := s.Get()
	return v == val
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

func (p *tickerDataType) Insert(rate string) {
	p.Set(rate)
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

	ost, _ := p.status.Get()
	nst := ost

	switch ost {
	case JC.STATE_LOADING, JC.STATE_FETCHING_NEW, JC.STATE_ERROR:
		nst = JC.STATE_LOADED
	case JC.STATE_LOADED:
	}

	p.status.Set(nst)
	if npk != opk || ost != nst {
		// JC.Logln("Updating Ticker:", npk, opk, p.status)
		p.oldKey = opk
		p.data.Set(npk)
		return true
	}

	return false
}

func (p *tickerDataType) FormatContent() string {
	raw := p.Get()
	format := p.GetFormat()

	val, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return raw
	}

	switch format {
	case TickerFormatNodecimal:
		return strconv.FormatFloat(val, 'f', 0, 64)

	case TickerFormatNumber:
		return strconv.FormatFloat(val, 'f', 2, 64)

	case TickerFormatCurrency:
		return "$" + JC.FormatNumberWithCommas(val, 2)

	case TickerFormatShortCurrency:
		return JC.FormatShortCurrency(raw)

	case TickerFormatShortCurrencyWithSign:
		sign := "+"
		if val < 0 {
			sign = "-"
		}
		return sign + JC.FormatShortCurrency(strconv.FormatFloat(math.Abs(val), 'f', -1, 64))

	case TickerFormatPercentage:
		return raw + "/100"

	case TickerFormatShortPercentage:
		return strconv.FormatFloat(val, 'f', 1, 64) + "%"

	default:
		return raw
	}
}

func (p *tickerDataType) DidChange() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.IsStatus(JC.STATE_LOADED) || p.oldKey == "" {
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

	return tickerDataCache{
		Type:   p.category,
		Title:  p.title,
		Format: p.format,
		Status: p.GetStatus(),
		Key:    p.Get(),
		OldKey: p.oldKey,
	}
}

func NewTickerDataCache() []tickerDataCache {
	return []tickerDataCache{}
}

func NewTickerData() *tickerDataType {
	return &tickerDataType{}
}
