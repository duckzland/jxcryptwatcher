package main

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
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

	bg := canvas.NewRectangle(JC.AppBG)
	bg.SetMinSize(fyne.NewSize(920, 600))

	JC.Window.SetContent(fynetooltip.AddWindowToolTipLayer(container.NewStack(
		bg,
		container.NewPadded(
			container.NewBorder(
				// Top action bar
				JW.NewTopBar(
					// Refreshing cryptos.json callback
					func() {
						Cryptos := JT.CryptosType{}
						JT.BP.SetMaps(Cryptos.CreateFile().LoadFile().ConvertToMap())
						JT.BP.InvalidatePanels()
						UpdateDisplay()
					},
					// Refreshing rates from exchange callbck
					func() {
						UpdateRates()
						UpdateDisplay()
					},
					// Saving configuration form callback
					func() {
						JW.NewSettingsForm(func() {
							JT.Config.SaveFile()
						})
					},
					// Open the new panel creation form callback
					func() {
						OpenNewPanelForm()
					},
				),
				nil, nil, nil,
				// The panel grid
				container.NewVScroll(JC.Grid),
			),
		),
	), JC.Window.Canvas()))

	JC.Window.Resize(fyne.NewSize(920, 400))

	// Worker to attempt to refresh the ui for every 3 seconds
	go func() {
		for {
			fyne.Do(func() {
				UpdateDisplay()
			})
			time.Sleep(3 * time.Second)
		}
	}()

	// Worker to update the exchange rates data
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
