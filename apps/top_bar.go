package apps

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"

	JC "jxwatcher/core"
	JT "jxwatcher/types"
	JW "jxwatcher/widgets"
)

const AppOpenSettings = "open_settings"
const AppToggleDrag = "toggle_drag"
const AppAddPanel = "add_panel"

func NewTopBar() *fyne.Container {

	topBg := canvas.NewRectangle(JC.UseTheme().GetColor(JC.ColorNamePanelBG))
	topBg.CornerRadius = 4

	return container.New(
		&topBarLayout{
			fixedWidth: JC.UseTheme().Size(JC.SizeActionBtnWidth),
			spacer:     JC.UseTheme().Size(JC.SizeActionBtnGap),
		},
		container.NewStack(
			topBg,
			JW.UseNotification(),
		),
		UseAction().Get(JT.CryptoRefresh),
		UseAction().Get(JT.ExchangeRefresh),
		UseAction().Get(AppOpenSettings),
		UseAction().Get(AppToggleDrag),
		UseAction().Get(AppAddPanel),
	)
}
