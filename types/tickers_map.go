package types

import (
	"sync"

	JC "jxwatcher/core"
)

var tickerMapsStorage *tickersMapType = &tickersMapType{}

type tickersMapType struct {
	mu   sync.RWMutex
	data []TickerData
}

func (pc *tickersMapType) Init() {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.data = make([]TickerData, 0, 10)
}

func (pc *tickersMapType) SetData(data []TickerData) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	pc.data = data
}

func (pc *tickersMapType) Add(ticker TickerData) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	if pc.data == nil {
		pc.data = []TickerData{}
	}

	ticker.Init()
	ticker.SetStatus(JC.STATE_LOADING)

	pc.data = append(pc.data, ticker)
}

func (pc *tickersMapType) Update(uuid string) bool {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	for _, tdt := range pc.data {
		if tdt.IsID(uuid) {
			return tdt.Update()
		}
	}

	return false
}

func (pc *tickersMapType) GetData() []TickerData {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	dataCopy := make([]TickerData, len(pc.data))
	copy(dataCopy, pc.data)
	return dataCopy
}

func (pc *tickersMapType) GetDataByID(uuid string) TickerData {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	for _, tdt := range pc.data {
		if tdt.IsID(uuid) {
			return tdt
		}
	}
	return nil
}

func (pc *tickersMapType) GetDataByType(tickerType string) []TickerData {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	var nd []TickerData
	for _, tdt := range pc.data {
		if tdt.IsType(tickerType) {
			nd = append(nd, tdt)
		}
	}
	return nd
}

func (pc *tickersMapType) IsEmpty() bool {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return len(pc.data) == 0
}

func (pc *tickersMapType) Reset() {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	for _, tdt := range pc.data {
		tdt.Set(JC.STRING_EMPTY)
		tdt.SetStatus(JC.STATE_LOADING)
	}
}

func (pc *tickersMapType) ChangeStatus(newStatus int, shouldChange func(pdt TickerData) bool) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	for _, pdt := range pc.data {
		if shouldChange != nil && !shouldChange(pdt) {
			continue
		}
		pdt.SetStatus(newStatus)
	}
}

func (pc *tickersMapType) Hydrate(data []TickerData) {
	for _, tkd := range data {
		ex := pc.GetDataByType(tkd.GetType())

		if len(ex) > 0 {
			for _, tkt := range ex {
				tkt.Set(tkd.Get())
				tkt.SetOldKey(tkd.GetOldKey())

				switch tkd.GetStatus() {
				case JC.STATE_ERROR:
					tkt.SetStatus(JC.STATE_FETCHING_NEW)
				default:
					tkt.SetStatus(tkd.GetStatus())
				}
			}
		} else {
			switch tkd.GetType() {
			case TickerTypeMarketCap:
				if UseConfig().CanDoMarketCap() {
					pc.mu.Lock()
					pc.data = append(pc.data, tkd)
					pc.mu.Unlock()
				}
			case TickerTypeCMC100:
				if UseConfig().CanDoCMC100() {
					pc.mu.Lock()
					pc.data = append(pc.data, tkd)
					pc.mu.Unlock()
				}
			case TickerTypeAltcoinIndex:
				if UseConfig().CanDoAltSeason() {
					pc.mu.Lock()
					pc.data = append(pc.data, tkd)
					pc.mu.Unlock()
				}
			case TickerTypeFearGreed:
				if UseConfig().CanDoFearGreed() {
					pc.mu.Lock()
					pc.data = append(pc.data, tkd)
					pc.mu.Unlock()
				}
			case TickerTypeRSI, TickerTypePulse:
				if UseConfig().CanDoRSI() {
					pc.mu.Lock()
					pc.data = append(pc.data, tkd)
					pc.mu.Unlock()
				}
			case TickerTypeETF:
				if UseConfig().CanDoETF() {
					pc.mu.Lock()
					pc.data = append(pc.data, tkd)
					pc.mu.Unlock()
				}
			case TickerTypeDominance:
				if UseConfig().CanDoDominance() {
					pc.mu.Lock()
					pc.data = append(pc.data, tkd)
					pc.mu.Unlock()
				}
			}
		}
	}
}

func (pc *tickersMapType) Serialize() []tickerDataCache {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	var out []tickerDataCache
	for _, t := range pc.data {
		if t.IsStatus(JC.STATE_LOADED) {
			out = append(out, t.Serialize())
		}
	}

	return out
}

func TickersInit() {
	UseTickerMaps().Init()

	if UseConfig().CanDoMarketCap() {
		tdt := NewTickerData()
		tdt.SetTitle("Market Cap")
		tdt.SetType(TickerTypeMarketCap)
		tdt.SetFormat(TickerFormatShortCurrency)
		UseTickerMaps().Add(tdt)
	}

	if UseConfig().CanDoRSI() {
		tdt := NewTickerData()
		tdt.SetTitle("Market Bias")
		tdt.SetType(TickerTypePulse)
		tdt.SetFormat(TickerFormatPulse)
		UseTickerMaps().Add(tdt)
	}

	if UseConfig().CanDoCMC100() {
		tdt := NewTickerData()
		tdt.SetTitle("CMC100")
		tdt.SetType(TickerTypeCMC100)
		tdt.SetFormat(TickerFormatCurrency)
		UseTickerMaps().Add(tdt)
	}

	if UseConfig().CanDoAltSeason() {
		tdt := NewTickerData()
		tdt.SetTitle("Altcoin Index")
		tdt.SetType(TickerTypeAltcoinIndex)
		tdt.SetFormat(TickerFormatPercentage)
		UseTickerMaps().Add(tdt)
	}

	if UseConfig().CanDoFearGreed() {
		tdt := NewTickerData()
		tdt.SetTitle("Fear & Greed")
		tdt.SetType(TickerTypeFearGreed)
		tdt.SetFormat(TickerFormatPercentage)
		UseTickerMaps().Add(tdt)
	}

	if UseConfig().CanDoRSI() {
		tdt := NewTickerData()
		tdt.SetTitle("Crypto RSI")
		tdt.SetType(TickerTypeRSI)
		tdt.SetFormat(TickerFormatNumber)
		UseTickerMaps().Add(tdt)
	}

	if UseConfig().CanDoETF() {
		tdt := NewTickerData()
		tdt.SetTitle("ETF Flow")
		tdt.SetType(TickerTypeETF)
		tdt.SetFormat(TickerFormatShortCurrencyWithSign)
		UseTickerMaps().Add(tdt)
	}

	if UseConfig().CanDoDominance() {
		tdt := NewTickerData()
		tdt.SetTitle("Dominance")
		tdt.SetType(TickerTypeDominance)
		tdt.SetFormat(TickerFormatShortPercentage)
		UseTickerMaps().Add(tdt)
	}
}

func UseTickerMaps() *tickersMapType {
	return tickerMapsStorage
}
