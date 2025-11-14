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
	if pa.canShow() {
		pa.container.Show()

		JA.UseAction().Add(pa.deleteBtn)
		JA.UseAction().Add(pa.editBtn)
	}
}

func (pa *panelAction) Hide() {
	pa.container.Hide()
	JA.UseAction().Remove(pa.deleteBtn)
	JA.UseAction().Remove(pa.editBtn)
}

func (pa *panelAction) canShow() bool {
	if JA.UseStatus().IsFetchingCryptos() {
		return false
	}

	if JA.UseStatus().IsDraggable() {
		return false
	}

	return true
}

func NewPanelAction(
	onEdit func(),
	onDelete func(),
) *panelAction {

	pa := &panelAction{}
	pa.editBtn = JW.NewActionButton(JC.ACT_PANEL_EDIT, JC.STRING_EMPTY, theme.DocumentCreateIcon(), "Edit panel", JW.ActionStateNormal,
		func(JW.ActionButton) {
			if onEdit != nil {
				onEdit()
			}

			pa.editBtn.MouseOut()
			pa.deleteBtn.MouseOut()

		}, func(btn JW.ActionButton) {
			if JA.UseStatus().IsOverlayShown() {
				btn.DisallowActions()
				return
			}

			if JA.UseStatus().IsFetchingCryptos() {
				pa.Hide()
				return
			}

			if JA.UseStatus().IsDraggable() {
				pa.Hide()
				return
			}

			btn.Enable()
		})

	pa.deleteBtn = JW.NewActionButton(JC.ACT_PANEL_DELETE, JC.STRING_EMPTY, theme.DeleteIcon(), "Delete panel", JW.ActionStateNormal,
		func(JW.ActionButton) {
			if onDelete != nil {
				onDelete()
			}

			pa.editBtn.MouseOut()
			pa.deleteBtn.MouseOut()

		}, func(btn JW.ActionButton) {
			if JA.UseStatus().IsOverlayShown() {
				btn.DisallowActions()
				return
			}

			if JA.UseStatus().IsFetchingCryptos() {
				pa.Hide()
				return
			}

			if JA.UseStatus().IsDraggable() {
				pa.Hide()
				return
			}

			btn.Enable()
		})

	pa.container = container.New(&panelActionLayout{height: 30, margin: 3}, pa.editBtn, pa.deleteBtn)

	return pa
}
