package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"

	JC "jxwatcher/core"
	JT "jxwatcher/types"
)

func NewInvalidPanel(pk string, onEdit func(pk string), onDelete func(index int)) fyne.CanvasObject {
	pi := JT.BP.GetIndex(pk)

	content := canvas.NewText("Invalid Panel", JC.TextColor)
	content.Alignment = fyne.TextAlignCenter
	content.TextStyle = fyne.TextStyle{Bold: true}
	content.TextSize = 16

	action := container.NewHBox(
		layout.NewSpacer(),
		NewHoverCursorIconButton("", theme.DocumentCreateIcon(), "Edit panel", func() {
			if onEdit != nil {
				onEdit(pk)
			}
		}),
		NewHoverCursorIconButton("", theme.DeleteIcon(), "Delete panel", func() {
			DoActionWithNotification("Removing Panel...", "Panel removed...", JC.NotificationBox, func() {
				if onEdit != nil {
					onDelete(pi)
				}
			})
		}),
	)

	return NewDoubleClickContainer(
		"InvalidPanel",
		NewPanelItem(
			container.NewStack(
				container.NewVBox(
					layout.NewSpacer(),
					content,
					layout.NewSpacer(),
				),
				container.NewVBox(action),
			),
			JC.PanelBG,
			6,
			[4]float32{0, 5, 10, 5},
		),
		action,
		false,
	)
}
