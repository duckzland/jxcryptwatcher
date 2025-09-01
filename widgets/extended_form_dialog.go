package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type ExtendedFormDialog struct {
	dialog         *dialog.CustomDialog
	confirm        *HoverCursorIconButton
	cancel         *HoverCursorIconButton
	items          []*widget.FormItem
	callback       func(bool)
	form           *widget.Form
	parent         fyne.Window
	topContent     []*fyne.Container
	bottomContent  []*fyne.Container
	overlayContent *fyne.Container
	popup          *widget.PopUp
}

func NewExtendedFormDialog(
	title string,
	items []*widget.FormItem,
	topContent []*fyne.Container,
	bottomContent []*fyne.Container,
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
		items:         items,
		callback:      func(resp bool) { callback(resp) },
		form:          form,
		parent:        parent,
		topContent:    topContent,
		bottomContent: bottomContent,
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

	if d.overlayContent == nil {
		for _, overlay := range d.parent.Canvas().Overlays().List() {
			if popup, ok := overlay.(*widget.PopUp); ok {
				d.popup = popup
				if cont, ok := popup.Content.(*fyne.Container); ok {
					d.overlayContent = cont
				}
			}
		}

		// We just prepend or append, its up to other to reposition them!
		if d.overlayContent != nil && d.topContent != nil {
			// Convert []*fyne.Container to []fyne.CanvasObject
			topObjects := make([]fyne.CanvasObject, len(d.topContent))
			for i, c := range d.topContent {
				topObjects[i] = c // Implicit interface conversion
			}

			// Append converted slice to overlayContent.Objects
			d.overlayContent.Objects = append(topObjects, d.overlayContent.Objects...)
		}
		if d.overlayContent != nil && d.bottomContent != nil {
			for _, c := range d.bottomContent {
				var obj fyne.CanvasObject = c // Explicit interface assignment
				d.overlayContent.Add(obj)
			}
		}
	}
}

func (d *ExtendedFormDialog) Resize(newSize fyne.Size) {
	d.dialog.Resize(newSize)
}

func (d *ExtendedFormDialog) GetContent() *fyne.Container {
	return d.overlayContent
}
