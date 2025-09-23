package apps

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"

	JC "jxwatcher/core"
	JW "jxwatcher/widgets"
)

func NewTopBar() *fyne.Container {

	topBg := canvas.NewRectangle(JC.ThemeColor(JC.ColorNamePanelBG))
	topBg.CornerRadius = 4

	return container.New(
		&topBarLayout{
			fixedWidth: JC.ThemeSize(JC.SizeActionBtnWidth),
			spacer:     JC.ThemeSize(JC.SizeActionBtnGap),
		},
		container.NewStack(
			topBg,
			JW.UseNotification(),
		),
		ActionManager.Get("refresh_cryptos"),
		ActionManager.Get("refresh_rates"),
		ActionManager.Get("open_settings"),
		ActionManager.Get("toggle_drag"),
		ActionManager.Get("add_panel"),
	)
}
