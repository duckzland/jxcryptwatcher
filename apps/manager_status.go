package apps

import (
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"

	JC "jxwatcher/core"
	JT "jxwatcher/types"
)

var statusManagerStorage *statusManager

type statusManager struct {
	ready            atomic.Bool
	paused           atomic.Bool
	bad_config       atomic.Bool
	bad_cryptos      atomic.Bool
	bad_tickers      atomic.Bool
	no_panels        atomic.Bool
	allow_dragging   atomic.Bool
	fetching_cryptos atomic.Bool
	fetching_rates   atomic.Bool
	fetching_tickers atomic.Bool
	network_status   atomic.Bool
	overlay_shown    atomic.Bool
	show_tickers     atomic.Bool
	panels_count     atomic.Int64
	lastChange       atomic.Int64
	lastRefresh      atomic.Int64
}

func (a *statusManager) Init() {
	a.ready.Store(false)
	a.paused.Store(false)
	a.bad_config.Store(false)
	a.bad_cryptos.Store(false)
	a.bad_tickers.Store(false)
	a.no_panels.Store(false)
	a.allow_dragging.Store(false)
	a.fetching_cryptos.Store(false)
	a.fetching_rates.Store(false)
	a.fetching_tickers.Store(false)
	a.network_status.Store(true)
	a.overlay_shown.Store(false)
	a.show_tickers.Store(true)

	a.panels_count.Store(0)

	now := time.Now().UnixNano()
	a.lastChange.Store(now)
	a.lastRefresh.Store(now)
}

func (a *statusManager) HasError() bool {
	return a.bad_config.Load() || a.bad_cryptos.Load()
}

func (a *statusManager) IsReady() bool {
	return a.ready.Load()
}

func (a *statusManager) IsPaused() bool {
	return a.paused.Load()
}

func (a *statusManager) IsDraggable() bool {
	return a.allow_dragging.Load()
}

func (a *statusManager) IsFetchingCryptos() bool {
	return a.fetching_cryptos.Load()
}

func (a *statusManager) IsFetchingRates() bool {
	return a.fetching_rates.Load()
}

func (a *statusManager) IsFetchingTickers() bool {
	return a.fetching_tickers.Load()
}

func (a *statusManager) IsOverlayShown() bool {
	return a.overlay_shown.Load()
}

func (a *statusManager) IsValidConfig() bool {
	return !a.bad_config.Load()
}

func (a *statusManager) IsValidCrypto() bool {
	return !a.bad_cryptos.Load()
}

func (a *statusManager) IsValidPanels() bool {
	return !a.no_panels.Load()
}

func (a *statusManager) IsValidTickers() bool {
	return !a.bad_tickers.Load()
}

func (a *statusManager) IsGoodNetworkStatus() bool {
	return a.network_status.Load()
}

func (a *statusManager) IsTickerShown() bool {
	return a.show_tickers.Load()
}

func (a *statusManager) IsDirty() bool {
	return a.lastChange.Load() > a.lastRefresh.Load()
}

func (a *statusManager) touch() {
	a.lastChange.Store(time.Now().UnixNano())
}

func (a *statusManager) AppReady() *statusManager {
	if !a.ready.Load() {
		a.ready.Store(true)
		a.touch()
		a.Refresh()
	}
	return a
}

func (a *statusManager) Pause() *statusManager {
	a.paused.Store(true)
	a.touch()
	a.Refresh()
	return a
}

func (a *statusManager) Resume() *statusManager {
	a.paused.Store(false)
	a.touch()
	a.Refresh()
	return a
}

func (a *statusManager) StartFetchingCryptos() *statusManager {
	if !a.fetching_cryptos.Load() {
		a.fetching_cryptos.Store(true)
		a.touch()
		a.Refresh()
	}
	return a
}

func (a *statusManager) EndFetchingCryptos() *statusManager {
	if a.fetching_cryptos.Load() {
		a.fetching_cryptos.Store(false)
		a.touch()
		a.DetectData()
		a.Refresh()
	}
	return a
}

func (a *statusManager) StartFetchingRates() *statusManager {
	if !a.fetching_rates.Load() {
		a.fetching_rates.Store(true)
		a.touch()
		a.Refresh()
	}
	return a
}

func (a *statusManager) EndFetchingRates() *statusManager {
	if a.fetching_rates.Load() {
		a.fetching_rates.Store(false)
		a.touch()
		a.Refresh()
	}
	return a
}

func (a *statusManager) StartFetchingTickers() *statusManager {
	if !a.fetching_tickers.Load() {
		a.fetching_tickers.Store(true)
		a.touch()
		a.Refresh()
	}
	return a
}

func (a *statusManager) EndFetchingTickers() *statusManager {
	if a.fetching_tickers.Load() {
		a.fetching_tickers.Store(false)
		a.touch()
		a.Refresh()
	}
	return a
}

