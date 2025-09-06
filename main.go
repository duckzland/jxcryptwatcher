package main

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	JA "jxwatcher/apps"
	JC "jxwatcher/core"
	JP "jxwatcher/panels"
	JX "jxwatcher/tickers"
	JT "jxwatcher/types"
)

func main() {
	JC.InitLogger()

	JT.ExchangeCache.Reset()

	a := app.NewWithID(JC.AppID)

	a.Settings().SetTheme(JA.NewTheme())

	JA.AppActionManager.Init()

	JA.AppStatusManager.Init()

	// Prevent locking when initialized at first install
	JC.MainDebouncer.Call("initializing", 33*time.Millisecond, func() {

		JT.ConfigInit()

		JT.PanelsInit()

		JT.TickersInit()

		fyne.Do(func() {

			JC.Tickers = JX.NewTickerGrid()
			JP.Grid = JP.NewPanelGrid(CreatePanel)

			JA.AppStatusManager.DetectData()
			JA.AppLayoutManager.SetPage(JP.Grid)
			JA.AppLayoutManager.SetTickers(JC.Tickers)
			JA.AppLayoutManager.Refresh()

			JC.Logln("App is ready: ", JA.AppStatusManager.IsReady())

			JP.Grid.Refresh()

			if !JA.AppStatusManager.HasError() {

				// Force Refresh
				JT.ExchangeCache.SoftReset()
				RequestRateUpdate(true)

				// Force Refresh
				JT.TickerCache.SoftReset()
				RequestTickersUpdate()
			}
		})
	})

	JC.Window = a.NewWindow("JXCrypto Watcher")

	JP.Grid = &JP.PanelGridContainer{}

	RegisterActions()

	topBar := JA.NewTopBar()

	JC.Window.SetContent(JA.NewAppLayoutManager(&topBar))

	JC.Window.Resize(fyne.NewSize(920, 600))

	StartWorkers()
	StartUpdateRatesWorker()
	StartUpdateTickersWorker()

	JC.Notify("Application is starting...")

	if JC.IsMobile {
		JC.Window.SetFixedSize(true)
	}

	JC.Window.ShowAndRun()
}
