package types

import (
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
	ticker.Status = 0

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

func (pc *TickersMapType) IsEmpty() bool {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	return len(pc.Data) == 0
}

func TickersInit() {
	BT.Init()

	if Config.CanDoMetrics() {
		BT.Add(&TickerDataType{
			Title:  "Market Cap",
			Type:   "market_cap",
			Format: "shortcurrency",
		})
	}

	if Config.CanDoCMC100() {
		BT.Add(&TickerDataType{
			Title:  "CMC100",
			Type:   "cmc100",
			Format: "currency",
		})
	}

	if Config.CanDoMetrics() {
		BT.Add(&TickerDataType{
			Title:  "Altcoin Index",
			Type:   "altcoin_index",
			Format: "percentage",
		})
	}

	if Config.CanDoFearGreed() {
		BT.Add(&TickerDataType{
			Title:  "Fear & Greed",
			Type:   "feargreed",
			Format: "percentage",
		})
	}
}
