package main

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type DoubleClickWrapper struct {
	widget.BaseWidget
	content   fyne.CanvasObject
	child     fyne.CanvasObject
	lastClick time.Time
	visible   bool
}

func NewDoubleClickWrapper(content, child fyne.CanvasObject) *DoubleClickWrapper {
	child.Hide()
	wrapper := &DoubleClickWrapper{
		content: content,
		child:   child,
		visible: false,
	}
	wrapper.ExtendBaseWidget(wrapper)
	return wrapper
}

func (h *DoubleClickWrapper) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(h.content)
}

func (h *DoubleClickWrapper) Tapped(_ *fyne.PointEvent) {
	now := time.Now()
	if now.Sub(h.lastClick) < 500*time.Millisecond {
		if h.visible {
			h.child.Hide()
			h.visible = false
			h.Refresh()
		} else {
			h.child.Show()
			h.visible = true
			h.Refresh()
		}
	}
	h.lastClick = now
}
