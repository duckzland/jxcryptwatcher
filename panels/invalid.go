package panels

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"

	JC "jxwatcher/core"
	JT "jxwatcher/types"
	JW "jxwatcher/widgets"
)

func NewInvalidPanel(
	pk string,
	onEdit func(pk string),
	onDelete func(index int),
) fyne.CanvasObject {

	pi := JT.BP.GetIndex(pk)

	content := canvas.NewText("Invalid Panel", JC.TextColor)
	content.Alignment = fyne.TextAlignCenter
	content.TextStyle = fyne.TextStyle{Bold: true}
	content.TextSize = JC.PanelTitleSize

	action := NewPanelActionBar(
		func() {
			pko := JT.BP.GetDataByIndex(pi)
			if pko != nil {
				onEdit(pko.Get())
			}
		},
		func() {
			onDelete(pi)
		},
	)

	return JW.NewDoubleClickContainer(
		"InvalidPanel",
		NewPanelContainer(
			container.NewStack(
				container.NewVBox(
					layout.NewSpacer(),
					content,
					layout.NewSpacer(),
				),
				container.NewVBox(action),
			),
		),
		action,
		false,
	)
}
