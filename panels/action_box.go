package panels

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	JA "jxwatcher/apps"
	JC "jxwatcher/core"
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
		obj.Move(fyne.NewPos(x, r.margin))
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
	return fyne.NewSize(totalWidth, r.height+r.margin)
}

func NewPanelActionBar(
	onEdit func(),
	onDelete func(),
) *PanelActionDisplay {

	pa := &PanelActionDisplay{
		editBtn: JW.NewHoverCursorIconButton("edit_panel", "", theme.DocumentCreateIcon(), "Edit panel", "normal",
			func(*JW.HoverCursorIconButton) {
				if onEdit != nil {
					onEdit()
				}
			}, func(btn *JW.HoverCursorIconButton) {
				if JA.AppStatusManager.IsOverlayShown() {
					btn.Disable()
					return
				}

				btn.Enable()
			}),
		deleteBtn: JW.NewHoverCursorIconButton("delete_panel", "", theme.DeleteIcon(), "Delete panel", "normal",
			func(*JW.HoverCursorIconButton) {
				if onDelete != nil {
					onDelete()
				}
			}, func(btn *JW.HoverCursorIconButton) {
				JC.Logln("Validation called")
				if JA.AppStatusManager.IsOverlayShown() {
					btn.Disable()
					return
				}

				btn.Enable()
			}),
	}

	pa.container = container.New(&PanelActionLayout{height: 30, margin: 3}, pa.editBtn, pa.deleteBtn)

	return pa
}

type PanelActionDisplay struct {
	widget.BaseWidget
	editBtn   *JW.HoverCursorIconButton
	deleteBtn *JW.HoverCursorIconButton
	container *fyne.Container
}

func (pa *PanelActionDisplay) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(pa.container)
}

func (pa *PanelActionDisplay) Show() {
	pa.container.Show()
	JA.AppActionManager.AddButton(pa.deleteBtn)
	JA.AppActionManager.AddButton(pa.editBtn)
}

func (pa *PanelActionDisplay) Hide() {
	pa.container.Hide()
	JA.AppActionManager.RemoveButton(pa.deleteBtn)
	JA.AppActionManager.RemoveButton(pa.editBtn)
}
