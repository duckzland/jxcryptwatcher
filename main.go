package main

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	JA "jxwatcher/apps"
	JC "jxwatcher/core"
	JP "jxwatcher/panels"
	JT "jxwatcher/types"
)

func main() {
	JC.InitLogger()

	JT.ExchangeCache.Reset()

	a := app.NewWithID(JC.AppID)

	a.Settings().SetTheme(JA.NewTheme())

	JA.AppActionManager.Init()

	// Prevent locking when initialized at first install
	JC.MainDebouncer.Call("initializing", 33*time.Millisecond, func() {

		JT.ConfigInit()

		JT.PanelsInit()

		fyne.Do(func() {

			JC.Grid = JP.NewPanelGrid(CreatePanel)

			JA.AppStatusManager.DetectData()
			JA.AppLayoutManager.SetContent(JC.Grid)
			JA.AppLayoutManager.Refresh()

			JC.Logln("App is ready: ", JA.AppStatusManager.IsReady())

			JC.Grid.Refresh()

			if !JA.AppStatusManager.HasError() {
				RequestRateUpdate(true)
			}
		})
	})

	JC.Window = a.NewWindow("JXCrypto Watcher")

	JC.Grid = &fyne.Container{}

	JC.AllowDragging = false

	RegisterActions()

	topBar := JA.NewTopBar()

	JC.Window.SetContent(JA.NewAppLayoutManager(&topBar, nil))

	JC.Window.Resize(fyne.NewSize(920, 400))

	StartWorkers()
	StartUpdateRatesWorker()

	JC.Notify("Application is starting...")

	if JC.IsMobile {
		JC.Window.SetFixedSize(true)
	}

	JC.Window.ShowAndRun()
}
