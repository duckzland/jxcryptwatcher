package apps

import (
	"encoding/json"
	"runtime"
	"sync"
	"time"

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
	defer runtime.GC()

	JC.PrintPerfStats("Loading snapshot for panels", time.Now())

	snapshot := JT.NewPanelDataCache()

	if !sm.load("snapshots-panels", &snapshot) {
		return JC.NO_SNAPSHOT
	}

	var restored []JT.PanelData
	for _, c := range snapshot {
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
	defer runtime.GC()

	JC.PrintPerfStats("Loading snapshot for cryptos", time.Now())

	snapshot := JT.NewCryptosMapCache()

	if !sm.load("snapshots-cryptos", &snapshot) {
		return JC.NO_SNAPSHOT
	}

	cm := JT.NewCryptosMap()
	cm.Hydrate(*snapshot)

	JT.UsePanelMaps().SetMaps(cm)
	JT.UsePanelMaps().GetOptions()

	return JC.HAVE_SNAPSHOT
}

func (sm *snapshotManager) LoadTickers() int {
	defer runtime.GC()

	JC.PrintPerfStats("Loading snapshot for tickers", time.Now())

	snapshot := JT.NewTickerDataCache()

	if !sm.load("snapshots-tickers", &snapshot) {
		return JC.NO_SNAPSHOT
	}

	var restored []JT.TickerData
	for _, c := range snapshot {
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
	defer runtime.GC()

	JC.PrintPerfStats("Loading snapshot for exchange data", time.Now())

	snapshot := JT.NewExchangeDataCacheSnapshot()

	if !sm.load("snapshots-exchange", &snapshot) {
		return JC.NO_SNAPSHOT
	}

	JT.UseExchangeCache().Hydrate(*snapshot)
	return JC.HAVE_SNAPSHOT
}

func (sm *snapshotManager) LoadTickerData() int {
	defer runtime.GC()

	JC.PrintPerfStats("Loading snapshot for ticker data", time.Now())

	snapshot := JT.NewTickerDataCacheSnapshot()

	if !sm.load("snapshots-ticker-cache", &snapshot) {
		return JC.NO_SNAPSHOT
	}

	JT.UseTickerCache().Hydrate(*snapshot)
	return JC.HAVE_SNAPSHOT
}

func (sm *snapshotManager) Delete() int {
	files := []string{
		"snapshots-panels.gob",
		"snapshots-cryptos.gob",
		"snapshots-tickers.gob",
		"snapshots-exchange.gob",
		"snapshots-ticker-cache.gob",
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

	JC.SaveGobToStorage("snapshots-exchange.gob", JT.UseExchangeCache().Serialize())
	JC.SaveGobToStorage("snapshots-ticker-cache.gob", JT.UseTickerCache().Serialize())
	JC.SaveGobToStorage("snapshots-panels.gob", JT.UsePanelMaps().Serialize())
	JC.SaveGobToStorage("snapshots-tickers.gob", JT.UseTickerMaps().Serialize())
	JC.SaveGobToStorage("snapshots-cryptos.gob", JT.UsePanelMaps().GetMaps().Serialize())

	sm.mu.Lock()
	sm.snapshotted = true
	sm.mu.Unlock()
}

func (sm *snapshotManager) load(filename string, snapshot any) bool {

	// Migrating from json to gob
	if JT.UseConfig().IsVersionLessThan("1.8.0") {
		content, ok := JC.LoadFileFromStorage(filename + ".json")
		if !ok || json.Unmarshal([]byte(content), snapshot) != nil {
			return false
		}
		_ = JC.EraseFileFromStorage(filename + ".json")

		return true
	}

	return JC.LoadGobFromStorage(filename+".gob", snapshot)
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
