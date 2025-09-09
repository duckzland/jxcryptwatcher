package apps

import (
	"encoding/json"
	"time"

	JC "jxwatcher/core"
	JT "jxwatcher/types"
)

const (
	NO_SNAPSHOT     = -1
	HAVE_SNAPSHOT   = 0
	MinSaveInterval = 30 * time.Minute
)

var AppSnapshotManager *SnapshotManager = &SnapshotManager{}

type SnapshotManager struct {
	lastPanels      string
	lastCryptos     string
	lastTickers     string
	lastExchange    string
	lastTickerCache string

	lastPanelsSaveTime      time.Time
	lastCryptosSaveTime     time.Time
	lastTickersSaveTime     time.Time
	lastExchangeSaveTime    time.Time
	lastTickerCacheSaveTime time.Time
}

func (sm *SnapshotManager) Init() {
	sm.lastPanels = ""
	sm.lastCryptos = ""
	sm.lastTickers = ""
	sm.lastExchange = ""
	sm.lastTickerCache = ""

	sm.lastPanelsSaveTime = time.Time{}
	sm.lastCryptosSaveTime = time.Time{}
	sm.lastTickersSaveTime = time.Time{}
	sm.lastExchangeSaveTime = time.Time{}
	sm.lastTickerCacheSaveTime = time.Time{}

}

func (sm *SnapshotManager) LoadPanels() int {
	raw, ok := JC.LoadFile("snapshots-panels.json")
	if !ok || raw == "" || raw == "null" {
		return NO_SNAPSHOT
	}

	var caches []JT.PanelDataCache
	if err := json.Unmarshal([]byte(raw), &caches); err != nil {
		return NO_SNAPSHOT
	}

	var restored []*JT.PanelDataType
	for _, c := range caches {
		p := &JT.PanelDataType{
			Status: c.Status,
			OldKey: c.OldKey,
		}
		p.Init()
		p.Data.Set(c.Key)
		restored = append(restored, p)
	}

	JT.BP.Init()
	JT.BP.Hydrate(restored)

	sm.lastPanels = raw
	sm.lastPanelsSaveTime = time.Now()

	return HAVE_SNAPSHOT
}

func (sm *SnapshotManager) LoadCryptos() int {
	raw, ok := JC.LoadFile("snapshots-cryptos.json")
	if !ok || raw == "" || raw == "null" {
		return NO_SNAPSHOT
	}

	var cache JT.CryptosMapCache
	if err := json.Unmarshal([]byte(raw), &cache); err != nil {
		return NO_SNAPSHOT
	}

	cm := &JT.CryptosMapType{}
	cm.Hydrate(cache.Data)
	JT.BP.SetMaps(cm)

	sm.lastCryptos = raw
	sm.lastCryptosSaveTime = time.Now()

	return HAVE_SNAPSHOT
}

func (sm *SnapshotManager) LoadTickers() int {
	raw, ok := JC.LoadFile("snapshots-tickers.json")
	JC.Logln("Ticker snapshot loaded", raw)
	if !ok || raw == "" || raw == "null" {
		return NO_SNAPSHOT
	}

	var caches []JT.TickerDataCache
	if err := json.Unmarshal([]byte(raw), &caches); err != nil {
		return NO_SNAPSHOT
	}

	var restored []*JT.TickerDataType
	for _, c := range caches {
		t := &JT.TickerDataType{
			Type:   c.Type,
			Title:  c.Title,
			Format: c.Format,
			Status: c.Status,
			OldKey: c.OldKey,
		}
		t.Init()
		t.Data.Set(c.Key)
		restored = append(restored, t)
	}

	JT.BT.Init()
	JT.BT.Hydrate(restored)

	sm.lastTickers = raw
	sm.lastTickersSaveTime = time.Now()

	return HAVE_SNAPSHOT
}

func (sm *SnapshotManager) LoadExchangeData() int {
	raw, ok := JC.LoadFile("snapshots-exchange.json")
	if !ok || raw == "" || raw == "null" {
		return NO_SNAPSHOT
	}

	var snapshot JT.ExchangeDataCacheSnapshot
	if err := json.Unmarshal([]byte(raw), &snapshot); err != nil {
		return NO_SNAPSHOT
	}

	JT.ExchangeCache.Hydrate(snapshot)
	return HAVE_SNAPSHOT
}

func (sm *SnapshotManager) LoadTickerData() int {
	raw, ok := JC.LoadFile("snapshots-ticker-cache.json")
	if !ok || raw == "" || raw == "null" {
		return NO_SNAPSHOT
	}

	var snapshot JT.TickerDataCacheSnapshot
	if err := json.Unmarshal([]byte(raw), &snapshot); err != nil {
		return NO_SNAPSHOT
	}

	JT.TickerCache.Hydrate(snapshot)
	return HAVE_SNAPSHOT
}

