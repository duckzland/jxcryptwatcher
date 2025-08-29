package main

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"

	JA "jxwatcher/apps"
	JC "jxwatcher/core"
	JP "jxwatcher/panels"
	JT "jxwatcher/types"
)

func main() {
	JC.InitLogger()

	JT.ExchangeCache.Reset()

	a := app.NewWithID(JC.AppID)

	a.Settings().SetTheme(theme.DarkTheme())

	// Prevent locking when initialized at first install
	JC.MainDebouncer.Call("initializing", 10*time.Millisecond, func() {

		JT.ConfigInit()

		JT.PanelsInit()

		fyne.Do(func() {
			JC.Grid = JP.NewPanelGrid(CreatePanel)
			JA.AppLayoutManager.SetContent(JC.Grid)
			JA.AppLayoutManager.Refresh()
			JC.Grid.Refresh()

			if !JT.BP.IsEmpty() && JT.Config.IsValid() {
				RequestRateUpdate(true)
			}
		})
	})

	JC.Window = a.NewWindow("JXCrypto Watcher")

	JC.Grid = &fyne.Container{}

	JC.AllowDragging = false

	topBar := JA.NewTopBar(
		func() {
			if !JT.Config.IsValid() {
				JC.Notify("Please configure app first")
				return
			}

			ResetCryptosMap()
		},
		func() {
			if JT.BP.IsEmpty() {
				JC.Notify("Please create panel first")
				return
			}

			if !JT.Config.IsValid() {
				JC.Notify("Please configure app first")
				return
			}

			RequestRateUpdate(true)
		},
		func() {
			OpenSettingForm()
		},
		func() {
			if JT.BP.Maps == nil {
				JC.Notify("App not ready yet")
				return
			}

			if JT.BP.Maps.IsEmpty() {
				JC.Notify("No Cryptos Map, please fetch from exchange first")
				return
			}

			OpenNewPanelForm()
		},
		func() {
			if JT.BP.IsEmpty() {
				JC.Notify("Please create panel first")
				return
			}

			if JT.BP.Maps == nil {
				JC.Notify("App not ready yet")
				return
			}

			JC.AllowDragging = !JC.AllowDragging
			JC.Grid.Refresh()

			if JC.AllowDragging {
				JA.AppActionManager.ChangeButtonState("toggle_drag", "in_progress")
			} else {
				JA.AppActionManager.ChangeButtonState("toggle_drag", "reset")
			}
		})

	JC.Window.SetContent(JA.NewAppLayoutManager(&topBar, nil, func() {

		if JT.BP.Maps.IsEmpty() {
			JC.Notify("No Cryptos Map, please fetch from exchange first")
			return
		}

		OpenNewPanelForm()
	}))

	JC.Window.Resize(fyne.NewSize(920, 400))

	StartWorkers()
	StartUpdateRatesWorker()

	JC.Notify("Application is starting...")

	if !JT.Config.IsValid() {
		JC.Notify("Configuration file is invalid.")
	}

	if JC.IsMobile {
		JC.Window.SetFixedSize(true)
	}

	JC.Window.ShowAndRun()
}
