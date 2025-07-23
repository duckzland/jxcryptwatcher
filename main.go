package main

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	fynetooltip "github.com/dweymouth/fyne-tooltip"

	JC "jxwatcher/core"
	JL "jxwatcher/layouts"
	JT "jxwatcher/types"
	JW "jxwatcher/widgets"
)

func main() {

	JT.ExchangeCache.Reset()

	JT.ConfigInit()

	a := app.New()
	a.Settings().SetTheme(theme.DarkTheme())

	// Don't invoke this before app.New(), binding.UntypedList will crash
	JT.PanelsInit()

	JC.Window = a.NewWindow("JXCrypto Watcher")

	JC.Grid = container.New(JL.NewDynamicGridWrapLayout(fyne.NewSize(300, 150)))

	list := JT.BP.Get()
	for range list {
		JC.Grid.Add(JW.NewEmptyPanel())
	}

	JC.NotificationBox = widget.NewLabel("")

	topBg := canvas.NewRectangle(JC.PanelBG)
	topBg.CornerRadius = 4
	topBg.SetMinSize(fyne.NewSize(860, 20))
	topBar := container.New(
		&JL.StretchLayout{Widths: []float32{0.798, 0.004, 0.048, 0.002, 0.048, 0.002, 0.048, 0.002, 0.048}},
		container.NewStack(
			topBg,
			JC.NotificationBox,
		),
		layout.NewSpacer(),

		// Reload cryptos.json
		JW.NewHoverCursorIconButton("", theme.ViewRestoreIcon(), "Refresh ticker data", func() {
			JW.DoActionWithNotification("Fetching new ticker data...", "Finished fetching ticker data", JC.NotificationBox, func() {
				Cryptos := JT.CryptosType{}
				JT.BP.SetMaps(Cryptos.CreateFile().LoadFile().ConvertToMap())
				JT.BP.InvalidatePanels()
				UpdateDisplay()
			})
		}),
		layout.NewSpacer(),

		// Refresh data from exchange
		JW.NewHoverCursorIconButton("", theme.ViewRefreshIcon(), "Update rates from exchange", func() {
			JW.DoActionWithNotification("Fetching exchange rates...", "Panel refreshed with new rates", JC.NotificationBox, func() {
				UpdateRates()
				UpdateDisplay()
			})
		}),
		layout.NewSpacer(),

		// Open settings form
		JW.NewHoverCursorIconButton("", theme.SettingsIcon(), "Open settings", func() {
			JW.NewSettingsForm(func() {
				JT.Config.SaveFile()
			})
		}),
		layout.NewSpacer(),

		// Add new panel
		JW.NewHoverCursorIconButton("", theme.ContentAddIcon(), "Add new panel", func() {
			OpenNewPanelForm()
		}),
	)

	bg := canvas.NewRectangle(JC.AppBG)
	bg.SetMinSize(fyne.NewSize(920, 600))

	JC.Window.SetContent(fynetooltip.AddWindowToolTipLayer(container.NewStack(
		bg,
		container.NewPadded(
			container.NewBorder(
				topBar, nil, nil, nil, container.NewVScroll(JC.Grid),
			),
		),
	), JC.Window.Canvas()))

	JC.Window.Resize(fyne.NewSize(920, 400))

	go func() {
		for {
			fyne.Do(func() {
				UpdateDisplay()
			})
			time.Sleep(3 * time.Second)
		}
	}()

	go func() {
		for {
			JW.DoActionWithNotification("Fetching exchange rate...", "Fetching rates from exchange...", JC.NotificationBox, func() {
				if UpdateRates() {
					fyne.Do(func() {
						UpdateDisplay()
					})
				}
			})

			time.Sleep(time.Duration(JT.Config.Delay) * time.Second)
		}
	}()

	JC.Window.ShowAndRun()

}
