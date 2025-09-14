package apps

import (
	"sync"
	"time"

	JC "jxwatcher/core"
	JT "jxwatcher/types"
)

var AppStatusManager = &AppStatus{}

type AppStatus struct {
	mu               sync.RWMutex
	ready            bool
	paused           bool
	bad_config       bool
	bad_cryptos      bool
	bad_tickers      bool
	no_panels        bool
	allow_dragging   bool
	fetching_cryptos bool
	fetching_rates   bool
	fetching_tickers bool
	is_dirty         bool
	panels_count     int
	valid_config     bool
	valid_cryptos    bool
	network_status   bool
	lastChange       time.Time
	lastRefresh      time.Time
}

func (a *AppStatus) Init() {
	a.mu.Lock()
	a.ready = false
	a.paused = false
	a.bad_config = false
	a.bad_cryptos = false
	a.bad_tickers = false
	a.no_panels = false
	a.allow_dragging = false
	a.is_dirty = false
	a.panels_count = 0
	a.valid_config = true
	a.valid_cryptos = true
	a.network_status = true
	a.lastChange = time.Now()
	a.mu.Unlock()
}

func (a *AppStatus) IsReady() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.ready
}

func (a *AppStatus) IsPaused() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.paused
}

func (a *AppStatus) IsDraggable() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.allow_dragging
}

func (a *AppStatus) IsFetchingCryptos() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.fetching_cryptos
}

func (a *AppStatus) IsFetchingRates() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.fetching_rates
}

func (a *AppStatus) IsFetchingTickers() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.fetching_tickers
}

func (a *AppStatus) IsValidConfig() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.valid_config
}

func (a *AppStatus) IsValidCrypto() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.valid_cryptos
}

func (a *AppStatus) IsGoodNetworkStatus() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.network_status
}

func (a *AppStatus) IsDirty() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.lastChange.After(a.lastRefresh)
}

func (a *AppStatus) AppReady() *AppStatus {
	a.mu.Lock()
	changed := !a.ready
	if changed {
		a.ready = true
		a.lastChange = time.Now()
	}
	a.mu.Unlock()

	if changed {
		a.DebounceRefresh()
	}
	return a
}

func (a *AppStatus) PauseApp() *AppStatus {
	a.mu.Lock()
	a.paused = true
	a.lastChange = time.Now()
	a.mu.Unlock()

	a.DebounceRefresh()
	return a
}

func (a *AppStatus) ContinueApp() *AppStatus {
	a.mu.Lock()
	a.paused = false
	a.lastChange = time.Now()
	a.mu.Unlock()

	a.DebounceRefresh()
	return a
}

func (a *AppStatus) StartFetchingCryptos() *AppStatus {
	a.mu.Lock()
	changed := !a.fetching_cryptos
	if changed {
		a.fetching_cryptos = true
		a.lastChange = time.Now()
	}
	a.mu.Unlock()

	if changed {
		a.DebounceRefresh()
	}
	return a
}

func (a *AppStatus) EndFetchingCryptos() *AppStatus {
	a.mu.Lock()
	changed := a.fetching_cryptos
	if changed {
		a.fetching_cryptos = false
		a.lastChange = time.Now()
	}
	a.mu.Unlock()

	if changed {
		a.DetectData()
		a.DebounceRefresh()
	}
	return a
}

func (a *AppStatus) StartFetchingRates() *AppStatus {
	a.mu.Lock()
	changed := !a.fetching_rates
	if changed {
		a.fetching_rates = true
		a.lastChange = time.Now()
	}
	a.mu.Unlock()

	if changed {
		a.DebounceRefresh()
	}
	return a
}

func (a *AppStatus) EndFetchingRates() *AppStatus {
	a.mu.Lock()
	changed := a.fetching_rates
	if changed {
		a.fetching_rates = false
		a.lastChange = time.Now()
	}
	a.mu.Unlock()

	if changed {
		a.DebounceRefresh()
	}
	return a
}

func (a *AppStatus) StartFetchingTickers() *AppStatus {
	a.mu.Lock()
	changed := !a.fetching_tickers
	if changed {
		a.fetching_tickers = true
		a.lastChange = time.Now()
	}
	a.mu.Unlock()

	if changed {
		a.DebounceRefresh()
	}
	return a
}

func (a *AppStatus) EndFetchingTickers() *AppStatus {
	a.mu.Lock()
	changed := a.fetching_tickers
	if changed {
		a.fetching_tickers = false
		a.lastChange = time.Now()
	}
	a.mu.Unlock()

	if changed {
		a.DebounceRefresh()
	}
	return a
}

