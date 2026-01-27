package types

import (
	"math"
	"strconv"

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
	UseData() *JC.DataBinding
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
	UpdateStatus() bool
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
	data     *JC.DataBinding
	oldKey   string
	category string
	title    string
	format   string
	id       string
}

func (p *tickerDataType) Init() {
	p.data = JC.NewDataBinding(JC.STRING_EMPTY, JC.STATE_ERROR)
	p.id = JC.STRING_EMPTY
	p.oldKey = JC.STRING_EMPTY
}

func (p *tickerDataType) Set(rate string) {
	if p.data != nil {
		p.oldKey = p.data.GetData()
		p.data.SetData(rate)
	}
}

func (p *tickerDataType) SetType(val string) {
	p.category = val
}

func (p *tickerDataType) SetTitle(val string) {
	p.title = val
}

func (p *tickerDataType) SetFormat(val string) {
	p.format = val
}

func (p *tickerDataType) SetStatus(val int) {
	if p.data != nil && p.data.GetStatus() != val {
		p.data.SetStatus(val)
	}
}

func (p *tickerDataType) SetID(val string) {
	p.id = val
}

func (p *tickerDataType) SetOldKey(val string) {
	p.oldKey = val
}

func (p *tickerDataType) Get() string {
	if p.data == nil {
		return JC.STRING_EMPTY
	}
	return p.data.GetData()
}

func (p *tickerDataType) GetType() string {
	return p.category
}

func (p *tickerDataType) GetTitle() string {
	return p.title
}

func (p *tickerDataType) GetFormat() string {
	return p.format
}

func (p *tickerDataType) GetStatus() int {
	if p.data == nil {
		return JC.STATE_ERROR
	}
	return p.data.GetStatus()
}

func (p *tickerDataType) GetID() string {
	return p.id
}

func (p *tickerDataType) GetOldKey() string {
	return p.oldKey
}

func (p *tickerDataType) UseData() *JC.DataBinding {
	return p.data
}

func (p *tickerDataType) HasData() bool {
	return p.Get() != JC.STRING_EMPTY
}

func (p *tickerDataType) IsType(val string) bool {
	return p.category == val
}

func (p *tickerDataType) IsTitle(val string) bool {
	return p.title == val
}

func (p *tickerDataType) IsFormat(val string) bool {
	return p.format == val
}

func (p *tickerDataType) IsStatus(val int) bool {
	return p.GetStatus() == val
}

func (p *tickerDataType) IsID(val string) bool {
	return p.id == val
}

func (p *tickerDataType) IsOldKey(val string) bool {
	return p.oldKey == val
}

func (p *tickerDataType) IsKey(val string) bool {
	return p.Get() == val
}

func (p *tickerDataType) Insert(rate string) {
	p.Set(rate)
}

func (p *tickerDataType) Update() bool {
	if JC.IsShuttingDown() {
		return false
	}

	if !tickerCacheStorage.Has(p.category) {
		return false
	}

	npk := tickerCacheStorage.Get(p.category)
	if npk == JC.STRING_EMPTY {
		return false
	}

	opk := p.Get()
	ost := p.GetStatus()
	nst := ost

	switch ost {
	case JC.STATE_LOADING, JC.STATE_FETCHING_NEW, JC.STATE_ERROR:
		nst = JC.STATE_LOADED
	}

	p.SetStatus(nst)

	if npk != opk || ost != nst {
		p.oldKey = opk
		p.Set(npk)
		return true
	}

	return false
}

func (p *tickerDataType) UpdateStatus() bool {
	if JC.IsShuttingDown() {
		return false
	}

	ost := p.GetStatus()

	switch ost {
	case JC.STATE_LOADING, JC.STATE_FETCHING_NEW, JC.STATE_ERROR:
		p.SetStatus(JC.STATE_LOADED)
		return true
	case JC.STATE_LOADED:
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
		return JC.STRING_DOLLAR + JC.FormatNumberWithCommas(val, 2)

	case TickerFormatShortCurrency:
		return JC.FormatShortCurrency(raw)

	case TickerFormatShortCurrencyWithSign:
		sign := JC.STRING_PLUS
		if val < 0 {
			sign = JC.STRING_MINUS
		}
		return sign + JC.FormatShortCurrency(strconv.FormatFloat(math.Abs(val), 'f', -1, 64))

	case TickerFormatPercentage:
		return raw + JC.STRING_PERCENTAGE_DIVIDE

	case TickerFormatShortPercentage:
		return strconv.FormatFloat(val, 'f', 1, 64) + JC.STRING_PERCENTAGE

	default:
		return raw
	}
}

func (p *tickerDataType) DidChange() bool {
	if p.oldKey == JC.STRING_EMPTY {
		return false
	}
	if !p.IsStatus(JC.STATE_LOADED) {
		return false
	}
	return p.Get() != p.oldKey
}

func (p *tickerDataType) Serialize() tickerDataCache {
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
