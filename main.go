package main

import (
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	JA "jxwatcher/apps"
	JC "jxwatcher/core"
	JP "jxwatcher/panels"
	JX "jxwatcher/tickers"
	JT "jxwatcher/types"
)

var initOnce sync.Once

func main() {
	JC.InitLogger()

	JC.Logln("App is booting...")

	JC.InitOnce(func() {
		JC.App = app.NewWithID(JC.AppID)
		JC.App.Settings().SetTheme(JA.NewTheme())
		JC.Window = JC.App.NewWindow("JXCrypto Watcher")
	})

	JT.RegisterExchangeCache().Init()

	JT.RegisterTickerCache().Init()

	JA.RegisterActionManager().Init()

	JA.RegisterStatusManager().Init()

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
	JC.UseDebouncer().Call("initializing", 1*time.Millisecond, func() {

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
			JT.UseExchangeCache().Reset()
		}

		if JA.SnapshotManager.LoadTickerData() == JC.NO_SNAPSHOT {
			JT.UseTickerCache().Reset()
		}

		fyne.Do(func() {

			JX.RegisterTickerGrid()
			JP.RegisterPanelGrid(CreatePanel)

			JA.UseStatusManager().InitData()
			JA.UseLayoutManager().SetPage(JP.UsePanelGrid())
			JP.UsePanelGrid().Refresh()
			JA.UseLayoutManager().Refresh()

			JC.Logln("App is ready: ", JA.UseStatusManager().IsReady())

			if !JA.UseStatusManager().HasError() {

				// Force Refresh
				JT.UseExchangeCache().SoftReset()
				JC.UseWorker().Call("update_rates", JC.CallImmediate)

				// Force Refresh
				JT.UseTickerCache().SoftReset()
				JC.UseWorker().Call("update_tickers", JC.CallImmediate)
			}
		})
	})

	JC.Window.ShowAndRun()
}