func (sm *SnapshotManager) ShouldSave(domain string) bool {

	// Save energy on mobile and only save when exit or on background
	if JC.IsMobile {
		return false
	}

	now := time.Now()

	switch domain {
	case "panels":
		if sm.lastPanels == "" {
			return true
		}

		if now.Sub(sm.lastPanelsSaveTime) <= MinSaveInterval {
			return false
		}

		data, _ := json.Marshal(JT.BP.Serialize())
		serialized := string(data)

		if serialized != sm.lastPanels {
			sm.lastPanels = serialized
			sm.lastPanelsSaveTime = now
			return true
		}

	case "cryptos":
		if sm.lastCryptos == "" {
			return true
		}

		if now.Sub(sm.lastCryptosSaveTime) <= MinSaveInterval {
			return false
		}

		data, _ := json.Marshal(JT.BP.Maps.Serialize())
		serialized := string(data)

		if serialized != sm.lastCryptos {
			sm.lastCryptos = serialized
			sm.lastCryptosSaveTime = now
			return true
		}

	case "tickers":
		if sm.lastTickers == "" {
			return true
		}

		if now.Sub(sm.lastTickersSaveTime) <= MinSaveInterval {
			return false
		}

		data, _ := json.Marshal(JT.BT.Serialize())
		serialized := string(data)

		if serialized != sm.lastTickers {
			sm.lastTickers = serialized
			sm.lastTickersSaveTime = now
			return true
		}

	case "exchange":
		if sm.lastExchange == "" {
			return true
		}

		if now.Sub(sm.lastExchangeSaveTime) <= MinSaveInterval {
			return false
		}

		data, _ := json.Marshal(JT.ExchangeCache.Serialize())
		serialized := string(data)

		if serialized != sm.lastExchange {
			sm.lastExchange = serialized
			sm.lastExchangeSaveTime = now
			return true
		}

	case "ticker_cache":
		if sm.lastTickerCache == "" {
			return true
		}

		if now.Sub(sm.lastTickerCacheSaveTime) <= MinSaveInterval {
			return false
		}

		data, _ := json.Marshal(JT.TickerCache.Serialize())
		serialized := string(data)

		if serialized != sm.lastTickerCache {
			sm.lastTickerCache = serialized
			sm.lastTickerCacheSaveTime = now
			return true
		}
	}

	return false
}

func (sm *SnapshotManager) SavePanels() {
	if !sm.ShouldSave("panels") {
		return
	}

	JC.MainDebouncer.Call("snapshots_save_panels", 2*time.Second, func() {
		JC.SaveFile("snapshots-panels.json", JT.BP.Serialize())
	})
}

func (sm *SnapshotManager) SaveCryptos() {
	if !sm.ShouldSave("cryptos") {
		return
	}

	JC.MainDebouncer.Call("snapshots_save_cryptos", 2*time.Second, func() {
		JC.SaveFile("snapshots-cryptos.json", JT.BP.Maps.Serialize())
	})
}

func (sm *SnapshotManager) SaveTickers() {
	if !sm.ShouldSave("tickers") {
		return
	}

	JC.MainDebouncer.Call("snapshots_save_tickers", 2*time.Second, func() {
		JC.SaveFile("snapshots-tickers.json", JT.BT.Serialize())
	})
}

func (sm *SnapshotManager) SaveExchangeData() {
	if !sm.ShouldSave("exchange") {
		return
	}

	JC.MainDebouncer.Call("snapshots_save_exchange", 2*time.Second, func() {
		JC.SaveFile("snapshots-exchange.json", JT.ExchangeCache.Serialize())
	})
}

func (sm *SnapshotManager) SaveTickerData() {
	if !sm.ShouldSave("ticker_cache") {
		return
	}

	JC.MainDebouncer.Call("snapshots_save_ticker_cache", 2*time.Second, func() {
		JC.SaveFile("snapshots-ticker-cache.json", JT.TickerCache.Serialize())
	})
}

func (sm *SnapshotManager) ForceSaveAll() {
	if !JT.BP.IsEmpty() {
		JC.SaveFile("snapshots-panels.json", JT.BP.Serialize())
	}

	if JT.BP.Maps != nil && !JT.BP.Maps.IsEmpty() {
		JC.SaveFile("snapshots-cryptos.json", JT.BP.Maps.Serialize())
	}

	if !JT.BT.IsEmpty() {
		JC.SaveFile("snapshots-tickers.json", JT.BT.Serialize())
	}

	if !JT.ExchangeCache.IsEmpty() {
		JC.SaveFile("snapshots-exchange.json", JT.ExchangeCache.Serialize())
	}

	if !JT.TickerCache.IsEmpty() {
		JC.SaveFile("snapshots-ticker-cache.json", JT.TickerCache.Serialize())
	}
}