func (a *statusManager) AllowDragging() *statusManager {
	if !a.allow_dragging.Load() {
		a.allow_dragging.Store(true)
		a.touch()
		a.Refresh()
	}
	return a
}

func (a *statusManager) DisallowDragging() *statusManager {
	if a.allow_dragging.Load() {
		a.allow_dragging.Store(false)
		a.touch()
		a.Refresh()
	}
	return a
}

func (a *statusManager) ToggleTickers() *statusManager {
	a.show_tickers.Store(!a.show_tickers.Load())
	a.touch()
	a.Refresh()
	return a
}

func (a *statusManager) HideTickers() *statusManager {
	a.show_tickers.Store(false)
	a.touch()
	a.Refresh()
	return a
}

func (a *statusManager) SetOverlayShownStatus(status bool) *statusManager {
	if a.overlay_shown.Load() != status {
		a.overlay_shown.Store(status)
		a.touch()
		a.Refresh()
	}
	return a
}

func (a *statusManager) SetConfigStatus(status bool) *statusManager {
	if a.IsValidConfig() != status {
		a.bad_config.Store(!status)
		a.touch()
		a.Refresh()
	}
	return a
}

func (a *statusManager) SetCryptoStatus(status bool) *statusManager {
	if a.IsValidCrypto() != status {
		a.bad_cryptos.Store(!status)
		a.touch()
		a.Refresh()
	}
	return a
}

func (a *statusManager) SetNetworkStatus(status bool) *statusManager {
	if a.network_status.Load() != status {
		a.network_status.Store(status)
		a.touch()
		a.Refresh()
	}
	return a
}

func (a *statusManager) PanelsCount() int {
	return int(a.panels_count.Load())
}

func (a *statusManager) InitData() *statusManager {
	ready := JT.UsePanelMaps().HasMaps()
	noPanels := JT.UsePanelMaps().IsEmpty()
	badConfig := !JT.UseConfig().IsValid()
	badCryptos := !JT.UsePanelMaps().HasMaps() || JT.UsePanelMaps().GetMaps().IsEmpty()
	panelsCount := JT.UsePanelMaps().TotalData()
	badTickers := !JT.UseConfig().IsValidTickers()

	a.ready.Store(ready)
	a.no_panels.Store(noPanels)
	a.bad_config.Store(badConfig)
	a.bad_cryptos.Store(badCryptos)
	a.bad_tickers.Store(badTickers)
	a.panels_count.Store(int64(panelsCount))
	a.touch()

	return a
}

func (a *statusManager) DetectData() *statusManager {
	newReady := JT.UsePanelMaps().HasMaps()
	newNoPanels := JT.UsePanelMaps().IsEmpty()
	newBadConfig := !JT.UseConfig().IsValid()
	newBadCryptos := !JT.UsePanelMaps().HasMaps() || JT.UsePanelMaps().GetMaps().IsEmpty()
	newPanelsCount := JT.UsePanelMaps().TotalData()
	newBadTickers := !JT.UseConfig().IsValidTickers()

	changed := newReady != a.ready.Load() ||
		newNoPanels != a.no_panels.Load() ||
		newBadConfig != a.bad_config.Load() ||
		newBadCryptos != a.bad_cryptos.Load() ||
		newBadTickers != a.bad_tickers.Load() ||
		int64(newPanelsCount) != a.panels_count.Load()

	if changed {
		a.ready.Store(newReady)
		a.no_panels.Store(newNoPanels)
		a.bad_config.Store(newBadConfig)
		a.bad_cryptos.Store(newBadCryptos)
		a.bad_tickers.Store(newBadTickers)
		a.panels_count.Store(int64(newPanelsCount))
		a.touch()
		a.Refresh()
	}

	return a
}

func (a *statusManager) Refresh() *statusManager {
	lastChange := a.lastChange.Load()
	lastRefresh := a.lastRefresh.Load()

	if lastChange <= lastRefresh {
		return a
	}

	now := time.Now().UnixNano()
	a.lastRefresh.Store(now)

	fyne.Do(func() {
		UseLayout().UpdateState()
	})

	JC.UseDebouncer().Call("refreshing_main_layout", 60*time.Millisecond, func() {
		fyne.Do(func() {
			UseAction().Refresh()
		})
	})

	if !a.IsReady() || a.HasError() {
		JC.Logf("Application Status: Ready: %v | NoPanels: %v|%d | BadConfig: %v | BadCryptos: %v | BadTickers: %v | LastChange: %d | LastRefresh: %d", a.IsReady(), a.IsValidPanels(), a.PanelsCount(), !a.IsValidConfig(), !a.IsValidCrypto(), a.bad_tickers.Load(), a.lastChange.Load(), a.lastRefresh.Load())
	}

	return a
}

func RegisterStatusManager() *statusManager {
	if statusManagerStorage == nil {
		statusManagerStorage = &statusManager{}
	}
	return statusManagerStorage
}

func UseStatus() *statusManager {
	return statusManagerStorage
}
