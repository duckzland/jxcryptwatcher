package apps

import (
	"encoding/json"

	JC "jxwatcher/core"
	JT "jxwatcher/types"
)

var AppSnapshotManager *SnapshotManager = &SnapshotManager{}

type SnapshotManager struct{}

func (sm *SnapshotManager) Init() {}

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
		p := &JT.PanelDataType{}
		p.Init()
		p.SetStatus(c.Status)
		p.SetOldKey(c.OldKey)
		p.Set(c.Key)
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
		t := &JT.TickerDataType{}
		t.Init()
		t.Set(c.Key)
		t.SetType(c.Type)
		t.SetTitle(c.Title)
		t.SetFormat(c.Format)
		t.SetStatus(c.Status)
		t.SetOldKey(c.OldKey)
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
	JC.SaveFile("snapshots-panels.json", JT.BP.Serialize())
	JC.SaveFile("snapshots-cryptos.json", JT.BP.Maps.Serialize())
	JC.SaveFile("snapshots-tickers.json", JT.BT.Serialize())
	JC.SaveFile("snapshots-exchange.json", JT.ExchangeCache.Serialize())
	JC.SaveFile("snapshots-ticker-cache.json", JT.TickerCache.Serialize())
}
