package types

import (
	JC "jxwatcher/core"
	"sync"
)

var BT TickersMapType

type TickersMapType struct {
	mu   sync.RWMutex
	Data []*TickerDataType
}

func (pc *TickersMapType) Init() {
	pc.Data = []*TickerDataType{}
}

func (pc *TickersMapType) Add(ticker *TickerDataType) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	if pc.Data == nil {
		pc.Data = []*TickerDataType{}
	}

	ticker.Init()
	ticker.Status = JC.STATE_LOADING

	pc.Data = append(pc.Data, ticker)
}

func (pc *TickersMapType) Update(uuid string) bool {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	tdt := pc.GetData(uuid)

	if tdt != nil {
		return tdt.Update()
	}

	return false
}

func (pc *TickersMapType) Get() []*TickerDataType {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	dataCopy := make([]*TickerDataType, len(pc.Data))
	copy(dataCopy, pc.Data)
	return dataCopy
}

func (pc *TickersMapType) GetData(uuid string) *TickerDataType {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	for i, tdt := range pc.Data {
		if tdt.ID == uuid {
			return pc.Data[i]
		}
	}

	return nil
}

func (pc *TickersMapType) GetDataByType(ticker_type string) []*TickerDataType {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	nd := []*TickerDataType{}

	for i, tdt := range pc.Data {
		if tdt.Type == ticker_type {
			nd = append(nd, pc.Data[i])
		}
	}

	return nd
}

func (pc *TickersMapType) IsEmpty() bool {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	return len(pc.Data) == 0
}

func (pc *TickersMapType) Reset() {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	for i, tdt := range pc.Data {
		tdt.Set("")
		tdt.Status = JC.STATE_LOADING

		pc.Data[i] = tdt
	}
}

func TickersInit() {
	BT.Init()

	if Config.CanDoMarketCap() {
		BT.Add(&TickerDataType{
			Title:  "Market Cap",
			Type:   "market_cap",
			Format: "shortcurrency",
			Status: JC.STATE_LOADING,
		})
	}

	if Config.CanDoCMC100() {
		BT.Add(&TickerDataType{
			Title:  "CMC100",
			Type:   "cmc100",
			Format: "currency",
			Status: JC.STATE_LOADING,
		})
	}

	if Config.CanDoAltSeason() {
		BT.Add(&TickerDataType{
			Title:  "Altcoin Index",
			Type:   "altcoin_index",
			Format: "percentage",
			Status: JC.STATE_LOADING,
		})
	}

	if Config.CanDoFearGreed() {
		BT.Add(&TickerDataType{
			Title:  "Fear & Greed",
			Type:   "feargreed",
			Format: "percentage",
			Status: JC.STATE_LOADING,
		})
	}
}
