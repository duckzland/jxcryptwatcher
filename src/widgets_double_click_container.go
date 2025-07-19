package main

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

type DoubleClickContainer struct {
	widget.BaseWidget
	content     fyne.CanvasObject
	child       fyne.CanvasObject
	lastClick   time.Time
	visible     bool
	doubleClick bool
}

func NewDoubleClickContainer(content, child fyne.CanvasObject, doubleClick bool) *DoubleClickContainer {
	child.Hide()
	wrapper := &DoubleClickContainer{
		content:     content,
		child:       child,
		visible:     false,
		doubleClick: doubleClick,
	}
	wrapper.ExtendBaseWidget(wrapper)
	return wrapper
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

		// Single Click mode
	} else {
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

// Implement desktop.Cursorable
func (b *DoubleClickContainer) Cursor() desktop.Cursor {
	return desktop.PointerCursor // Shows hand icon
}
