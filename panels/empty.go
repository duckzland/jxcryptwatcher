package panels

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"

	JC "jxwatcher/core"
)

func NewEmptyPanel() fyne.CanvasObject {
	content := canvas.NewText("Loading...", JC.TextColor)
	content.Alignment = fyne.TextAlignCenter
	content.TextStyle = fyne.TextStyle{Bold: true}
	content.TextSize = JC.PanelTitleSize

	return NewPanelContainer(
		container.New(
			layout.NewCustomPaddedVBoxLayout(6),
			layout.NewSpacer(),
			content,
			layout.NewSpacer(),
		),
	)
}
