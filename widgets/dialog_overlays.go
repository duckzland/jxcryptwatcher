package widgets

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	JC "jxwatcher/core"
)

type dialogOverlays struct {
	widget.BaseWidget
	bgcolor color.Color
}

func (h *dialogOverlays) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(canvas.NewRectangle(h.bgcolor))
}

func (h *dialogOverlays) Scrolled(e *fyne.ScrollEvent) {
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

func NewDialogOverlays() *dialogOverlays {
	wrapper := &dialogOverlays{
		bgcolor: JC.UseTheme().GetColor(theme.ColorNameOverlayBackground),
	}

	wrapper.ExtendBaseWidget(wrapper)

	return wrapper
}
