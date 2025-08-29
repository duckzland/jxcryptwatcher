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
	timestamp        time.Time
}

func (a *AppStatus) Init() {
	a.ready = false
	a.timestamp = time.Now()
	a.bad_config = false
	a.bad_cryptos = false
	a.no_panels = false
	a.allow_dragging = false
	a.is_dirty = false
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

func (a *AppStatus) StartFetchingCryptos() *AppStatus {
	a.fetching_cryptos = true
	a.is_dirty = true
	return a
}

func (a *AppStatus) StartFetchingRates() *AppStatus {
	a.fetching_rates = true
	a.is_dirty = true
	return a
}

func (a *AppStatus) EndFetchingCryptos() *AppStatus {
	a.fetching_cryptos = false
	a.is_dirty = true
	return a
}

func (a *AppStatus) EndFetchingRates() *AppStatus {
	a.fetching_rates = false
	a.is_dirty = true
	return a
}

func (a *AppStatus) AllowDragging() *AppStatus {
	a.allow_dragging = true
	a.is_dirty = true
	return a
}

func (a *AppStatus) DisallowDragging() *AppStatus {
	a.allow_dragging = false
	a.is_dirty = true
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

func (a *AppStatus) Refresh() *AppStatus {
	a.is_dirty = false

	newReady := JT.BP.Maps != nil
	newNoPanels := JT.BP.IsEmpty()
	newBadConfig := !JT.Config.IsValid()
	newBadCryptos := JT.BP.Maps == nil || JT.BP.Maps.IsEmpty()
	newTimestamp := time.Now()

	if a.ready != newReady {
		a.ready = newReady
		a.is_dirty = true
	}
	if a.no_panels != newNoPanels {
		a.no_panels = newNoPanels
		a.is_dirty = true
	}
	if a.bad_config != newBadConfig {
		a.bad_config = newBadConfig
		a.is_dirty = true
	}
	if a.bad_cryptos != newBadCryptos {
		a.bad_cryptos = newBadCryptos
		a.is_dirty = true
	}
	if !a.timestamp.Equal(newTimestamp) {
		a.timestamp = newTimestamp
		a.is_dirty = true
	}

	// If app is not ready, disable everything
	if !a.ready {
		AppActionManager.DisableAllButton("open_settings")
		return a
	}

	// Attempt to refresh main layout to change content
	if a.is_dirty {
		AppLayoutManager.Refresh()
		AppActionManager.Refresh()
	}

	JC.Logf("Application Status: Ready: %v | NoPanels: %v | BadConfig: %v | BadCryptos: %v | Timestamp: %s",
		a.ready, a.no_panels, a.bad_config, a.bad_cryptos, a.timestamp.Format(time.RFC3339))

	return a
}
