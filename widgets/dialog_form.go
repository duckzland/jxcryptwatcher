package widgets

import (
	"runtime"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	JC "jxwatcher/core"
)

type actionEntry interface {
	SetAction(func(bool))
}

type DialogForm interface {
	Show()
	Hide()
	GetContent() *fyne.Container
	GetForm() *widget.Form
	Submit()
	Refresh()
}

type dialogForm struct {
	layer           *fyne.Container
	content         *fyne.Container
	inner           *dialogContentLayout
	confirm         ActionButton
	cancel          ActionButton
	form            *widget.Form
	items           []*widget.FormItem
	callback        func() bool
	parent          fyne.Window
	validationTimer *time.Timer
	render          func(*fyne.Container)
	destroy         func(*fyne.Container)
	validationDone  chan struct{}
}

func (d *dialogForm) Resize(size fyne.Size) {
	d.layer.Resize(d.parent.Canvas().Size())
}

func (d *dialogForm) Show() {
	d.layer.Refresh()
}

func (d *dialogForm) Hide() {

	defer runtime.GC()

	d.confirm.Destroy()
	d.cancel.Destroy()

	if d.destroy != nil {
		d.destroy(d.layer)
		return
	}

	d.parent.Canvas().Overlays().Remove(d.layer)
}

func (d *dialogForm) Submit() {
	if err := d.form.Validate(); err != nil {
		return
	}

	if d.confirm.IsDisabled() {
		return
	}

	if d.callback != nil {
		if d.callback() {
			d.Hide()
			return
		}

		return
	}

	d.Hide()
}

func (d *dialogForm) GetContent() *fyne.Container {
	return d.content
}

func (d *dialogForm) GetForm() *widget.Form {
	return d.form
}

func (d *dialogForm) Refresh() {
	d.inner.ClearCache()

	if d.content != nil {
		if d.content.Layout != nil {
			d.content.Layout.Layout(d.content.Objects, d.content.Size())
		}
		d.content.Refresh()
	}
	if d.layer != nil {
		if d.layer.Layout != nil {
			d.layer.Layout.Layout(d.layer.Objects, d.layer.Size())
		}
		d.layer.Refresh()
	}
	if d.form != nil {
		d.form.Refresh()
	}
}

func NewDialogForm(
	titleText string,
	items []*widget.FormItem,
	topContent []*fyne.Container,
	bottomContent []*fyne.Container,
	absolutePositionedContent []*fyne.Container,
	customAction ActionButton,
	callback func() bool,
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

	fd.form.Orientation = widget.Vertical

	fd.cancel = NewActionButton(
		"cancel_save_panel",
		"Cancel",
		theme.CancelIcon(),
		"Close Form",
		ActionStateNormal,
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
		ActionStateNormal,
		func(ActionButton) {
			fd.Submit()
		},
		nil,
	)

	formLayout := &dialogFormLayout{
		form:       fd.form,
		dispatcher: NewScrollDispatcher(),
	}

	spacer := canvas.NewRectangle(nil)
	spacer.SetMinSize(fyne.NewSize(10, 10))

	objs := []fyne.CanvasObject{}

	if customAction != nil {
		objs = append(objs, customAction, spacer)
	}

	objs = append(objs, fd.cancel, spacer, fd.confirm)

	buttons := container.NewHBox(objs...)

	innerLayout := &dialogContentLayout{
		background:    canvas.NewRectangle(JC.UseTheme().GetColor(theme.ColorNameBackground)),
		title:         widget.NewLabelWithStyle(titleText, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		topContent:    topContent,
		form:          fd.form,
		content:       container.NewVScroll(container.New(formLayout, container.NewThemeOverride(formLayout.form, &dialogFormTheme{base: JC.UseTheme()}), formLayout.dispatcher)),
		buttons:       buttons,
		bottomContent: bottomContent,
		padding:       16,
	}

	fd.inner = innerLayout

	formLayout.container = innerLayout.content
	formLayout.dispatcher.SetScroller(innerLayout.content)

	for _, item := range items {
		if item.HintText == JC.STRING_EMPTY {
			item.HintText = " "
		}

		if ae, ok := item.Widget.(actionEntry); ok {
			ae.SetAction(func(active bool) {
				if active {
					formLayout.dispatcher.Hide()
					return
				}

				formLayout.dispatcher.Show()
			})
		}
	}

	content := []fyne.CanvasObject{innerLayout.background}
	content = append(content, innerLayout.title)

	for _, c := range topContent {
		content = append(content, c)
	}

	content = append(content, innerLayout.content)

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
