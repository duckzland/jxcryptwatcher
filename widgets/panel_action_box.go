package widgets

import (
	JC "jxwatcher/core"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
)

func NewPanelActionBar(
	onEdit func(),
	onDelete func(),
) fyne.CanvasObject {

	editBtn := NewHoverCursorIconButton("", theme.DocumentCreateIcon(), "Edit panel", func() {
		if onEdit != nil {
			onEdit()
		}
	})

	deleteBtn := NewHoverCursorIconButton("", theme.DeleteIcon(), "Delete panel", func() {
		DoActionWithNotification("Removing Panel...", "Panel removed...", JC.NotificationBox, func() {
			if onDelete != nil {
				onDelete()
			}
		})
	})

	return container.NewHBox(
		layout.NewSpacer(),
		editBtn,
		deleteBtn,
	)
}