func (a *AppStatus) AllowDragging() *AppStatus {
	a.mu.Lock()
	changed := !a.allow_dragging
	if changed {
		a.allow_dragging = true
		a.lastChange = time.Now()
	}
	a.mu.Unlock()

	if changed {
		a.DebounceRefresh()
	}
	return a
}

func (a *AppStatus) DisallowDragging() *AppStatus {
	a.mu.Lock()
	changed := a.allow_dragging
	if changed {
		a.allow_dragging = false
		a.lastChange = time.Now()
	}
	a.mu.Unlock()

	if changed {
		a.DebounceRefresh()
	}
	return a
}

func (a *AppStatus) SetConfigStatus(status bool) *AppStatus {
	a.mu.Lock()
	changed := a.valid_config != status
	if changed {
		a.valid_config = status
		a.lastChange = time.Now()
	}
	a.mu.Unlock()

	if changed {
		a.DebounceRefresh()
	}
	return a
}

func (a *AppStatus) SetCryptoStatus(status bool) *AppStatus {
	a.mu.Lock()
	changed := a.valid_cryptos != status
	if changed {
		a.valid_cryptos = status
		a.lastChange = time.Now()
	}
	a.mu.Unlock()

	if changed {
		a.DebounceRefresh()
	}
	return a
}

func (a *AppStatus) SetNetworkStatus(status bool) *AppStatus {
	a.mu.Lock()
	changed := a.network_status != status
	if changed {
		a.network_status = status
		a.lastChange = time.Now()
	}
	a.mu.Unlock()

	if changed {
		a.DebounceRefresh()
	}
	return a
}

func (a *AppStatus) InitData() *AppStatus {
	ready := JT.BP.HasMaps()
	noPanels := JT.BP.IsEmpty()
	badConfig := !JT.Config.IsValid()
	badCryptos := !JT.BP.HasMaps() || JT.BP.GetMaps().IsEmpty()
	panelsCount := JT.BP.TotalData()
	badTickers := !JT.Config.IsValidTickers()

	a.mu.Lock()
	a.ready = ready
	a.no_panels = noPanels
	a.bad_config = badConfig
	a.bad_cryptos = badCryptos
	a.panels_count = panelsCount
	a.bad_tickers = badTickers
	a.lastChange = time.Now()
	a.mu.Unlock()

	return a
}

func (a *AppStatus) DetectData() *AppStatus {
	newReady := JT.BP.HasMaps()
	newNoPanels := JT.BP.IsEmpty()
	newBadConfig := !JT.Config.IsValid()
	newBadCryptos := !JT.BP.HasMaps() || JT.BP.GetMaps().IsEmpty()
	newPanelsCount := JT.BP.TotalData()
	newBadTickers := !JT.Config.IsValidTickers()

	a.mu.Lock()
	changed := a.ready != newReady ||
		a.no_panels != newNoPanels ||
		a.bad_config != newBadConfig ||
		a.bad_cryptos != newBadCryptos ||
		a.bad_tickers != newBadTickers ||
		a.panels_count != newPanelsCount

	if changed {
		a.ready = newReady
		a.no_panels = newNoPanels
		a.bad_config = newBadConfig
		a.bad_cryptos = newBadCryptos
		a.bad_tickers = newBadTickers
		a.panels_count = newPanelsCount
		a.lastChange = time.Now()
	}
	a.mu.Unlock()

	if changed {
		a.DebounceRefresh()
	}
	return a
}

func (a *AppStatus) HasError() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.bad_config || a.bad_cryptos
}

func (a *AppStatus) ValidConfig() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return !a.bad_config
}

func (a *AppStatus) ValidCryptos() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return !a.bad_cryptos
}

func (a *AppStatus) ValidPanels() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return !a.no_panels
}

func (a *AppStatus) ValidTickers() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return !a.bad_tickers
}

func (a *AppStatus) DebounceRefresh() *AppStatus {
	JC.MainDebouncer.Call("refreshing_status", 8*time.Millisecond, func() {
		a.Refresh()
	})
	return a
}

func (a *AppStatus) Refresh() *AppStatus {
	a.mu.Lock()
	shouldUpdate := a.lastChange.After(a.lastRefresh)
	a.lastRefresh = time.Now()
	a.mu.Unlock()

	if shouldUpdate {
		AppLayoutManager.Refresh()
		AppActionManager.Refresh()
	}

	if !a.IsReady() || a.HasError() {
		JC.Logf("Application Status: Ready: %v | NoPanels: %v | BadConfig: %v | BadCryptos: %v | BadTickers: %v | LastChange: %d | LastRefresh: %d",
			a.IsReady(), a.ValidPanels(), !a.ValidConfig(), !a.ValidCryptos(), a.bad_tickers, a.lastChange.UnixNano(), a.lastRefresh.UnixNano())
	}

	return a
}
