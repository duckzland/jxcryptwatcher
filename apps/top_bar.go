package apps

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"

	JC "jxwatcher/core"
)

func NewTopBar() *fyne.Container {

	topBg := canvas.NewRectangle(JC.PanelBG)
	topBg.CornerRadius = 4

	return container.New(
		&topBarLayout{
			fixedWidth: JC.ActionBtnWidth,
			spacer:     JC.ActionBtnGap,
		},
		container.NewStack(
			topBg,
			JC.NotificationContainer,
		),
		ActionManager.Get("refresh_cryptos"),
		ActionManager.Get("refresh_rates"),
		ActionManager.Get("open_settings"),
		ActionManager.Get("toggle_drag"),
		ActionManager.Get("add_panel"),
	)
}
