package main

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type DoubleClickContainer struct {
	widget.BaseWidget
	content   fyne.CanvasObject
	child     fyne.CanvasObject
	lastClick time.Time
	visible   bool
}

func NewDoubleClickContainer(content, child fyne.CanvasObject) *DoubleClickContainer {
	child.Hide()
	wrapper := &DoubleClickContainer{
		content: content,
		child:   child,
		visible: false,
	}
	wrapper.ExtendBaseWidget(wrapper)
	return wrapper
}

func (h *DoubleClickContainer) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(h.content)
}

func (h *DoubleClickContainer) Tapped(_ *fyne.PointEvent) {
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
