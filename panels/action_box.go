package panels

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"

	JW "jxwatcher/widgets"
)

func NewPanelActionBar(
	onEdit func(),
	onDelete func(),
) fyne.CanvasObject {

	editBtn := JW.NewHoverCursorIconButton("", theme.DocumentCreateIcon(), "Edit panel", func() {
		if onEdit != nil {
			onEdit()
		}
	})

	deleteBtn := JW.NewHoverCursorIconButton("", theme.DeleteIcon(), "Delete panel", func() {
		if onDelete != nil {
			onDelete()
		}
	})

	return container.NewHBox(
		layout.NewSpacer(),
		editBtn,
		deleteBtn,
	)
}
