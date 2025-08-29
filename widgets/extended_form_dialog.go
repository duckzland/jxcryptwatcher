package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type ExtendedFormDialog struct {
	dialog   *dialog.CustomDialog
	confirm  *HoverCursorIconButton
	cancel   *HoverCursorIconButton
	items    []*widget.FormItem
	callback func(bool)
	form     *widget.Form
}

func NewExtendedFormDialog(
	title string,
	items []*widget.FormItem,
	callback func(bool),
	parent fyne.Window,
) *ExtendedFormDialog {

	var fd *ExtendedFormDialog
	form := widget.NewForm(items...)

	fd = &ExtendedFormDialog{
		dialog: dialog.NewCustomWithoutButtons(title, form, parent),
		confirm: NewHoverCursorIconButton("save_panel", "Save", theme.ConfirmIcon(), "", func(*HoverCursorIconButton) {
			fd.hideWithResponse(true)
		}, nil),
		cancel: NewHoverCursorIconButton("cancel_save_panel", "Cancel", theme.CancelIcon(), "", func(*HoverCursorIconButton) {
			fd.dialog.Hide()
		}, nil),
		items:    items,
		callback: func(resp bool) { callback(resp) },
		form:     form,
	}

	fd.dialog.SetButtons([]fyne.CanvasObject{fd.cancel, fd.confirm})
	fd.setSubmitState(fd.form.Validate())
	fd.form.SetOnValidationChanged(fd.setSubmitState)

	return fd
}

func (d *ExtendedFormDialog) setSubmitState(err error) {
	if err != nil {
		d.confirm.Disable()
		return
	}

	d.confirm.Enable()
}

func (d *ExtendedFormDialog) Submit() {
	if d.confirm.Disabled() {
		return
	}

	d.hideWithResponse(true)
}

func (d *ExtendedFormDialog) hideWithResponse(resp bool) {
	d.dialog.Hide()
	if d.callback != nil {
		d.callback(resp)
	}
}

func (d *ExtendedFormDialog) Show() {
	d.dialog.Show()
}

func (d *ExtendedFormDialog) Resize(newSize fyne.Size) {
	d.dialog.Resize(newSize)
}
