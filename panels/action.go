package panels

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	JA "jxwatcher/apps"
	JW "jxwatcher/widgets"
)

func NewPanelAction(
	onEdit func(),
	onDelete func(),
) *panelAction {

	pa := &panelAction{
		editBtn: JW.NewActionButton("edit_panel", "", theme.DocumentCreateIcon(), "Edit panel", "normal",
			func(JW.ActionButton) {
				if onEdit != nil {
					onEdit()
				}
			}, func(btn JW.ActionButton) {
				if JA.StatusManager.IsOverlayShown() {
					btn.DisallowActions()
					return
				}

				if JA.StatusManager.IsDraggable() {
					btn.Hide()
					return
				}

				btn.Enable()
			}),
		deleteBtn: JW.NewActionButton("delete_panel", "", theme.DeleteIcon(), "Delete panel", "normal",
			func(JW.ActionButton) {
				if onDelete != nil {
					onDelete()
				}
			}, func(btn JW.ActionButton) {
				if JA.StatusManager.IsOverlayShown() {
					btn.DisallowActions()
					return
				}

				if JA.StatusManager.IsDraggable() {
					btn.Hide()
					return
				}

				btn.Enable()
			}),
	}

	pa.container = container.New(&panelActionLayout{height: 30, margin: 3}, pa.editBtn, pa.deleteBtn)

	return pa
}

type panelAction struct {
	widget.BaseWidget
	editBtn   JW.ActionButton
	deleteBtn JW.ActionButton
	container *fyne.Container
}

func (pa *panelAction) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(pa.container)
}

func (pa *panelAction) Show() {
	pa.container.Show()
	JA.UseActionManager().Add(pa.deleteBtn)
	JA.UseActionManager().Add(pa.editBtn)
}

func (pa *panelAction) Hide() {
	pa.container.Hide()
	JA.UseActionManager().Remove(pa.deleteBtn)
	JA.UseActionManager().Remove(pa.editBtn)
}
