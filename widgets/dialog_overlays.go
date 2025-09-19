package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	JC "jxwatcher/core"
)

type dialogOverlays struct {
	widget.BaseWidget
}

func NewDialogOverlays() *dialogOverlays {
	wrapper := &dialogOverlays{}
	wrapper.ExtendBaseWidget(wrapper)

	return wrapper
}

func (h *dialogOverlays) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(canvas.NewRectangle(JC.OverlayBG))
}

func (h *dialogOverlays) Tapped(e *fyne.PointEvent) {
}

func (h *dialogOverlays) DoubleTapped(e *fyne.PointEvent) {
}

func (h *dialogOverlays) TappedSecondary(e *fyne.PointEvent) {
}

func (h *dialogOverlays) MouseIn(*desktop.MouseEvent) {
}

func (h *dialogOverlays) MouseMoved(*desktop.MouseEvent) {
}

func (h *dialogOverlays) MouseOut() {
}

func (h *dialogOverlays) MouseDown(*desktop.MouseEvent) {
}

func (h *dialogOverlays) MouseUp(*desktop.MouseEvent) {
}

func (h *dialogOverlays) Dragged(ev *fyne.DragEvent) {
}

func (h *dialogOverlays) DragEnd() {
}

func (h *dialogOverlays) Cursor() desktop.StandardCursor {
	return desktop.HiddenCursor
}
