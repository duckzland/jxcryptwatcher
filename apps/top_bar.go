package apps

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"

	JC "jxwatcher/core"
	JL "jxwatcher/layouts"
	JW "jxwatcher/widgets"
)

func NewTopBar(
	onCryptosRefresh func(),
	onRatesRefresh func(),
	onSettingSave func(),
	onAddNewPanel func(),
) fyne.CanvasObject {

	topBg := canvas.NewRectangle(JC.PanelBG)
	topBg.CornerRadius = 4
	topBg.SetMinSize(fyne.NewSize(860, 20))

	return container.New(
		&JL.StretchLayout{
			Widths: []float32{0.798, 0.004, 0.048, 0.002, 0.048, 0.002, 0.048, 0.002, 0.048},
		},

		container.NewStack(
			topBg,
			JC.NotificationBox,
		),

		layout.NewSpacer(),

		// Refresh ticker data
		JW.NewHoverCursorIconButton("", theme.ViewRestoreIcon(), "Refresh ticker data", func() {
			JW.DoActionWithNotification("Fetching new ticker data...", "Finished fetching ticker data", JC.NotificationBox, func() {
				if onCryptosRefresh != nil {
					onCryptosRefresh()
				}
			})
		}),

		layout.NewSpacer(),

		// Refresh exchange rates
		JW.NewHoverCursorIconButton("", theme.ViewRefreshIcon(), "Update rates from exchange", func() {
			JW.DoActionWithNotification("Fetching exchange rates...", "Panel refreshed with new rates", JC.NotificationBox, func() {
				if onRatesRefresh != nil {
					onRatesRefresh()
				}
			})
		}),

		layout.NewSpacer(),

		// Open settings
		JW.NewHoverCursorIconButton("", theme.SettingsIcon(), "Open settings", func() {
			if onSettingSave != nil {
				onSettingSave()
			}
		}),

		layout.NewSpacer(),

		// Add new panel
		JW.NewHoverCursorIconButton("", theme.ContentAddIcon(), "Add new panel", func() {
			if onAddNewPanel != nil {
				onAddNewPanel()
			}
		}),
	)
}
