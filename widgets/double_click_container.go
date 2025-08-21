package widgets

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

var Groups = make(map[string]*DoubleClickContainer)

type DoubleClickContainer struct {
	widget.BaseWidget
	tag         string
	content     fyne.CanvasObject
	child       fyne.CanvasObject
	lastClick   time.Time
	visible     bool
	doubleClick bool
	disabled    bool
	group       *string
}

func NewDoubleClickContainer(
	tag string,
	content fyne.CanvasObject,
	child fyne.CanvasObject,
	doubleClick bool,
	group *string,
) *DoubleClickContainer {

	child.Hide()
	wrapper := &DoubleClickContainer{
		tag:         tag,
		content:     content,
		child:       child,
		visible:     false,
		doubleClick: doubleClick,
		disabled:    false,
		group:       group,
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

	if h.disabled {
		return
	}

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

	if h.group != nil {
		hg := *h.group

		if Groups[hg] != nil {
			Groups[hg].HideTarget()
		}

		Groups[hg] = h
	}
}

func (h *DoubleClickContainer) HideTarget() {
	h.child.Hide()
	h.visible = false
	h.Refresh()

	if h.group != nil {
		hg := *h.group

		if Groups[hg] != nil {
			delete(Groups, hg)
		}
	}
}

func (h *DoubleClickContainer) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}

func (h *DoubleClickContainer) DisableClick() {
	h.disabled = true
}

func (h *DoubleClickContainer) EnableClick() {
	h.disabled = false
}
