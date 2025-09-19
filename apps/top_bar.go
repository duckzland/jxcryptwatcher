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
		&TopBarLayout{
			fixedWidth: JC.ActionBtnWidth,
			spacer:     JC.ActionBtnGap,
		},
		container.NewStack(
			topBg,
			JC.NotificationContainer,
		),
		AppActionManager.GetButton("refresh_cryptos"),
		AppActionManager.GetButton("refresh_rates"),
		AppActionManager.GetButton("open_settings"),
		AppActionManager.GetButton("toggle_drag"),
		AppActionManager.GetButton("add_panel"),
	)
}
