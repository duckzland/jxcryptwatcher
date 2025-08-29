package panels

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"

	JW "jxwatcher/widgets"
)

type PanelActionLayout struct {
	margin float32
	height float32
}

func (r *PanelActionLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	x := size.Width
	for i := len(objects) - 1; i >= 0; i-- {
		obj := objects[i]
		objSize := obj.MinSize()
		x -= objSize.Width + r.margin
		obj.Move(fyne.NewPos(x, r.margin)) // Vertically center
		obj.Resize(objSize)
	}
}

func (r *PanelActionLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	totalWidth := float32(0)
	maxHeight := float32(0)
	for _, obj := range objects {
		size := obj.MinSize()
		totalWidth += size.Width
		if size.Height > maxHeight {
			maxHeight = size.Height
		}
	}
	return fyne.NewSize(totalWidth, r.height+r.margin) // Fixed height
}

func NewPanelActionBar(
	onEdit func(),
	onDelete func(),
) fyne.CanvasObject {

	editBtn := JW.NewHoverCursorIconButton("edit_panel", "", theme.DocumentCreateIcon(), "Edit panel", func(*JW.HoverCursorIconButton) {
		if onEdit != nil {
			onEdit()
		}
	}, nil)

	deleteBtn := JW.NewHoverCursorIconButton("delete_panel", "", theme.DeleteIcon(), "Delete panel", func(*JW.HoverCursorIconButton) {
		if onDelete != nil {
			onDelete()
		}
	}, nil)

	return container.New(&PanelActionLayout{height: 30, margin: 3}, editBtn, deleteBtn)
}
