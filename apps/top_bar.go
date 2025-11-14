package apps

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"

	JC "jxwatcher/core"
	JW "jxwatcher/widgets"
)

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
		UseAction().Get(JC.ACT_CRYPTO_REFRESH_MAP),
		UseAction().Get(JC.ACT_EXCHANGE_REFRESH_RATES),
		UseAction().Get(JC.ACT_OPEN_SETTINGS),
		UseAction().Get(JC.ACT_PANEL_DRAG),
		UseAction().Get(JC.ACT_PANEL_ADD),
	)
}
