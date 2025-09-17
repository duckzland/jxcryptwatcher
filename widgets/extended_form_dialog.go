package widgets

import (
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type ExtendedFormDialog struct {
	layer           *fyne.Container
	confirm         *HoverCursorIconButton
	cancel          *HoverCursorIconButton
	items           []*widget.FormItem
	callback        func(bool) bool
	form            *widget.Form
	parent          fyne.Window
	validationTimer *time.Timer
	content         *fyne.Container
}

func NewExtendedFormDialog(
	titleText string,
	items []*widget.FormItem,
	topContent []*fyne.Container,
	bottomContent []*fyne.Container,
	absolutePositionedContent []*fyne.Container,
	callback func(bool) bool,
	parent fyne.Window,
) *ExtendedFormDialog {
	fd := &ExtendedFormDialog{
		items:    items,
		form:     widget.NewForm(items...),
		parent:   parent,
		callback: callback,
	}

	fd.cancel = NewHoverCursorIconButton(
		"cancel_save_panel",
		"Cancel",
		theme.CancelIcon(),
		"",
		"normal",
		func(*HoverCursorIconButton) {
			fd.Hide()
		},
		nil,
	)

	fd.confirm = NewHoverCursorIconButton(
		"cancel_save_panel",
		"Save",
		theme.ConfirmIcon(),
		"",
		"normal",
		func(*HoverCursorIconButton) {
			fd.hideWithResponse(true)
		},
		nil,
	)

	fd.setSubmitState(fd.form.Validate())

	fd.form.SetOnValidationChanged(func(err error) {
		if fd.validationTimer != nil {
			fd.validationTimer.Stop()
		}
		fd.validationTimer = time.AfterFunc(300*time.Millisecond, func() {
			fd.setSubmitState(err)
		})
	})

	innerLayout := &ExtendedDialogContentLayout{
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

	outerLayout := &ExtendedDialogOverlayLayout{
		background: canvas.NewRectangle(color.NRGBA{R: 0, G: 0, B: 0, A: 128}),
		dialogBox:  fd.content,
	}

	fd.layer = container.New(outerLayout, outerLayout.background, fd.content)

	fd.parent.Canvas().Overlays().Add(fd.layer)

	return fd
}

func (d *ExtendedFormDialog) Resize(size fyne.Size) {
	d.layer.Resize(d.parent.Canvas().Size())
}

func (d *ExtendedFormDialog) Show() {
	d.layer.Refresh()
}

func (d *ExtendedFormDialog) Hide() {
	d.parent.Canvas().Overlays().Remove(d.layer)
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
			d.Hide()
			return
		}
	}

	d.Hide()
}

func (d *ExtendedFormDialog) setSubmitState(err error) {
	if err != nil {
		d.confirm.Disable()
	} else {
		d.confirm.Enable()
	}
}

type ExtendedDialogContentLayout struct {
	background    *canvas.Rectangle
	title         fyne.CanvasObject
	topContent    []*fyne.Container
	form          fyne.CanvasObject
	buttons       fyne.CanvasObject
	bottomContent []*fyne.Container
	padding       float32
}

func (l *ExtendedDialogContentLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	l.background.Resize(size)
	l.background.Move(fyne.NewPos(0, 0))

	x := l.padding
	y := l.padding
	w := size.Width - 2*l.padding

	titleSize := l.title.MinSize()
	l.title.Resize(fyne.NewSize(w, titleSize.Height))
	l.title.Move(fyne.NewPos(x, y))
	y += titleSize.Height + l.padding

	for _, top := range l.topContent {
		topSize := top.MinSize()
		top.Resize(fyne.NewSize(w, topSize.Height))
		top.Move(fyne.NewPos(x, y))
		y += topSize.Height + l.padding
	}

	formSize := l.form.MinSize()
	l.form.Resize(fyne.NewSize(w, formSize.Height))
	l.form.Move(fyne.NewPos(x, y))
	y += formSize.Height + l.padding

	for _, bottom := range l.bottomContent {
		bottomSize := bottom.MinSize()
		bottom.Resize(fyne.NewSize(w, bottomSize.Height))
		bottom.Move(fyne.NewPos(x, y))
		y += bottomSize.Height + l.padding
	}

	buttonSize := l.buttons.MinSize()
	buttonX := x + (w-buttonSize.Width)/2
	l.buttons.Resize(buttonSize)
	l.buttons.Move(fyne.NewPos(buttonX, y))
	y += buttonSize.Height + l.padding
}

func (l *ExtendedDialogContentLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	h := l.title.MinSize().Height +
		l.form.MinSize().Height +
		l.buttons.MinSize().Height +
		4*l.padding

	return fyne.NewSize(0, h)
}

type ExtendedDialogOverlayLayout struct {
	background *canvas.Rectangle
	dialogBox  fyne.CanvasObject
	cHeight    float32
	cWidth     float32
}

func (l *ExtendedDialogOverlayLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if l.cWidth != size.Width {
		l.cHeight = 0
	}

	l.cWidth = size.Width

	if l.cHeight == 0 {
		l.cHeight = size.Height
	}

	l.background.Resize(size)
	l.background.Move(fyne.NewPos(0, 0))
	l.background.Show()

	var dialogWidth float32
	switch {
	case l.cWidth <= 560:
		dialogWidth = l.cWidth - 10
	case l.cWidth > 560 && l.cWidth <= 1200:
		dialogWidth = l.cWidth * 0.8
	default:
		dialogWidth = 800
	}

	dialogHeight := l.dialogBox.MinSize().Height
	emptySpace := l.cHeight - dialogHeight
	posX := (l.cWidth - dialogWidth) / 2
	posY := emptySpace / 4

	if posY < 0 {
		posY = 0
	}

	l.dialogBox.Resize(fyne.NewSize(dialogWidth, dialogHeight))
	l.dialogBox.Move(fyne.NewPos(posX, posY))
}

func (l *ExtendedDialogOverlayLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(300, 300)
}
