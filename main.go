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

	JC.App = app.NewWithID(JC.AppID)

	JC.App.Settings().SetTheme(JA.NewTheme())

	JC.Window = JC.App.NewWindow("JXCrypto Watcher")

	JT.ConfigInit()

	JT.ExchangeCache.Init()

	JT.TickerCache.Init()

	JA.AppActionManager.Init()

	JA.AppStatusManager.Init()

	JA.AppSnapshotManager.Init()

	RegisterActions()

	RegisterFetchers()

	RegisterWorkers()

	RegisterLifecycle()

	JC.Window.SetContent(JA.NewAppLayoutManager())

	JC.Window.Resize(fyne.NewSize(920, 600))

	if JC.IsMobile {
		JC.Window.SetFixedSize(true)
	}

	// Prevent locking when initialized at first install
	JC.MainDebouncer.Call("initializing", 33*time.Millisecond, func() {

		if JA.AppSnapshotManager.LoadCryptos() == JC.NO_SNAPSHOT {
			JT.CryptosInit()
		}

		if JA.AppSnapshotManager.LoadPanels() == JC.NO_SNAPSHOT {
			JT.PanelsInit()
		}

		if JA.AppSnapshotManager.LoadTickers() == JC.NO_SNAPSHOT {
			JT.TickersInit()
		}

		if JA.AppSnapshotManager.LoadExchangeData() == JC.NO_SNAPSHOT {
			JT.ExchangeCache.Reset()
		}

		if JA.AppSnapshotManager.LoadTickerData() == JC.NO_SNAPSHOT {
			JT.TickerCache.Reset()
		}

		fyne.Do(func() {

			JX.Grid = JX.NewTickerGrid()
			JP.Grid = JP.NewPanelGrid(CreatePanel)

			JA.AppStatusManager.InitData()
			JA.AppLayoutManager.SetPage(JP.Grid)
			JA.AppLayoutManager.SetTickers(JC.Tickers)
			JA.AppLayoutManager.Refresh()

			JC.Logln("App is ready: ", JA.AppStatusManager.IsReady())
			JC.Notify("Application is starting...")

			JP.Grid.Refresh()

			if !JA.AppStatusManager.HasError() {

				// Force Refresh
				JT.ExchangeCache.SoftReset()
				JC.WorkerManager.Call("update_rates", JC.CallImmediate)

				// Force Refresh
				JT.TickerCache.SoftReset()
				JC.WorkerManager.Call("update_tickers", JC.CallImmediate)
			}
		})
	})

	JC.Window.ShowAndRun()
}
