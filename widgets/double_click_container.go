package widgets

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

type DoubleClickContainer struct {
	widget.BaseWidget
	tag         string
	content     fyne.CanvasObject
	child       fyne.CanvasObject
	lastClick   time.Time
	visible     bool
	doubleClick bool
}

func NewDoubleClickContainer(
	tag string,
	content fyne.CanvasObject,
	child fyne.CanvasObject,
	doubleClick bool,
) *DoubleClickContainer {

	child.Hide()
	wrapper := &DoubleClickContainer{
		tag:         tag,
		content:     content,
		child:       child,
		visible:     false,
		doubleClick: doubleClick,
	}
	wrapper.ExtendBaseWidget(wrapper)
	return wrapper
}

func (h *DoubleClickContainer) GetTag() string {
	return h.tag
}

func (h *DoubleClickContainer) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(h.content)
}

func (h *DoubleClickContainer) Tapped(_ *fyne.PointEvent) {

	// Double Click mode
	if h.doubleClick {
		now := time.Now()
		if now.Sub(h.lastClick) < 500*time.Millisecond {
			if h.visible {
				h.HideTarget()
			} else {
				h.ShowTarget()
			}
		}
		h.lastClick = now
	}

	// Single Click mode
	if !h.doubleClick {
		if h.visible {
			h.HideTarget()
		} else {
			h.ShowTarget()
		}
	}
}

func (h *DoubleClickContainer) ShowTarget() {
	h.child.Show()
	h.visible = true
	h.Refresh()
}

func (h *DoubleClickContainer) HideTarget() {
	h.child.Hide()
	h.visible = false
	h.Refresh()
}

func (b *DoubleClickContainer) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}
