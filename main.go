package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	JA "jxwatcher/apps"
	JC "jxwatcher/core"
	JP "jxwatcher/panels"
	JT "jxwatcher/types"
)

func main() {
	JC.InitLogger()
	JT.ExchangeCache.Reset()
	JT.ConfigInit()

	a := app.New()
	a.Settings().SetTheme(theme.DarkTheme())

	JT.PanelsInit()
	JC.Window = a.NewWindow("JXCrypto Watcher")

	JC.Grid = JP.NewPanelGrid(CreatePanel)

	JC.NotificationBox = widget.NewLabel("")
	bg := canvas.NewRectangle(JC.AppBG)
	bg.SetMinSize(fyne.NewSize(920, 600))

	topBar := JA.NewTopBar(
		ResetCryptosMap,
		RefreshRates,
		OpenSettingForm,
		OpenNewPanelForm,
	)

	JC.Window.SetContent(JA.NewAppLayout(bg, &topBar, JC.Grid))

	JC.Window.Resize(fyne.NewSize(920, 400))

	StartWorkers()
	StartUpdateDisplayWorker()
	StartUpdateRatesWorker()

	JC.Window.ShowAndRun()
}
