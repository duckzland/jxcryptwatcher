package apps

import (
	"time"

	JC "jxwatcher/core"
	JT "jxwatcher/types"
)

var AppStatusManager *AppStatus = &AppStatus{}

type AppStatus struct {
	ready       bool
	bad_config  bool
	bad_cryptos bool
	no_panels   bool
	timestamp   time.Time
}

func (a *AppStatus) Init() {
	a.ready = false
	a.timestamp = time.Now()
	a.bad_config = false
	a.bad_cryptos = false
	a.no_panels = false
}

func (a *AppStatus) IsReady() bool {
	return a.ready
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
	a.ready = !(JT.BP.Maps == nil)
	a.no_panels = JT.BP.IsEmpty()
	a.bad_config = !JT.Config.IsValid()
	a.bad_cryptos = JT.BP.Maps == nil || JT.BP.Maps.IsEmpty()
	a.timestamp = time.Now()

	// If app is not ready, disable everything
	if !a.ready {
		AppActionManager.DisableAllButton("open_settings")
		return a
	}

	AppActionManager.EnableButton("open_settings")

	if a.bad_cryptos {
		AppActionManager.DisableButton("add_panel")
	} else {
		AppActionManager.EnableButton("add_panel")
	}

	if a.no_panels {
		AppActionManager.DisableButton("toggle_drag")
	} else {
		AppActionManager.EnableButton("toggle_drag")

		if JC.AllowDragging {
			AppActionManager.ChangeButtonState("toggle_drag", "active")
		} else {
			AppActionManager.ChangeButtonState("toggle_drag", "reset")
		}
	}

	if !a.bad_config && !a.bad_cryptos && !a.no_panels {
		AppActionManager.EnableButton("refresh_rates")
	} else {
		AppActionManager.DisableButton("refresh_rates")
	}

	if a.bad_cryptos {
		AppActionManager.ChangeButtonState("refresh_cryptos", "error")
	} else {
		AppActionManager.ChangeButtonState("refresh_cryptos", "reset")
	}

	if a.bad_config {
		AppActionManager.DisableButton("refresh_cryptos")
		AppActionManager.ChangeButtonState("open_settings", "error")
	} else {
		AppActionManager.EnableButton("refresh_cryptos")
		AppActionManager.ChangeButtonState("open_settings", "reset")
	}

	JC.Logf("Application Status: Ready: %v | NoPanels: %v | BadConfig: %v | BadCryptos: %v | Timestamp: %s",
		a.ready, a.no_panels, a.bad_config, a.bad_cryptos, a.timestamp.Format(time.RFC3339))

	return a
}
