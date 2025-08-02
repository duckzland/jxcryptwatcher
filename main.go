package main

import (
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

	JT.ConfigInit()

	JT.PanelsInit()

	JC.Window = a.NewWindow("JXCrypto Watcher")

	JC.Grid = JP.NewPanelGrid(CreatePanel)

	topBar := JA.NewTopBar(ResetCryptosMap, func() { RefreshRates() }, OpenSettingForm, OpenNewPanelForm)

	JC.Window.SetContent(JA.NewAppLayout(&topBar, JC.Grid))

	JC.Window.Resize(fyne.NewSize(920, 400))

	StartWorkers()
	StartUpdateDisplayWorker()
	StartUpdateRatesWorker()

	JC.Notify("Starting Application...")

	if !JT.Config.IsValid() {
		JC.Notify("Bad configuration file")
	}

	JC.Window.ShowAndRun()
}
