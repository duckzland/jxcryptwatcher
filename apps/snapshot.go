package apps

import (
	"encoding/json"

	JC "jxwatcher/core"
	JT "jxwatcher/types"
)

var AppSnapshotManager *SnapshotManager = &SnapshotManager{}

type SnapshotManager struct{}

func (sm *SnapshotManager) Init() {
	// Nothing to do
}

func (sm *SnapshotManager) LoadPanels() int {
	raw, ok := JC.LoadFile("snapshots-panels.json")
	if !ok || raw == "" || raw == "null" {
		return JC.NO_SNAPSHOT
	}

	var caches []JT.PanelDataCache
	if err := json.Unmarshal([]byte(raw), &caches); err != nil {
		return JC.NO_SNAPSHOT
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

	return JC.HAVE_SNAPSHOT
}

func (sm *SnapshotManager) LoadCryptos() int {
	raw, ok := JC.LoadFile("snapshots-cryptos.json")
	if !ok || raw == "" || raw == "null" {
		return JC.NO_SNAPSHOT
	}

	var cache JT.CryptosMapCache
	if err := json.Unmarshal([]byte(raw), &cache); err != nil {
		return JC.NO_SNAPSHOT
	}

	cm := &JT.CryptosMapType{}
	cm.Hydrate(cache.Data)
	JT.BP.SetMaps(cm)

	return JC.HAVE_SNAPSHOT
}

func (sm *SnapshotManager) LoadTickers() int {
	raw, ok := JC.LoadFile("snapshots-tickers.json")
	if !ok || raw == "" || raw == "null" {
		return JC.NO_SNAPSHOT
	}

	var caches []JT.TickerDataCache
	if err := json.Unmarshal([]byte(raw), &caches); err != nil {
		return JC.NO_SNAPSHOT
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

	return JC.HAVE_SNAPSHOT
}

func (sm *SnapshotManager) LoadExchangeData() int {
	raw, ok := JC.LoadFile("snapshots-exchange.json")
	if !ok || raw == "" || raw == "null" {
		return JC.NO_SNAPSHOT
	}

	var snapshot JT.ExchangeDataCacheSnapshot
	if err := json.Unmarshal([]byte(raw), &snapshot); err != nil {
		return JC.NO_SNAPSHOT
	}

	JT.ExchangeCache.Hydrate(snapshot)
	return JC.HAVE_SNAPSHOT
}

func (sm *SnapshotManager) LoadTickerData() int {
	raw, ok := JC.LoadFile("snapshots-ticker-cache.json")
	if !ok || raw == "" || raw == "null" {
		return JC.NO_SNAPSHOT
	}

	var snapshot JT.TickerDataCacheSnapshot
	if err := json.Unmarshal([]byte(raw), &snapshot); err != nil {
		return JC.NO_SNAPSHOT
	}

	JT.TickerCache.Hydrate(snapshot)
	return JC.HAVE_SNAPSHOT
}

func (sm *SnapshotManager) Save() {
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
