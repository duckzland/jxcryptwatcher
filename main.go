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

	JC.RegisterWorkerManager().Init()

	JC.RegisterFetcherManager().Init()

	JC.RegisterDispatcher().Init()

	JC.RegisterDebouncer().Init()

	JC.RegisterCharWidthCache().Init()

	JT.RegisterExchangeCache().Init()

	JT.RegisterTickerCache().Init()

	JA.RegisteLayoutManager().Init()

	JA.RegisterActionManager().Init()

	JA.RegisterStatusManager().Init()

	JA.RegisterSnapshotManager().Init()

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

		if JA.UseSnapshot().LoadCryptos() == JC.NO_SNAPSHOT {
			JT.CryptosLoaderInit()
		}

		if JA.UseSnapshot().LoadPanels() == JC.NO_SNAPSHOT {
			JT.PanelsInit()
		}

		if JA.UseSnapshot().LoadTickers() == JC.NO_SNAPSHOT {
			JT.TickersInit()
		}

		if JA.UseSnapshot().LoadExchangeData() == JC.NO_SNAPSHOT {
			JT.UseExchangeCache().Reset()
		}

		if JA.UseSnapshot().LoadTickerData() == JC.NO_SNAPSHOT {
			JT.UseTickerCache().Reset()
		}

		fyne.Do(func() {

			JX.RegisterTickerGrid()
			JP.RegisterPanelGrid(CreatePanel)

			JA.UseStatus().InitData()
			JA.UseLayout().SetPage(JP.UsePanelGrid())
			JP.UsePanelGrid().Refresh()
			JA.UseLayout().Refresh()

			JC.Logln("App is ready: ", JA.UseStatus().IsReady())

			if !JA.UseStatus().HasError() {

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
