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
	pc.data = []TickerData{}
}

func (pc *tickersMapType) Set(data []TickerData) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	for _, tdt := range data {
		tdt.Init()
		tdt.SetStatus(JC.STATE_LOADING)
	}
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

	tdt := pc.getDataUnsafe(uuid)
	if tdt != nil {
		return tdt.Update()
	}
	return false
}

func (pc *tickersMapType) Get() []TickerData {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	dataCopy := make([]TickerData, len(pc.data))
	copy(dataCopy, pc.data)
	return dataCopy
}

func (pc *tickersMapType) GetData(uuid string) TickerData {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return pc.getDataUnsafe(uuid)
}

func (pc *tickersMapType) getDataUnsafe(uuid string) TickerData {
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
		tdt.Set("")
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
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.data = data
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
		tdt.SetType("market_cap")
		tdt.SetFormat("shortcurrency")
		UseTickerMaps().Add(tdt)
	}

	if UseConfig().CanDoCMC100() {
		tdt := NewTickerData()
		tdt.SetTitle("CMC100")
		tdt.SetType("cmc100")
		tdt.SetFormat("currency")
		UseTickerMaps().Add(tdt)
	}

	if UseConfig().CanDoAltSeason() {
		tdt := NewTickerData()
		tdt.SetTitle("Altcoin Index")
		tdt.SetType("altcoin_index")
		tdt.SetFormat("percentage")
		UseTickerMaps().Add(tdt)
	}

	if UseConfig().CanDoFearGreed() {
		tdt := NewTickerData()
		tdt.SetTitle("Fear & Greed")
		tdt.SetType("feargreed")
		tdt.SetFormat("percentage")
		UseTickerMaps().Add(tdt)
	}
}

func UseTickerMaps() *tickersMapType {
	return tickerMapsStorage
}
