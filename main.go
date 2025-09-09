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

	a := app.NewWithID(JC.AppID)

	a.Settings().SetTheme(JA.NewTheme())

	JC.Window = a.NewWindow("JXCrypto Watcher")

	JT.ExchangeCache.Init()

	JT.TickerCache.Init()

	JT.ConfigInit()

	JA.AppActionManager.Init()

	JA.AppStatusManager.Init()

	JA.AppSnapshotManager.Init()

	RegisterActions()
	RegisterFetchers()
	RegisterWorkers()

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

		// Do saving as this isnt fired frequently
		if JT.BP.Maps != nil && !JT.BP.Maps.IsEmpty() {
			JA.AppSnapshotManager.SaveCryptos()
		}

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
				JC.WorkerManager.Call("update_rates", JC.CallImmediate)

				// Force Refresh
				JT.TickerCache.SoftReset()
				JC.WorkerManager.Call("update_tickers", JC.CallImmediate)
			}
		})
	})

	JC.Notify("Application is starting...")

	// Hook into lifecycle events
	if lc := a.Lifecycle(); lc != nil {

		var snapshotSaved bool = false

		lc.SetOnEnteredForeground(func() {
			JC.Logln("App entered foreground")

			snapshotSaved = false

			if !JA.AppStatusManager.IsReady() {
				JC.Logln("Refused to fetch data as app is not ready yet")
				return
			}

			if !JA.AppStatusManager.HasError() && JC.IsMobile {
				// Force Refresh
				JT.ExchangeCache.SoftReset()
				JC.WorkerManager.Call("update_rates", JC.CallImmediate)

				// Force Refresh
				JT.TickerCache.SoftReset()
				JC.WorkerManager.Call("update_tickers", JC.CallImmediate)
			}
		})
		lc.SetOnExitedForeground(func() {
			JC.Logln("App exited foreground — snapshot time!")

			if !JA.AppStatusManager.IsReady() {
				JC.Logln("Refused to take snapshot as app is not ready yet")
				return
			}

			if !snapshotSaved && JC.IsMobile {
				JA.AppSnapshotManager.ForceSaveAll()
				snapshotSaved = true
			}
		})
		lc.SetOnStopped(func() {
			JC.Logln("App stopped")

			if !JA.AppStatusManager.IsReady() {
				JC.Logln("Refused to take snapshot as app is not ready yet")
				return
			}

			if !snapshotSaved {
				JA.AppSnapshotManager.ForceSaveAll()
				snapshotSaved = true
			}
		})
	}

	JC.Window.ShowAndRun()
}
