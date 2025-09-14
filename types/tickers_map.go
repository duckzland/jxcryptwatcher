package types

import (
	"sync"

	JC "jxwatcher/core"
)

var BT TickersMapType

type TickersMapType struct {
	mu   sync.RWMutex
	data []*TickerDataType
}

func (pc *TickersMapType) Init() {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.data = []*TickerDataType{}
}

func (pc *TickersMapType) Set(data []*TickerDataType) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	for _, tdt := range data {
		tdt.Init()
		tdt.SetStatus(JC.STATE_LOADING)
	}
	pc.data = data
}

func (pc *TickersMapType) Add(ticker *TickerDataType) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	if pc.data == nil {
		pc.data = []*TickerDataType{}
	}

	ticker.Init()
	ticker.SetStatus(JC.STATE_LOADING)

	pc.data = append(pc.data, ticker)
}

func (pc *TickersMapType) Update(uuid string) bool {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	tdt := pc.getDataUnsafe(uuid)
	if tdt != nil {
		return tdt.Update()
	}
	return false
}

func (pc *TickersMapType) Get() []*TickerDataType {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	dataCopy := make([]*TickerDataType, len(pc.data))
	copy(dataCopy, pc.data)
	return dataCopy
}

func (pc *TickersMapType) GetData(uuid string) *TickerDataType {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return pc.getDataUnsafe(uuid)
}

func (pc *TickersMapType) getDataUnsafe(uuid string) *TickerDataType {
	for _, tdt := range pc.data {
		if tdt.IsID(uuid) {
			return tdt
		}
	}
	return nil
}

func (pc *TickersMapType) GetDataByType(tickerType string) []*TickerDataType {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	var nd []*TickerDataType
	for _, tdt := range pc.data {
		if tdt.IsType(tickerType) {
			nd = append(nd, tdt)
		}
	}
	return nd
}

func (pc *TickersMapType) IsEmpty() bool {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return len(pc.data) == 0
}

func (pc *TickersMapType) Reset() {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	for _, tdt := range pc.data {
		tdt.Set("")
		tdt.SetStatus(JC.STATE_LOADING)
	}
}

func (pc *TickersMapType) ChangeStatus(newStatus int, shouldChange func(pdt *TickerDataType) bool) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	for _, pdt := range pc.data {
		if shouldChange != nil && !shouldChange(pdt) {
			continue
		}
		pdt.SetStatus(newStatus)
	}
}

func (pc *TickersMapType) Hydrate(data []*TickerDataType) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.data = data
}

func (pc *TickersMapType) Serialize() []TickerDataCache {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	var out []TickerDataCache
	for _, t := range pc.data {
		if t.IsStatus(JC.STATE_LOADED) {
			out = append(out, t.Serialize())
		}
	}
	return out
}

func TickersInit() {
	BT.Init()

	if Config.CanDoMarketCap() {
		tdt := &TickerDataType{}
		tdt.SetTitle("Market Cap")
		tdt.SetType("market_cap")
		tdt.SetFormat("shortcurrency")
		BT.Add(tdt)
	}

	if Config.CanDoCMC100() {
		tdt := &TickerDataType{}
		tdt.SetTitle("CMC100")
		tdt.SetType("cmc100")
		tdt.SetFormat("currency")
		BT.Add(tdt)
	}

	if Config.CanDoAltSeason() {
		tdt := &TickerDataType{}
		tdt.SetTitle("Altcoin Index")
		tdt.SetType("altcoin_index")
		tdt.SetFormat("percentage")
		BT.Add(tdt)
	}

	if Config.CanDoFearGreed() {
		tdt := &TickerDataType{}
		tdt.SetTitle("Fear & Greed")
		tdt.SetType("feargreed")
		tdt.SetFormat("percentage")
		BT.Add(tdt)
	}
}
