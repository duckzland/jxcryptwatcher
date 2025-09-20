package apps

import (
	"sync"
	"time"

	JC "jxwatcher/core"
	JT "jxwatcher/types"
)

var AppStatus = &appStatus{}

type appStatus struct {
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
	overlay_shown    bool
	lastChange       time.Time
	lastRefresh      time.Time
}

func (a *appStatus) Init() {
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

func (a *appStatus) IsReady() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.ready
}

func (a *appStatus) IsPaused() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.paused
}

func (a *appStatus) IsDraggable() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.allow_dragging
}

func (a *appStatus) IsFetchingCryptos() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.fetching_cryptos
}

func (a *appStatus) IsFetchingRates() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.fetching_rates
}

func (a *appStatus) IsFetchingTickers() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.fetching_tickers
}

func (a *appStatus) IsOverlayShown() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.overlay_shown
}

func (a *appStatus) IsValidConfig() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.valid_config
}

func (a *appStatus) IsValidCrypto() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.valid_cryptos
}

func (a *appStatus) IsGoodNetworkStatus() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.network_status
}

func (a *appStatus) IsDirty() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.lastChange.After(a.lastRefresh)
}

func (a *appStatus) AppReady() *appStatus {
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

func (a *appStatus) PauseApp() *appStatus {
	a.mu.Lock()
	a.paused = true
	a.lastChange = time.Now()
	a.mu.Unlock()

	a.DebounceRefresh()
	return a
}

func (a *appStatus) ContinueApp() *appStatus {
	a.mu.Lock()
	a.paused = false
	a.lastChange = time.Now()
	a.mu.Unlock()

	a.DebounceRefresh()
	return a
}

func (a *appStatus) StartFetchingCryptos() *appStatus {
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

func (a *appStatus) EndFetchingCryptos() *appStatus {
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

func (a *appStatus) StartFetchingRates() *appStatus {
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

func (a *appStatus) EndFetchingRates() *appStatus {
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

func (a *appStatus) StartFetchingTickers() *appStatus {
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

func (a *appStatus) EndFetchingTickers() *appStatus {
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

func (a *appStatus) AllowDragging() *appStatus {
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

func (a *appStatus) DisallowDragging() *appStatus {
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

func (a *appStatus) SetOverlayShownStatus(status bool) *appStatus {
	a.mu.Lock()
	changed := a.overlay_shown != status
	if changed {
		a.overlay_shown = status
		a.lastChange = time.Now()
	}
	a.mu.Unlock()

	if changed {
		a.DebounceRefresh()
	}
	return a
}

func (a *appStatus) SetConfigStatus(status bool) *appStatus {
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

func (a *appStatus) SetCryptoStatus(status bool) *appStatus {
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

func (a *appStatus) SetNetworkStatus(status bool) *appStatus {
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

func (a *appStatus) InitData() *appStatus {
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

func (a *appStatus) DetectData() *appStatus {
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

func (a *appStatus) HasError() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.bad_config || a.bad_cryptos
}

func (a *appStatus) ValidConfig() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return !a.bad_config
}

func (a *appStatus) ValidCryptos() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return !a.bad_cryptos
}

func (a *appStatus) ValidPanels() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return !a.no_panels
}

func (a *appStatus) ValidTickers() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return !a.bad_tickers
}

func (a *appStatus) DebounceRefresh() *appStatus {
	JC.MainDebouncer.Call("refreshing_status", 8*time.Millisecond, func() {
		a.Refresh()
	})
	return a
}

func (a *appStatus) Refresh() *appStatus {
	a.mu.Lock()
	shouldUpdate := a.lastChange.After(a.lastRefresh)
	a.lastRefresh = time.Now()
	a.mu.Unlock()

	if shouldUpdate {
		AppLayout.Refresh()
		AppActions.Refresh()
	}

	if !a.IsReady() || a.HasError() {
		JC.Logf("Application Status: Ready: %v | NoPanels: %v | BadConfig: %v | BadCryptos: %v | BadTickers: %v | LastChange: %d | LastRefresh: %d",
			a.IsReady(), a.ValidPanels(), !a.ValidConfig(), !a.ValidCryptos(), a.bad_tickers, a.lastChange.UnixNano(), a.lastRefresh.UnixNano())
	}

	return a
}
