package widgets

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type ExtendedFormDialog struct {
	dialog          *dialog.CustomDialog
	confirm         *HoverCursorIconButton
	cancel          *HoverCursorIconButton
	items           []*widget.FormItem
	callback        func(bool) bool
	form            *widget.Form
	parent          fyne.Window
	topContent      []*fyne.Container
	bottomContent   []*fyne.Container
	overlayContent  *fyne.Container
	popup           *widget.PopUp
	validationTimer *time.Timer
}

func NewExtendedFormDialog(
	title string,
	items []*widget.FormItem,
	topContent []*fyne.Container,
	bottomContent []*fyne.Container,
	callback func(bool) bool,
	parent fyne.Window,
) *ExtendedFormDialog {

	var fd *ExtendedFormDialog
	form := widget.NewForm(items...)

	fd = &ExtendedFormDialog{
		dialog: dialog.NewCustomWithoutButtons(title, form, parent),
		confirm: NewHoverCursorIconButton("save_panel", "Save", theme.ConfirmIcon(), "", "normal", func(*HoverCursorIconButton) {
			fd.hideWithResponse(true)
		}, nil),
		cancel: NewHoverCursorIconButton("cancel_save_panel", "Cancel", theme.CancelIcon(), "", "normal", func(*HoverCursorIconButton) {
			fd.dialog.Hide()
		}, nil),
		items:         items,
		callback:      func(resp bool) bool { return callback(resp) },
		form:          form,
		parent:        parent,
		topContent:    topContent,
		bottomContent: bottomContent,
	}

	fd.dialog.SetButtons([]fyne.CanvasObject{fd.cancel, fd.confirm})
	fd.setSubmitState(fd.form.Validate())

	fd.form.SetOnValidationChanged(func(err error) {
		if fd.validationTimer != nil {
			fd.validationTimer.Stop()
		}
		fd.validationTimer = time.AfterFunc(300*time.Millisecond, func() {
			fd.setSubmitState(err)
		})
	})

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

	err := d.form.Validate()
	d.setSubmitState(err)

	if err != nil {
		return
	}

	if d.callback != nil {
		if d.callback(resp) {
			d.dialog.Hide()
			return
		}
	}

	d.dialog.Hide()

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

		if d.overlayContent != nil && d.topContent != nil {
			topObjects := make([]fyne.CanvasObject, len(d.topContent))
			for i, c := range d.topContent {
				topObjects[i] = c
			}

			d.overlayContent.Objects = append(topObjects, d.overlayContent.Objects...)
		}
		if d.overlayContent != nil && d.bottomContent != nil {
			for _, c := range d.bottomContent {
				var obj fyne.CanvasObject = c
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
