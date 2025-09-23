package apps

import (
	"encoding/json"

	JC "jxwatcher/core"
	JT "jxwatcher/types"
)

var SnapshotManager *snapshotManager = &snapshotManager{}

type snapshotManager struct{}

func (sm *snapshotManager) Init() {}

func (sm *snapshotManager) LoadPanels() int {
	raw, ok := JC.LoadFile("snapshots-panels.json")
	if !ok || raw == "" || raw == "null" {
		return JC.NO_SNAPSHOT
	}

	caches := JT.NewPanelDataCache()
	if err := json.Unmarshal([]byte(raw), &caches); err != nil {
		return JC.NO_SNAPSHOT
	}

	var restored []JT.PanelData
	for _, c := range caches {
		p := JT.NewPanelData()

		if !p.GetParent().ValidateKey(c.Key) {
			continue
		}

		p.Init()
		p.SetStatus(c.Status)
		p.SetOldKey(c.OldKey)
		p.Set(p.RefreshKey(c.Key))
		restored = append(restored, p)
	}

	JT.UsePanelMaps().Init()
	JT.UsePanelMaps().Hydrate(restored)

	return JC.HAVE_SNAPSHOT
}

func (sm *snapshotManager) LoadCryptos() int {
	raw, ok := JC.LoadFile("snapshots-cryptos.json")
	if !ok || raw == "" || raw == "null" {
		return JC.NO_SNAPSHOT
	}

	var cache JT.CryptosMapCache
	if err := json.Unmarshal([]byte(raw), &cache); err != nil {
		return JC.NO_SNAPSHOT
	}

	cm := JT.NewCryptosMap()
	cm.Hydrate(cache.Data)
	JT.UsePanelMaps().SetMaps(cm)
	JT.UsePanelMaps().GetOptions()

	return JC.HAVE_SNAPSHOT
}

func (sm *snapshotManager) LoadTickers() int {
	raw, ok := JC.LoadFile("snapshots-tickers.json")
	if !ok || raw == "" || raw == "null" {
		return JC.NO_SNAPSHOT
	}

	caches := JT.NewTickerDataCache()
	if err := json.Unmarshal([]byte(raw), &caches); err != nil {
		return JC.NO_SNAPSHOT
	}

	var restored []JT.TickerData
	for _, c := range caches {
		t := JT.NewTickerData()
		t.Init()
		t.Set(c.Key)
		t.SetType(c.Type)
		t.SetTitle(c.Title)
		t.SetFormat(c.Format)
		t.SetStatus(c.Status)
		t.SetOldKey(c.OldKey)
		restored = append(restored, t)
	}

	JT.UseTickerMaps().Init()
	JT.UseTickerMaps().Hydrate(restored)

	return JC.HAVE_SNAPSHOT
}

func (sm *snapshotManager) LoadExchangeData() int {
	raw, ok := JC.LoadFile("snapshots-exchange.json")
	if !ok || raw == "" || raw == "null" {
		return JC.NO_SNAPSHOT
	}

	snapshot := JT.NewExchangeDataCacheSnapshot()
	if err := json.Unmarshal([]byte(raw), &snapshot); err != nil {
		return JC.NO_SNAPSHOT
	}

	JT.UseExchangeCache().Hydrate(*snapshot)
	return JC.HAVE_SNAPSHOT
}

func (sm *snapshotManager) LoadTickerData() int {
	raw, ok := JC.LoadFile("snapshots-ticker-cache.json")
	if !ok || raw == "" || raw == "null" {
		return JC.NO_SNAPSHOT
	}

	snapshot := JT.NewTickerDataCacheSnapshot()
	if err := json.Unmarshal([]byte(raw), &snapshot); err != nil {
		return JC.NO_SNAPSHOT
	}

	JT.TickerCache.Hydrate(*snapshot)
	return JC.HAVE_SNAPSHOT
}

func (sm *snapshotManager) Save() {
	JC.SaveFile("snapshots-panels.json", JT.UsePanelMaps().Serialize())
	JC.SaveFile("snapshots-cryptos.json", JT.UsePanelMaps().GetMaps().Serialize())
	JC.SaveFile("snapshots-tickers.json", JT.UseTickerMaps().Serialize())
	JC.SaveFile("snapshots-exchange.json", JT.UseExchangeCache().Serialize())
	JC.SaveFile("snapshots-ticker-cache.json", JT.TickerCache.Serialize())
}
