package apps

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	JC "jxwatcher/core"
)

type dragPlaceholder struct {
	widget.BaseWidget
	bgcolor color.Color
	canvas  *canvas.Rectangle
}

func (h *dragPlaceholder) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(h.canvas)
}

func (h *dragPlaceholder) Cursor() desktop.StandardCursor {
	return desktop.PointerCursor
}

func (h *dragPlaceholder) Resize(size fyne.Size) {
	if h.Size() != size {
		h.BaseWidget.Resize(size)
	}
}

func (h *dragPlaceholder) Move(pos fyne.Position) {
	if h.Position() != pos {
		h.BaseWidget.Move(pos)
	}
}

func (h *dragPlaceholder) SetColor(c color.Color) {
	if h.canvas == nil {
		return
	}
	if !h.IsColor(c) {
		h.canvas.FillColor = c
		canvas.Refresh(h.canvas)
	}
}

func (h *dragPlaceholder) IsColor(c color.Color) bool {
	if h.canvas == nil {
		return false
	}
	return h.canvas.FillColor == c
}

func NewDragPlaceholder() *dragPlaceholder {
	wrapper := &dragPlaceholder{
		bgcolor: JC.UseTheme().GetColor(JC.ColorNameTransparent),
		canvas:  canvas.NewRectangle(JC.UseTheme().GetColor(JC.ColorNameTransparent)),
	}

	wrapper.canvas.CornerRadius = JC.UseTheme().Size(JC.SizePanelBorderRadius)
	wrapper.ExtendBaseWidget(wrapper)

	return wrapper
}
