package apps

import (
	"encoding/json"
	"sync"

	JC "jxwatcher/core"
	JT "jxwatcher/types"
)

var snapshotManagerStorage *snapshotManager = nil

type snapshotManager struct {
	mu          sync.RWMutex
	snapshotted bool
}

func (sm *snapshotManager) Init() {
	sm.mu.Lock()
	sm.snapshotted = false
	sm.mu.Unlock()
}

func (sm *snapshotManager) Reset() {
	sm.mu.Lock()
	sm.snapshotted = false
	sm.mu.Unlock()
}

func (sm *snapshotManager) IsSnapshotted() bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return sm.snapshotted
}

func (sm *snapshotManager) LoadPanels() int {
	raw, ok := JC.LoadFileFromStorage("snapshots-panels.json")
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
		p.Set(p.RefreshKey(c.Key))
		p.SetOldKey(c.OldKey)
		p.SetStatus(c.Status)
		restored = append(restored, p)
	}

	JT.PanelsInit()
	JT.UsePanelMaps().Hydrate(restored)

	return JC.HAVE_SNAPSHOT
}

func (sm *snapshotManager) LoadCryptos() int {
	raw, ok := JC.LoadFileFromStorage("snapshots-cryptos.json")
	if !ok || raw == "" || raw == "null" {
		return JC.NO_SNAPSHOT
	}

	cache := JT.NewCryptosMapCache()
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
	raw, ok := JC.LoadFileFromStorage("snapshots-tickers.json")
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

	JT.TickersInit()
	JT.UseTickerMaps().Hydrate(restored)

	return JC.HAVE_SNAPSHOT
}

func (sm *snapshotManager) LoadExchangeData() int {
	raw, ok := JC.LoadFileFromStorage("snapshots-exchange.json")
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
	raw, ok := JC.LoadFileFromStorage("snapshots-ticker-cache.json")
	if !ok || raw == "" || raw == "null" {
		return JC.NO_SNAPSHOT
	}

	snapshot := JT.NewTickerDataCacheSnapshot()
	if err := json.Unmarshal([]byte(raw), &snapshot); err != nil {
		return JC.NO_SNAPSHOT
	}

	JT.UseTickerCache().Hydrate(*snapshot)
	return JC.HAVE_SNAPSHOT
}

func (sm *snapshotManager) Delete() int {
	files := []string{
		"snapshots-panels.json",
		"snapshots-cryptos.json",
		"snapshots-tickers.json",
		"snapshots-exchange.json",
		"snapshots-ticker-cache.json",
	}

	success := true
	for _, file := range files {
		if !JC.EraseFileFromStorage(file) {
			success = false
		}
	}

	if success {
		return JC.SNAPSHOT_DELETED
	}
	return JC.SNAPSHOT_DELETE_FAILED
}

func (sm *snapshotManager) Save() {
	JC.SaveFileToStorage("snapshots-panels.json", JT.UsePanelMaps().Serialize())
	JC.SaveFileToStorage("snapshots-cryptos.json", JT.UsePanelMaps().GetMaps().Serialize())
	JC.SaveFileToStorage("snapshots-tickers.json", JT.UseTickerMaps().Serialize())
	JC.SaveFileToStorage("snapshots-exchange.json", JT.UseExchangeCache().Serialize())
	JC.SaveFileToStorage("snapshots-ticker-cache.json", JT.UseTickerCache().Serialize())

	sm.mu.Lock()
	sm.snapshotted = true
	sm.mu.Unlock()
}

func RegisterSnapshotManager() *snapshotManager {
	if snapshotManagerStorage == nil {
		snapshotManagerStorage = &snapshotManager{}
	}
	return snapshotManagerStorage
}

func UseSnapshot() *snapshotManager {
	return snapshotManagerStorage
}
