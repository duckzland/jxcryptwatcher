package widgets

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	JC "jxwatcher/core"
)

type DialogForm interface {
	Show()
	Hide()
	GetContent() *fyne.Container
	GetForm() *widget.Form
}

type dialogForm struct {
	layer           *fyne.Container
	confirm         ActionButton
	cancel          ActionButton
	items           []*widget.FormItem
	callback        func(bool) bool
	form            *widget.Form
	parent          fyne.Window
	validationTimer *time.Timer
	content         *fyne.Container
	render          func(*fyne.Container)
	destroy         func(*fyne.Container)
	validationDone  chan struct{}
}

func NewDialogForm(
	titleText string,
	items []*widget.FormItem,
	topContent []*fyne.Container,
	bottomContent []*fyne.Container,
	absolutePositionedContent []*fyne.Container,
	callback func(bool) bool,
	render func(*fyne.Container),
	destroy func(*fyne.Container),
	parent fyne.Window,
) DialogForm {

	fd := &dialogForm{
		items:    items,
		form:     widget.NewForm(items...),
		parent:   parent,
		callback: callback,
		render:   render,
		destroy:  destroy,
	}

	fd.cancel = NewActionButton(
		"cancel_save_panel",
		"Cancel",
		theme.CancelIcon(),
		"Close Form",
		"normal",
		func(ActionButton) {
			fd.Hide()
		},
		nil,
	)

	fd.confirm = NewActionButton(
		"save_panel",
		"Save",
		theme.ConfirmIcon(),
		"Save and Close Form",
		"normal",
		func(ActionButton) {
			fd.hideWithResponse(true)
		},
		nil,
	)

	delay := 300 * time.Millisecond
	if JC.IsMobile {
		delay = 1000 * time.Millisecond
	}

	fd.setSubmitState(fd.form.Validate())

	fd.form.SetOnValidationChanged(func(err error) {
		if fd.validationTimer != nil {
			fd.validationTimer.Stop()
		}
		fd.validationTimer = time.AfterFunc(delay, func() {
			fd.setSubmitState(err)
		})
	})

	innerLayout := &dialogContentLayout{
		background:    canvas.NewRectangle(theme.DefaultTheme().Color(theme.ColorNameBackground, theme.VariantDark)),
		title:         widget.NewLabelWithStyle(titleText, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		topContent:    topContent,
		form:          fd.form,
		buttons:       container.NewHBox(fd.cancel, widget.NewLabel(" "), fd.confirm),
		bottomContent: bottomContent,
		padding:       16,
	}

	content := []fyne.CanvasObject{innerLayout.background}
	content = append(content, innerLayout.title)

	for _, c := range topContent {
		content = append(content, c)
	}

	content = append(content, fd.form)

	for _, c := range bottomContent {
		content = append(content, c)
	}

	content = append(content, innerLayout.buttons)

	for _, c := range absolutePositionedContent {
		content = append(content, c)
	}

	fd.content = container.New(innerLayout, content...)

	outerLayout := &dialogOverlaysLayout{
		background: NewDialogOverlays(),
		dialogBox:  fd.content,
	}

	fd.layer = container.New(outerLayout, outerLayout.background, fd.content)

	if fd.render != nil {
		fd.render(fd.layer)
	} else {
		fd.parent.Canvas().Overlays().Add(fd.layer)
	}

	return fd
}

func (d *dialogForm) Resize(size fyne.Size) {
	d.layer.Resize(d.parent.Canvas().Size())
}

func (d *dialogForm) Show() {
	d.layer.Refresh()
}

func (d *dialogForm) Hide() {
	if d.destroy != nil {
		d.destroy(d.layer)
		return
	}

	d.parent.Canvas().Overlays().Remove(d.layer)
}

func (d *dialogForm) Submit() {
	if d.confirm.IsDisabled() {
		return
	}
	d.hideWithResponse(true)
}

func (d *dialogForm) GetContent() *fyne.Container {
	return d.content
}

func (d *dialogForm) GetForm() *widget.Form {
	return d.form
}

func (d *dialogForm) hideWithResponse(resp bool) {
	err := d.form.Validate()
	d.setSubmitState(err)

	if err != nil {
		return
	}

	if d.callback != nil {
		if d.callback(resp) {
			d.Hide()
			return
		}
	}

	d.Hide()
}

func (d *dialogForm) setSubmitState(err error) {
	if err != nil {
		d.confirm.Disable()
	} else {
		d.confirm.Enable()
	}
}
