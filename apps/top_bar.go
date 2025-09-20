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
		AppActions.GetButton("refresh_cryptos"),
		AppActions.GetButton("refresh_rates"),
		AppActions.GetButton("open_settings"),
		AppActions.GetButton("toggle_drag"),
		AppActions.GetButton("add_panel"),
	)
}
