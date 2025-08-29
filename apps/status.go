package apps

import (
	"time"

	JC "jxwatcher/core"
	JT "jxwatcher/types"
)

var AppStatusManager *AppStatus = &AppStatus{}

type AppStatus struct {
	ready            bool
	bad_config       bool
	bad_cryptos      bool
	no_panels        bool
	allow_dragging   bool
	fetching_cryptos bool
	fetching_rates   bool
	is_dirty         bool
	panels_count     int
	lastChange       time.Time
	lastRefresh      time.Time
}

func (a *AppStatus) Init() {
	a.ready = false
	a.bad_config = false
	a.bad_cryptos = false
	a.no_panels = false
	a.allow_dragging = false
	a.is_dirty = false
	a.panels_count = 0
	a.lastChange = time.Now()
}

func (a *AppStatus) IsReady() bool {
	return a.ready
}

func (a *AppStatus) IsDraggable() bool {
	return a.allow_dragging
}

func (a *AppStatus) IsFetchingCryptos() bool {
	return a.fetching_cryptos
}

func (a *AppStatus) IsFetchingRates() bool {
	return a.fetching_rates
}

func (a *AppStatus) IsDirty() bool {
	return a.lastChange.After(a.lastRefresh)
}

func (a *AppStatus) AppReady() *AppStatus {
	a.ready = true
	a.lastChange = time.Now()
	a.DebounceRefresh()
	return a
}

func (a *AppStatus) StartFetchingCryptos() *AppStatus {
	a.fetching_cryptos = true
	a.lastChange = time.Now()
	a.DebounceRefresh()
	return a
}

func (a *AppStatus) StartFetchingRates() *AppStatus {
	a.fetching_rates = true
	a.lastChange = time.Now()
	a.DebounceRefresh()
	return a
}

func (a *AppStatus) EndFetchingCryptos() *AppStatus {
	a.fetching_cryptos = false
	a.lastChange = time.Now()
	a.DetectData()
	return a
}

func (a *AppStatus) EndFetchingRates() *AppStatus {
	a.fetching_rates = false
	a.lastChange = time.Now()
	a.DebounceRefresh()
	return a
}

func (a *AppStatus) AllowDragging() *AppStatus {
	a.allow_dragging = true
	a.lastChange = time.Now()
	a.DebounceRefresh()
	return a
}

func (a *AppStatus) DisallowDragging() *AppStatus {
	a.allow_dragging = false
	a.lastChange = time.Now()
	a.DebounceRefresh()
	return a
}

func (a *AppStatus) DetectData() *AppStatus {
	// Capture new state
	newReady := JT.BP.Maps != nil
	newNoPanels := JT.BP.IsEmpty()
	newBadConfig := !JT.Config.IsValid()
	newBadCryptos := JT.BP.Maps == nil || JT.BP.Maps.IsEmpty()
	newPanelsCount := JT.BP.TotalData()

	// Detect changes
	if a.ready != newReady ||
		a.no_panels != newNoPanels ||
		a.bad_config != newBadConfig ||
		a.bad_cryptos != newBadCryptos ||
		a.panels_count != newPanelsCount {

		// Apply changes
		a.ready = newReady
		a.no_panels = newNoPanels
		a.bad_config = newBadConfig
		a.bad_cryptos = newBadCryptos
		a.lastChange = time.Now()
		a.DebounceRefresh()
	}

	return a
}

func (a *AppStatus) HasError() bool {
	return a.bad_config || a.bad_cryptos
}

func (a *AppStatus) ValidConfig() bool {
	return !a.bad_config
}

func (a *AppStatus) ValidCryptos() bool {
	return !a.bad_cryptos
}

func (a *AppStatus) ValidPanels() bool {
	return !a.no_panels
}

func (a *AppStatus) DebounceRefresh() *AppStatus {
	JC.MainDebouncer.Call("refreshing_status", 33*time.Millisecond, func() {
		a.Refresh()
	})

	return a
}

func (a *AppStatus) Refresh() *AppStatus {

	// Attempt to refresh main layout to change content
	if a.IsDirty() {
		AppLayoutManager.Refresh()
		AppActionManager.Refresh()
	}

	a.lastRefresh = time.Now()

	JC.Logf("Application Status: Ready: %v | NoPanels: %v | BadConfig: %v | BadCryptos: %v | LastChange: %s | LastRefresh: %s",
		a.ready, a.no_panels, a.bad_config, a.bad_cryptos, a.lastChange.UnixNano(), a.lastRefresh.UnixNano())

	return a
}
