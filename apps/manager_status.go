package apps

import (
	"sync"
	"time"

	JC "jxwatcher/core"
	JT "jxwatcher/types"
)

var statusManagerStorage *statusManager = nil

type statusManager struct {
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

func (a *statusManager) Init() {
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

func (a *statusManager) IsReady() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.ready
}

func (a *statusManager) IsPaused() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.paused
}

func (a *statusManager) IsDraggable() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.allow_dragging
}

func (a *statusManager) IsFetchingCryptos() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.fetching_cryptos
}

func (a *statusManager) IsFetchingRates() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.fetching_rates
}

func (a *statusManager) IsFetchingTickers() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.fetching_tickers
}

func (a *statusManager) IsOverlayShown() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.overlay_shown
}

func (a *statusManager) IsValidConfig() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.valid_config
}

func (a *statusManager) IsValidCrypto() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.valid_cryptos
}

func (a *statusManager) IsGoodNetworkStatus() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.network_status
}

func (a *statusManager) IsDirty() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.lastChange.After(a.lastRefresh)
}

func (a *statusManager) AppReady() *statusManager {
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

func (a *statusManager) Pause() *statusManager {
	a.mu.Lock()
	a.paused = true
	a.lastChange = time.Now()
	a.mu.Unlock()

	a.DebounceRefresh()
	return a
}

func (a *statusManager) Resume() *statusManager {
	a.mu.Lock()
	a.paused = false
	a.lastChange = time.Now()
	a.mu.Unlock()

	a.DebounceRefresh()
	return a
}

func (a *statusManager) StartFetchingCryptos() *statusManager {
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

func (a *statusManager) EndFetchingCryptos() *statusManager {
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

func (a *statusManager) StartFetchingRates() *statusManager {
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

func (a *statusManager) EndFetchingRates() *statusManager {
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

func (a *statusManager) StartFetchingTickers() *statusManager {
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

func (a *statusManager) EndFetchingTickers() *statusManager {
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

func (a *statusManager) AllowDragging() *statusManager {
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

func (a *statusManager) DisallowDragging() *statusManager {
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

func (a *statusManager) SetOverlayShownStatus(status bool) *statusManager {
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

func (a *statusManager) SetConfigStatus(status bool) *statusManager {
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

func (a *statusManager) SetCryptoStatus(status bool) *statusManager {
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

func (a *statusManager) SetNetworkStatus(status bool) *statusManager {
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

func (a *statusManager) InitData() *statusManager {
	ready := JT.UsePanelMaps().HasMaps()
	noPanels := JT.UsePanelMaps().IsEmpty()
	badConfig := !JT.UseConfig().IsValid()
	badCryptos := !JT.UsePanelMaps().HasMaps() || JT.UsePanelMaps().GetMaps().IsEmpty()
	panelsCount := JT.UsePanelMaps().TotalData()
	badTickers := !JT.UseConfig().IsValidTickers()

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

func (a *statusManager) DetectData() *statusManager {
	newReady := JT.UsePanelMaps().HasMaps()
	newNoPanels := JT.UsePanelMaps().IsEmpty()
	newBadConfig := !JT.UseConfig().IsValid()
	newBadCryptos := !JT.UsePanelMaps().HasMaps() || JT.UsePanelMaps().GetMaps().IsEmpty()
	newPanelsCount := JT.UsePanelMaps().TotalData()
	newBadTickers := !JT.UseConfig().IsValidTickers()

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

func (a *statusManager) HasError() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.bad_config || a.bad_cryptos
}

func (a *statusManager) ValidConfig() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return !a.bad_config
}

func (a *statusManager) ValidCryptos() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return !a.bad_cryptos
}

func (a *statusManager) ValidPanels() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return !a.no_panels
}

func (a *statusManager) ValidTickers() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return !a.bad_tickers
}

func (a *statusManager) DebounceRefresh() *statusManager {
	JC.UseDebouncer().Call("refreshing_status", 8*time.Millisecond, func() {
		a.Refresh()
	})
	return a
}

func (a *statusManager) Refresh() *statusManager {
	a.mu.Lock()
	shouldUpdate := a.lastChange.After(a.lastRefresh)
	a.lastRefresh = time.Now()
	a.mu.Unlock()

	if shouldUpdate {
		UseLayoutManager().Refresh()
		UseActionManager().Refresh()
	}

	if !a.IsReady() || a.HasError() {
		JC.Logf("Application Status: Ready: %v | NoPanels: %v | BadConfig: %v | BadCryptos: %v | BadTickers: %v | LastChange: %d | LastRefresh: %d",
			a.IsReady(), a.ValidPanels(), !a.ValidConfig(), !a.ValidCryptos(), a.bad_tickers, a.lastChange.UnixNano(), a.lastRefresh.UnixNano())
	}

	return a
}

func RegisterStatusManager() *statusManager {
	if statusManagerStorage == nil {
		JC.InitOnce(func() {
			statusManagerStorage = &statusManager{}
		})
	}
	return statusManagerStorage
}

func UseStatusManager() *statusManager {
	return statusManagerStorage
}
