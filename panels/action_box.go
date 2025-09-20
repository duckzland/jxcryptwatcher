package panels

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	JA "jxwatcher/apps"
	JW "jxwatcher/widgets"
)

func NewPanelActionBar(
	onEdit func(),
	onDelete func(),
) *panelActionDisplay {

	pa := &panelActionDisplay{
		editBtn: JW.NewActionButton("edit_panel", "", theme.DocumentCreateIcon(), "Edit panel", "normal",
			func(*JW.ActionButton) {
				if onEdit != nil {
					onEdit()
				}
			}, func(btn *JW.ActionButton) {
				if JA.AppStatusManager.IsOverlayShown() {
					btn.DisallowActions()
					return
				}

				if JA.AppStatusManager.IsDraggable() {
					btn.Hide()
					return
				}
				
				btn.Enable()
			}),
		deleteBtn: JW.NewActionButton("delete_panel", "", theme.DeleteIcon(), "Delete panel", "normal",
			func(*JW.ActionButton) {
				if onDelete != nil {
					onDelete()
				}
			}, func(btn *JW.ActionButton) {
				if JA.AppStatusManager.IsOverlayShown() {
					btn.DisallowActions()
					return
				}

				if JA.AppStatusManager.IsDraggable() {
					btn.Hide()
					return
				}

				btn.Enable()
			}),
	}

	pa.container = container.New(&panelActionLayout{height: 30, margin: 3}, pa.editBtn, pa.deleteBtn)

	return pa
}

type panelActionDisplay struct {
	widget.BaseWidget
	editBtn   *JW.ActionButton
	deleteBtn *JW.ActionButton
	container *fyne.Container
}

func (pa *panelActionDisplay) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(pa.container)
}

func (pa *panelActionDisplay) Show() {
	pa.container.Show()
	JA.AppActionManager.AddButton(pa.deleteBtn)
	JA.AppActionManager.AddButton(pa.editBtn)
}

func (pa *panelActionDisplay) Hide() {
	pa.container.Hide()
	JA.AppActionManager.RemoveButton(pa.deleteBtn)
	JA.AppActionManager.RemoveButton(pa.editBtn)
}
