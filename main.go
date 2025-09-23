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

	JC.Logln("App is booting...")

	JC.App = app.NewWithID(JC.AppID)

	JC.App.Settings().SetTheme(JA.NewTheme())

	JC.Window = JC.App.NewWindow("JXCrypto Watcher")

	JT.ExchangeCache.Init()

	JT.TickerCache.Init()

	JA.ActionManager.Init()

	JA.StatusManager.Init()

	JA.SnapshotManager.Init()

	RegisterCache()

	RegisterActions()

	RegisterFetchers()

	RegisterWorkers()

	RegisterLifecycle()

	RegisterDispatcher()

	JC.Window.SetContent(JA.NewAppLayout())

	JC.Window.Resize(fyne.NewSize(920, 600))

	if JC.IsMobile {
		JC.Window.SetFixedSize(true)
	}

	JC.Notify("Application is starting...")

	// Prevent locking when initialized at first install
	JC.MainDebouncer.Call("initializing", 1*time.Millisecond, func() {

		JT.ConfigInit()

		if JA.SnapshotManager.LoadCryptos() == JC.NO_SNAPSHOT {
			JT.CryptosLoaderInit()
		}

		if JA.SnapshotManager.LoadPanels() == JC.NO_SNAPSHOT {
			JT.PanelsInit()
		}

		if JA.SnapshotManager.LoadTickers() == JC.NO_SNAPSHOT {
			JT.TickersInit()
		}

		if JA.SnapshotManager.LoadExchangeData() == JC.NO_SNAPSHOT {
			JT.ExchangeCache.Reset()
		}

		if JA.SnapshotManager.LoadTickerData() == JC.NO_SNAPSHOT {
			JT.TickerCache.Reset()
		}

		fyne.Do(func() {

			JX.Grid = JX.NewTickerGrid()
			JP.Grid = JP.NewPanelGrid(CreatePanel)

			JA.StatusManager.InitData()
			JA.LayoutManager.SetPage(JP.Grid)
			JA.LayoutManager.SetTickers(JC.Tickers)
			JP.Grid.Refresh()
			JA.LayoutManager.Refresh()

			JC.Logln("App is ready: ", JA.StatusManager.IsReady())

			if !JA.StatusManager.HasError() {

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
